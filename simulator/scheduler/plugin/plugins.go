package plugin

import (
	"context"
	"encoding/json"
	"strings"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	configv1 "k8s.io/kube-scheduler/config/v1"
	schedulerConfig "k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
	schedulingresultstore "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin/resultstore"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/storereflector"
)

// ResultStoreKey represents key name of plugins results on sharedstore.
const ResultStoreKey = "PluginResultStoreKey"

func NewRegistry(sharedStore storereflector.Reflector, cfg *schedulerConfig.KubeSchedulerConfiguration, pluginExtenders map[string]PluginExtenderInitializer) (map[string]schedulerRuntime.PluginFactory, error) {
	scorePluginWeight := getScorePluginWeight(cfg)
	store := schedulingresultstore.New(scorePluginWeight)
	// Add the resultStore to the sharedStore to store the results and share it.
	sharedStore.AddResultStore(store, ResultStoreKey)

	ret, err := newPluginFactories(store, pluginExtenders)
	if err != nil {
		return nil, xerrors.Errorf("New pluginFactories: %w", err)
	}

	return ret, nil
}

func newPluginFactories(store *schedulingresultstore.Store, pluginExtenders map[string]PluginExtenderInitializer) (map[string]schedulerRuntime.PluginFactory, error) {
	intreeRegistries := config.InTreeRegistries()
	outoftreeRegistries := config.OutOfTreeRegistries()
	pls, err := config.RegisteredMultiPointPluginNames()
	if err != nil {
		return nil, xerrors.Errorf("failed to get registered plugin names: %w", err)
	}
	ret := map[string]schedulerRuntime.PluginFactory{}
	for _, pluginname := range pls {
		pluginname := pluginname

		r, ok := intreeRegistries[pluginname]
		if !ok {
			// not found in intreeRegistries. search registry in outoftreeRegistries.
			r, ok = outoftreeRegistries[pluginname]
			if !ok {
				return nil, xerrors.Errorf("registry for %s is not found", pluginname)
			}
			// For out-of-tree plugins, we need to add original registry to registries.
			// (For in-tree plugins, schedulers add original registry to registries internally.)
			ret[pluginname] = r
		}

		if _, ok := ret[pluginName(pluginname)]; ok {
			// already created
			continue
		}

		factory := func(ctx context.Context, configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
			p, err := r(ctx, configuration, f)
			if err != nil {
				return nil, xerrors.Errorf("create original plugin: %w", err)
			}

			opts := []Option{}
			extender, ok := pluginExtenders[pluginname]
			if ok {
				opts = append(opts, WithExtendersOption(extender))
			}

			return NewWrappedPlugin(store, p, opts...), nil
		}
		ret[pluginName(pluginname)] = factory
	}

	return ret, nil
}

// NewPluginConfig converts []configv1.PluginConfig for simulator.
// Passed []v1beta.PluginConfig overrides default config values.
//
// NewPluginConfig expects that either PluginConfig.Args.Raw or PluginConfig.Args.Object has data
// in the passed configv1.PluginConfig parameter.
// If data exists in both PluginConfig.Args.Raw and PluginConfig.Args.Object, PluginConfig.Args.Raw would be ignored
// since PluginConfig.Args.Object has higher priority.
//
//nolint:funlen,cyclop
func NewPluginConfig(pc []configv1.PluginConfig) ([]configv1.PluginConfig, error) {
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

		// configv1.PluginConfig may have data in pc[i].Args.Raw as []byte.
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

	ret := make([]configv1.PluginConfig, 0, len(pluginConfig))
	for name, arg := range pluginConfig {
		// add plugin configs for default plugins.
		ret = append(ret, configv1.PluginConfig{
			Name: name,
			Args: *arg,
		})
	}

	defaultpls, err := config.RegisteredMultiPointPluginNames()
	if err != nil {
		return nil, xerrors.Errorf("get default score/filter plugins: %w", err)
	}

	for _, name := range defaultpls {
		pc, ok := pluginConfig[name]
		if !ok {
			continue
		}

		ret = append(ret, configv1.PluginConfig{
			Name: pluginName(name),
			Args: *pc,
		})

		// avoid adding same plugin config.
		delete(pluginConfig, name)
	}

	return ret, nil
}

// ConvertForSimulator convert configv1.Plugins for simulator.
func ConvertForSimulator(pls *configv1.Plugins) (*configv1.Plugins, error) {
	newpls := pls.DeepCopy()

	applyPluginSet(&newpls.PreFilter, pls.PreFilter, configv1.PluginSet{})
	applyPluginSet(&newpls.Filter, pls.Filter, configv1.PluginSet{})
	applyPluginSet(&newpls.PostFilter, pls.PostFilter, configv1.PluginSet{})
	applyPluginSet(&newpls.PreScore, pls.PreScore, configv1.PluginSet{})
	applyPluginSet(&newpls.Score, pls.Score, configv1.PluginSet{})
	applyPluginSet(&newpls.Reserve, pls.Reserve, configv1.PluginSet{})
	applyPluginSet(&newpls.Permit, pls.Permit, configv1.PluginSet{})
	applyPluginSet(&newpls.PreBind, pls.PreBind, configv1.PluginSet{})
	applyPluginSet(&newpls.Bind, pls.Bind, configv1.PluginSet{})
	applyPluginSet(&newpls.PostBind, pls.PostBind, configv1.PluginSet{})
	inTreeMultiPointPls, err := config.InTreeMultiPointPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get in tree multi point plugins: %w", err)
	}
	applyPluginSet(&newpls.MultiPoint, pls.MultiPoint, inTreeMultiPointPls)
	// The default MultiPoint PluginSets should be disable to "*" here
	// so that the scheduler won't enable all default plugins.
	disableAllPluginSet(&newpls.MultiPoint)

	return newpls, nil
}

