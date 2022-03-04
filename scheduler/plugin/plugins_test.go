package plugin

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kube-scheduler/config/v1beta2"
)

func TestConvertForSimulator(t *testing.T) {
	t.Parallel()
	var weight1 int32 = 1
	var weight2 int32 = 2

	tests := []struct {
		name    string
		arg     *v1beta2.Plugins
		want    *v1beta2.Plugins
		wantErr bool
	}{
		{
			name: "success",
			arg: &v1beta2.Plugins{
				Filter: v1beta2.PluginSet{
					Disabled: []v1beta2.Plugin{
						{Name: "EBSLimits"},
						{Name: "NodeUnschedulable"},
						{Name: "NodeName"},
						{Name: "TaintToleration"},
						{Name: "NodeAffinity"},
						{Name: "GCEPDLimits"},
						{Name: "NodeVolumeLimits"},
						{Name: "AzureDiskLimits"},
						{Name: "VolumeBinding"},
						{Name: "VolumeZone"},
						{Name: "NodePorts"},
						{Name: "NodeResourcesFit"},
						{Name: "VolumeRestrictions"},
					},
				},
				Score: v1beta2.PluginSet{
					Disabled: []v1beta2.Plugin{
						{Name: "NodeResourcesFit"},
						{Name: "NodeResourcesBalancedAllocation"},
						{Name: "ImageLocality"},
						{Name: "InterPodAffinity"},
						{Name: "NodeAffinity"},
					},
				},
			},
			want: &v1beta2.Plugins{
				Filter: v1beta2.PluginSet{
					Enabled: []v1beta2.Plugin{
						{Name: "PodTopologySpreadWrapped"},
						{Name: "InterPodAffinityWrapped"},
					},
					Disabled: []v1beta2.Plugin{
						{
							Name: "*",
						},
					},
				},
				Score: v1beta2.PluginSet{
					Enabled: []v1beta2.Plugin{
						{Name: "PodTopologySpreadWrapped", Weight: &weight2},
						{Name: "TaintTolerationWrapped", Weight: &weight1},
					},
					Disabled: []v1beta2.Plugin{
						{
							Name: "*",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success when user disable all plugins with '*'",
			arg: &v1beta2.Plugins{
				Filter: v1beta2.PluginSet{
					Disabled: []v1beta2.Plugin{
						{Name: "*"},
					},
				},
				Score: v1beta2.PluginSet{
					Disabled: []v1beta2.Plugin{
						{Name: "NodeResourcesFit"},
						{Name: "NodeResourcesBalancedAllocation"},
						{Name: "ImageLocality"},
						{Name: "InterPodAffinity"},
						{Name: "NodeAffinity"},
					},
				},
			},
			want: &v1beta2.Plugins{
				Filter: v1beta2.PluginSet{
					Disabled: []v1beta2.Plugin{
						{
							Name: "*",
						},
					},
				},
				Score: v1beta2.PluginSet{
					Enabled: []v1beta2.Plugin{
						{Name: "PodTopologySpreadWrapped", Weight: &weight2},
						{Name: "TaintTolerationWrapped", Weight: &weight1},
					},
					Disabled: []v1beta2.Plugin{
						{
							Name: "*",
						},
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
		pc      []v1beta2.PluginConfig
		want    []v1beta2.PluginConfig
		wantErr bool
	}{
		{
			name:    "success with empty arg",
			pc:      nil,
			want:    defaultPluginConfig(),
			wantErr: false,
		},
		{
			name: "success with plugin config of not filter/score",
			pc: []v1beta2.PluginConfig{
				{
					Name: "DefaultPreemption",
					Args: runtime.RawExtension{
						Object: &v1beta2.DefaultPreemptionArgs{
							TypeMeta: metav1.TypeMeta{
								Kind:       "DefaultPreemptionArgs",
								APIVersion: "kubescheduler.config.k8s.io/v1beta2",
							},
							MinCandidateNodesPercentage: &minCandidateNodesPercentage,
							MinCandidateNodesAbsolute:   &minCandidateNodesAbsolute,
						},
					},
				},
			},
			want: func() []v1beta2.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name != "DefaultPreemption" {
						continue
					}

					pc[i] = v1beta2.PluginConfig{
						Name: "DefaultPreemption",
						Args: runtime.RawExtension{
							Object: &v1beta2.DefaultPreemptionArgs{
								TypeMeta: metav1.TypeMeta{
									Kind:       "DefaultPreemptionArgs",
									APIVersion: "kubescheduler.config.k8s.io/v1beta2",
								},
								MinCandidateNodesPercentage: &minCandidateNodesPercentage,
								MinCandidateNodesAbsolute:   &minCandidateNodesAbsolute,
							},
						},
					}
				}

				return pc
			}(),
			wantErr: false,
		},
		{
			name: "success with plugin config on Args.Object",
			pc: []v1beta2.PluginConfig{
				{
					Name: "InterPodAffinity",
					Args: runtime.RawExtension{
						Object: &v1beta2.InterPodAffinityArgs{
							TypeMeta: metav1.TypeMeta{
								Kind:       "InterPodAffinityArgs",
								APIVersion: "kubescheduler.config.k8s.io/v1beta2",
							},
							HardPodAffinityWeight: &hardPodAffinityWeight,
						},
					},
				},
			},
			want: func() []v1beta2.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = v1beta2.PluginConfig{
							Name: "InterPodAffinity",
							Args: runtime.RawExtension{
								Object: &v1beta2.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta2",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = v1beta2.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &v1beta2.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta2",
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
			name: "Success: if data exists in both PluginConfig.Args.Raw and PluginConfig.Args.Object," +
				"PluginConfig.Args.Raw would be ignored",
			pc: []v1beta2.PluginConfig{
				{
					Name: "InterPodAffinity",
					Args: runtime.RawExtension{
						Object: &v1beta2.InterPodAffinityArgs{
							TypeMeta: metav1.TypeMeta{
								Kind:       "InterPodAffinityArgs",
								APIVersion: "kubescheduler.config.k8s.io/v1beta2",
							},
							HardPodAffinityWeight: &hardPodAffinityWeight,
						},
						Raw: func() []byte {
							anotherHardPodAffinityWeight := hardPodAffinityWeight + 1
							cfg := v1beta2.InterPodAffinityArgs{
								TypeMeta: metav1.TypeMeta{
									Kind:       "InterPodAffinityArgs",
									APIVersion: "kubescheduler.config.k8s.io/v1beta2",
								},
								HardPodAffinityWeight: &anotherHardPodAffinityWeight,
							}
							d, _ := json.Marshal(cfg)
							return d
						}(),
					},
				},
			},
			want: func() []v1beta2.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = v1beta2.PluginConfig{
							Name: "InterPodAffinity",
							Args: runtime.RawExtension{
								Object: &v1beta2.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta2",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = v1beta2.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &v1beta2.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta2",
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
			pc: []v1beta2.PluginConfig{
				{
					Name: "InterPodAffinity",
					Args: runtime.RawExtension{
						Raw: func() []byte {
							cfg := v1beta2.InterPodAffinityArgs{
								TypeMeta: metav1.TypeMeta{
									Kind:       "InterPodAffinityArgs",
									APIVersion: "kubescheduler.config.k8s.io/v1beta2",
								},
								HardPodAffinityWeight: &hardPodAffinityWeight,
							}
							d, _ := json.Marshal(cfg)
							return d
						}(),
					},
				},
			},
			want: func() []v1beta2.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = v1beta2.PluginConfig{
							Name: "InterPodAffinity",
							Args: runtime.RawExtension{
								Object: &v1beta2.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta2",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = v1beta2.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &v1beta2.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta2",
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

func Test_defaultFilterScorePlugins(t *testing.T) {
	t.Parallel()
	var weight1 int32 = 1
	var weight2 int32 = 2
	tests := []struct {
		name    string
		want    []v1beta2.Plugin
		wantErr bool
	}{
		{
			name: "success",
			want: []v1beta2.Plugin{
				{Name: "NodeResourcesBalancedAllocation", Weight: &weight1},
				{Name: "ImageLocality", Weight: &weight1},
				{Name: "InterPodAffinity", Weight: &weight1},
				{Name: "NodeResourcesFit", Weight: &weight1},
				{Name: "NodeAffinity", Weight: &weight1},
				{Name: "PodTopologySpread", Weight: &weight2},
				{Name: "TaintToleration", Weight: &weight1},
				{Name: "NodeUnschedulable"},
				{Name: "NodeName"},
				{Name: "TaintToleration"},
				{Name: "NodeAffinity"},
				{Name: "NodePorts"},
				{Name: "NodeResourcesFit"},
				{Name: "VolumeRestrictions"},
				{Name: "EBSLimits"},
				{Name: "GCEPDLimits"},
				{Name: "NodeVolumeLimits"},
				{Name: "AzureDiskLimits"},
				{Name: "VolumeBinding"},
				{Name: "VolumeZone"},
				{Name: "PodTopologySpread"},
				{Name: "InterPodAffinity"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := defaultFilterScorePlugins()
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultFilterScorePlugins() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func defaultPluginConfig() []v1beta2.PluginConfig {
	var minCandidateNodesPercentage int32 = 10
	var minCandidateNodesAbsolute int32 = 100
	var hardPodAffinityWeight int32 = 1
	var bindTimeoutSeconds int64 = 600

	return []v1beta2.PluginConfig{
		{
			Name: "DefaultPreemption",
			Args: runtime.RawExtension{
				Object: &v1beta2.DefaultPreemptionArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "DefaultPreemptionArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					MinCandidateNodesPercentage: &minCandidateNodesPercentage,
					MinCandidateNodesAbsolute:   &minCandidateNodesAbsolute,
				},
			},
		},
		{
			Name: "InterPodAffinity",
			Args: runtime.RawExtension{
				Object: &v1beta2.InterPodAffinityArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "InterPodAffinityArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					HardPodAffinityWeight: &hardPodAffinityWeight,
				},
			},
		},
		{
			Name: "NodeAffinity",
			Args: runtime.RawExtension{
				Object: &v1beta2.NodeAffinityArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeAffinityArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
				},
			},
		},
		{
			Name: "NodeResourcesBalancedAllocation",
			Args: runtime.RawExtension{
				Object: &v1beta2.NodeResourcesBalancedAllocationArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesBalancedAllocationArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					Resources: []v1beta2.ResourceSpec{
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
				Object: &v1beta2.NodeResourcesFitArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesFitArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					ScoringStrategy: &v1beta2.ScoringStrategy{
						Type: "LeastAllocated",
						Resources: []v1beta2.ResourceSpec{
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
				Object: &v1beta2.PodTopologySpreadArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PodTopologySpreadArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					DefaultingType: "System",
				},
			},
		},
		{
			Name: "VolumeBinding",
			Args: runtime.RawExtension{
				Object: &v1beta2.VolumeBindingArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "VolumeBindingArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					BindTimeoutSeconds: &bindTimeoutSeconds,
				},
			},
		},
		{
			Name: "NodeResourcesBalancedAllocationWrapped",
			Args: runtime.RawExtension{
				Object: &v1beta2.NodeResourcesBalancedAllocationArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesBalancedAllocationArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					Resources: []v1beta2.ResourceSpec{
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
				Object: &v1beta2.InterPodAffinityArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "InterPodAffinityArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					HardPodAffinityWeight: &hardPodAffinityWeight,
				},
			},
		},
		{
			Name: "NodeResourcesFitWrapped",
			Args: runtime.RawExtension{
				Object: &v1beta2.NodeResourcesFitArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesFitArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					ScoringStrategy: &v1beta2.ScoringStrategy{
						Type: "LeastAllocated",
						Resources: []v1beta2.ResourceSpec{
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
				Object: &v1beta2.NodeAffinityArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeAffinityArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
				},
			},
		},
		{
			Name: "PodTopologySpreadWrapped",
			Args: runtime.RawExtension{
				Object: &v1beta2.PodTopologySpreadArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PodTopologySpreadArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					DefaultingType: "System",
				},
			},
		},
		{
			Name: "VolumeBindingWrapped",
			Args: runtime.RawExtension{
				Object: &v1beta2.VolumeBindingArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "VolumeBindingArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta2",
					},
					BindTimeoutSeconds: &bindTimeoutSeconds,
				},
			},
		},
	}
}
