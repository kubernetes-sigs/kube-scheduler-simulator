package plugin

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	schedulingresultstore "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin/resultstore"
)

//go:generate mockgen -destination=./mock/$GOFILE -package=plugin . Store,PreFilterPluginExtender,FilterPluginExtender,PostFilterPluginExtender,PreScorePluginExtender,ScorePluginExtender,NormalizeScorePluginExtender,ReservePluginExtender,PermitPluginExtender,PreBindPluginExtender,BindPluginExtender,PostBindPluginExtender
//go:generate mockgen -destination=./mock/framework.go -package=plugin k8s.io/kubernetes/pkg/scheduler/framework PreFilterPlugin,FilterPlugin,PostFilterPlugin,PreScorePlugin,ScorePlugin,ScoreExtensions,PermitPlugin,BindPlugin,PreBindPlugin,PostBindPlugin,ReservePlugin
type Store interface {
	AddNormalizedScoreResult(namespace, podName, nodeName, pluginName string, normalizedscore int64)
	AddPreFilterResult(namespace, podName, pluginName, reason string, preFilterResult *framework.PreFilterResult)
	AddFilterResult(namespace, podName, nodeName, pluginName, reason string)
	AddPreScoreResult(namespace, podName, pluginName, reason string)
	AddScoreResult(namespace, podName, nodeName, pluginName string, score int64)
	AddPostFilterResult(namespace, podName, nominatedNodeName, pluginName string, nodeNames []string)
	AddPermitResult(namespace, podName, pluginName, status string, timeout time.Duration)
	AddReserveResult(namespace, podName, pluginName, status string)
	AddSelectedNode(namespace, podName, nodeName string)
	AddBindResult(namespace, podName, pluginName, status string)
	AddPreBindResult(namespace, podName, pluginName, status string)
	// AddCustomResult is intended to be used from outside of simulator.
	AddCustomResult(namespace, podName, annotationKey, result string)
}

//nolint:revive
type PluginExtenderInitializer func(handle SimulatorHandle) PluginExtenders

type SimulatorHandle interface {
	// AddCustomResult adds user defined data.
	// The results added through this func is reflected on the Pod's annotation eventually like other scheduling results.
	// This function is intended to be called from the plugin.PluginExtender; allow users to export some internal state on Pods for debugging purpose.
	// For example,
	// Calling AddCustomResult in NodeAffinity's PreFilterPluginExtender:
	// AddCustomResult("namespace", "incomingPod", "node-affinity-filter-internal-state-anno-key", "internal-state")
	// Then, "incomingPod" Pod will get {"node-affinity-filter-internal-state-anno-key": "internal-state"} annotation after scheduling.
	AddCustomResult(namespace, podName, annotationKey, result string)
}

// PreFilterPluginExtender is the extender for PreFilter plugin.
type PreFilterPluginExtender interface {
	// BeforePreFilter is a function that runs before the PreFilter method of the original plugin.
	// If BeforePreFilter returns non-success status, the simulator plugin doesn't run the PreFilter method of the original plugin and return that status.
	BeforePreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (*framework.PreFilterResult, *framework.Status)
	// AfterPreFilter is a function that is run after the PreFilter method of the original plugin.
	// A PreFilter of the simulator plugin finally returns the PreFilterResult and the status returned from AfterPreFilter.
	AfterPreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, preFilterResult *framework.PreFilterResult, preFilterStatus *framework.Status) (*framework.PreFilterResult, *framework.Status)
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

type PostFilterPluginExtender interface {
	// BeforePostFilter is a function that is run before the PostFilter method of the original plugin.
	// If BeforePostFilter return non-success status, the simulator plugin doesn't run the PostFilter method of the original plugin and return that status.
	BeforePostFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status)
	// AfterPostFilter is a function that is run after the PostFilter method of the original plugin.
	// A PostFilter of the simulator plugin finally returns the status returned from PostFilter.
	AfterPostFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap, postFilterResult *framework.PostFilterResult, status *framework.Status) (*framework.PostFilterResult, *framework.Status)
}
type PreScorePluginExtender interface {
	// BeforePreScore is a function that runs before the PreFilter method of the original plugin.
	// If BeforePreScore returns non-success status, the simulator plugin doesn't run the PreScore method of the original plugin and return that status.
	BeforePreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status
	// AfterPreScore is a function that is run after the PreScore method of the original plugin.
	// A PreScore of the simulator plugin finally returns the status returned from AfterPreScore.
	AfterPreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node, preScoreStatus *framework.Status) *framework.Status
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

