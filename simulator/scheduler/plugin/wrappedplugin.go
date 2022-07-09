package plugin

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	schedulingresultstore "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin/resultstore"
)

//go:generate mockgen -destination=./mock/$GOFILE -package=plugin . Store,FilterPluginExtender,ScorePluginExtender,NormalizeScorePluginExtender
//go:generate mockgen -destination=./mock/framework.go -package=plugin k8s.io/kubernetes/pkg/scheduler/framework FilterPlugin,ScorePlugin,ScoreExtensions
type Store interface {
	AddNormalizedScoreResult(namespace, podName, nodeName, pluginName string, normalizedscore int64)
	AddFilterResult(namespace, podName, nodeName, pluginName, reason string)
	AddScoreResult(namespace, podName, nodeName, pluginName string, score int64)
}

// FilterPluginExtender is the extender for Filter plugin.
type FilterPluginExtender interface {
	// BeforeFilter is a function that runs before the Filter method of the original plugin.
	// If BeforeFilter returns non-success status, the simulator plugin doesn't run the Filter method of the original plugin and return that status.
	BeforeFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status
	// AfterFilter is a function that is run after the Filter method of the original plugin.
	// A Filter of the simulator plugin finally returns the status returned from AfterFilter.
	AfterFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo, filterResult *framework.Status) *framework.Status
}

// ScorePluginExtender is the extender for Score plugin.
type ScorePluginExtender interface {
	// BeforeScore is a function that runs before the Score method of the original plugin.
	// If BeforeScore returns non-success status, the simulator plugin doesn't run the Score method of the original plugin and return that score & status.
	BeforeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status)
	// AfterScore is a function that runs after the Score method of the original plugin.
	// A Score of the simulator plugin finally returns the score & status returned from AfterScore.
	AfterScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string, score int64, scoreResult *framework.Status) (int64, *framework.Status)
}

// NormalizeScorePluginExtender is the extender for NormalizeScore plugin.
type NormalizeScorePluginExtender interface {
	// BeforeNormalizeScore is a function that runs before the NormalizeScore method of the original plugin.
	// If BeforeNormalizeScore returns non-success status, the simulator plugin doesn't run the NormalizeScore method of the original plugin and return that status.
	BeforeNormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status
	// AfterNormalizeScore is a function that runs after the NormalizeScore method of the original plugin.
	// A NormalizeScore of the simulator plugins finally returns the status returned from AfterNormalizeScore.
	AfterNormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList, normalizeScoreResult *framework.Status) *framework.Status
}

// Extenders contains XXXXPluginExtenders.
// Each extender will intercept a calling to target method call of scheduler plugins,
// and you can check/modify requests and/or results.
type Extenders struct {
	FilterPluginExtender         FilterPluginExtender
	ScorePluginExtender          ScorePluginExtender
	NormalizeScorePluginExtender NormalizeScorePluginExtender
}

type options struct {
	extenderOption   Extenders
	pluginNameOption string
	weightOption     int32
}

type (
	extendersOption  Extenders
	pluginNameOption string
	weightOption     int32
)

type Option interface {
	apply(*options)
}

func (e extendersOption) apply(opts *options) {
	opts.extenderOption = Extenders(e)
}

func (p pluginNameOption) apply(opts *options) {
	opts.pluginNameOption = string(p)
}

func (w weightOption) apply(opts *options) {
	opts.weightOption = int32(w)
}

// WithExtendersOption provides an easy way to extend the behavior of the plugin.
// These containing functions in Extenders should be run before and after the original plugin of Scheduler Framework.
func WithExtendersOption(opt *Extenders) Option {
	return extendersOption(*opt)
}

// WithPluginNameOption contains configuration options for the name field of a wrappedPlugin.
func WithPluginNameOption(opt *string) Option {
	return pluginNameOption(*opt)
}

// WithWeightOption contains configuration options for the weight field of a wrappedPlugin.
func WithWeightOption(opt *int32) Option {
	return weightOption(*opt)
}

// wrappedPlugin behaves as if it is original plugin, but it records result of plugin.
type wrappedPlugin struct {
	// name is plugin's name returned by Name() method.
	// This name is default to original plugin name + pluginSuffix.
	// You can change this name by WithPluginNameOption.
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
	pluginSuffix = "Wrapped"
)

