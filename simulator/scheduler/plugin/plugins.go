package plugin

import (
	"encoding/json"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/simulator/scheduler/config"
	schedulingresultstore "github.com/kubernetes-sigs/kube-scheduler-simulator/simulator/scheduler/plugin/resultstore"
)

//nolint: cyclop
func NewRegistry(informerFactory informers.SharedInformerFactory, client clientset.Interface) (map[string]schedulerRuntime.PluginFactory, error) {
	scorePluginWeight := map[string]int32{}
	registeredScorePlugin, err := config.RegisteredScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get registered score plugins: %w", err)
	}
	for _, p := range registeredScorePlugin {
		scorePluginWeight[p.Name] = 0
		if p.Weight != nil {
			scorePluginWeight[p.Name] = *p.Weight
		}
	}

	registeredpls, err := registeredFilterScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default score/filter plugins: %w", err)
	}

	store := schedulingresultstore.New(informerFactory, client, scorePluginWeight)
	intreeRegistries := config.InTreeRegistries()
	outoftreeRegistries := config.OutOfTreeRegistries()
	ret := map[string]schedulerRuntime.PluginFactory{}
	for _, pl := range registeredpls {
		pl := pl

		r, ok := intreeRegistries[pl.Name]
		if !ok {
			// not found in intreeRegistries. search registry in outoftreeRegistries.
			r, ok = outoftreeRegistries[pl.Name]
			if !ok {
				return nil, xerrors.Errorf("registry for %s is not found", pl.Name)
			}
			// For out-of-tree plugins, we need to add original registry to registries.
			// (For in-tree plugins, schedulers add original registry to registries internally.)
			ret[pl.Name] = r
		}

		if _, ok := ret[pluginName(pl.Name)]; ok {
			// already created
			continue
		}

		factory := func(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
			p, err := r(configuration, f)
			if err != nil {
				return nil, xerrors.Errorf("create original plugin: %w", err)
			}

			var weight int32
			if pl.Weight != nil {
				weight = *pl.Weight
			}

			return NewWrappedPlugin(store, p, WithWeightOption(&weight)), nil
		}
		ret[pluginName(pl.Name)] = factory
	}

	return ret, nil
}

// NewPluginConfig converts []v1beta2.PluginConfig for simulator.
// Passed []v1beta.PluginConfig overrides default config values.
//
// NewPluginConfig expects that either PluginConfig.Args.Raw or PluginConfig.Args.Object has data
// in the passed v1beta2.PluginConfig parameter.
// If data exists in both PluginConfig.Args.Raw and PluginConfig.Args.Object, PluginConfig.Args.Raw would be ignored
// since PluginConfig.Args.Object has higher priority.
//nolint:funlen,cyclop
func NewPluginConfig(pc []v1beta2.PluginConfig) ([]v1beta2.PluginConfig, error) {
	defaultcfg, err := config.DefaultSchedulerConfig()
	if err != nil || len(defaultcfg.Profiles) != 1 {
		return nil, xerrors.Errorf("get default scheduler configuration: %w", err)
	}

	pluginConfig := make(map[string]*runtime.RawExtension, len(defaultcfg.Profiles[0].PluginConfig))
	for i := range defaultcfg.Profiles[0].PluginConfig {
		name := defaultcfg.Profiles[0].PluginConfig[i].Name
		pluginConfig[name] = &defaultcfg.Profiles[0].PluginConfig[i].Args
	}

	for i := range pc {
		name := pc[i].Name
		if _, ok := pluginConfig[name]; !ok {
			// it's non-in-tree's plugin's config.
			pluginConfig[name] = &pc[i].Args
			continue
		}

		ret := pluginConfig[name].DeepCopy()
		// If ret is nil, to reference ret.Object is occurred invalid memory address or nil pointer dereference.
		// To avoid this error, if ret is nil, we continue to next loop.
		if ret == nil {
			continue
		}

		// v1beta2.PluginConfig may have data in pc[i].Args.Raw as []byte.
		// We have to encoding it in this case.
		if len(pc[i].Args.Raw) != 0 {
			// override default configuration
			if err := json.Unmarshal(pc[i].Args.Raw, &ret.Object); err != nil {
				return nil, err
			}
		}

		if pc[i].Args.Object != nil {
			// If data exists in both PluginConfig.Args.Raw and PluginConfig.Args.Object,
			// PluginConfig.Args.Raw would be ignored
			ret.Object = pc[i].Args.Object
		}

		pluginConfig[name] = ret
	}

	ret := make([]v1beta2.PluginConfig, 0, len(pluginConfig))
	for name, arg := range pluginConfig {
		// add plugin configs for default plugins.
		ret = append(ret, v1beta2.PluginConfig{
			Name: name,
			Args: *arg,
		})
	}

	defaultpls, err := registeredFilterScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default score/filter plugins: %w", err)
	}

	for _, p := range defaultpls {
		name := p.Name
		pc, ok := pluginConfig[name]
		if !ok {
			continue
		}

		ret = append(ret, v1beta2.PluginConfig{
			Name: pluginName(name),
			Args: *pc,
		})

		// avoid adding same plugin config.
		delete(pluginConfig, name)
	}

	return ret, nil
}

