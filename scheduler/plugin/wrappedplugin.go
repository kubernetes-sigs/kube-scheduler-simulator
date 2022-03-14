package plugin

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	schedulingresultstore "github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/plugin/resultstore"
)

//go:generate mockgen -destination=./mock/$GOFILE -package=plugin . Store,FilterPluginExtender,ScorePluginExtender,NormalizeScorePluginExtender
//go:generate mockgen -destination=./mock/framework.go -package=plugin k8s.io/kubernetes/pkg/scheduler/framework FilterPlugin,ScorePlugin,ScoreExtensions
type Store interface {
	AddNormalizedScoreResult(namespace, podName, nodeName, pluginName string, normalizedscore int64)
	AddFilterResult(namespace, podName, nodeName, pluginName, reason string)
	AddScoreResult(namespace, podName, nodeName, pluginName string, score int64)
}

type FilterPluginExtender interface {
	// BeforeFilter is a function that runs before the Filter method of the original plugin.
	// If BeforeFilter returns non-success status, the simulator plugin doesn't run the Filter method of the original plugin and return that status.
	BeforeFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status
	// AfterFilter is a function that is run after the Filter method of the original plugin.
	// AfterFilter must be run regardless of the result from a Filter method of the original plugin,
	// and a Filter method of the simulator plugin returns the status returned from AfterFilter.
	AfterFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo, filterResult *framework.Status) *framework.Status
}

type ScorePluginExtender interface {
	// BeforeScore is a function that runs before the Score method of the original plugin.
	// If BeforeScore returns non-success status, the simulator plugin doesn't run the Score method of the original plugin and return that score & status.
	BeforeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status)
	// AfterScore is a function that runs after the Score method of the original plugin.
	// A Score of the simulator plugin finally returns the score & status returned from AfterScore.
	AfterScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string, score int64, scoreResult *framework.Status) (int64, *framework.Status)
}

type NormalizeScorePluginExtender interface {
	// BeforeNormalizeScore is a function that runs before the NormalizeScore method of the original plugin.
	// If BeforeNormalizeScore returns non-success status, the simulator plugin doesn't run the NormalizeScore method of the original plugin and return that status.
	BeforeNormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status
	// AfterNormalizeScore is a function that runs after the NormalizeScore method of the original plugin.
	// If original plugin returns non-success status, a NormalizeScore of the simulator plugin finally returns the status returned from AfterNormalizeScore or original score.
	// Otherwise, return nil.
	AfterNormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList, normalizeScoreResult *framework.Status) *framework.Status
}

// Extenders is a option for some pluginExtenders.
// It will contain some arbitrary processing function defined by a user.
type Extenders struct {
	filterPluginExtender         FilterPluginExtender
	scorePluginExtender          ScorePluginExtender
	normalizeScorePluginExtender NormalizeScorePluginExtender
}

type options struct {
	extenderOption Extenders
}

type extendersOption Extenders

type Option interface {
	apply(*options)
}

func (e extendersOption) apply(opts *options) {
	opts.extenderOption = Extenders(e)
}

func WithExtendersOption(opt *Extenders) Option {
	return extendersOption(*opt)
}

// wrappedPlugin behaves as if it is original plugin, but it records result of plugin.
// All wrappedPlugin's name is originalPlugin name + pluginSuffix.
type wrappedPlugin struct {
	name                         string
	originalFilterPlugin         framework.FilterPlugin
	originalScorePlugin          framework.ScorePlugin
	filterPluginExtender         FilterPluginExtender
	scorePluginExtender          ScorePluginExtender
	normalizeScorePluginExtender NormalizeScorePluginExtender
	weight                       int32
	store                        Store
}

const (
	pluginSuffix       = "Wrapped"
	resultBeforePrefix = "Before"
	resultAfterPrefix  = "After"
)

func pluginName(pluginName string) string {
	return pluginName + pluginSuffix
}

// newWrappedPlugin makes wrappedPlugin from score or/and filter plugin.
func newWrappedPlugin(s Store, p framework.Plugin, weight int32, opts ...Option) framework.Plugin {
	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}

	plg := &wrappedPlugin{
		name:   pluginName(p.Name()),
		weight: weight,
		store:  s,
	}
	if options.extenderOption.filterPluginExtender != nil {
		plg.filterPluginExtender = options.extenderOption.filterPluginExtender
	}
	if options.extenderOption.scorePluginExtender != nil {
		plg.scorePluginExtender = options.extenderOption.scorePluginExtender
	}
	if options.extenderOption.normalizeScorePluginExtender != nil {
		plg.normalizeScorePluginExtender = options.extenderOption.normalizeScorePluginExtender
	}

	fp, ok := p.(framework.FilterPlugin)
	if ok {
		plg.originalFilterPlugin = fp
	}

	sp, ok := p.(framework.ScorePlugin)
	if ok {
		plg.originalScorePlugin = sp
	}

	return plg
}

func (w *wrappedPlugin) Name() string { return w.name }
func (w *wrappedPlugin) ScoreExtensions() framework.ScoreExtensions {
	if w.originalScorePlugin != nil && w.originalScorePlugin.ScoreExtensions() != nil {
		return w
	}
	return nil
}

