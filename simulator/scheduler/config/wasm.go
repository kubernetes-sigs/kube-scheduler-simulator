package config

import (
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/util/sets"
	configv1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	wasm "sigs.k8s.io/kube-scheduler-wasm-extension/scheduler/plugin"
)

// RegisterWasmPlugins registers wasm plugins from the given configuration.
func RegisterWasmPlugins(versionedCfg *configv1.KubeSchedulerConfiguration) error {
	cfg := config.KubeSchedulerConfiguration{}
	if err := scheme.Scheme.Convert(versionedCfg, &cfg, nil); err != nil {
		return xerrors.Errorf("convert configuration: %w", err)
	}

	registry, err := getWasmRegistryFromUnversionedConfig(&cfg)
	if err != nil {
		return err
	}

	SetOutOfTreeRegistries(registry)

	return nil
}

// getWasmRegistryFromUnversionedConfig registers wasm plugins from the given unversioned configuration.
func getWasmRegistryFromUnversionedConfig(cfg *config.KubeSchedulerConfiguration) (runtime.Registry, error) {
	registry := runtime.Registry{}

	for _, profile := range cfg.Profiles {
		wasmplugins := sets.New[string]()
		// look for the wasm plugin in the plugin config.
		for _, config := range profile.PluginConfig {
			if err := runtime.DecodeInto(config.Args, &wasm.PluginConfig{}); err != nil {
				// not wasm plugin.
				continue
			}

			wasmplugins.Insert(config.Name)
		}

		// look for the wasm plugin in the enabled plugins.
		// (assuming that the wasm plugin is specified as a multi-point plugin.)
		for _, plugin := range profile.Plugins.MultiPoint.Enabled {
			if wasmplugins.Has(plugin.Name) {
				if err := registry.Register(plugin.Name, wasm.PluginFactory(plugin.Name)); err != nil {
					return nil, xerrors.Errorf("register plugin %s: %w", plugin.Name, err)
				}
			}
		}
	}

	return registry, nil
}
