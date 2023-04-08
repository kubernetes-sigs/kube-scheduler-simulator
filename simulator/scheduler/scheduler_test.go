package scheduler

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	configv1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"

	schedConfig "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
)

var (
	weight1 int32 = 1
	weight2 int32 = 2
	weight3 int32 = 3
)

//nolint:gocognit // For test case.
func Test_convertConfigurationForSimulator(t *testing.T) {
	t.Parallel()

	var nondefaultParallelism int32 = 3
	defaultschedulername := v1.DefaultSchedulerName
	nondefaultschedulername := v1.DefaultSchedulerName + "2"

	var minCandidateNodesPercentage int32 = 20
	var minCandidateNodesAbsolute int32 = 100
	var hardPodAffinityWeight int32 = 2

	type args struct {
		versioned *configv1.KubeSchedulerConfiguration
		port      int
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
				versioned: &configv1.KubeSchedulerConfiguration{},
				port:      80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with no-disabled plugin",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &configv1.Plugins{},
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with empty Profiles",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{},
				port:      80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "changes of field other than Profiles and Extenders does not affects result",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &configv1.Plugins{},
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "changes of field other than Profiles.Plugins and Extenders does not affects result",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &configv1.Plugins{},
							PluginConfig:  nil,
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with multiple profiles/applied disabled setting",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
						},
						{
							SchedulerName: &nondefaultschedulername,
							Plugins: &configv1.Plugins{
								MultiPoint: configv1.PluginSet{
									Disabled: []configv1.Plugin{
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
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				profile2 := cfg.Profiles[0].DeepCopy()
				profile2.SchedulerName = nondefaultschedulername
				profile2.Plugins.MultiPoint.Enabled = []config.Plugin{
					{Name: "PrioritySortWrapped"},
					{Name: "NodeUnschedulableWrapped"},
					{Name: "NodeNameWrapped"},
					{Name: "TaintTolerationWrapped", Weight: weight3},
					{Name: "NodeAffinityWrapped", Weight: weight2},
					{Name: "NodePortsWrapped"},
					{Name: "VolumeRestrictionsWrapped"},
					{Name: "EBSLimitsWrapped"},
					{Name: "GCEPDLimitsWrapped"},
					{Name: "NodeVolumeLimitsWrapped"},
					{Name: "AzureDiskLimitsWrapped"},
					{Name: "VolumeBindingWrapped"},
					{Name: "VolumeZoneWrapped"},
					{Name: "PodTopologySpreadWrapped", Weight: weight2},
					{Name: "InterPodAffinityWrapped", Weight: weight2},
					{Name: "DefaultPreemptionWrapped"},
					{Name: "NodeResourcesBalancedAllocationWrapped", Weight: weight1},
					{Name: "DefaultBinderWrapped"},
				}
				cfg.Profiles = append(cfg.Profiles, *profile2)
				return &cfg
			}(),
		},
		{
			name: "success with Extender",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &configv1.Plugins{},
						},
					},
					Extenders: []configv1.Extender{
						{
							URLPrefix:      "http://example.com/extender/",
							PreemptVerb:    "PreemptVerb/",
							FilterVerb:     "FilterVerb/",
							PrioritizeVerb: "PrioritizeVerb/",
							BindVerb:       "BindVerb/",
							Weight:         1,
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				cfg.Extenders = []config.Extender{
					{
						URLPrefix:      "http://localhost:80/api/v1/extender/",
						PreemptVerb:    "preempt/0",
						FilterVerb:     "filter/0",
						PrioritizeVerb: "prioritize/0",
						BindVerb:       "bind/0",
						Weight:         1,
					},
				}
				return &cfg
			}(),
		},
		{
			name: "success with multiple profiles and custom-pluginconfig",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							PluginConfig: []configv1.PluginConfig{
								{
									Name: "DefaultPreemption",
									Args: runtime.RawExtension{
										Object: &configv1.DefaultPreemptionArgs{
											TypeMeta: metav1.TypeMeta{
												Kind:       "DefaultPreemptionArgs",
												APIVersion: "kubescheduler.config.k8s.io/v1",
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
							PluginConfig: []configv1.PluginConfig{
								{
									Name: "InterPodAffinity",
									Args: runtime.RawExtension{
										Object: &configv1.InterPodAffinityArgs{
											TypeMeta: metav1.TypeMeta{
												Kind:       "InterPodAffinityArgs",
												APIVersion: "kubescheduler.config.k8s.io/v1",
											},
											HardPodAffinityWeight: &hardPodAffinityWeight,
										},
									},
								},
							},
						},
					},
				},
				port: 80,
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
					if cfg.Profiles[0].PluginConfig[i].Name == "DefaultPreemptionWrapped" {
						cfg.Profiles[0].PluginConfig[i] = config.PluginConfig{
							Name: "DefaultPreemptionWrapped",
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
			name: "success with multiplugin plugin setting/manual setting weights have priority.",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins: &configv1.Plugins{
								MultiPoint: configv1.PluginSet{
									Enabled: []configv1.Plugin{
										{
											Name:   "NodeResourcesFit",
											Weight: &weight3,
										},
									},
								},
							},
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				cfg.Profiles[0].Plugins.MultiPoint.Disabled = []config.Plugin{
					{Name: "*"},
				}
				cfg.Profiles[0].Plugins.MultiPoint.Enabled = []config.Plugin{
					{Name: "PrioritySortWrapped"},
					{Name: "NodeUnschedulableWrapped"},
					{Name: "NodeNameWrapped"},
					{Name: "TaintTolerationWrapped", Weight: weight3},
					{Name: "NodeAffinityWrapped", Weight: weight2},
					{Name: "NodePortsWrapped"},
					{Name: "NodeResourcesFitWrapped", Weight: weight3},
					{Name: "VolumeRestrictionsWrapped"},
					{Name: "EBSLimitsWrapped"},
					{Name: "GCEPDLimitsWrapped"},
					{Name: "NodeVolumeLimitsWrapped"},
					{Name: "AzureDiskLimitsWrapped"},
					{Name: "VolumeBindingWrapped"},
					{Name: "VolumeZoneWrapped"},
					{Name: "PodTopologySpreadWrapped", Weight: weight2},
					{Name: "InterPodAffinityWrapped", Weight: weight2},
					{Name: "DefaultPreemptionWrapped"},
					{Name: "NodeResourcesBalancedAllocationWrapped", Weight: weight1},
					{Name: "ImageLocalityWrapped", Weight: weight1},
					{Name: "DefaultBinderWrapped"},
				}
				return &cfg
			}(),
		},
		{
			name: "success with multiplugin plugin setting/multi manual setting weights have priority.",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins: &configv1.Plugins{
								MultiPoint: configv1.PluginSet{
									Enabled: []configv1.Plugin{
										{
											Name:   "NodeResourcesFit",
											Weight: &weight2,
										},
									},
								},
								Score: configv1.PluginSet{
									Enabled: []configv1.Plugin{
										{
											Name:   "NodeResourcesFit",
											Weight: &weight3,
										},
									},
								},
							},
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				cfg.Profiles[0].Plugins.Score.Enabled = []config.Plugin{
					{
						Name:   "NodeResourcesFitWrapped",
						Weight: weight3,
					},
				}
				cfg.Profiles[0].Plugins.MultiPoint.Disabled = []config.Plugin{
					{Name: "*"},
				}
				cfg.Profiles[0].Plugins.MultiPoint.Enabled = []config.Plugin{
					{Name: "PrioritySortWrapped"},
					{Name: "NodeUnschedulableWrapped"},
					{Name: "NodeNameWrapped"},
					{Name: "TaintTolerationWrapped", Weight: weight3},
					{Name: "NodeAffinityWrapped", Weight: weight2},
					{Name: "NodePortsWrapped"},
					{Name: "NodeResourcesFitWrapped", Weight: weight2},
					{Name: "VolumeRestrictionsWrapped"},
					{Name: "EBSLimitsWrapped"},
					{Name: "GCEPDLimitsWrapped"},
					{Name: "NodeVolumeLimitsWrapped"},
					{Name: "AzureDiskLimitsWrapped"},
					{Name: "VolumeBindingWrapped"},
					{Name: "VolumeZoneWrapped"},
					{Name: "PodTopologySpreadWrapped", Weight: weight2},
					{Name: "InterPodAffinityWrapped", Weight: weight2},
					{Name: "DefaultPreemptionWrapped"},
					{Name: "NodeResourcesBalancedAllocationWrapped", Weight: weight1},
					{Name: "ImageLocalityWrapped", Weight: weight1},
					{Name: "DefaultBinderWrapped"},
				}
				return &cfg
			}(),
		},
		{
			name: "success with multiplugin plugin setting/disable a specific default multipoint plugin on a extension point",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins: &configv1.Plugins{
								Score: configv1.PluginSet{
									Disabled: []configv1.Plugin{
										{
											Name: "NodeResourcesFit",
										},
									},
								},
							},
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				cfg.Profiles[0].Plugins.Score.Disabled = []config.Plugin{
					{
						Name: "NodeResourcesFitWrapped",
					},
				}
				cfg.Profiles[0].Plugins.MultiPoint.Disabled = []config.Plugin{
					{Name: "*"},
				}
				return &cfg
			}(),
		},
		{
			name: "success with multiplugin plugin setting/disable a specific default multipoint plugin",
			args: args{
				versioned: &configv1.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []configv1.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins: &configv1.Plugins{
								MultiPoint: configv1.PluginSet{
									Disabled: []configv1.Plugin{
										{
											Name: "NodeResourcesFit",
										},
									},
								},
							},
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				cfg.Profiles[0].Plugins.MultiPoint.Disabled = []config.Plugin{
					{Name: "*"},
				}
				cfg.Profiles[0].Plugins.MultiPoint.Enabled = []config.Plugin{
					{Name: "PrioritySortWrapped"},
					{Name: "NodeUnschedulableWrapped"},
					{Name: "NodeNameWrapped"},
					{Name: "TaintTolerationWrapped", Weight: weight3},
					{Name: "NodeAffinityWrapped", Weight: weight2},
					{Name: "NodePortsWrapped"},
					{Name: "VolumeRestrictionsWrapped"},
					{Name: "EBSLimitsWrapped"},
					{Name: "GCEPDLimitsWrapped"},
					{Name: "NodeVolumeLimitsWrapped"},
					{Name: "AzureDiskLimitsWrapped"},
					{Name: "VolumeBindingWrapped"},
					{Name: "VolumeZoneWrapped"},
					{Name: "PodTopologySpreadWrapped", Weight: weight2},
					{Name: "InterPodAffinityWrapped", Weight: weight2},
					{Name: "DefaultPreemptionWrapped"},
					{Name: "NodeResourcesBalancedAllocationWrapped", Weight: weight1},
					{Name: "ImageLocalityWrapped", Weight: weight1},
					{Name: "DefaultBinderWrapped"},
				}
				return &cfg
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := convertConfigurationForSimulator(tt.args.versioned, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertConfigurationForSimulator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Profiles) != len(tt.want.Profiles) {
				t.Errorf("unmatch length of profiles, want: %v, got: %v", len(tt.want.Profiles), len(got.Profiles))
				return
			}
			if len(got.Extenders) != len(tt.want.Extenders) {
				t.Errorf("unmatch length of extenders, want: %v, got: %v", len(tt.want.Extenders), len(got.Extenders))
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

			for i := range tt.want.Profiles {
				assert.Equal(t, tt.want.Profiles[i].Plugins, got.Profiles[i].Plugins)
				assert.Equal(t, tt.want.Profiles[i].PluginConfig, got.Profiles[i].PluginConfig)
			}
			assert.Equal(t, tt.want.Extenders, got.Extenders)
		})
	}
}

func configGeneratedFromDefault() config.KubeSchedulerConfiguration {
	versioned, _ := schedConfig.DefaultSchedulerConfig()
	cfg := versioned.DeepCopy()

	cfg.Profiles[0].Plugins.MultiPoint.Enabled = []configv1.Plugin{
		{Name: "PrioritySortWrapped"},
		{Name: "NodeUnschedulableWrapped"},
		{Name: "NodeNameWrapped"},
		{Name: "TaintTolerationWrapped", Weight: &weight3},
		{Name: "NodeAffinityWrapped", Weight: &weight2},
		{Name: "NodePortsWrapped"},
		{Name: "NodeResourcesFitWrapped", Weight: &weight1},
		{Name: "VolumeRestrictionsWrapped"},
		{Name: "EBSLimitsWrapped"},
		{Name: "GCEPDLimitsWrapped"},
		{Name: "NodeVolumeLimitsWrapped"},
		{Name: "AzureDiskLimitsWrapped"},
		{Name: "VolumeBindingWrapped"},
		{Name: "VolumeZoneWrapped"},
		{Name: "PodTopologySpreadWrapped", Weight: &weight2},
		{Name: "InterPodAffinityWrapped", Weight: &weight2},
		{Name: "DefaultPreemptionWrapped"},
		{Name: "NodeResourcesBalancedAllocationWrapped", Weight: &weight1},
		{Name: "ImageLocalityWrapped", Weight: &weight1},
		{Name: "DefaultBinderWrapped"},
	}
	cfg.Profiles[0].Plugins.MultiPoint.Disabled = []configv1.Plugin{
		{Name: "*"},
	}

	pcMap := map[string]runtime.RawExtension{}
	for _, c := range cfg.Profiles[0].PluginConfig {
		pcMap[c.Name] = c.Args
	}

	var newpc []configv1.PluginConfig
	newpc = append(newpc, configv1.PluginConfig{
		Name: "NodeResourcesBalancedAllocationWrapped",
		Args: pcMap["NodeResourcesBalancedAllocation"],
	})
	newpc = append(newpc, configv1.PluginConfig{
		Name: "InterPodAffinityWrapped",
		Args: pcMap["InterPodAffinity"],
	})
	newpc = append(newpc, configv1.PluginConfig{
		Name: "NodeResourcesFitWrapped",
		Args: pcMap["NodeResourcesFit"],
	})
	newpc = append(newpc, configv1.PluginConfig{
		Name: "NodeAffinityWrapped",
		Args: pcMap["NodeAffinity"],
	})
	newpc = append(newpc, configv1.PluginConfig{
		Name: "PodTopologySpreadWrapped",
		Args: pcMap["PodTopologySpread"],
	})
	newpc = append(newpc, configv1.PluginConfig{
		Name: "VolumeBindingWrapped",
		Args: pcMap["VolumeBinding"],
	})
	newpc = append(newpc, configv1.PluginConfig{
		Name: "DefaultPreemptionWrapped",
		Args: pcMap["DefaultPreemption"],
	})

	cfg.Profiles[0].PluginConfig = append(cfg.Profiles[0].PluginConfig, newpc...)

	converted := config.KubeSchedulerConfiguration{}
	scheme.Scheme.Convert(cfg, &converted, nil)
	converted.SetGroupVersionKind(configv1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))
	return converted
}