// NormalizeScore wraps original NormalizeScore plugin of Scheduler Framework.
// Before and after the execution of original NormalizeScore plugin,
// we will run arbitrary processing as functions from normalizeScorePluginExtender.
func (w *wrappedPlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	if w.originalScorePlugin == nil || w.originalScorePlugin.ScoreExtensions() == nil {
		// return nil not to affect scoring
		return nil
	}

	if w.normalizeScorePluginExtender != nil {
		if s := w.doBeforeNormalizeScoreAndStoreResult(ctx, state, pod, scores); !s.IsSuccess() {
			return s
		}
	}
	orignstatus := w.originalScorePlugin.ScoreExtensions().NormalizeScore(ctx, state, pod, scores)
	if !orignstatus.IsSuccess() {
		klog.Errorf("failed to run normalize score: %v, %v", orignstatus.Code(), orignstatus.Message())
		// If it is not nil, we wiil return the results of AfterNormalizeScore.
		if w.normalizeScorePluginExtender != nil {
			return w.doAfterNormalizeScoreAndStoreResult(ctx, state, pod, scores, orignstatus)
		}
		// Otherwise, we will return the original results.
		return orignstatus
	}

	for _, s := range scores {
		w.store.AddNormalizedScoreResult(pod.Namespace, pod.Name, s.Name, w.originalScorePlugin.Name(), s.Score)
	}
	// If it is not nil, we wiil run AfterNormalizeScore.
	if w.normalizeScorePluginExtender != nil {
		_ = w.doAfterNormalizeScoreAndStoreResult(ctx, state, pod, scores, orignstatus)
	}

	return nil
}

func (w *wrappedPlugin) doBeforeNormalizeScoreAndStoreResult(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	s := w.normalizeScorePluginExtender.BeforeNormalizeScore(ctx, state, pod, scores)
	for _, score := range scores {
		w.store.AddNormalizedScoreResult(pod.Namespace, pod.Name, score.Name, resultBeforePrefix+w.originalScorePlugin.Name(), score.Score)
	}
	return s
}

func (w *wrappedPlugin) doAfterNormalizeScoreAndStoreResult(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList, orignstatus *framework.Status) *framework.Status {
	s := w.normalizeScorePluginExtender.AfterNormalizeScore(ctx, state, pod, scores, orignstatus)
	for _, score := range scores {
		w.store.AddNormalizedScoreResult(pod.Namespace, pod.Name, score.Name, resultAfterPrefix+w.originalScorePlugin.Name(), score.Score)
	}
	return s
}

// Score wraps original Score plugin of Scheduler Framework.
// Before and after the execution of original Score plugin,
// we will run arbitrary processing as functions from scorePluginExtender.
func (w *wrappedPlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	if w.originalScorePlugin == nil {
		// return zero-score and nil not to affect scoring
		return 0, nil
	}

	if w.scorePluginExtender != nil {
		score, s := w.doBeforeScoreAndStoreResult(ctx, state, pod, nodeName)
		if !s.IsSuccess() {
			return score, s
		}
	}
	score, s := w.originalScorePlugin.Score(ctx, state, pod, nodeName)
	if !s.IsSuccess() {
		klog.Errorf("failed to run score plugin: %v, %v", s.Code(), s.Message())
	} else {
		w.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, w.originalScorePlugin.Name(), score)
	}
	// If it is not nil, we wiil return the results of AfterScore.
	if w.scorePluginExtender != nil {
		return w.doAfterScoreAndStoreResult(ctx, state, pod, nodeName, score, s)
	}
	return score, s
}

func (w *wrappedPlugin) doBeforeScoreAndStoreResult(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	score, s := w.scorePluginExtender.BeforeScore(ctx, state, pod, nodeName)
	w.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, resultBeforePrefix+w.originalScorePlugin.Name(), score)
	return score, s
}

func (w *wrappedPlugin) doAfterScoreAndStoreResult(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string, score int64, status *framework.Status) (int64, *framework.Status) {
	score, s := w.scorePluginExtender.AfterScore(ctx, state, pod, nodeName, score, status)
	w.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, resultAfterPrefix+w.originalScorePlugin.Name(), score)
	return score, s
}

// Filter wraps original Filter plugin of Scheduler Framework.
// Before and after the execution of original Filter plugin,
// we will run arbitrary processing as functions from filterPluginExtender.
func (w *wrappedPlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if w.originalFilterPlugin == nil {
		// return nil not to affect filtering
		return nil
	}

	if w.filterPluginExtender != nil {
		if s := w.doBeforeFilterAndStoreResult(ctx, state, pod, nodeInfo); !s.IsSuccess() {
			return s
		}
	}
	s := w.originalFilterPlugin.Filter(ctx, state, pod, nodeInfo)
	if s.IsSuccess() {
		w.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, w.originalFilterPlugin.Name(), schedulingresultstore.PassedFilterMessage)
	} else {
		w.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, w.originalFilterPlugin.Name(), s.Message())
	}
	// If it is not nil, we wiil return the results of AfterFilter.
	if w.filterPluginExtender != nil {
		return w.doAfterFilterAndStoreResult(ctx, state, pod, nodeInfo, s)
	}
	return s
}

func (w *wrappedPlugin) doBeforeFilterAndStoreResult(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	s := w.filterPluginExtender.BeforeFilter(ctx, state, pod, nodeInfo)
	w.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, resultBeforePrefix+w.originalFilterPlugin.Name(), s.Message())
	return s
}

func (w *wrappedPlugin) doAfterFilterAndStoreResult(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo, status *framework.Status) *framework.Status {
	s := w.filterPluginExtender.AfterFilter(ctx, state, pod, nodeInfo, status)
	w.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, resultAfterPrefix+w.originalFilterPlugin.Name(), s.Message())
	return s
}