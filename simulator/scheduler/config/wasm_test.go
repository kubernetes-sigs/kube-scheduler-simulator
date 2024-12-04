package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
)

func TestGetWasmRegistryFromUnversionedConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      *config.KubeSchedulerConfiguration
		expected int
	}{
		{
			name:     "no profiles",
			cfg:      &config.KubeSchedulerConfiguration{},
			expected: 0,
		},
		{
			name: "no wasm plugins",
			cfg: &config.KubeSchedulerConfiguration{
				Profiles: []config.KubeSchedulerProfile{
					{
						PluginConfig: []config.PluginConfig{
							{
								Name: "DefaultPreemption",
								Args: &config.DefaultPreemptionArgs{},
							},
						},
						Plugins: &config.Plugins{
							MultiPoint: config.PluginSet{
								Enabled: []config.Plugin{
									{Name: "DefaultPreemption"},
								},
							},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "one wasm plugin",
			cfg: &config.KubeSchedulerConfiguration{
				Profiles: []config.KubeSchedulerProfile{
					{
						PluginConfig: []config.PluginConfig{
							{
								Name: "DefaultPreemption",
								Args: &config.DefaultPreemptionArgs{},
							},
							{Name: "wasmPlugin", Args: &runtime.Unknown{
								ContentType: runtime.ContentTypeJSON,
								Raw:         []byte(`{"guestURL":"http://example.com/plugin.wasm"}`),
							}},
						},
						Plugins: &config.Plugins{
							MultiPoint: config.PluginSet{
								Enabled: []config.Plugin{
									{Name: "DefaultPreemption"},
									{Name: "wasmPlugin"},
								},
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple wasm plugins",
			cfg: &config.KubeSchedulerConfiguration{
				Profiles: []config.KubeSchedulerProfile{
					{
						PluginConfig: []config.PluginConfig{
							{
								Name: "DefaultPreemption",
								Args: &config.DefaultPreemptionArgs{},
							},
							{Name: "wasmPlugin1", Args: &runtime.Unknown{
								ContentType: runtime.ContentTypeJSON,
								Raw:         []byte(`{"guestURL":"http://example.com/plugin1.wasm"}`),
							}},
							{Name: "wasmPlugin2", Args: &runtime.Unknown{
								ContentType: runtime.ContentTypeJSON,
								Raw:         []byte(`{"guestURL":"http://example.com/plugin2.wasm"}`),
							}},
						},
						Plugins: &config.Plugins{
							MultiPoint: config.PluginSet{
								Enabled: []config.Plugin{
									{Name: "DefaultPreemption"},
									{Name: "wasmPlugin1"},
									{Name: "wasmPlugin2"},
								},
							},
						},
					},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			registry, err := getWasmRegistryFromUnversionedConfig(tt.cfg)
			require.NoError(t, err, "check error")
			if len(registry) != tt.expected {
				t.Errorf("expected %d plugins, got %d", tt.expected, len(registry))
			}
		})
	}
}