// ReservePluginExtender is the extender for Reserve plugin.
type ReservePluginExtender interface {
	// BeforeReserve is a function that runs before the Reserve method of the original plugin.
	// If BeforeReserve returns non-success status, the simulator plugin doesn't run the Reserve method of the original plugin and return that status.
	BeforeReserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status
	// AfterReserve is a function that is run after the Reserve method of the original plugin.
	// A Reserve of the simulator plugin finally returns the status returned from AfterReserve.
	AfterReserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string, reserveStatus *framework.Status) *framework.Status
	// BeforeUnreserve is a function that runs before the Reserve method of the original plugin.
	// If BeforeUnreserve returns non-success status, the simulator plugin doesn't run the Reserve method of the original plugin.
	BeforeUnreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status
	// AfterUnreserve is a function that is run after the Unreserve method of the original plugin.
	AfterUnreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string)
}

// PermitPluginExtender is the extender for Permit plugin.
type PermitPluginExtender interface {
	// BeforePermit is a function that runs before the Permit method of the original plugin.
	// If BeforePermit returns non-success status, the simulator plugin doesn't run the Permit method of the original plugin and return that status.
	BeforePermit(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (*framework.Status, time.Duration)
	// AfterPermit is a function that runs after the Permit method of the original plugin.
	// A Permit of the simulator plugins finally returns the status returned from AfterPermit.
	AfterPermit(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string, permitResult *framework.Status, timeout time.Duration) (*framework.Status, time.Duration)
}

type PreBindPluginExtender interface {
	// BeforePreBind is a function that runs before the PreBind method of the original plugin.
	// If BeforePreBind returns non-success status, the simulator plugin doesn't run the PreBind method of the original plugin and return that status.
	BeforePreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status
	// AfterPreBind is a function that is run after the Bind method of the original plugin.
	// A PreBind of the simulator plugin finally returns the status returned from AfterBind.
	AfterPreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string, bindResult *framework.Status) *framework.Status
}

type BindPluginExtender interface {
	// BeforeBind is a function that runs before the Bind method of the original plugin.
	// If BeforeBind returns non-success status, the simulator plugin doesn't run the Bind method of the original plugin and return that status.
	BeforeBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status
	// AfterBind is a function that is run after the Bind method of the original plugin.
	// A Bind of the simulator plugin finally returns the status returned from AfterBind.
	AfterBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string, bindResult *framework.Status) *framework.Status
}

type PostBindPluginExtender interface {
	// BeforePostBind is a function that runs before the PostBind method of the original plugin.
	// If BeforePostBind returns non-success status, the simulator plugin doesn't run the PostBind method of the original plugin and return that status.
	BeforePostBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status
	// AfterPostBind is a function that is run after the PostBind method of the original plugin.
	AfterPostBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string)
}

// PluginExtenders contains XXXXPluginExtenders.
// Each extender will intercept a calling to target method call of scheduler plugins,
// and you can check/modify requests and/or results.
//
//nolint:revive // intended to name it PluginExtenders to distinguish from the scheduler's extender.
type PluginExtenders struct {
	PreFilterPluginExtender      PreFilterPluginExtender
	FilterPluginExtender         FilterPluginExtender
	PostFilterPluginExtender     PostFilterPluginExtender
	PreScorePluginExtender       PreScorePluginExtender
	ScorePluginExtender          ScorePluginExtender
	NormalizeScorePluginExtender NormalizeScorePluginExtender
	PermitPluginExtender         PermitPluginExtender
	ReservePluginExtender        ReservePluginExtender
	PreBindPluginExtender        PreBindPluginExtender
	BindPluginExtender           BindPluginExtender
	PostBindPluginExtender       PostBindPluginExtender
}

type options struct {
	extenderInitializerOption PluginExtenderInitializer
	pluginNameOption          string
}

type (
	extendersOption  PluginExtenderInitializer
	pluginNameOption string
)

type Option interface {
	apply(*options)
}

func (e extendersOption) apply(opts *options) {
	opts.extenderInitializerOption = PluginExtenderInitializer(e)
}

func (p pluginNameOption) apply(opts *options) {
	opts.pluginNameOption = string(p)
}

