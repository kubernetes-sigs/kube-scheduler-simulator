package plugin

import (
	"encoding/json"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	schedulerRuntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/defaultconfig"
	schedulingresultstore "github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/plugin/resultstore"
)

func NewRegistry(informerFactory informers.SharedInformerFactory, client clientset.Interface) (map[string]schedulerRuntime.PluginFactory, error) {
	defaultScorePluginWeight := map[string]int32{}
	defaultScorePlugin, err := defaultconfig.DefaultScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default score plugins: %w", err)
	}
	for _, p := range defaultScorePlugin {
		defaultScorePluginWeight[p.Name] = 0
		if p.Weight != nil {
			defaultScorePluginWeight[p.Name] = *p.Weight
		}
	}

	defaultpls, err := defaultFilterScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default score/filter plugins: %w", err)
	}

	store := schedulingresultstore.New(informerFactory, client, defaultScorePluginWeight)
	rs := plugins.NewInTreeRegistry()
	ret := map[string]schedulerRuntime.PluginFactory{}
	for _, pl := range defaultpls {
		pl := pl
		r := rs[pl.Name]
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

			return NewWrappedPlugin(store, p, weight), nil
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
	defaultcfg, err := defaultconfig.DefaultSchedulerConfig()
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

	defaultpls, err := defaultFilterScorePlugins()
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
//nolint:cyclop
func ConvertForSimulator(pls *v1beta2.Plugins) (*v1beta2.Plugins, error) {
	newpls := pls.DeepCopy()
	// true means the plugin is disabled
	disabledMapForScore := map[string]bool{}
	for _, p := range pls.Score.Disabled {
		disabledMapForScore[p.Name] = true
	}
	if !disabledMapForScore["*"] {
		// user wants not to disable all plugin.
		defaultscorepls, err := defaultconfig.DefaultScorePlugins()
		if err != nil {
			return nil, xerrors.Errorf("get default score plugins: %w", err)
		}
		var retscorepls []v1beta2.Plugin
		for _, dp := range defaultscorepls {
			if !disabledMapForScore[dp.Name] {
				retscorepls = append(retscorepls, v1beta2.Plugin{Name: pluginName(dp.Name), Weight: dp.Weight})
			}
		}
		newpls.Score.Enabled = retscorepls
	}

	// disable default plugins whatever scheduler configuration value is
	newpls.Score.Disabled = []v1beta2.Plugin{
		{
			Name: "*",
		},
	}

	disabledMapForFilter := map[string]bool{}
	for _, p := range pls.Filter.Disabled {
		disabledMapForFilter[p.Name] = true
	}
	if !disabledMapForFilter["*"] {
		// user wants not to disable all plugin.
		defaultfilterpls, err := defaultconfig.DefaultFilterPlugins()
		if err != nil {
			return nil, xerrors.Errorf("get default filter plugins: %w", err)
		}
		var retfilterpls []v1beta2.Plugin
		for _, dp := range defaultfilterpls {
			if !disabledMapForFilter[dp.Name] {
				retfilterpls = append(retfilterpls, v1beta2.Plugin{Name: pluginName(dp.Name), Weight: dp.Weight})
			}
		}
		newpls.Filter.Enabled = retfilterpls
	}

	// disable default plugins whatever scheduler configuration value is
	newpls.Filter.Disabled = []v1beta2.Plugin{
		{
			Name: "*",
		},
	}

	return newpls, nil
}

// defaultFilterScorePlugins are score plugin and/or filter plugin.
func defaultFilterScorePlugins() ([]v1beta2.Plugin, error) {
	defaultfilterpls, err := defaultconfig.DefaultFilterPlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}
	defaultscorepls, err := defaultconfig.DefaultScorePlugins()
	if err != nil {
		return nil, xerrors.Errorf("get default score plugins: %w", err)
	}

	// defaultpls type must be score plugin and/or filter plugin
	defaultpls := append(defaultscorepls, defaultfilterpls...)

	return defaultpls, nil
}
