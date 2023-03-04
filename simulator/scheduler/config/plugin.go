package config

import (
	"golang.org/x/xerrors"
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

// RegisteredPreScorePlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreePreScorePlugins.
func RegisteredPreScorePlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreePreScorePluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePreScorePlugins()...), nil
}

func InTreePreScorePluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PreScore, nil
}

func OutOfTreePreScorePlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPermitPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreePermitPlugins.
func RegisteredPermitPlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreePermitPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePermitPlugins()...), nil
}

func InTreePermitPluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Permit, nil
}

func OutOfTreePermitPlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredReservePlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeReservePlugins.
func RegisteredReservePlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreeReservePluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeReservePlugins()...), nil
}

func InTreeReservePluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Reserve, nil
}

func OutOfTreeReservePlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPreBindPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreePreBindPlugins.
func RegisteredPreBindPlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreePreBindPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePreBindPlugins()...), nil
}

func InTreePreBindPluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PreBind, nil
}

func OutOfTreePreBindPlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPostBindPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreePostBindPlugins.
func RegisteredPostBindPlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreePostBindPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePostBindPlugins()...), nil
}

func InTreePostBindPluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PostBind, nil
}

func OutOfTreePostBindPlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredBindPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeBindPlugins.
func RegisteredBindPlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreeBindPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeBindPlugins()...), nil
}

func InTreeBindPluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Bind, nil
}

func OutOfTreeBindPlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPreFilterPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeFilterPlugins.
func RegisteredPreFilterPlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreePreFilterPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default prefilter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePreFilterPlugins()...), nil
}

func InTreePreFilterPluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PreFilter, nil
}

func OutOfTreePreFilterPlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredFilterPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeFilterPlugins.
func RegisteredFilterPlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreeFilterPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeFilterPlugins()...), nil
}

func InTreeFilterPluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Filter, nil
}

func OutOfTreeFilterPlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your filter plugins here.
	}
}

func RegisteredPostFilterPlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreePostFilterPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default post filter plugins: %w", err)
	}
	return append(def.Enabled, OutOfTreePostFilterPlugins()...), nil
}

func InTreePostFilterPluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PostFilter, nil
}

func OutOfTreePostFilterPlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your post filter plugins here.
	}
}

// RegisteredScorePlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeScorePlugins.
func RegisteredScorePlugins() ([]v1beta2.Plugin, error) {
	def, err := InTreeScorePluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default score plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeScorePlugins()...), nil
}

func InTreeScorePluginSet() (v1beta2.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return v1beta2.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Score, nil
}

func OutOfTreeScorePlugins() []v1beta2.Plugin {
	return []v1beta2.Plugin{
		// Note: add your score plugins here.
	}
}

func InTreeRegistries() runtime.Registry {
	return plugins.NewInTreeRegistry()
}

func OutOfTreeRegistries() runtime.Registry {
	return runtime.Registry{
		// Note: add your plugins registries here.
	}
}