// ConvertForSimulator convert v1beta2.Plugins for simulator.
// It ignores non-default plugin.
func ConvertForSimulator(pls *v1beta2.Plugins) (*v1beta2.Plugins, error) {
	newpls := pls.DeepCopy()

	defaultScorePls, err := config.InTreeScorePluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default score plugins: %w", err)
	}
	merged := mergePluginSet(defaultScorePls, pls.Score)

	retscorepls := make([]v1beta2.Plugin, 0, len(merged.Enabled))
	for _, p := range merged.Enabled {
		retscorepls = append(retscorepls, v1beta2.Plugin{Name: pluginName(p.Name), Weight: p.Weight})
	}
	newpls.Score.Enabled = retscorepls

	// disable default plugins whatever scheduler configuration value is
	newpls.Score.Disabled = []v1beta2.Plugin{
		{
			Name: "*",
		},
	}

	defaultFilterPls, err := config.InTreeFilterPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default score plugins: %w", err)
	}
	merged = mergePluginSet(defaultFilterPls, pls.Filter)

	retfilterpls := make([]v1beta2.Plugin, 0, len(merged.Enabled))
	for _, p := range merged.Enabled {
		retfilterpls = append(retfilterpls, v1beta2.Plugin{Name: pluginName(p.Name), Weight: p.Weight})
	}
	newpls.Filter.Enabled = retfilterpls

	// disable default plugins whatever scheduler configuration value is
	newpls.Filter.Disabled = []v1beta2.Plugin{
		{
			Name: "*",
		},
	}

	return newpls, nil
}

// mergePluginsSet merges two plugin sets.
// This function is copied from k8s.io/kubernetes/pkg/scheduler/apis/config/v1beta2/default_config.go.
func mergePluginSet(inTreePluginSet, outOfTreePluginSet v1beta2.PluginSet) v1beta2.PluginSet {
	type pluginIndex struct {
		index  int
		plugin v1beta2.Plugin
	}

	disabledPlugins := sets.NewString()
	enabledCustomPlugins := make(map[string]pluginIndex)
	// replacedPluginIndex is a set of index of plugins, which have replaced the default plugins.
	replacedPluginIndex := sets.NewInt()
	for _, disabledPlugin := range outOfTreePluginSet.Disabled {
		disabledPlugins.Insert(disabledPlugin.Name)
	}
	for index, enabledPlugin := range outOfTreePluginSet.Enabled {
		enabledCustomPlugins[enabledPlugin.Name] = pluginIndex{index, enabledPlugin}
	}
	var enabledPlugins []v1beta2.Plugin
	if !disabledPlugins.Has("*") {
		for _, defaultEnabledPlugin := range inTreePluginSet.Enabled {
			if disabledPlugins.Has(defaultEnabledPlugin.Name) {
				continue
			}
			// The default plugin is explicitly re-configured, update the default plugin accordingly.
			if customPlugin, ok := enabledCustomPlugins[defaultEnabledPlugin.Name]; ok {
				klog.InfoS("Default plugin is explicitly re-configured; overriding", "plugin", defaultEnabledPlugin.Name)
				// Update the default plugin in place to preserve order.
				defaultEnabledPlugin = customPlugin.plugin
				replacedPluginIndex.Insert(customPlugin.index)
			}
			enabledPlugins = append(enabledPlugins, defaultEnabledPlugin)
		}
	}

	// Append all the custom plugins which haven't replaced any default plugins.
	// Note: duplicated custom plugins will still be appended here.
	// If so, the instantiation of scheduler framework will detect it and abort.
	for index, plugin := range outOfTreePluginSet.Enabled {
		if !replacedPluginIndex.Has(index) {
			enabledPlugins = append(enabledPlugins, plugin)
		}
	}
	return v1beta2.PluginSet{Enabled: enabledPlugins}
}

// registeredFilterScorePlugins returns all registered score plugin and filter plugin.
func registeredFilterScorePlugins() ([]v1beta2.Plugin, error) {
	registeredfilterpls, err := config.RegisteredFilterPlugins()
	if err != nil {
		return nil, xerrors.Errorf("get registered filter plugins: %w", err)
	}
	registeredscorepls, err := config.RegisteredScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get registered score plugins: %w", err)
	}

	return append(registeredscorepls, registeredfilterpls...), nil
}
