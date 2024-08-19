package plugin

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	configv1 "k8s.io/kube-scheduler/config/v1"
	schedulerConfig "k8s.io/kubernetes/pkg/scheduler/apis/config"
)

func TestConvertForSimulator(t *testing.T) {
	t.Parallel()
	var weight1 int32 = 1
	var weight2 int32 = 2
	var weight3 int32 = 3

	tests := []struct {
		name    string
		arg     *configv1.Plugins
		want    *configv1.Plugins
		wantErr bool
	}{
		{
			name: "success",
			arg: &configv1.Plugins{
				PreFilter: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreScore: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Reserve: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Permit: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreBind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Bind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostBind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Filter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "NodeUnschedulable"},
						{Name: "NodeName"},
					},
				},
				PostFilter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "DefaultPreemption"},
					},
				},
				Score: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				MultiPoint: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "EBSLimits",
						},
					},
				},
			},
			want: &configv1.Plugins{
				PreFilter: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreScore: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Reserve: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Permit: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreBind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Bind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostBind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Filter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "NodeUnschedulableWrapped"},
						{Name: "NodeNameWrapped"},
					},
					Disabled: []configv1.Plugin{},
				},
				PostFilter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "DefaultPreemptionWrapped"},
					},
					Disabled: []configv1.Plugin{},
				},
				Score: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				MultiPoint: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "SchedulingGatesWrapped"},
						{Name: "PrioritySortWrapped"},
						{Name: "NodeUnschedulableWrapped"},
						{Name: "NodeNameWrapped"},
						{Name: "TaintTolerationWrapped", Weight: &weight3},
						{Name: "NodeAffinityWrapped", Weight: &weight2},
						{Name: "NodePortsWrapped"},
						{Name: "NodeResourcesFitWrapped", Weight: &weight1},
						{Name: "VolumeRestrictionsWrapped"},
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
					},
					Disabled: []configv1.Plugin{
						{Name: "*"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success when user disable all plugins with '*'",
			arg: &configv1.Plugins{
				PreFilter: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreScore: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Reserve: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Permit: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreBind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Bind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostBind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Filter: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{Name: "*"},
					},
				},
				PostFilter: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{Name: "*"},
					},
				},
				Score: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{Name: "*"},
					},
				},
				MultiPoint: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{Name: "*"},
					},
				},
			},
			want: &configv1.Plugins{
				PreFilter: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreScore: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Reserve: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Permit: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreBind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Bind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostBind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Filter: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostFilter: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Score: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				MultiPoint: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success with non in-tree plugins",
			arg: &configv1.Plugins{
				PreFilter: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreScore: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Reserve: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Permit: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreBind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Bind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostBind: configv1.PluginSet{
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Filter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "CustomPlugin1"},
					},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostFilter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "CustomPlugin1"},
					},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Score: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "CustomPlugin1"},
					},
					Disabled: []configv1.Plugin{
						{Name: "*"},
					},
				},
				MultiPoint: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{Name: "*"},
					},
				},
			},
			want: &configv1.Plugins{
				PreFilter: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreScore: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Reserve: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Permit: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PreBind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Bind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostBind: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Filter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "CustomPlugin1Wrapped"},
					},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				PostFilter: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "CustomPlugin1Wrapped"},
					},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				Score: configv1.PluginSet{
					Enabled: []configv1.Plugin{
						{Name: "CustomPlugin1Wrapped"},
					},
					Disabled: []configv1.Plugin{
						{
							Name: "*",
						},
					},
				},
				MultiPoint: configv1.PluginSet{
					Enabled: []configv1.Plugin{},
					Disabled: []configv1.Plugin{
						{Name: "*"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertForSimulator(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertWrapped() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

//nolint:gocognit // it is because of huge test cases.
func Test_NewPluginConfig(t *testing.T) {
	t.Parallel()

	var minCandidateNodesPercentage int32 = 20
	var minCandidateNodesAbsolute int32 = 100
	var hardPodAffinityWeight int32 = 2

	tests := []struct {
		name    string
		pc      []configv1.PluginConfig
		want    []configv1.PluginConfig
		wantErr bool
	}{
		{
			name:    "success with empty arg",
			pc:      nil,
			want:    defaultPluginConfig(),
			wantErr: false,
		},
		{
			name: "success with plugin config of postFilter",
			pc: []configv1.PluginConfig{
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
			want: func() []configv1.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "DefaultPreemption" {
						pc[i] = configv1.PluginConfig{
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
						}
					}
					if pc[i].Name == "DefaultPreemptionWrapped" {
						pc[i] = configv1.PluginConfig{
							Name: "DefaultPreemptionWrapped",
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
						}
					}
				}

				return pc
			}(),
			wantErr: false,
		},
		{
			name: "success with plugin config on Args.Object",
			pc: []configv1.PluginConfig{
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
			want: func() []configv1.PluginConfig {
				pc := defaultPluginConfig()
				var defaultMinCandidateNodesPercentage int32 = 10
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = configv1.PluginConfig{
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
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = configv1.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &configv1.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
					if pc[i].Name == "DefaultPreemption" {
						pc[i] = configv1.PluginConfig{
							Name: "DefaultPreemption",
							Args: runtime.RawExtension{
								Object: &configv1.DefaultPreemptionArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "DefaultPreemptionArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1",
									},
									MinCandidateNodesPercentage: &defaultMinCandidateNodesPercentage,
									MinCandidateNodesAbsolute:   &minCandidateNodesAbsolute,
								},
							},
						}
					}
					if pc[i].Name == "DefaultPreemptionWrapped" {
						pc[i] = configv1.PluginConfig{
							Name: "DefaultPreemptionWrapped",
							Args: runtime.RawExtension{
								Object: &configv1.DefaultPreemptionArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "DefaultPreemptionArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1",
									},
									MinCandidateNodesPercentage: &defaultMinCandidateNodesPercentage,
									MinCandidateNodesAbsolute:   &minCandidateNodesAbsolute,
								},
							},
						}
					}
				}

				return pc
			}(),
			wantErr: false,
		},
		{
			name: "Success: if data exists in both PluginConfig.Args.Raw and PluginConfig.Args.Object," +
				"PluginConfig.Args.Raw would be ignored",
			pc: []configv1.PluginConfig{
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
						Raw: func() []byte {
							anotherHardPodAffinityWeight := hardPodAffinityWeight + 1
							cfg := configv1.InterPodAffinityArgs{
								TypeMeta: metav1.TypeMeta{
									Kind:       "InterPodAffinityArgs",
									APIVersion: "kubescheduler.config.k8s.io/v1",
								},
								HardPodAffinityWeight: &anotherHardPodAffinityWeight,
							}
							d, _ := json.Marshal(cfg)
							return d
						}(),
					},
				},
			},
			want: func() []configv1.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = configv1.PluginConfig{
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
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = configv1.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &configv1.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
				}

				return pc
			}(),
			wantErr: false,
		},
		{
			name: "success with plugin config on Args.Raw ",
			pc: []configv1.PluginConfig{
				{
					Name: "InterPodAffinity",
					Args: runtime.RawExtension{
						Raw: func() []byte {
							cfg := configv1.InterPodAffinityArgs{
								TypeMeta: metav1.TypeMeta{
									Kind:       "InterPodAffinityArgs",
									APIVersion: "kubescheduler.config.k8s.io/v1",
								},
								HardPodAffinityWeight: &hardPodAffinityWeight,
							}
							d, _ := json.Marshal(cfg)
							return d
						}(),
					},
				},
			},
			want: func() []configv1.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = configv1.PluginConfig{
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
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = configv1.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &configv1.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
				}

				return pc
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewPluginConfig(tt.pc)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPluginConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.SliceStable(got, func(i, j int) bool { return got[i].Name < got[j].Name })
			sort.SliceStable(tt.want, func(i, j int) bool { return tt.want[i].Name < tt.want[j].Name })

			assert.Equal(t, tt.want, got)
		})
	}
}

