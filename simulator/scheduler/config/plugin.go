package config

import (
	"golang.org/x/xerrors"
	configv1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

// RegisteredMultiPointPlugins returns all registered multipoint plugins.
// in-tree plugins and your original plugins listed in OutOfTreeMultiPointPlugins.
func RegisteredMultiPointPlugins() ([]configv1.Plugin, error) {
	def, err := InTreeMultiPointPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default multi point plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeMultiPointPlugins()...), nil
}

// InTreeMultiPointPluginSet returns default multipoint plugins.
// See also: https://github.com/kubernetes/kubernetes/blob/475f9010f5faa7bdd439944a6f5f1ec206297602/pkg/scheduler/apis/config/v1/default_plugins.go#L30https://github.com/kubernetes/kubernetes/blob/475f9010f5faa7bdd439944a6f5f1ec206297602/pkg/scheduler/apis/config/v1/default_plugins.go#L30
func InTreeMultiPointPluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.MultiPoint, nil
}

func OutOfTreeMultiPointPlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPreScorePlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreePreScorePlugins.
func RegisteredPreScorePlugins() ([]configv1.Plugin, error) {
	def, err := InTreePreScorePluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePreScorePlugins()...), nil
}

func InTreePreScorePluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PreScore, nil
}

func OutOfTreePreScorePlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPermitPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreePermitPlugins.
func RegisteredPermitPlugins() ([]configv1.Plugin, error) {
	def, err := InTreePermitPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePermitPlugins()...), nil
}

func InTreePermitPluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Permit, nil
}

func OutOfTreePermitPlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredReservePlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeReservePlugins.
func RegisteredReservePlugins() ([]configv1.Plugin, error) {
	def, err := InTreeReservePluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeReservePlugins()...), nil
}

func InTreeReservePluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Reserve, nil
}

func OutOfTreeReservePlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPreBindPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreePreBindPlugins.
func RegisteredPreBindPlugins() ([]configv1.Plugin, error) {
	def, err := InTreePreBindPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePreBindPlugins()...), nil
}

func InTreePreBindPluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PreBind, nil
}

func OutOfTreePreBindPlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPostBindPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreePostBindPlugins.
func RegisteredPostBindPlugins() ([]configv1.Plugin, error) {
	def, err := InTreePostBindPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePostBindPlugins()...), nil
}

func InTreePostBindPluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PostBind, nil
}

func OutOfTreePostBindPlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredBindPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeBindPlugins.
func RegisteredBindPlugins() ([]configv1.Plugin, error) {
	def, err := InTreeBindPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeBindPlugins()...), nil
}

func InTreeBindPluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Bind, nil
}

func OutOfTreeBindPlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredPreFilterPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeFilterPlugins.
func RegisteredPreFilterPlugins() ([]configv1.Plugin, error) {
	def, err := InTreePreFilterPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default prefilter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreePreFilterPlugins()...), nil
}

func InTreePreFilterPluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PreFilter, nil
}

func OutOfTreePreFilterPlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

// RegisteredFilterPlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeFilterPlugins.
func RegisteredFilterPlugins() ([]configv1.Plugin, error) {
	def, err := InTreeFilterPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default filter plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeFilterPlugins()...), nil
}

func InTreeFilterPluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Filter, nil
}

func OutOfTreeFilterPlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your filter plugins here.
	}
}

func RegisteredPostFilterPlugins() ([]configv1.Plugin, error) {
	def, err := InTreePostFilterPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default post filter plugins: %w", err)
	}
	return append(def.Enabled, OutOfTreePostFilterPlugins()...), nil
}

func InTreePostFilterPluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.PostFilter, nil
}

func OutOfTreePostFilterPlugins() []configv1.Plugin {
	return []configv1.Plugin{
		// Note: add your post filter plugins here.
	}
}

// RegisteredScorePlugins returns all registered plugins.
// in-tree plugins and your original plugins listed in OutOfTreeScorePlugins.
func RegisteredScorePlugins() ([]configv1.Plugin, error) {
	def, err := InTreeScorePluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default score plugins: %w", err)
	}

	return append(def.Enabled, OutOfTreeScorePlugins()...), nil
}

func InTreeScorePluginSet() (configv1.PluginSet, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return configv1.PluginSet{}, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Score, nil
}

func OutOfTreeScorePlugins() []configv1.Plugin {
	return []configv1.Plugin{
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