func pluginName(pluginName string) string {
	return pluginName + pluginSuffix
}

// NewWrappedPlugin makes wrappedPlugin from score or/and filter plugin.
func NewWrappedPlugin(s Store, p framework.Plugin, opts ...Option) framework.Plugin {
	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}
	pName := pluginName(p.Name())
	if options.pluginNameOption != "" {
		pName = options.pluginNameOption
	}

	plg := &wrappedPlugin{
		name:   pName,
		weight: options.weightOption,
		store:  s,
	}
	if options.extenderOption.FilterPluginExtender != nil {
		plg.filterPluginExtender = options.extenderOption.FilterPluginExtender
	}
	if options.extenderOption.ScorePluginExtender != nil {
		plg.scorePluginExtender = options.extenderOption.ScorePluginExtender
	}
	if options.extenderOption.NormalizeScorePluginExtender != nil {
		plg.normalizeScorePluginExtender = options.extenderOption.NormalizeScorePluginExtender
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
// You can run your function before and/or after the execution of original NormalizeScore plugin
// by configuring with WithExtendersOption.
func (w *wrappedPlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	if w.originalScorePlugin == nil || w.originalScorePlugin.ScoreExtensions() == nil {
		// return nil not to affect scoring
		return nil
	}

	if w.normalizeScorePluginExtender != nil {
		if s := w.normalizeScorePluginExtender.BeforeNormalizeScore(ctx, state, pod, scores); !s.IsSuccess() {
			return s
		}
	}

	s := w.originalScorePlugin.ScoreExtensions().NormalizeScore(ctx, state, pod, scores)
	if !s.IsSuccess() {
		klog.Errorf("failed to run normalize score: %v, %v", s.Code(), s.Message())
	} else {
		// TODO: move to AfterNormalizeScore.
		for _, s := range scores {
			w.store.AddNormalizedScoreResult(pod.Namespace, pod.Name, s.Name, w.originalScorePlugin.Name(), s.Score)
		}
	}

	if w.normalizeScorePluginExtender != nil {
		return w.normalizeScorePluginExtender.AfterNormalizeScore(ctx, state, pod, scores, s)
	}

	return s
}

// Score wraps original Score plugin of Scheduler Framework.
// You can run your function before and/or after the execution of original Score plugin
// by configuring with WithExtendersOption.
func (w *wrappedPlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	if w.originalScorePlugin == nil {
		// return zero-score and nil not to affect scoring
		return 0, nil
	}

	if w.scorePluginExtender != nil {
		score, s := w.scorePluginExtender.BeforeScore(ctx, state, pod, nodeName)
		if !s.IsSuccess() {
			return score, s
		}
	}

	score, s := w.originalScorePlugin.Score(ctx, state, pod, nodeName)
	if !s.IsSuccess() {
		klog.Errorf("failed to run score plugin: %v, %v", s.Code(), s.Message())
	} else {
		// TODO: move to AfterScore.
		w.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, w.originalScorePlugin.Name(), score)
	}

	if w.scorePluginExtender != nil {
		return w.scorePluginExtender.AfterScore(ctx, state, pod, nodeName, score, s)
	}
	return score, s
}

// Filter wraps original Filter plugin of Scheduler Framework.
// You can run your function before and/or after the execution of original Filter plugin
// by configuring with WithExtendersOption.
func (w *wrappedPlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if w.originalFilterPlugin == nil {
		// return nil not to affect filtering
		return nil
	}

	if w.filterPluginExtender != nil {
		if s := w.filterPluginExtender.BeforeFilter(ctx, state, pod, nodeInfo); !s.IsSuccess() {
			return s
		}
	}

	s := w.originalFilterPlugin.Filter(ctx, state, pod, nodeInfo)
	var msg string
	if s.IsSuccess() {
		msg = schedulingresultstore.PassedFilterMessage
	} else {
		msg = s.Message()
	}
	// TODO: move to AfterFilter.
	w.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, w.originalFilterPlugin.Name(), msg)

	if w.filterPluginExtender != nil {
		return w.filterPluginExtender.AfterFilter(ctx, state, pod, nodeInfo, s)
	}
	return s
}
