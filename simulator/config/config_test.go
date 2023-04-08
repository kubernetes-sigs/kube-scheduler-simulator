package config

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	configv1 "k8s.io/kube-scheduler/config/v1"
)

func Test_decodeSchedulerCfg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		buf     []byte
		want    *configv1.KubeSchedulerConfiguration
		wantErr bool
	}{
		{
			name: "success with normal configuration",
			buf: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1
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
			want: &configv1.KubeSchedulerConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KubeSchedulerConfiguration",
					APIVersion: "kubescheduler.config.k8s.io/v1",
				},
				Profiles: []configv1.KubeSchedulerProfile{
					{
						PluginConfig: []configv1.PluginConfig{
							{
								Name: "NodeResourcesFit",
								Args: runtime.RawExtension{
									Object: &configv1.NodeResourcesFitArgs{
										ScoringStrategy: &configv1.ScoringStrategy{
											Resources: []configv1.ResourceSpec{
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
apiVersion: kubescheduler.config.k8s.io/v1beta1
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

func Test_parseStringListEnv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg  string
		want []string
	}{
		{
			name: "happy path: can parse the list which has multiple elements",
			arg:  "hoge,fuga,foo",
			want: []string{
				"hoge",
				"fuga",
				"foo",
			},
		},
		{
			name: "happy path: can parse the list which has the space between elements",
			arg:  "hoge,         fuga, foo    ",
			want: []string{
				"hoge",
				"fuga",
				"foo",
			},
		},
		{
			name: "happy path: do nothing with non-list string",
			arg:  "hoge",
			want: []string{
				"hoge",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, parseStringListEnv(tt.arg), "parseStringListEnv(%v)", tt.arg)
		})
	}
}

func Test_validateURLs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		urls    []string
		wantErr bool
	}{
		{
			name: "all urls are valid",
			urls: []string{
				"https://hoge.com/hoge",
				"http://hoge2.com/hoge",
			},
			wantErr: false,
		},
		{
			name: "one url is invalid",
			urls: []string{
				"https://hoge.com/hoge",
				"invalid",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateURLs(tt.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error result is returned. got: %v, wantErr: %v", err, tt.wantErr)
			}
		})
	}
}