func defaultPluginConfig() []configv1.PluginConfig {
	var minCandidateNodesPercentage int32 = 10
	var minCandidateNodesAbsolute int32 = 100
	var hardPodAffinityWeight int32 = 1
	var bindTimeoutSeconds int64 = 600

	return []configv1.PluginConfig{
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
		{
			Name: "NodeAffinity",
			Args: runtime.RawExtension{
				Object: &configv1.NodeAffinityArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeAffinityArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
				},
			},
		},
		{
			Name: "NodeResourcesBalancedAllocation",
			Args: runtime.RawExtension{
				Object: &configv1.NodeResourcesBalancedAllocationArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesBalancedAllocationArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
					Resources: []configv1.ResourceSpec{
						{
							Name:   "cpu",
							Weight: 1,
						},
						{
							Name:   "memory",
							Weight: 1,
						},
					},
				},
			},
		},
		{
			Name: "NodeResourcesFit",
			Args: runtime.RawExtension{
				Object: &configv1.NodeResourcesFitArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesFitArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
					ScoringStrategy: &configv1.ScoringStrategy{
						Type: "LeastAllocated",
						Resources: []configv1.ResourceSpec{
							{
								Name:   "cpu",
								Weight: 1,
							},
							{
								Name:   "memory",
								Weight: 1,
							},
						},
					},
				},
			},
		},
		{
			Name: "PodTopologySpread",
			Args: runtime.RawExtension{
				Object: &configv1.PodTopologySpreadArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PodTopologySpreadArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
					DefaultingType: "System",
				},
			},
		},
		{
			Name: "VolumeBinding",
			Args: runtime.RawExtension{
				Object: &configv1.VolumeBindingArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "VolumeBindingArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
					BindTimeoutSeconds: &bindTimeoutSeconds,
				},
			},
		},
		{
			Name: "DefaultPreemptionWrapped",
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
		{
			Name: "NodeResourcesBalancedAllocationWrapped",
			Args: runtime.RawExtension{
				Object: &configv1.NodeResourcesBalancedAllocationArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesBalancedAllocationArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
					Resources: []configv1.ResourceSpec{
						{
							Name:   "cpu",
							Weight: 1,
						},
						{
							Name:   "memory",
							Weight: 1,
						},
					},
				},
			},
		},
		{
			Name: "InterPodAffinityWrapped",
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
		{
			Name: "NodeResourcesFitWrapped",
			Args: runtime.RawExtension{
				Object: &configv1.NodeResourcesFitArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesFitArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
					ScoringStrategy: &configv1.ScoringStrategy{
						Type: "LeastAllocated",
						Resources: []configv1.ResourceSpec{
							{
								Name:   "cpu",
								Weight: 1,
							},
							{
								Name:   "memory",
								Weight: 1,
							},
						},
					},
				},
			},
		},
		{
			Name: "NodeAffinityWrapped",
			Args: runtime.RawExtension{
				Object: &configv1.NodeAffinityArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeAffinityArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
				},
			},
		},
		{
			Name: "PodTopologySpreadWrapped",
			Args: runtime.RawExtension{
				Object: &configv1.PodTopologySpreadArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PodTopologySpreadArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
					DefaultingType: "System",
				},
			},
		},
		{
			Name: "VolumeBindingWrapped",
			Args: runtime.RawExtension{
				Object: &configv1.VolumeBindingArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "VolumeBindingArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1",
					},
					BindTimeoutSeconds: &bindTimeoutSeconds,
				},
			},
		},
	}
}