// WithExtendersOption provides an easy way to extend the behavior of the plugin.
// These containing functions in PluginExtenders should be run before and after the original plugin of Scheduler Framework.
func WithExtendersOption(opt PluginExtenderInitializer) Option {
	return extendersOption(opt)
}

// WithPluginNameOption contains configuration options for the name field of a wrappedPlugin.
func WithPluginNameOption(opt *string) Option {
	return pluginNameOption(*opt)
}

// wrappedPlugin behaves as if it is original plugin, but it records result of plugin.
type wrappedPlugin struct {
	// name is plugin's name returned by Name() method.
	// This name is default to original plugin name + pluginSuffix.
	// You can change this name by WithPluginNameOption.
	name string
	// store records plugin's result.
	// TODO: move store's logic to plugin extender.
	store Store

	originalPreEnqueuePlugin framework.PreEnqueuePlugin
	originalPreFilterPlugin  framework.PreFilterPlugin
	originalFilterPlugin     framework.FilterPlugin
	originalPreScorePlugin   framework.PreScorePlugin
	originalPostFilterPlugin framework.PostFilterPlugin
	originalScorePlugin      framework.ScorePlugin
	originalPermitPlugin     framework.PermitPlugin
	originalReservePlugin    framework.ReservePlugin
	originalPreBindPlugin    framework.PreBindPlugin
	originalBindPlugin       framework.BindPlugin
	originalPostBindPlugin   framework.PostBindPlugin

	// plugin extenders
	preFilterPluginExtender      PreFilterPluginExtender
	filterPluginExtender         FilterPluginExtender
	postFilterPluginExtender     PostFilterPluginExtender
	scorePluginExtender          ScorePluginExtender
	preScorePluginExtender       PreScorePluginExtender
	normalizeScorePluginExtender NormalizeScorePluginExtender
	permitPluginExtender         PermitPluginExtender
	reservePluginExtender        ReservePluginExtender
	preBindPluginExtender        PreBindPluginExtender
	bindPluginExtender           BindPluginExtender
	postBindPluginExtender       PostBindPluginExtender
}

const (
	pluginSuffix = "Wrapped"
)

func pluginName(pluginName string) string {
	return pluginName + pluginSuffix
}

