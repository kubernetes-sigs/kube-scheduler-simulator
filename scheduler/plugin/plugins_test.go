package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	v1beta3config "k8s.io/kube-scheduler/config/v1beta3"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	mock_plugin "github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/plugin/mock"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/plugin/resultstore"
)

func TestConvertWrapped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		arg     *v1beta3config.Plugins
		want    *v1beta3config.Plugins
		wantErr bool
	}{
		{
			name: "success",
			arg: &v1beta3config.Plugins{
				Filter: v1beta3config.PluginSet{
					Disabled: []v1beta3config.Plugin{
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
				Score: v1beta3config.PluginSet{
					Disabled: []v1beta3config.Plugin{
						{Name: "NodeResourcesFit"},
						{Name: "NodeResourcesBalancedAllocation"},
						{Name: "ImageLocality"},
						{Name: "InterPodAffinity"},
						{Name: "NodeAffinity"},
					},
				},
			},
			want: &v1beta3config.Plugins{
				Filter: v1beta3config.PluginSet{
					Enabled: nil,
					Disabled: []v1beta3config.Plugin{
						{
							Name: "*",
						},
					},
				},
				Score: v1beta3config.PluginSet{
					Enabled: nil,
					Disabled: []v1beta3config.Plugin{
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
			arg: &v1beta3config.Plugins{
				Filter: v1beta3config.PluginSet{
					Disabled: []v1beta3config.Plugin{
						{Name: "*"},
					},
				},
				Score: v1beta3config.PluginSet{
					Disabled: []v1beta3config.Plugin{
						{Name: "NodeResourcesFit"},
						{Name: "NodeResourcesBalancedAllocation"},
						{Name: "ImageLocality"},
						{Name: "InterPodAffinity"},
						{Name: "NodeAffinity"},
					},
				},
			},
			want: &v1beta3config.Plugins{
				Filter: v1beta3config.PluginSet{
					Disabled: []v1beta3config.Plugin{
						{
							Name: "*",
						},
					},
				},
				Score: v1beta3config.PluginSet{
					Enabled: nil,
					Disabled: []v1beta3config.Plugin{
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
		pc      []v1beta3config.PluginConfig
		want    []v1beta3config.PluginConfig
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
			pc: []v1beta3config.PluginConfig{
				{
					Name: "DefaultPreemption",
					Args: runtime.RawExtension{
						Object: &v1beta3config.DefaultPreemptionArgs{
							TypeMeta: metav1.TypeMeta{
								Kind:       "DefaultPreemptionArgs",
								APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
							},
							MinCandidateNodesPercentage: &minCandidateNodesPercentage,
							MinCandidateNodesAbsolute:   &minCandidateNodesAbsolute,
						},
					},
				},
			},
			want: func() []v1beta3config.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name != "DefaultPreemption" {
						continue
					}

					pc[i] = v1beta3config.PluginConfig{
						Name: "DefaultPreemption",
						Args: runtime.RawExtension{
							Object: &v1beta3config.DefaultPreemptionArgs{
								TypeMeta: metav1.TypeMeta{
									Kind:       "DefaultPreemptionArgs",
									APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
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
			pc: []v1beta3config.PluginConfig{
				{
					Name: "InterPodAffinity",
					Args: runtime.RawExtension{
						Object: &v1beta3config.InterPodAffinityArgs{
							TypeMeta: metav1.TypeMeta{
								Kind:       "InterPodAffinityArgs",
								APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
							},
							HardPodAffinityWeight: &hardPodAffinityWeight,
						},
					},
				},
			},
			want: func() []v1beta3config.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = v1beta3config.PluginConfig{
							Name: "InterPodAffinity",
							Args: runtime.RawExtension{
								Object: &v1beta3config.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = v1beta3config.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &v1beta3config.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
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
			pc: []v1beta3config.PluginConfig{
				{
					Name: "InterPodAffinity",
					Args: runtime.RawExtension{
						Object: &v1beta3config.InterPodAffinityArgs{
							TypeMeta: metav1.TypeMeta{
								Kind:       "InterPodAffinityArgs",
								APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
							},
							HardPodAffinityWeight: &hardPodAffinityWeight,
						},
						Raw: func() []byte {
							anotherHardPodAffinityWeight := hardPodAffinityWeight + 1
							cfg := v1beta3config.InterPodAffinityArgs{
								TypeMeta: metav1.TypeMeta{
									Kind:       "InterPodAffinityArgs",
									APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
								},
								HardPodAffinityWeight: &anotherHardPodAffinityWeight,
							}
							d, _ := json.Marshal(cfg)
							return d
						}(),
					},
				},
			},
			want: func() []v1beta3config.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = v1beta3config.PluginConfig{
							Name: "InterPodAffinity",
							Args: runtime.RawExtension{
								Object: &v1beta3config.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = v1beta3config.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &v1beta3config.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
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
			pc: []v1beta3config.PluginConfig{
				{
					Name: "InterPodAffinity",
					Args: runtime.RawExtension{
						Raw: func() []byte {
							cfg := v1beta3config.InterPodAffinityArgs{
								TypeMeta: metav1.TypeMeta{
									Kind:       "InterPodAffinityArgs",
									APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
								},
								HardPodAffinityWeight: &hardPodAffinityWeight,
							}
							d, _ := json.Marshal(cfg)
							return d
						}(),
					},
				},
			},
			want: func() []v1beta3config.PluginConfig {
				pc := defaultPluginConfig()
				for i := range pc {
					if pc[i].Name == "InterPodAffinity" {
						pc[i] = v1beta3config.PluginConfig{
							Name: "InterPodAffinity",
							Args: runtime.RawExtension{
								Object: &v1beta3config.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
									},
									HardPodAffinityWeight: &hardPodAffinityWeight,
								},
							},
						}
					}
					if pc[i].Name == "InterPodAffinityWrapped" {
						pc[i] = v1beta3config.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: runtime.RawExtension{
								Object: &v1beta3config.InterPodAffinityArgs{
									TypeMeta: metav1.TypeMeta{
										Kind:       "InterPodAffinityArgs",
										APIVersion: "kubescheduler.config.k8s.io/v1beta3config",
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
	tests := []struct {
		name    string
		want    []v1beta3config.Plugin
		wantErr bool
	}{
		{
			name:    "success",
			want:    nil,
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

func Test_newWrappedPlugin(t *testing.T) {
	t.Parallel()
	fakeclientset := fake.NewSimpleClientset()
	store := resultstore.New(informers.NewSharedInformerFactory(fakeclientset, 0), nil, nil)

	type args struct {
		s      *resultstore.Store
		p      framework.Plugin
		weight int32
	}
	tests := []struct {
		name string
		args args
		want framework.Plugin
	}{
		{
			name: "success with filter plugin",
			args: args{
				s:      store,
				p:      fakeFilterPlugin{},
				weight: 0,
			},
			want: &wrappedPlugin{
				name:                 "fakeFilterPluginWrapped",
				originalFilterPlugin: fakeFilterPlugin{},
				originalScorePlugin:  nil,
				weight:               0,
				store:                store,
			},
		},
		{
			name: "success with score plugin",
			args: args{
				s:      store,
				p:      fakeScorePlugin{},
				weight: 1,
			},
			want: &wrappedPlugin{
				name:                 "fakeScorePluginWrapped",
				originalFilterPlugin: nil,
				originalScorePlugin:  fakeScorePlugin{},
				weight:               1,
				store:                store,
			},
		},
		{
			name: "success with score/filter plugin",
			args: args{
				s:      store,
				p:      fakeFilterScorePlugin{},
				weight: 1,
			},
			want: &wrappedPlugin{
				name:                 "fakeFilterScorePluginWrapped",
				originalFilterPlugin: fakeFilterScorePlugin{},
				originalScorePlugin:  fakeFilterScorePlugin{},
				weight:               1,
				store:                store,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := newWrappedPlugin(tt.args.s, tt.args.p, tt.args.weight)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_pluginName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		pluginName string
		want       string
	}{
		{
			name:       "success",
			pluginName: "pluginname",
			want:       "pluginnameWrapped",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := pluginName(tt.pluginName); got != tt.want {
				t.Errorf("pluginName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wrappedPlugin_Filter(t *testing.T) {
	t.Parallel()

	type args struct {
		pod      *v1.Pod
		nodeInfo *framework.NodeInfo
	}
	tests := []struct {
		name                 string
		prepareStoreFn       func(m *mock_plugin.Mockstore)
		originalFilterPlugin framework.FilterPlugin
		args                 args
		want                 *framework.Status
	}{
		{
			name: "success",
			prepareStoreFn: func(m *mock_plugin.Mockstore) {
				m.EXPECT().AddFilterResult("default", "pod1", "node1", "fakeFilterPlugin", resultstore.PassedFilterMessage)
			},
			originalFilterPlugin: fakeFilterPlugin{},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodeInfo: func() *framework.NodeInfo {
					n := &framework.NodeInfo{}
					n.SetNode(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})
					return n
				}(),
			},
			want: nil,
		},
		{
			name:                 "success when it is not filter plugin",
			prepareStoreFn:       func(m *mock_plugin.Mockstore) {},
			originalFilterPlugin: nil, // don't have filter plugin
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				nodeInfo: func() *framework.NodeInfo {
					n := &framework.NodeInfo{}
					n.SetNode(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})
					return n
				}(),
			},
			want: nil,
		},
		{
			name: "fail when original plugin return non-success",
			prepareStoreFn: func(m *mock_plugin.Mockstore) {
				m.EXPECT().AddFilterResult("default", "pod1", "node1", "fakeMustFailFilterScorePlugin", "filter failed")
			},
			originalFilterPlugin: fakeMustFailFilterScorePlugin{},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodeInfo: func() *framework.NodeInfo {
					n := &framework.NodeInfo{}
					n.SetNode(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})
					return n
				}(),
			},
			want: framework.AsStatus(errors.New("filter failed")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			s := mock_plugin.NewMockstore(ctrl)
			tt.prepareStoreFn(s)
			pl := &wrappedPlugin{
				originalFilterPlugin: tt.originalFilterPlugin,
				store:                s,
			}
			got := pl.Filter(context.Background(), nil, tt.args.pod, tt.args.nodeInfo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_wrappedPlugin_Name(t *testing.T) {
	t.Parallel()
	type fields struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "success",
			fields: fields{name: "pluginWrapped"},
			want:   "pluginWrapped",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pl := &wrappedPlugin{
				name: tt.fields.name,
			}
			if got := pl.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wrappedPlugin_NormalizeScore(t *testing.T) {
	t.Parallel()

	type args struct {
		pod    *v1.Pod
		scores framework.NodeScoreList
	}
	tests := []struct {
		name                string
		prepareStoreFn      func(m *mock_plugin.Mockstore)
		originalScorePlugin framework.ScorePlugin
		args                args
		want                *framework.Status
	}{
		{
			name: "success",
			prepareStoreFn: func(m *mock_plugin.Mockstore) {
				m.EXPECT().AddNormalizedScoreResult("default", "pod1", "node1", "fakeScorePlugin", int64(10))
				m.EXPECT().AddNormalizedScoreResult("default", "pod1", "node1", "fakeScorePlugin", int64(200))
			},
			originalScorePlugin: fakeScorePlugin{},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				scores: []framework.NodeScore{
					{
						Name:  "node1",
						Score: 10,
					},
					{
						Name:  "node1",
						Score: 200,
					},
				},
			},
			want: nil,
		},
		{
			name:                "return score 0 when it is not filter plugin",
			prepareStoreFn:      func(m *mock_plugin.Mockstore) {},
			originalScorePlugin: nil, // don't have filter plugin
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				scores: []framework.NodeScore{
					{
						Name:  "node1",
						Score: 10,
					},
				},
			},
			want: nil,
		},
		{
			name:                "fail when original plugin return non-success",
			prepareStoreFn:      func(m *mock_plugin.Mockstore) {},
			originalScorePlugin: fakeMustFailFilterScorePlugin{},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				scores: []framework.NodeScore{
					{
						Name:  "node1",
						Score: 10,
					},
				},
			},
			want: framework.AsStatus(errors.New("normalize failed")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			s := mock_plugin.NewMockstore(ctrl)
			tt.prepareStoreFn(s)
			pl := &wrappedPlugin{
				originalScorePlugin: tt.originalScorePlugin,
				store:               s,
			}
			got := pl.NormalizeScore(context.Background(), nil, tt.args.pod, tt.args.scores)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_wrappedPlugin_Score(t *testing.T) {
	t.Parallel()

	type args struct {
		pod      *v1.Pod
		nodename string
	}
	tests := []struct {
		name                string
		prepareStoreFn      func(m *mock_plugin.Mockstore)
		originalScorePlugin framework.ScorePlugin
		args                args
		want                int64
		wantstatus          *framework.Status
	}{
		{
			name: "success",
			prepareStoreFn: func(m *mock_plugin.Mockstore) {
				m.EXPECT().AddScoreResult("default", "pod1", "node1", "fakeScorePlugin", int64(1))
			},
			originalScorePlugin: fakeScorePlugin{},
			args: args{
				pod:      &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodename: "node1",
			},
			want:       1,
			wantstatus: nil,
		},
		{
			name:                "return score 0 when it is not filter plugin",
			prepareStoreFn:      func(m *mock_plugin.Mockstore) {},
			originalScorePlugin: nil, // don't have filter plugin
			args: args{
				pod:      &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				nodename: "node1",
			},
			want:       0,
			wantstatus: nil,
		},
		{
			name:                "fail when original plugin return non-success",
			prepareStoreFn:      func(m *mock_plugin.Mockstore) {},
			originalScorePlugin: fakeMustFailFilterScorePlugin{},
			args: args{
				pod:      &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodename: "node1",
			},
			want:       0,
			wantstatus: framework.AsStatus(errors.New("score failed")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			s := mock_plugin.NewMockstore(ctrl)
			tt.prepareStoreFn(s)
			pl := &wrappedPlugin{
				originalScorePlugin: tt.originalScorePlugin,
				store:               s,
			}
			got, gotstatus := pl.Score(context.Background(), nil, tt.args.pod, tt.args.nodename)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantstatus, gotstatus)
		})
	}
}

func Test_wrappedPlugin_ScoreExtensions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		originalScorePlugin framework.ScorePlugin
		want                framework.ScoreExtensions
	}{
		{
			name:                "success",
			originalScorePlugin: fakeScorePlugin{},
			want: &wrappedPlugin{
				originalScorePlugin: fakeScorePlugin{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pl := &wrappedPlugin{
				originalScorePlugin: tt.originalScorePlugin,
			}
			got := pl.ScoreExtensions()
			assert.Equal(t, tt.want, got)
		})
	}
}

func defaultPluginConfig() []v1beta3config.PluginConfig {
	var minCandidateNodesPercentage int32 = 10
	var minCandidateNodesAbsolute int32 = 100
	var hardPodAffinityWeight int32 = 1
	var bindTimeoutSeconds int64 = 600

	return []v1beta3config.PluginConfig{
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
		{
			Name: "NodeAffinity",
			Args: runtime.RawExtension{
				Object: &v1beta3config.NodeAffinityArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeAffinityArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta3",
					},
				},
			},
		},
		{
			Name: "NodeResourcesBalancedAllocation",
			Args: runtime.RawExtension{
				Object: &v1beta3config.NodeResourcesBalancedAllocationArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesBalancedAllocationArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta3",
					},
					Resources: []v1beta3config.ResourceSpec{
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
				Object: &v1beta3config.NodeResourcesFitArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "NodeResourcesFitArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta3",
					},
					ScoringStrategy: &v1beta3config.ScoringStrategy{
						Type: "LeastAllocated",
						Resources: []v1beta3config.ResourceSpec{
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
				Object: &v1beta3config.PodTopologySpreadArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PodTopologySpreadArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta3",
					},
					DefaultingType: "System",
				},
			},
		},
		{
			Name: "VolumeBinding",
			Args: runtime.RawExtension{
				Object: &v1beta3config.VolumeBindingArgs{
					TypeMeta: metav1.TypeMeta{
						Kind:       "VolumeBindingArgs",
						APIVersion: "kubescheduler.config.k8s.io/v1beta3",
					},
					BindTimeoutSeconds: &bindTimeoutSeconds,
				},
			},
		},
	}
}

// fake plugins for test

type fakeFilterPlugin struct{}

func (fakeFilterPlugin) Name() string { return "fakeFilterPlugin" }
func (fakeFilterPlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

type fakeScorePlugin struct{}

func (fakeScorePlugin) Name() string { return "fakeScorePlugin" }
func (pl fakeScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func (fakeScorePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	return nil
}

func (fakeScorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	return 1, nil
}

type fakeFilterScorePlugin struct{}

func (fakeFilterScorePlugin) Name() string { return "fakeFilterScorePlugin" }
func (fakeFilterScorePlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

func (pl fakeFilterScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func (fakeFilterScorePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	return nil
}

func (fakeFilterScorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	return 0, nil
}

// all method on this plugin will fail.
type fakeMustFailFilterScorePlugin struct{}

func (fakeMustFailFilterScorePlugin) Name() string { return "fakeMustFailFilterScorePlugin" }
func (fakeMustFailFilterScorePlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return framework.AsStatus(errors.New("filter failed"))
}

func (pl fakeMustFailFilterScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func (fakeMustFailFilterScorePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	return framework.AsStatus(errors.New("normalize failed"))
}

func (fakeMustFailFilterScorePlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	return 0, framework.AsStatus(errors.New("score failed"))
}