func TestGetScorePluginWeight(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  *schedulerConfig.KubeSchedulerConfiguration
		want map[string]int32
	}{
		{
			name: "score and multipoint plugins",
			cfg: &schedulerConfig.KubeSchedulerConfiguration{
				Profiles: []schedulerConfig.KubeSchedulerProfile{
					{
						Plugins: &schedulerConfig.Plugins{
							Score: schedulerConfig.PluginSet{
								Enabled: []schedulerConfig.Plugin{
									{
										Name:   "score1",
										Weight: 1,
									},
									{
										Name:   "score2",
										Weight: 2,
									},
								},
							},
							MultiPoint: schedulerConfig.PluginSet{
								Enabled: []schedulerConfig.Plugin{
									{
										Name:   "multipoint1",
										Weight: 1,
									},
									{
										Name:   "multipoint2",
										Weight: 2,
									},
								},
							},
						},
					},
				},
			},
			want: map[string]int32{
				"multipoint1": 1,
				"multipoint2": 2,
				"score1":      1,
				"score2":      2,
			},
		},
		{
			name: "only score plugins",
			cfg: &schedulerConfig.KubeSchedulerConfiguration{
				Profiles: []schedulerConfig.KubeSchedulerProfile{
					{
						Plugins: &schedulerConfig.Plugins{
							Score: schedulerConfig.PluginSet{
								Enabled: []schedulerConfig.Plugin{
									{
										Name:   "score1",
										Weight: 1,
									},
									{
										Name:   "score2",
										Weight: 2,
									},
								},
							},
						},
					},
				},
			},
			want: map[string]int32{
				"score1": 1,
				"score2": 2,
			},
		},
		{
			name: "only multipoint plugins",
			cfg: &schedulerConfig.KubeSchedulerConfiguration{
				Profiles: []schedulerConfig.KubeSchedulerProfile{
					{
						Plugins: &schedulerConfig.Plugins{
							MultiPoint: schedulerConfig.PluginSet{
								Enabled: []schedulerConfig.Plugin{
									{
										Name:   "multipoint1",
										Weight: 1,
									},
									{
										Name:   "multipoint2",
										Weight: 2,
									},
								},
							},
						},
					},
				},
			},
			want: map[string]int32{
				"multipoint1": 1,
				"multipoint2": 2,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := getScorePluginWeight(tt.cfg)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("unexpected plugins map: (-want, +got):\n%s", diff)
			}
		})
	}
}