// NewWrappedPlugin makes wrappedPlugin from score or/and filter plugin.
//
//nolint:funlen,cyclop
func NewWrappedPlugin(s Store, p framework.Plugin, opts ...Option) framework.Plugin {
	options := options{
		// default value to create empty extenders.
		extenderInitializerOption: func(handle SimulatorHandle) PluginExtenders { return PluginExtenders{} },
	}
	for _, o := range opts {
		o.apply(&options)
	}
	pName := pluginName(p.Name())
	if options.pluginNameOption != "" {
		pName = options.pluginNameOption
	}

	plg := &wrappedPlugin{
		name:  pName,
		store: s,
	}

	extender := options.extenderInitializerOption(s)

	if extender.PreFilterPluginExtender != nil {
		plg.preFilterPluginExtender = extender.PreFilterPluginExtender
	}
	if extender.FilterPluginExtender != nil {
		plg.filterPluginExtender = extender.FilterPluginExtender
	}
	if extender.PostFilterPluginExtender != nil {
		plg.postFilterPluginExtender = extender.PostFilterPluginExtender
	}
	if extender.ScorePluginExtender != nil {
		plg.scorePluginExtender = extender.ScorePluginExtender
	}
	if extender.PreScorePluginExtender != nil {
		plg.preScorePluginExtender = extender.PreScorePluginExtender
	}
	if extender.NormalizeScorePluginExtender != nil {
		plg.normalizeScorePluginExtender = extender.NormalizeScorePluginExtender
	}
	if extender.PermitPluginExtender != nil {
		plg.permitPluginExtender = extender.PermitPluginExtender
	}
	if extender.ReservePluginExtender != nil {
		plg.reservePluginExtender = extender.ReservePluginExtender
	}
	if extender.PreBindPluginExtender != nil {
		plg.preBindPluginExtender = extender.PreBindPluginExtender
	}
	if extender.BindPluginExtender != nil {
		plg.bindPluginExtender = extender.BindPluginExtender
	}
	if extender.PostBindPluginExtender != nil {
		plg.postBindPluginExtender = extender.PostBindPluginExtender
	}

	peqp, ok := p.(framework.PreEnqueuePlugin)
	if ok {
		plg.originalPreEnqueuePlugin = peqp
	}
	prefp, ok := p.(framework.PreFilterPlugin)
	if ok {
		plg.originalPreFilterPlugin = prefp
	}
	fp, ok := p.(framework.FilterPlugin)
	if ok {
		plg.originalFilterPlugin = fp
	}
	pfp, ok := p.(framework.PostFilterPlugin)
	if ok {
		plg.originalPostFilterPlugin = pfp
	}
	presp, ok := p.(framework.PreScorePlugin)
	if ok {
		plg.originalPreScorePlugin = presp
	}
	sp, ok := p.(framework.ScorePlugin)
	if ok {
		plg.originalScorePlugin = sp
	}
	pp, ok := p.(framework.PermitPlugin)
	if ok {
		plg.originalPermitPlugin = pp
	}
	rp, ok := p.(framework.ReservePlugin)
	if ok {
		plg.originalReservePlugin = rp
	}

	bp, ok := p.(framework.BindPlugin)
	if ok {
		plg.originalBindPlugin = bp
	}

	prebp, ok := p.(framework.PreBindPlugin)
	if ok {
		plg.originalPreBindPlugin = prebp
	}

	postbp, ok := p.(framework.PostBindPlugin)
	if ok {
		plg.originalPostBindPlugin = postbp
	}

	queuesortp, ok := p.(framework.QueueSortPlugin)
	if ok {
		// There must be only one in each profile for which the QueueSortPlugin interface is implemented.
		newplug := &wrappedPluginWithQueueSort{wrappedPlugin: *plg}
		newplug.originalQueueSortPlugin = queuesortp
		return newplug
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

// PreEnqueue wraps original PreEnqueue plugin of Scheduler Framework.
// TODO: Implements before/after PreEnqueue function.
func (w *wrappedPlugin) PreEnqueue(ctx context.Context, p *v1.Pod) *framework.Status {
	if w.originalPreEnqueuePlugin == nil {
		// return nil not to affect queuing
		return nil
	}

	return w.originalPreEnqueuePlugin.PreEnqueue(ctx, p)
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
		klog.Errorf("failed to run normalize score. Normalized scores won't be recorded on Pod annotation: %v, %v", s.Code(), s.Message())
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
		klog.Errorf("failed to run score plugin. Scores won't be recorded on Pod annotation: %v, %v", s.Code(), s.Message())
	} else {
		// TODO: move to AfterScore.
		w.store.AddScoreResult(pod.Namespace, pod.Name, nodeName, w.originalScorePlugin.Name(), score)
	}

	if w.scorePluginExtender != nil {
		return w.scorePluginExtender.AfterScore(ctx, state, pod, nodeName, score, s)
	}
	return score, s
}

func (w *wrappedPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	if w.originalPreFilterPlugin == nil {
		// return nils not to affect scoring
		return nil
	}

	return w.originalPreFilterPlugin.PreFilterExtensions()
}

// PreScore wraps original PreScore plugin of Scheduler Framework.
// You can run your function before and/or after the execution of original PreScore plugin
// by configuring with WithExtendersOption.
func (w *wrappedPlugin) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	if w.originalPreScorePlugin == nil {
		// return nil not to affect scoring
		return nil
	}

	if w.preScorePluginExtender != nil {
		s := w.preScorePluginExtender.BeforePreScore(ctx, state, pod, nodes)
		if !s.IsSuccess() {
			return s
		}
	}

	s := w.originalPreScorePlugin.PreScore(ctx, state, pod, nodes)
	var msg string
	if s.IsSuccess() {
		msg = schedulingresultstore.SuccessMessage
	} else {
		msg = s.Message()
	}
	w.store.AddPreScoreResult(pod.Namespace, pod.Name, w.originalPreScorePlugin.Name(), msg)

	if w.preScorePluginExtender != nil {
		return w.preScorePluginExtender.AfterPreScore(ctx, state, pod, nodes, s)
	}

	return s
}

