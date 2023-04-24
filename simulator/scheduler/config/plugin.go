package config

import (
	"golang.org/x/xerrors"
	configv1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

var (
	outOfTreeRegistries = runtime.Registry{
		// TODO(user): add your plugins registries here.
	}

	registeredOutOfTreeMultiPointName = []string{}
)

// RegisteredMultiPointPluginNames returns all registered multipoint plugin names.
// in-tree plugins and your original plugins listed in outOfTreeRegistries above.
func RegisteredMultiPointPluginNames() ([]string, error) {
	def, err := InTreeMultiPointPluginSet()
	if err != nil {
		return nil, xerrors.Errorf("get default multi point plugins: %w", err)
	}

	enabledPls := make([]string, 0, len(def.Enabled))
	for _, e := range def.Enabled {
		enabledPls = append(enabledPls, e.Name)
	}

	return append(enabledPls, OutOfTreeMultiPointPluginNames()...), nil
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

func OutOfTreeMultiPointPluginNames() []string {
	return registeredOutOfTreeMultiPointName
}

func InTreeRegistries() runtime.Registry {
	return plugins.NewInTreeRegistry()
}

func OutOfTreeRegistries() runtime.Registry {
	return outOfTreeRegistries
}

func SetOutOfTreeRegistries(r runtime.Registry) {
	for k, v := range r {
		outOfTreeRegistries[k] = v
		registeredOutOfTreeMultiPointName = append(registeredOutOfTreeMultiPointName, k)
	}
}
