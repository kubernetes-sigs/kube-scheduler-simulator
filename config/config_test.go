package config

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1beta3config "k8s.io/kube-scheduler/config/v1beta3"
)

func Test_decodeSchedulerCfg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buf     []byte
		want    *v1beta3config.KubeSchedulerConfiguration
		wantErr bool
	}{
		{
			name: "success with normal configuration",
			buf: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
profiles:
- pluginConfig:
  - args:
      scoringStrategy:
        resources:
        - name: cpu
          weight: 1
        type: MostAllocated
    name: NodeResourcesFit
`),
			want: &v1beta3config.KubeSchedulerConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KubeSchedulerConfiguration",
					APIVersion: "kubescheduler.config.k8s.io/v1beta3",
				},
				Profiles: []v1beta3config.KubeSchedulerProfile{
					{
						PluginConfig: []v1beta3config.PluginConfig{
							{
								Name: "NodeResourcesFit",
								Args: runtime.RawExtension{
									Object: &v1beta3config.NodeResourcesFitArgs{
										ScoringStrategy: &v1beta3config.ScoringStrategy{
											Resources: []v1beta3config.ResourceSpec{
												{
													Name:   "cpu",
													Weight: 1,
												},
											},
											Type: "MostAllocated",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "fail because of wrong apiVersion",
			buf: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
profiles:
- pluginConfig:
  - args:
      scoringStrategy:
        resources:
        - name: cpu
          weight: 1
        type: MostAllocated
    name: NodeResourcesFit
`),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := decodeSchedulerCfg(tt.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeSchedulerCfg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				if len(got.Profiles) != len(tt.want.Profiles) {
					t.Errorf("unmatch length of profiles, want: %v, got: %v", len(tt.want.Profiles), len(got.Profiles))
					return
				}

				for k := range got.Profiles {
					sort.SliceStable(got.Profiles[k].PluginConfig, func(i, j int) bool {
						return got.Profiles[k].PluginConfig[i].Name < got.Profiles[k].PluginConfig[j].Name
					})
					sort.SliceStable(tt.want.Profiles[k].PluginConfig, func(i, j int) bool {
						return tt.want.Profiles[k].PluginConfig[i].Name < tt.want.Profiles[k].PluginConfig[j].Name
					})
				}
				// remove Raw to assert
				for i := range got.Profiles {
					prof := &got.Profiles[i]
					for j := range prof.PluginConfig {
						prof.PluginConfig[j].Args.Raw = nil
					}
				}
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
