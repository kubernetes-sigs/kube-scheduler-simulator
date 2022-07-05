package config

import (
	"golang.org/x/xerrors"
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

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
