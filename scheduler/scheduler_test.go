package scheduler

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1beta3config "k8s.io/kube-scheduler/config/v1beta3"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/defaultconfig"
)

func Test_convertConfigurationForSimulator(t *testing.T) {
	t.Parallel()

	var nondefaultParallelism int32 = 3
	defaultschedulername := v1.DefaultSchedulerName
	nondefaultschedulername := v1.DefaultSchedulerName + "2"

	var minCandidateNodesPercentage int32 = 20
	var minCandidateNodesAbsolute int32 = 100
	var hardPodAffinityWeight int32 = 2

	type args struct {
		versioned *v1beta3config.KubeSchedulerConfiguration
	}
	tests := []struct {
		name    string
		args    args
		want    *config.KubeSchedulerConfiguration
		wantErr bool
	}{
		{
			name: "success with empty-configuration",
			args: args{
				versioned: &v1beta3config.KubeSchedulerConfiguration{},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with no-disabled plugin",
			args: args{
				versioned: &v1beta3config.KubeSchedulerConfiguration{
					Profiles: []v1beta3config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta3config.Plugins{},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with empty Profiles",
			args: args{
				versioned: &v1beta3config.KubeSchedulerConfiguration{},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "changes of field other than Profiles does not affects result",
			args: args{
				versioned: &v1beta3config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta3config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta3config.Plugins{},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "changes of field other than Profiles.Plugins does not affects result",
			args: args{
				versioned: &v1beta3config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta3config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta3config.Plugins{},
							PluginConfig:  nil,
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with multiple profiles",
			args: args{
				versioned: &v1beta3config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta3config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
						},
						{
							SchedulerName: &nondefaultschedulername,
							Plugins: &v1beta3config.Plugins{
								Score: v1beta3config.PluginSet{
									Disabled: []v1beta3config.Plugin{
										{
											Name: "ImageLocality",
										},
										{
											Name: "NodeResourcesFit",
										},
									},
								},
							},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				profile2 := cfg.Profiles[0].DeepCopy()
				profile2.SchedulerName = nondefaultschedulername
				profile2.Plugins.Score.Enabled = nil
				cfg.Profiles = append(cfg.Profiles, *profile2)
				return &cfg
			}(),
		},
		{
			name: "success with multiple profiles and custom-pluginconfig",
			args: args{
				versioned: &v1beta3config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta3config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							PluginConfig: []v1beta3config.PluginConfig{
								{
									Name: "DefaultPreemption",
									Args: runtime.RawExtension{
										Object: &v1beta3config.DefaultPreemptionArgs{
											TypeMeta: metav1.TypeMeta{
												Kind:       "DefaultPreemptionArgs",
												APIVersion: "kubescheduler.config.k8s.io/v1beta3",
											},
											MinCandidateNodesPercentage: &minCandidateNodesPercentage,
											MinCandidateNodesAbsolute:   &minCandidateNodesAbsolute,
										},
									},
								},
							},
						},
						{
							SchedulerName: &nondefaultschedulername,
							PluginConfig: []v1beta3config.PluginConfig{
								{
									Name: "InterPodAffinity",
									Args: runtime.RawExtension{
										Object: &v1beta3config.InterPodAffinityArgs{
											TypeMeta: metav1.TypeMeta{
												Kind:       "InterPodAffinityArgs",
												APIVersion: "kubescheduler.config.k8s.io/v1beta3",
											},
											HardPodAffinityWeight: &hardPodAffinityWeight,
										},
									},
								},
							},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				profile2 := cfg.Profiles[0].DeepCopy()
				profile2.SchedulerName = nondefaultschedulername
				for i := range cfg.Profiles[0].PluginConfig {
					if cfg.Profiles[0].PluginConfig[i].Name == "DefaultPreemption" {
						cfg.Profiles[0].PluginConfig[i] = config.PluginConfig{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{
								MinCandidateNodesPercentage: minCandidateNodesPercentage,
								MinCandidateNodesAbsolute:   minCandidateNodesAbsolute,
							},
						}
					}
				}

				for i := range profile2.PluginConfig {
					if profile2.PluginConfig[i].Name == "InterPodAffinity" {
						profile2.PluginConfig[i] = config.PluginConfig{
							Name: "InterPodAffinity",
							Args: &config.InterPodAffinityArgs{
								HardPodAffinityWeight: hardPodAffinityWeight,
							},
						}
					}
					if profile2.PluginConfig[i].Name == "InterPodAffinityWrapped" {
						profile2.PluginConfig[i] = config.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: &config.InterPodAffinityArgs{
								HardPodAffinityWeight: hardPodAffinityWeight,
							},
						}
					}
				}

				cfg.Profiles = append(cfg.Profiles, *profile2)
				return &cfg
			}(),
		},
		{
			name: "success with some plugin disabled",
			args: args{
				versioned: &v1beta3config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta3config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins: &v1beta3config.Plugins{
								Score: v1beta3config.PluginSet{
									Disabled: []v1beta3config.Plugin{
										{
											Name: "ImageLocality",
										},
										{
											Name: "NodeResourcesFit",
										},
									},
								},
							},
						},
					},
				},
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				cfg.Profiles[0].Plugins.Score.Disabled = []config.Plugin{
					{
						Name:   "*",
						Weight: 0,
					},
				}
				cfg.Profiles[0].Plugins.Score.Enabled = nil
				return &cfg
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := convertConfigurationForSimulator(tt.args.versioned)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertConfigurationForSimulator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
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

			assert.Equal(t, tt.want, got)
		})
	}
}

func configGeneratedFromDefault() config.KubeSchedulerConfiguration {
	var weight0 int32
	versioned, _ := defaultconfig.DefaultSchedulerConfig()
	cfg := versioned.DeepCopy()
	cfg.Profiles[0].Plugins.Filter.Enabled = nil
	cfg.Profiles[0].Plugins.Filter.Disabled = []v1beta3config.Plugin{
		{
			Name:   "*",
			Weight: &weight0,
		},
	}
	cfg.Profiles[0].Plugins.Score.Enabled = nil
	cfg.Profiles[0].Plugins.Score.Disabled = []v1beta3config.Plugin{
		{
			Name:   "*",
			Weight: &weight0,
		},
	}
	pcMap := map[string]runtime.RawExtension{}
	for _, c := range cfg.Profiles[0].PluginConfig {
		pcMap[c.Name] = c.Args
	}

	var newpc []v1beta3config.PluginConfig

	cfg.Profiles[0].PluginConfig = append(cfg.Profiles[0].PluginConfig, newpc...)

	converted := config.KubeSchedulerConfiguration{}
	scheme.Scheme.Convert(cfg, &converted, nil)
	converted.SetGroupVersionKind(v1beta3config.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))
	return converted
}