// PreFilter wraps original PreFilter plugin of Scheduler Framework.
// You can run your function before and/or after the execution of original PreFilter plugin
// by configuring with WithExtendersOption.
func (w *wrappedPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	if w.originalPreFilterPlugin == nil {
		// return nils not to affect scoring
		return nil, nil
	}

	if w.preFilterPluginExtender != nil {
		r, s := w.preFilterPluginExtender.BeforePreFilter(ctx, state, p)
		if !s.IsSuccess() {
			return r, s
		}
	}

	result, s := w.originalPreFilterPlugin.PreFilter(ctx, state, p)
	var msg string
	if s.IsSuccess() {
		msg = schedulingresultstore.SuccessMessage
	} else {
		msg = s.Message()
	}
	w.store.AddPreFilterResult(p.Namespace, p.Name, w.originalPreFilterPlugin.Name(), msg, result)

	if w.preFilterPluginExtender != nil {
		return w.preFilterPluginExtender.AfterPreFilter(ctx, state, p, result, s)
	}

	return result, s
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
	w.store.AddFilterResult(pod.Namespace, pod.Name, nodeInfo.Node().Name, w.originalFilterPlugin.Name(), msg)

	if w.filterPluginExtender != nil {
		return w.filterPluginExtender.AfterFilter(ctx, state, pod, nodeInfo, s)
	}
	return s
}

func (w *wrappedPlugin) PostFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	if w.originalPostFilterPlugin == nil {
		// return Unschedulable not to affect post filtering.
		// (If return Unschedulable, the scheduler will execute next PostFilter plugin.)
		return nil, framework.NewStatus(framework.Unschedulable)
	}
	if w.postFilterPluginExtender != nil {
		r, s := w.postFilterPluginExtender.BeforePostFilter(ctx, state, pod, filteredNodeStatusMap)
		if !s.IsSuccess() {
			return r, s
		}
	}
	r, s := w.originalPostFilterPlugin.PostFilter(ctx, state, pod, filteredNodeStatusMap)
	var nominatedNodeName string
	if s.IsSuccess() {
		nominatedNodeName = r.NominatedNodeName
	}
	nodeNames := make([]string, 0, len(filteredNodeStatusMap))
	for k := range filteredNodeStatusMap {
		nodeNames = append(nodeNames, k)
	}
	w.store.AddPostFilterResult(pod.Namespace, pod.Name, nominatedNodeName, w.originalPostFilterPlugin.Name(), nodeNames)

	if w.postFilterPluginExtender != nil {
		return w.postFilterPluginExtender.AfterPostFilter(ctx, state, pod, filteredNodeStatusMap, r, s)
	}
	return r, s
}

// Permit wraps original Permit plugin of Scheduler Framework.
// You can run your function before and/or after the execution of original Permit plugin
// by configuring with WithExtendersOption.
func (w *wrappedPlugin) Permit(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	if w.originalPermitPlugin == nil {
		// return zero-score and nil not to affect scoring
		return nil, 0
	}

	if w.permitPluginExtender != nil {
		s, d := w.permitPluginExtender.BeforePermit(ctx, state, pod, nodeName)
		if !s.IsSuccess() {
			return s, d
		}
	}

	s, timeout := w.originalPermitPlugin.Permit(ctx, state, pod, nodeName)
	msg := s.Message()
	if s.IsSuccess() {
		msg = schedulingresultstore.SuccessMessage
	}
	if s.IsWait() {
		msg = schedulingresultstore.WaitMessage
	}

	w.store.AddPermitResult(pod.Namespace, pod.Name, w.originalPermitPlugin.Name(), msg, timeout)

	if w.permitPluginExtender != nil {
		return w.permitPluginExtender.AfterPermit(ctx, state, pod, nodeName, s, timeout)
	}

	return s, timeout
}

// Reserve wraps original Reserve plugin of Scheduler Framework.
// You can run your function before and/or after the execution of original Reserve plugin
// by configuring with WithExtendersOption.
func (w *wrappedPlugin) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status {
	w.store.AddSelectedNode(pod.Namespace, pod.Name, nodename)

	if w.originalReservePlugin == nil {
		// return nil not to affect scoring
		return nil
	}

	if w.reservePluginExtender != nil {
		s := w.reservePluginExtender.BeforeReserve(ctx, state, pod, nodename)
		if !s.IsSuccess() {
			return s
		}
	}

	s := w.originalReservePlugin.Reserve(ctx, state, pod, nodename)
	var msg string
	if s.IsSuccess() {
		msg = schedulingresultstore.SuccessMessage
	} else {
		msg = s.Message()
	}
	w.store.AddReserveResult(pod.Namespace, pod.Name, w.originalReservePlugin.Name(), msg)

	if w.reservePluginExtender != nil {
		return w.reservePluginExtender.AfterReserve(ctx, state, pod, nodename, s)
	}

	return s
}