// disableAllPluginSet set target PluginSet to "*".
func disableAllPluginSet(targetPlsSet *configv1.PluginSet) {
	targetPlsSet.Disabled = []configv1.Plugin{
		{
			Name: "*",
		},
	}
}

// applyPluginSet merges inTree and outOfTree PluginSet.
func applyPluginSet(targetPlsSet *configv1.PluginSet, plsSet configv1.PluginSet, inTreePls configv1.PluginSet) {
	merged := mergePluginSet(inTreePls, plsSet)
	enabledPls := make([]configv1.Plugin, 0, len(merged.Enabled))
	for _, p := range merged.Enabled {
		enabledPls = append(enabledPls, configv1.Plugin{Name: pluginName(p.Name), Weight: p.Weight})
	}
	targetPlsSet.Enabled = enabledPls

	disabledPls := make([]configv1.Plugin, 0, len(merged.Disabled))
	for _, p := range merged.Disabled {
		wName := pluginName(p.Name)
		if p.Name == "*" {
			wName = p.Name
		}
		disabledPls = append(disabledPls, configv1.Plugin{Name: wName, Weight: p.Weight})
	}
	targetPlsSet.Disabled = disabledPls
}

// mergePluginsSet merges two plugin sets.
// This function is copied from https://github.com/kubernetes/kubernetes/blob/release-1.27/pkg/scheduler/apis/config/v1/default_plugins.go.
func mergePluginSet(defaultPluginSet, customPluginSet configv1.PluginSet) configv1.PluginSet {
	type pluginIndex struct {
		index  int
		plugin configv1.Plugin
	}

	disabledPlugins := sets.New[string]()
	enabledCustomPlugins := make(map[string]pluginIndex)
	// replacedPluginIndex is a set of index of plugins, which have replaced the default plugins.
	replacedPluginIndex := sets.NewInt()
	disabled := make([]configv1.Plugin, 0, len(customPluginSet.Disabled))
	for _, disabledPlugin := range customPluginSet.Disabled {
		// if the user is manually disabling any (or all, with "*") default plugins for an extension point,
		// we need to track that so that the MultiPoint extension logic in the framework can know to skip
		// inserting unspecified default plugins to this point.
		disabled = append(disabled, configv1.Plugin{Name: disabledPlugin.Name})
		disabledPlugins.Insert(disabledPlugin.Name)
	}

	// With MultiPoint, we may now have some disabledPlugins in the default registry
	// For example, we enable PluginX with Filter+Score through MultiPoint but disable its Score plugin by default.
	for _, disabledPlugin := range defaultPluginSet.Disabled {
		disabled = append(disabled, configv1.Plugin{Name: disabledPlugin.Name})
		disabledPlugins.Insert(disabledPlugin.Name)
	}

	for index, enabledPlugin := range customPluginSet.Enabled {
		enabledCustomPlugins[enabledPlugin.Name] = pluginIndex{index, enabledPlugin}
	}
	var enabledPlugins []configv1.Plugin
	if !disabledPlugins.Has("*") {
		for _, defaultEnabledPlugin := range defaultPluginSet.Enabled {
			if disabledPlugins.Has(defaultEnabledPlugin.Name) {
				continue
			}
			// The default plugin is explicitly re-configured, update the default plugin accordingly.
			if customPlugin, ok := enabledCustomPlugins[defaultEnabledPlugin.Name]; ok {
				klog.Info("Default plugin is explicitly re-configured; overriding", "plugin", defaultEnabledPlugin.Name)
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
	for index, plugin := range customPluginSet.Enabled {
		if !replacedPluginIndex.Has(index) {
			enabledPlugins = append(enabledPlugins, plugin)
		}
	}
	return configv1.PluginSet{Enabled: enabledPlugins, Disabled: disabled}
}

// getScorePluginWeight get weights of enabled score plugins in the scheduler configuration.
// It only supports the scheduler configuration with one profile -- scheduler with multiple profiles isn't supported.
func getScorePluginWeight(cfg *schedulerConfig.KubeSchedulerConfiguration) map[string]int32 {
	scorePluginWeight := make(map[string]int32)
	enabledScorePlugins := cfg.Profiles[0].Plugins.Score.Enabled
	enabledScorePlugins = append(enabledScorePlugins, cfg.Profiles[0].Plugins.MultiPoint.Enabled...)
	for _, p := range enabledScorePlugins {
		if p.Weight != 0 {
			scorePluginWeight[strings.TrimSuffix(p.Name, pluginSuffix)] = p.Weight
		} else {
			// a weight of zero is not permitted, plugins can be disabled explicitly
			// when configured.
			scorePluginWeight[strings.TrimSuffix(p.Name, pluginSuffix)] = 1
		}
	}

	return scorePluginWeight
}