// Unreserve wraps original Unreserve plugin of Scheduler Framework.
// You can run your function before and/or after the execution of original Unreserve plugin
// by configuring with WithExtendersOption.
func (w *wrappedPlugin) Unreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) {
	if w.originalReservePlugin == nil {
		return
	}

	if w.reservePluginExtender != nil {
		s := w.reservePluginExtender.BeforeUnreserve(ctx, state, pod, nodename)
		if !s.IsSuccess() {
			klog.ErrorS(nil, "reservePluginExtender.BeforeUnreserve returned non success status, won't run Unreserve", "status_message", s.Message(), "plugin", w.originalReservePlugin.Name())
			return
		}
	}

	w.originalReservePlugin.Unreserve(ctx, state, pod, nodename)

	if w.reservePluginExtender != nil {
		w.reservePluginExtender.AfterUnreserve(ctx, state, pod, nodename)
	}
}

func (w *wrappedPlugin) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status {
	if w.originalPreBindPlugin == nil {
		// return nil not to affect scoring
		return nil
	}

	if w.preBindPluginExtender != nil {
		s := w.preBindPluginExtender.BeforePreBind(ctx, state, pod, nodename)
		if !s.IsSuccess() {
			return s
		}
	}

	s := w.originalPreBindPlugin.PreBind(ctx, state, pod, nodename)
	var msg string
	if s.IsSuccess() {
		msg = schedulingresultstore.SuccessMessage
	} else {
		msg = s.Message()
	}
	w.store.AddPreBindResult(pod.Namespace, pod.Name, w.originalPreBindPlugin.Name(), msg)

	if w.preBindPluginExtender != nil {
		return w.preBindPluginExtender.AfterPreBind(ctx, state, pod, nodename, s)
	}

	return s
}

func (w *wrappedPlugin) Bind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) *framework.Status {
	if w.originalBindPlugin == nil {
		// return skip not to affect other bind plugins.
		return framework.NewStatus(framework.Skip, "called wrapped bind plugin is nil")
	}

	if w.bindPluginExtender != nil {
		s := w.bindPluginExtender.BeforeBind(ctx, state, pod, nodename)
		if !s.IsSuccess() {
			return s
		}
	}

	s := w.originalBindPlugin.Bind(ctx, state, pod, nodename)
	var msg string
	if s.IsSuccess() {
		msg = schedulingresultstore.SuccessMessage
	} else {
		msg = s.Message()
	}
	w.store.AddBindResult(pod.Namespace, pod.Name, w.originalBindPlugin.Name(), msg)

	if w.bindPluginExtender != nil {
		return w.bindPluginExtender.AfterBind(ctx, state, pod, nodename, s)
	}

	return s
}

func (w *wrappedPlugin) PostBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodename string) {
	if w.originalPostBindPlugin == nil {
		return
	}

	if w.postBindPluginExtender != nil {
		s := w.postBindPluginExtender.BeforePostBind(ctx, state, pod, nodename)
		if !s.IsSuccess() {
			klog.ErrorS(nil, "postBindPluginExtender.BeforePostBind returned non success status, won't run PostBind", "status_message", s.Message(), "plugin", w.originalPostBindPlugin.Name())
			return
		}
	}

	w.originalPostBindPlugin.PostBind(ctx, state, pod, nodename)

	if w.postBindPluginExtender != nil {
		w.postBindPluginExtender.AfterPostBind(ctx, state, pod, nodename)
	}
}

// wrappedPluginWithQueueSort behaves as if it is original plugin and QueueSort plugin.
// To support MultiPoint field, we are required to separate WrappedPlugin and the implementation of QueueSort interface.
type wrappedPluginWithQueueSort struct {
	wrappedPlugin

	originalQueueSortPlugin framework.QueueSortPlugin
}

func (w *wrappedPluginWithQueueSort) Name() string { return w.wrappedPlugin.Name() }

// Less  wraps original Less plugin of Scheduler Framework.
func (w *wrappedPluginWithQueueSort) Less(pod1 *framework.QueuedPodInfo, pod2 *framework.QueuedPodInfo) bool {
	if w.originalQueueSortPlugin == nil {
		return false
	}

	return w.originalQueueSortPlugin.Less(pod1, pod2)
}
