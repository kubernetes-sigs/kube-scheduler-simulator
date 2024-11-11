package plugin

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	mock_plugin "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin/mock"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin/resultstore"
)

func Test_NewWrappedPlugin(t *testing.T) {
	t.Parallel()
	store := resultstore.New(nil)

	type args struct {
		s *resultstore.Store
		p framework.Plugin
	}
	tests := []struct {
		name string
		args args
		want framework.Plugin
	}{
		{
			name: "success with filter plugin",
			args: args{
				s: store,
				p: fakeFilterPlugin{},
			},
			want: &wrappedPlugin{
				name:                 "fakeFilterPluginWrapped",
				originalFilterPlugin: fakeFilterPlugin{},
				originalScorePlugin:  nil,
				store:                store,
			},
		},
		{
			name: "success with postFilter plugin",
			args: args{
				s: store,
				p: fakePostFilterPlugin{},
			},
			want: &wrappedPlugin{
				name:                     "fakePostFilterPluginWrapped",
				originalFilterPlugin:     nil,
				originalPostFilterPlugin: fakePostFilterPlugin{},
				originalScorePlugin:      nil,
				store:                    store,
			},
		},
		{
			name: "success with score plugin",
			args: args{
				s: store,
				p: fakeScorePlugin{},
			},
			want: &wrappedPlugin{
				name:                 "fakeScorePluginWrapped",
				originalFilterPlugin: nil,
				originalScorePlugin:  fakeScorePlugin{},
				store:                store,
			},
		},
		{
			name: "success with score/filter/postFilter plugin",
			args: args{
				s: store,
				p: fakeWrappedPlugin{},
			},
			want: &wrappedPlugin{
				name:                     "fakeWrappedPluginWrapped",
				originalFilterPlugin:     fakeWrappedPlugin{},
				originalScorePlugin:      fakeWrappedPlugin{},
				originalPostFilterPlugin: fakeWrappedPlugin{},
				store:                    store,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewWrappedPlugin(tt.args.s, tt.args.p)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_NewWrappedPlugin_WithPluginNameOption(t *testing.T) {
	t.Parallel()
	store := resultstore.New(nil)

	type args struct {
		s    *resultstore.Store
		p    framework.Plugin
		name string
	}
	tests := []struct {
		name string
		args args
		want framework.Plugin
	}{
		{
			name: "plugin name is named by user",
			args: args{
				s:    store,
				p:    fakeFilterPlugin{},
				name: "userNamedPlugin",
			},
			want: &wrappedPlugin{
				name:                 "userNamedPlugin",
				originalFilterPlugin: fakeFilterPlugin{},
				originalScorePlugin:  nil,
				store:                store,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewWrappedPlugin(tt.args.s, tt.args.p, WithPluginNameOption(&tt.args.name))
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
		prepareStoreFn       func(m *mock_plugin.MockStore)
		originalFilterPlugin framework.FilterPlugin
		args                 args
		want                 *framework.Status
	}{
		{
			name: "success",
			prepareStoreFn: func(m *mock_plugin.MockStore) {
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
			prepareStoreFn:       func(m *mock_plugin.MockStore) {},
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
			prepareStoreFn: func(m *mock_plugin.MockStore) {
				m.EXPECT().AddFilterResult("default", "pod1", "node1", "fakeMustFailWrappedPlugin", "filter failed")
			},
			originalFilterPlugin: fakeMustFailWrappedPlugin{},
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

			s := mock_plugin.NewMockStore(ctrl)
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

func Test_wrappedPlugin_Filter_WithPluginExtender(t *testing.T) {
	t.Parallel()

	type args struct {
		pod      *v1.Pod
		nodeInfo *framework.NodeInfo
	}
	tests := []struct {
		name              string
		prepareEachMockFn func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockFilterPlugin, fe *mock_plugin.MockFilterPluginExtender, as args)
		args              args
		wantstatus        *framework.Status
	}{
		{
			name: "return AfterFilter's results when Filter is successful",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockFilterPlugin, fe *mock_plugin.MockFilterPluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforeFilter returned")
				success2 := framework.NewStatus(framework.Success, "Filter returned")
				success3 := framework.NewStatus(framework.Success, "AfterFilter returned")
				fe.EXPECT().BeforeFilter(ctx, nil, as.pod, as.nodeInfo).Return(success1)
				p.EXPECT().Filter(ctx, nil, as.pod, as.nodeInfo).Return(success2)
				fe.EXPECT().AfterFilter(ctx, nil, as.pod, as.nodeInfo, success2).Return(success3)
				p.EXPECT().Name().Return("fakeFilterPlugin").AnyTimes()
				// Filter sotres resultstore.PassedFilterMessage if it is successful.
				s.EXPECT().AddFilterResult("default", "pod1", "node1", "fakeFilterPlugin", resultstore.PassedFilterMessage)
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodeInfo: func() *framework.NodeInfo {
					n := &framework.NodeInfo{}
					n.SetNode(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})
					return n
				}(),
			},
			wantstatus: framework.NewStatus(framework.Success, "AfterFilter returned"),
		},
		{
			name: "return AfterFilter's results if Filter is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockFilterPlugin, fe *mock_plugin.MockFilterPluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforeFilter returned")
				failure := framework.NewStatus(framework.Error, "Filter returned")
				success3 := framework.NewStatus(framework.Success, "AfterFilter returned")
				fe.EXPECT().BeforeFilter(ctx, nil, as.pod, as.nodeInfo).Return(success1)
				p.EXPECT().Filter(ctx, nil, as.pod, as.nodeInfo).Return(failure)
				fe.EXPECT().AfterFilter(ctx, nil, as.pod, as.nodeInfo, failure).Return(success3)
				p.EXPECT().Name().Return("fakeFilterPlugin").AnyTimes()
				// Filter stores own message if it is successful.
				s.EXPECT().AddFilterResult("default", "pod1", "node1", "fakeFilterPlugin", failure.Message())
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodeInfo: func() *framework.NodeInfo {
					n := &framework.NodeInfo{}
					n.SetNode(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})
					return n
				}(),
			},
			wantstatus: framework.NewStatus(framework.Success, "AfterFilter returned"),
		},
		{
			name: "return BeforeFilter's results when BeforeFilter is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockFilterPlugin, fe *mock_plugin.MockFilterPluginExtender, as args) {
				failure := framework.NewStatus(framework.Error, "BeforeFilter returned")
				fe.EXPECT().BeforeFilter(ctx, nil, as.pod, as.nodeInfo).Return(failure)
				p.EXPECT().Name().Return("fakeFilterPlugin").AnyTimes()
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodeInfo: func() *framework.NodeInfo {
					n := &framework.NodeInfo{}
					n.SetNode(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})
					return n
				}(),
			},
			wantstatus: framework.NewStatus(framework.Error, "BeforeFilter returned"),
		},
		{
			name: "return AfterFilter's results when AfterFilter is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockFilterPlugin, fe *mock_plugin.MockFilterPluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforeFilter returned")
				success2 := framework.NewStatus(framework.Success, "Filter returned")
				failure := framework.NewStatus(framework.Error, "AfterFilter returned")
				fe.EXPECT().BeforeFilter(ctx, nil, as.pod, as.nodeInfo).Return(success1)
				p.EXPECT().Filter(ctx, nil, as.pod, as.nodeInfo).Return(success2)
				fe.EXPECT().AfterFilter(ctx, nil, as.pod, as.nodeInfo, success2).Return(failure)
				p.EXPECT().Name().Return("fakeFilterPlugin").AnyTimes()
				s.EXPECT().AddFilterResult("default", "pod1", "node1", "fakeFilterPlugin", resultstore.PassedFilterMessage)
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodeInfo: func() *framework.NodeInfo {
					n := &framework.NodeInfo{}
					n.SetNode(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})
					return n
				}(),
			},
			wantstatus: framework.NewStatus(framework.Error, "AfterFilter returned"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockFilterPlugin(ctrl)
			fe := mock_plugin.NewMockFilterPluginExtender(ctrl)
			ctx := context.Background()
			tt.prepareEachMockFn(ctx, s, p, fe, tt.args)
			pl, ok := NewWrappedPlugin(s, p, WithExtendersOption(func(_ SimulatorHandle) PluginExtenders {
				return PluginExtenders{
					FilterPluginExtender: fe,
				}
			})).(*wrappedPlugin)
			if !ok { // should never happen
				t.Fatalf("Assert to wrapped plugin: %v", ok)
			}
			gotstatus := pl.Filter(ctx, nil, tt.args.pod, tt.args.nodeInfo)
			assert.Equal(t, tt.wantstatus, gotstatus)
		})
	}
}

func Test_wrappedPlugin_PostFilter(t *testing.T) {
	t.Parallel()

	type args struct {
		pod                   *v1.Pod
		filteredNodeStatusMap framework.NodeToStatusMap
	}
	tests := []struct {
		name                     string
		prepareStoreFn           func(m *mock_plugin.MockStore)
		originalPostFilterPlugin framework.PostFilterPlugin
		args                     args
		wantResult               *framework.PostFilterResult
		wantStatus               *framework.Status
	}{
		{
			name: "success",
			prepareStoreFn: func(m *mock_plugin.MockStore) {
				m.EXPECT().AddPostFilterResult("default", "pod1", "node1", "fakePostFilterPlugin", gomock.Any()).Do(func(_, _, _, _ string, nodeNames []string) {
					sort.SliceStable(nodeNames, func(i, j int) bool {
						return nodeNames[i] < nodeNames[j]
					})
					assert.Equal(t, []string{"node1", "node2"}, nodeNames)
				})
			},
			originalPostFilterPlugin: fakePostFilterPlugin{},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				filteredNodeStatusMap: framework.NodeToStatusMap{
					"node1": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
					"node2": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
				},
			},
			wantResult: &framework.PostFilterResult{
				NominatingInfo: &framework.NominatingInfo{
					NominatedNodeName: "node1",
				},
			},
			wantStatus: framework.NewStatus(framework.Success, "postfilter success"),
		},
		{
			name:                     "success when it is not post filter plugin",
			prepareStoreFn:           func(m *mock_plugin.MockStore) {},
			originalPostFilterPlugin: nil, // don't have post filter plugin
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
				filteredNodeStatusMap: framework.NodeToStatusMap{
					"node1": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
					"node2": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
				},
			},
			wantResult: nil,
			wantStatus: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "fail when original plugin return non-success",
			prepareStoreFn: func(m *mock_plugin.MockStore) {
				m.EXPECT().AddPostFilterResult("default", "pod1", "", "fakeMustFailWrappedPlugin", gomock.Any()).Do(func(_, _, _, _ string, nodeNames []string) {
					sort.SliceStable(nodeNames, func(i, j int) bool {
						return nodeNames[i] < nodeNames[j]
					})
					assert.Equal(t, []string{"node1", "node2"}, nodeNames)
				})
			},
			originalPostFilterPlugin: fakeMustFailWrappedPlugin{},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				filteredNodeStatusMap: framework.NodeToStatusMap{
					"node1": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
					"node2": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
				},
			},
			wantResult: nil,
			wantStatus: framework.AsStatus(errors.New("postFilter failed")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			s := mock_plugin.NewMockStore(ctrl)
			tt.prepareStoreFn(s)
			pl := &wrappedPlugin{
				originalPostFilterPlugin: tt.originalPostFilterPlugin,
				store:                    s,
			}
			gotResult, gotStatus := pl.PostFilter(context.Background(), nil, tt.args.pod, tt.args.filteredNodeStatusMap)
			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantStatus, gotStatus)
		})
	}
}

func Test_wrappedPlugin_PostFilter_WithPluginExtender(t *testing.T) {
	t.Parallel()

	type args struct {
		pod                   *v1.Pod
		filteredNodeStatusMap framework.NodeToStatusMap
	}
	tests := []struct {
		name              string
		prepareEachMockFn func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockPostFilterPlugin, fe *mock_plugin.MockPostFilterPluginExtender, as args)
		args              args
		wantResult        *framework.PostFilterResult
		wantStatus        *framework.Status
	}{
		{
			name: "return AfterPostFilter's results when PostFilter is successful",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockPostFilterPlugin, fe *mock_plugin.MockPostFilterPluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforePostFilter returned")
				success2 := framework.NewStatus(framework.Success, "PostFilter returned")
				success3 := framework.NewStatus(framework.Success, "AfterPostFilter returned")
				result1 := &framework.PostFilterResult{
					NominatingInfo: &framework.NominatingInfo{
						NominatedNodeName: "node1",
					},
				}
				result2 := &framework.PostFilterResult{
					NominatingInfo: &framework.NominatingInfo{
						NominatedNodeName: "node2",
					},
				}
				fe.EXPECT().BeforePostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap).Return(nil, success1)
				p.EXPECT().PostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap).Return(result1, success2)
				fe.EXPECT().AfterPostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap, result1, success2).Return(result2, success3)
				p.EXPECT().Name().Return("fakePostFilterPlugin").AnyTimes()
				// PostFilter sotres resultstore.PassedFilterMessage if it is successful.
				s.EXPECT().AddPostFilterResult("default", "pod1", "node1", "fakePostFilterPlugin", gomock.Any()).Do(func(_, _, _, _ string, nodeNames []string) {
					sort.SliceStable(nodeNames, func(i, j int) bool {
						return nodeNames[i] < nodeNames[j]
					})
					assert.Equal(t, []string{"node1", "node2"}, nodeNames)
				})
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				filteredNodeStatusMap: framework.NodeToStatusMap{
					"node1": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
					"node2": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
				},
			},
			wantResult: &framework.PostFilterResult{
				NominatingInfo: &framework.NominatingInfo{
					NominatedNodeName: "node2",
				},
			},
			wantStatus: framework.NewStatus(framework.Success, "AfterPostFilter returned"),
		},
		{
			name: "return AfterPostFilter's results if Filter is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockPostFilterPlugin, fe *mock_plugin.MockPostFilterPluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforePostFilter returned")
				failure := framework.NewStatus(framework.Error, "PostFilter returned")
				success3 := framework.NewStatus(framework.Success, "AfterPostFilter returned")
				result2 := &framework.PostFilterResult{
					NominatingInfo: &framework.NominatingInfo{
						NominatedNodeName: "node2",
					},
				}
				fe.EXPECT().BeforePostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap).Return(nil, success1)
				p.EXPECT().PostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap).Return(nil, failure)
				fe.EXPECT().AfterPostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap, nil, failure).Return(result2, success3)
				p.EXPECT().Name().Return("fakePostFilterPlugin").AnyTimes()
				// Filter stores own message if it is successful.
				s.EXPECT().AddPostFilterResult("default", "pod1", "", "fakePostFilterPlugin", gomock.Any()).Do(func(_, _, _, _ string, nodeNames []string) {
					sort.SliceStable(nodeNames, func(i, j int) bool {
						return nodeNames[i] < nodeNames[j]
					})
					assert.Equal(t, []string{"node1", "node2"}, nodeNames)
				})
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				filteredNodeStatusMap: framework.NodeToStatusMap{
					"node1": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
					"node2": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
				},
			},
			wantResult: &framework.PostFilterResult{
				NominatingInfo: &framework.NominatingInfo{
					NominatedNodeName: "node2",
				},
			},
			wantStatus: framework.NewStatus(framework.Success, "AfterPostFilter returned"),
		},
		{
			name: "return BeforeFilter's results when BeforeFilter is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockPostFilterPlugin, fe *mock_plugin.MockPostFilterPluginExtender, as args) {
				failure := framework.NewStatus(framework.Error, "BeforePostFilter returned")
				fe.EXPECT().BeforePostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap).Return(nil, failure)
				p.EXPECT().Name().Return("fakePostFilterPlugin").AnyTimes()
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				filteredNodeStatusMap: framework.NodeToStatusMap{
					"node1": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
					"node2": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
				},
			},
			wantResult: nil,
			wantStatus: framework.NewStatus(framework.Error, "BeforePostFilter returned"),
		},
		{
			name: "return AfterFilter's results when AfterFilter is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockPostFilterPlugin, fe *mock_plugin.MockPostFilterPluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforePostFilter returned")
				success2 := framework.NewStatus(framework.Success, "PostFilter returned")
				result1 := &framework.PostFilterResult{
					NominatingInfo: &framework.NominatingInfo{
						NominatedNodeName: "node1",
					},
				}
				failure := framework.NewStatus(framework.Error, "AfterPostFilter returned")
				fe.EXPECT().BeforePostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap).Return(nil, success1)
				p.EXPECT().PostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap).Return(result1, success2)
				fe.EXPECT().AfterPostFilter(ctx, nil, as.pod, as.filteredNodeStatusMap, result1, success2).Return(nil, failure)
				p.EXPECT().Name().Return("fakePostFilterPlugin").AnyTimes()
				// PostFilter sotres resultstore.PassedFilterMessage if it is successful.
				s.EXPECT().AddPostFilterResult("default", "pod1", "node1", "fakePostFilterPlugin", gomock.Any()).Do(func(_, _, _, _ string, nodeNames []string) {
					sort.SliceStable(nodeNames, func(i, j int) bool {
						return nodeNames[i] < nodeNames[j]
					})
					assert.Equal(t, []string{"node1", "node2"}, nodeNames)
				})
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				filteredNodeStatusMap: framework.NodeToStatusMap{
					"node1": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
					"node2": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
				},
			},
			wantResult: nil,
			wantStatus: framework.NewStatus(framework.Error, "AfterPostFilter returned"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockPostFilterPlugin(ctrl)
			fe := mock_plugin.NewMockPostFilterPluginExtender(ctrl)
			ctx := context.Background()
			tt.prepareEachMockFn(ctx, s, p, fe, tt.args)
			pl, ok := NewWrappedPlugin(s, p, WithExtendersOption(func(_ SimulatorHandle) PluginExtenders {
				return PluginExtenders{
					PostFilterPluginExtender: fe,
				}
			})).(*wrappedPlugin)
			if !ok { // should never happen
				t.Fatalf("Assert to wrapped plugin: %v", ok)
			}
			gotResult, gotStatus := pl.PostFilter(context.Background(), nil, tt.args.pod, tt.args.filteredNodeStatusMap)
			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantStatus, gotStatus)
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
		prepareStoreFn      func(m *mock_plugin.MockStore)
		originalScorePlugin framework.ScorePlugin
		args                args
		want                *framework.Status
	}{
		{
			name: "success",
			prepareStoreFn: func(m *mock_plugin.MockStore) {
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
			prepareStoreFn:      func(m *mock_plugin.MockStore) {},
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
			prepareStoreFn:      func(m *mock_plugin.MockStore) {},
			originalScorePlugin: fakeMustFailWrappedPlugin{},
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

			s := mock_plugin.NewMockStore(ctrl)
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

func Test_wrappedPlugin_NormalizeScore_WithPluginExtender(t *testing.T) {
	t.Parallel()

	type args struct {
		pod    *v1.Pod
		scores framework.NodeScoreList
	}
	tests := []struct {
		name                      string
		prepareEachMockFn         func(ctx context.Context, s *mock_plugin.MockStore, se *mock_plugin.MockScoreExtensions, sp *mock_plugin.MockScorePlugin, spe *mock_plugin.MockNormalizeScorePluginExtender, as args)
		calOnBeforeNormalizeScore func(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList)
		args                      args
		wantScores                framework.NodeScoreList
		wantstatus                *framework.Status
	}{
		{
			name: "return AfterNormalizeScore's results when NormalizeScore is successful",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, se *mock_plugin.MockScoreExtensions, sp *mock_plugin.MockScorePlugin, spe *mock_plugin.MockNormalizeScorePluginExtender, as args) {
				calOnNormalizeScore := func(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) {
					for i := range scores {
						scores[i].Score += 1000
					}
				}
				calOnAfterNormalizeScore := func(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList, fstatus *framework.Status) {
					for i := range scores {
						scores[i].Score += 1000
					}
				}
				success1 := framework.NewStatus(framework.Success, "BeforeNormalizeScore returned")
				success2 := framework.NewStatus(framework.Success, "NormalizeScore returned")
				success3 := framework.NewStatus(framework.Success, "AfterNormalizeScore returned")
				spe.EXPECT().BeforeNormalizeScore(ctx, nil, as.pod, as.scores).Return(success1).Do(calOnNormalizeScore)
				sp.EXPECT().ScoreExtensions().Return(se).Times(2)
				se.EXPECT().NormalizeScore(ctx, nil, as.pod, as.scores).Return(success2).Do(calOnNormalizeScore)
				spe.EXPECT().AfterNormalizeScore(ctx, nil, as.pod, as.scores, success2).Return(success3).Do(calOnAfterNormalizeScore)
				sp.EXPECT().Name().Return("fakeNormalizeScorePlugin").AnyTimes()
				s.EXPECT().AddNormalizedScoreResult("default", "pod1", "node1", "fakeNormalizeScorePlugin", int64(2000))
				s.EXPECT().AddNormalizedScoreResult("default", "pod1", "node2", "fakeNormalizeScorePlugin", int64(2010))
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				scores: []framework.NodeScore{
					{
						Name:  "node1",
						Score: 0,
					},
					{
						Name:  "node2",
						Score: 10,
					},
				},
			},
			wantScores: []framework.NodeScore{
				{
					Name:  "node1",
					Score: 3000,
				},
				{
					Name:  "node2",
					Score: 3010,
				},
			},
			wantstatus: framework.NewStatus(framework.Success, "AfterNormalizeScore returned"),
		},
		{
			name: "return AfterNormalizeScore's results when NormalizeScore is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, se *mock_plugin.MockScoreExtensions, sp *mock_plugin.MockScorePlugin, spe *mock_plugin.MockNormalizeScorePluginExtender, as args) {
				calOnNormalizeScore := func(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) {
					for i := range scores {
						scores[i].Score += 1000
					}
				}
				calOnAfterNormalizeScore := func(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList, fstatus *framework.Status) {
					for i := range scores {
						scores[i].Score += 1000
					}
				}
				success1 := framework.NewStatus(framework.Success, "BeforeNormalizeScore returned")
				failure := framework.NewStatus(framework.Error, "NormalizeScore returned")
				success3 := framework.NewStatus(framework.Success, "AfterNormalizeScore returned")
				spe.EXPECT().BeforeNormalizeScore(ctx, nil, as.pod, as.scores).Return(success1).Do(calOnNormalizeScore)
				sp.EXPECT().ScoreExtensions().Return(se).Times(2)
				se.EXPECT().NormalizeScore(ctx, nil, as.pod, as.scores).Return(failure).Do(calOnNormalizeScore)
				spe.EXPECT().AfterNormalizeScore(ctx, nil, as.pod, as.scores, failure).Return(success3).Do(calOnAfterNormalizeScore)
				sp.EXPECT().Name().Return("fakeNormalizeScorePlugin").AnyTimes()
				// NormalizeScore isnt't stores own results if return error.
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				scores: []framework.NodeScore{
					{
						Name:  "node1",
						Score: 0,
					},
					{
						Name:  "node2",
						Score: 10,
					},
				},
			},
			wantScores: []framework.NodeScore{
				{
					Name:  "node1",
					Score: 3000,
				},
				{
					Name:  "node2",
					Score: 3010,
				},
			},
			wantstatus: framework.NewStatus(framework.Success, "AfterNormalizeScore returned"),
		},
		{
			name: "return AfterNormalizeScore's results, when NormalizeScore is successful and AfterNormalizeScore is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, se *mock_plugin.MockScoreExtensions, sp *mock_plugin.MockScorePlugin, spe *mock_plugin.MockNormalizeScorePluginExtender, as args) {
				calOnNormalizeScore := func(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) {
					for i := range scores {
						scores[i].Score += 1000
					}
				}
				calOnAfterNormalizeScore := func(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList, fstatus *framework.Status) {
					for i := range scores {
						scores[i].Score += 1000
					}
				}
				success1 := framework.NewStatus(framework.Success, "BeforeNormalizeScore returned")
				success2 := framework.NewStatus(framework.Success, "NormalizeScore returned")
				failure := framework.NewStatus(framework.Error, "AfterNormalizeScore returned")
				spe.EXPECT().BeforeNormalizeScore(ctx, nil, as.pod, as.scores).Return(success1).Do(calOnNormalizeScore)
				sp.EXPECT().ScoreExtensions().Return(se).Times(2)
				se.EXPECT().NormalizeScore(ctx, nil, as.pod, as.scores).Return(success2).Do(calOnNormalizeScore)
				spe.EXPECT().AfterNormalizeScore(ctx, nil, as.pod, as.scores, success2).Return(failure).Do(calOnAfterNormalizeScore)
				sp.EXPECT().Name().Return("fakeNormalizeScorePlugin").AnyTimes()
				s.EXPECT().AddNormalizedScoreResult("default", "pod1", "node1", "fakeNormalizeScorePlugin", int64(2000))
				s.EXPECT().AddNormalizedScoreResult("default", "pod1", "node2", "fakeNormalizeScorePlugin", int64(2010))
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				scores: []framework.NodeScore{
					{
						Name:  "node1",
						Score: 0,
					},
					{
						Name:  "node2",
						Score: 10,
					},
				},
			},
			wantScores: []framework.NodeScore{
				{
					Name:  "node1",
					Score: 3000,
				},
				{
					Name:  "node2",
					Score: 3010,
				},
			},
			wantstatus: framework.NewStatus(framework.Error, "AfterNormalizeScore returned"),
		},
		{
			name: "return BeforeNormalizeScore when BeforeNormalizeScore is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, se *mock_plugin.MockScoreExtensions, sp *mock_plugin.MockScorePlugin, spe *mock_plugin.MockNormalizeScorePluginExtender, as args) {
				calOnNormalizeScore := func(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) {
					for i := range scores {
						scores[i].Score += 1000
					}
				}
				success1 := framework.NewStatus(framework.Error, "BeforeNormalizeScore returned")
				spe.EXPECT().BeforeNormalizeScore(ctx, nil, as.pod, as.scores).Return(success1).Do(calOnNormalizeScore)
				sp.EXPECT().ScoreExtensions().Return(se).Times(1)
				sp.EXPECT().Name().Return("fakeNormalizeScorePlugin").AnyTimes()
			},
			args: args{
				pod: &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				scores: []framework.NodeScore{
					{
						Name:  "node1",
						Score: 0,
					},
					{
						Name:  "node2",
						Score: 10,
					},
				},
			},
			wantScores: []framework.NodeScore{
				{
					Name:  "node1",
					Score: 1000,
				},
				{
					Name:  "node2",
					Score: 1010,
				},
			},
			wantstatus: framework.NewStatus(framework.Error, "BeforeNormalizeScore returned"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			se := mock_plugin.NewMockScoreExtensions(ctrl)
			sp := mock_plugin.NewMockScorePlugin(ctrl)

			spe := mock_plugin.NewMockNormalizeScorePluginExtender(ctrl)
			ctx := context.Background()
			tt.prepareEachMockFn(ctx, s, se, sp, spe, tt.args)
			pl, ok := NewWrappedPlugin(s, sp, WithExtendersOption(func(_ SimulatorHandle) PluginExtenders {
				return PluginExtenders{
					NormalizeScorePluginExtender: spe,
				}
			})).(*wrappedPlugin)
			if !ok { // should never happen
				t.Fatalf("Assert to wrapped plugin: %v", ok)
			}
			gotstatus := pl.NormalizeScore(ctx, nil, tt.args.pod, tt.args.scores)
			assert.Equal(t, tt.wantScores, tt.args.scores)
			assert.Equal(t, tt.wantstatus, gotstatus)
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
		prepareStoreFn      func(m *mock_plugin.MockStore)
		originalScorePlugin framework.ScorePlugin
		args                args
		want                int64
		wantstatus          *framework.Status
	}{
		{
			name: "success",
			prepareStoreFn: func(m *mock_plugin.MockStore) {
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
			prepareStoreFn:      func(m *mock_plugin.MockStore) {},
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
			prepareStoreFn:      func(m *mock_plugin.MockStore) {},
			originalScorePlugin: fakeMustFailWrappedPlugin{},
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

			s := mock_plugin.NewMockStore(ctrl)
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

func Test_wrappedPlugin_Score_WithPluginExtender(t *testing.T) {
	t.Parallel()

	type args struct {
		pod      *v1.Pod
		nodename string
	}
	tests := []struct {
		name              string
		prepareEachMockFn func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockScorePlugin, se *mock_plugin.MockScorePluginExtender, as args)
		args              args
		want              int64
		wantstatus        *framework.Status
	}{
		{
			name: "return AfterScore's results when Score is successful",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockScorePlugin, se *mock_plugin.MockScorePluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforeScore returned")
				success2 := framework.NewStatus(framework.Success, "Score returned")
				success3 := framework.NewStatus(framework.Success, "AfterScore returned")
				se.EXPECT().BeforeScore(ctx, nil, as.pod, "node1").Return(int64(1111), success1)
				p.EXPECT().Score(ctx, nil, as.pod, "node1").Return(int64(2222), success2)
				se.EXPECT().AfterScore(ctx, nil, as.pod, "node1", int64(2222), success2).Return(int64(3333), success3)
				p.EXPECT().Name().Return("fakeScorePlugin").AnyTimes()
				s.EXPECT().AddScoreResult("default", "pod1", "node1", "fakeScorePlugin", int64(2222))
			},
			args: args{
				pod:      &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodename: "node1",
			},
			want:       3333,
			wantstatus: framework.NewStatus(framework.Success, "AfterScore returned"),
		},
		{
			name: "return AfterScore's results & does not call AddScoreResult after Score, if Score fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockScorePlugin, se *mock_plugin.MockScorePluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforeScore returned")
				failure := framework.NewStatus(framework.Error, "Score returned")
				success3 := framework.NewStatus(framework.Success, "AfterScore returned")
				se.EXPECT().BeforeScore(ctx, nil, as.pod, "node1").Return(int64(1111), success1)
				p.EXPECT().Score(ctx, nil, as.pod, "node1").Return(int64(2222), failure)
				se.EXPECT().AfterScore(ctx, nil, as.pod, "node1", int64(2222), failure).Return(int64(3333), success3)
				p.EXPECT().Name().Return("fakeScorePlugin").AnyTimes()
				s.EXPECT().AddScoreResult(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			args: args{
				pod:      &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodename: "node1",
			},
			want:       3333,
			wantstatus: framework.NewStatus(framework.Success, "AfterScore returned"),
		},
		{
			name: "return Before's results & does not call Score, if BeforeScore fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockScorePlugin, se *mock_plugin.MockScorePluginExtender, as args) {
				failure := framework.NewStatus(framework.Error, "BeforeScore returned")
				se.EXPECT().BeforeScore(ctx, nil, as.pod, "node1").Return(int64(1111), failure)
				p.EXPECT().Name().Return("fakeScorePlugin").AnyTimes()
			},
			args: args{
				pod:      &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodename: "node1",
			},
			want:       1111,
			wantstatus: framework.NewStatus(framework.Error, "BeforeScore returned"),
		},
		{
			name: "return AfterScore's results when AfterScore is fails",
			prepareEachMockFn: func(ctx context.Context, s *mock_plugin.MockStore, p *mock_plugin.MockScorePlugin, se *mock_plugin.MockScorePluginExtender, as args) {
				success1 := framework.NewStatus(framework.Success, "BeforeScore returned")
				success2 := framework.NewStatus(framework.Success, "Score returned")
				failure := framework.NewStatus(framework.Error, "AfterScore returned")
				se.EXPECT().BeforeScore(ctx, nil, as.pod, "node1").Return(int64(1111), success1)
				p.EXPECT().Score(ctx, nil, as.pod, "node1").Return(int64(2222), success2)
				se.EXPECT().AfterScore(ctx, nil, as.pod, "node1", int64(2222), success2).Return(int64(3333), failure)
				p.EXPECT().Name().Return("fakeScorePlugin").AnyTimes()
				s.EXPECT().AddScoreResult("default", "pod1", "node1", "fakeScorePlugin", int64(2222))
			},
			args: args{
				pod:      &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
				nodename: "node1",
			},
			want:       3333,
			wantstatus: framework.NewStatus(framework.Error, "AfterScore returned"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockScorePlugin(ctrl)
			se := mock_plugin.NewMockScorePluginExtender(ctrl)
			ctx := context.Background()
			tt.prepareEachMockFn(ctx, s, p, se, tt.args)
			pl, ok := NewWrappedPlugin(s, p, WithExtendersOption(func(_ SimulatorHandle) PluginExtenders {
				return PluginExtenders{
					ScorePluginExtender: se,
				}
			})).(*wrappedPlugin)
			if !ok { // should never happen
				t.Fatalf("Assert to wrapped plugin: %v", ok)
			}
			gotscore, gotstatus := pl.Score(ctx, nil, tt.args.pod, tt.args.nodename)
			assert.Equal(t, tt.want, gotscore)
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

func Test_wrappedPlugin_PreScore(t *testing.T) {
	t.Parallel()

	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "namespace"}}
	testNodes := []*v1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "node"}}}
	testNodeInfos := make([]*framework.NodeInfo, len(testNodes))
	for i, node := range testNodes {
		nodeInfo := &framework.NodeInfo{}
		nodeInfo.SetNode(node)
		testNodeInfos[i] = nodeInfo
	}

	tests := []struct {
		name           string
		prepareMocksFn func(s *mock_plugin.MockStore, se *mock_plugin.MockPreScorePlugin, extender *mock_plugin.MockPreScorePluginExtender)
		noExtender     bool
		want           *framework.Status
	}{
		{
			name: "happy with extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreScorePlugin, extender *mock_plugin.MockPreScorePluginExtender) {
				extender.EXPECT().BeforePreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreScoreResult("namespace", "pod", "name", resultstore.SuccessMessage)
				extender.EXPECT().AfterPreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos, framework.NewStatus(framework.Success)).Return(framework.NewStatus(framework.Success))
			},
			want: framework.NewStatus(framework.Success),
		},
		{
			name: "unhappy: BeforePreScore returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreScorePlugin, extender *mock_plugin.MockPreScorePluginExtender) {
				extender.EXPECT().BeforePreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: AfterPreScore returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreScorePlugin, extender *mock_plugin.MockPreScorePluginExtender) {
				extender.EXPECT().BeforePreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreScoreResult("namespace", "pod", "name", resultstore.SuccessMessage)
				extender.EXPECT().AfterPreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos, framework.NewStatus(framework.Success)).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: PreScore and AfterPreScore return non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreScorePlugin, extender *mock_plugin.MockPreScorePluginExtender) {
				extender.EXPECT().BeforePreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreScoreResult("namespace", "pod", "name", "error")
				extender.EXPECT().AfterPreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos, framework.NewStatus(framework.Unschedulable, "error")).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "happy: PreScore returns non-success, but AfterPreScore return success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreScorePlugin, extender *mock_plugin.MockPreScorePluginExtender) {
				extender.EXPECT().BeforePreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreScoreResult("namespace", "pod", "name", "error")
				extender.EXPECT().AfterPreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos, framework.NewStatus(framework.Unschedulable, "error")).Return(framework.NewStatus(framework.Success))
			},
			want: framework.NewStatus(framework.Success),
		},
		{
			name: "happy without extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreScorePlugin, extender *mock_plugin.MockPreScorePluginExtender) {
				se.EXPECT().PreScore(gomock.Any(), gomock.Any(), testPod, testNodeInfos).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreScoreResult("namespace", "pod", "name", resultstore.SuccessMessage)
			},
			noExtender: true,
			want:       framework.NewStatus(framework.Success),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockPreScorePlugin(ctrl)
			ex := mock_plugin.NewMockPreScorePluginExtender(ctrl)
			tt.prepareMocksFn(s, p, ex)

			w := &wrappedPlugin{
				store:                  s,
				originalPreScorePlugin: p,
			}
			if !tt.noExtender {
				w.preScorePluginExtender = ex
			}
			assert.Equalf(t, tt.want, w.PreScore(context.Background(), framework.NewCycleState(), testPod, testNodeInfos), "PreScore(ctx, cyclestate, %v, %v)", testPod, testNodes)
		})
	}
}

func Test_wrappedPlugin_PreFilter(t *testing.T) {
	t.Parallel()
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "namespace"}}

	tests := []struct {
		name           string
		prepareMocksFn func(s *mock_plugin.MockStore, se *mock_plugin.MockPreFilterPlugin, extender *mock_plugin.MockPreFilterPluginExtender)
		noExtender     bool
		want           *framework.PreFilterResult
		want1          *framework.Status
	}{
		{
			name: "happy with extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreFilterPlugin, extender *mock_plugin.MockPreFilterPluginExtender) {
				extender.EXPECT().BeforePreFilter(gomock.Any(), gomock.Any(), testPod).Return(nil, framework.NewStatus(framework.Success))
				se.EXPECT().PreFilter(gomock.Any(), gomock.Any(), testPod).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreFilterResult("namespace", "pod", "name", resultstore.SuccessMessage, &framework.PreFilterResult{NodeNames: sets.New("hoge")})
				extender.EXPECT().AfterPreFilter(gomock.Any(), gomock.Any(), testPod, &framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Success)).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Success))
			},
			want:  &framework.PreFilterResult{NodeNames: sets.New("hoge")},
			want1: framework.NewStatus(framework.Success),
		},
		{
			name: "unhappy: BeforePreFilter returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreFilterPlugin, extender *mock_plugin.MockPreFilterPluginExtender) {
				extender.EXPECT().BeforePreFilter(gomock.Any(), gomock.Any(), testPod).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Unschedulable))
			},
			want:  &framework.PreFilterResult{NodeNames: sets.New("hoge")},
			want1: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: AfterPreFilter returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreFilterPlugin, extender *mock_plugin.MockPreFilterPluginExtender) {
				extender.EXPECT().BeforePreFilter(gomock.Any(), gomock.Any(), testPod).Return(nil, framework.NewStatus(framework.Success))
				se.EXPECT().PreFilter(gomock.Any(), gomock.Any(), testPod).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreFilterResult("namespace", "pod", "name", resultstore.SuccessMessage, &framework.PreFilterResult{NodeNames: sets.New("hoge")})
				extender.EXPECT().AfterPreFilter(gomock.Any(), gomock.Any(), testPod, &framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Success)).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Unschedulable))
			},
			want:  &framework.PreFilterResult{NodeNames: sets.New("hoge")},
			want1: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: PreFilter and AfterPreFilter return non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreFilterPlugin, extender *mock_plugin.MockPreFilterPluginExtender) {
				extender.EXPECT().BeforePreFilter(gomock.Any(), gomock.Any(), testPod).Return(nil, framework.NewStatus(framework.Success))
				se.EXPECT().PreFilter(gomock.Any(), gomock.Any(), testPod).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreFilterResult("namespace", "pod", "name", "error", &framework.PreFilterResult{NodeNames: sets.New("hoge")})
				extender.EXPECT().AfterPreFilter(gomock.Any(), gomock.Any(), testPod, &framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Unschedulable, "error")).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Unschedulable))
			},
			want:  &framework.PreFilterResult{NodeNames: sets.New("hoge")},
			want1: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "happy: PreFilter returns non-success, but AfterPreFilter return success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreFilterPlugin, extender *mock_plugin.MockPreFilterPluginExtender) {
				extender.EXPECT().BeforePreFilter(gomock.Any(), gomock.Any(), testPod).Return(nil, framework.NewStatus(framework.Success))
				se.EXPECT().PreFilter(gomock.Any(), gomock.Any(), testPod).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreFilterResult("namespace", "pod", "name", "error", &framework.PreFilterResult{NodeNames: sets.New("hoge")})
				extender.EXPECT().AfterPreFilter(gomock.Any(), gomock.Any(), testPod, &framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Unschedulable, "error")).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge2")}, framework.NewStatus(framework.Success))
			},
			want:  &framework.PreFilterResult{NodeNames: sets.New("hoge2")},
			want1: framework.NewStatus(framework.Success),
		},
		{
			name: "happy without extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreFilterPlugin, extender *mock_plugin.MockPreFilterPluginExtender) {
				se.EXPECT().PreFilter(gomock.Any(), gomock.Any(), testPod).Return(&framework.PreFilterResult{NodeNames: sets.New("hoge")}, framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreFilterResult("namespace", "pod", "name", resultstore.SuccessMessage, &framework.PreFilterResult{NodeNames: sets.New("hoge")})
			},
			noExtender: true,
			want:       &framework.PreFilterResult{NodeNames: sets.New("hoge")},
			want1:      framework.NewStatus(framework.Success),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockPreFilterPlugin(ctrl)
			ex := mock_plugin.NewMockPreFilterPluginExtender(ctrl)
			tt.prepareMocksFn(s, p, ex)

			w := &wrappedPlugin{
				store:                   s,
				originalPreFilterPlugin: p,
			}

			if !tt.noExtender {
				w.preFilterPluginExtender = ex
			}

			got, got1 := w.PreFilter(context.Background(), framework.NewCycleState(), testPod)
			assert.Equalf(t, tt.want, got, "PreFilter(ctx, cyclestate, %v)", testPod)
			assert.Equalf(t, tt.want1, got1, "PreFilter(ctx, cyclestate, %v)", testPod)
		})
	}
}

func Test_wrappedPlugin_Permit(t *testing.T) {
	t.Parallel()
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "namespace"}}
	testNodeName := "node"

	tests := []struct {
		name           string
		prepareMocksFn func(s *mock_plugin.MockStore, se *mock_plugin.MockPermitPlugin, extender *mock_plugin.MockPermitPluginExtender)
		noExtender     bool
		want           *framework.Status
		want1          time.Duration
	}{
		{
			name: "happy with extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPermitPlugin, extender *mock_plugin.MockPermitPluginExtender) {
				extender.EXPECT().BeforePermit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success), time.Duration(1))
				se.EXPECT().Permit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success), time.Duration(1))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPermitResult("namespace", "pod", "name", resultstore.SuccessMessage, time.Duration(1))
				extender.EXPECT().AfterPermit(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Success), time.Duration(1)).Return(framework.NewStatus(framework.Success), time.Duration(1))
			},
			want:  framework.NewStatus(framework.Success),
			want1: time.Duration(1),
		},
		{
			name: "unhappy: BeforePermit returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPermitPlugin, extender *mock_plugin.MockPermitPluginExtender) {
				extender.EXPECT().BeforePermit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable), time.Duration(1))
			},
			want:  framework.NewStatus(framework.Unschedulable),
			want1: time.Duration(1),
		},
		{
			name: "unhappy: AfterPermit returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPermitPlugin, extender *mock_plugin.MockPermitPluginExtender) {
				extender.EXPECT().BeforePermit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success), time.Duration(1))
				se.EXPECT().Permit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success), time.Duration(1))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPermitResult("namespace", "pod", "name", resultstore.SuccessMessage, time.Duration(1))
				extender.EXPECT().AfterPermit(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Success), time.Duration(1)).Return(framework.NewStatus(framework.Unschedulable), time.Duration(2))
			},
			want:  framework.NewStatus(framework.Unschedulable),
			want1: time.Duration(2),
		},
		{
			name: "unhappy: Permit and AfterPermit return non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPermitPlugin, extender *mock_plugin.MockPermitPluginExtender) {
				extender.EXPECT().BeforePermit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success), time.Duration(1))
				se.EXPECT().Permit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable, "error"), time.Duration(1))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPermitResult("namespace", "pod", "name", "error", time.Duration(1))
				extender.EXPECT().AfterPermit(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Unschedulable, "error"), time.Duration(1)).Return(framework.NewStatus(framework.Unschedulable), time.Duration(2))
			},
			want:  framework.NewStatus(framework.Unschedulable),
			want1: time.Duration(2),
		},
		{
			name: "happy: Permit returns non-success, but AfterPermit return success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPermitPlugin, extender *mock_plugin.MockPermitPluginExtender) {
				extender.EXPECT().BeforePermit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success), time.Duration(1))
				se.EXPECT().Permit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable, "error"), time.Duration(1))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPermitResult("namespace", "pod", "name", "error", time.Duration(1))
				extender.EXPECT().AfterPermit(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Unschedulable, "error"), time.Duration(1)).Return(framework.NewStatus(framework.Success), time.Duration(2))
			},
			want:  framework.NewStatus(framework.Success),
			want1: time.Duration(2),
		},
		{
			name: "happy without extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPermitPlugin, extender *mock_plugin.MockPermitPluginExtender) {
				se.EXPECT().Permit(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success), time.Duration(1))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPermitResult("namespace", "pod", "name", resultstore.SuccessMessage, time.Duration(1))
			},
			noExtender: true,
			want:       framework.NewStatus(framework.Success),
			want1:      time.Duration(1),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockPermitPlugin(ctrl)
			ex := mock_plugin.NewMockPermitPluginExtender(ctrl)
			tt.prepareMocksFn(s, p, ex)

			w := &wrappedPlugin{
				store:                s,
				originalPermitPlugin: p,
			}
			if !tt.noExtender {
				w.permitPluginExtender = ex
			}
			got, got1 := w.Permit(context.Background(), framework.NewCycleState(), testPod, testNodeName)
			assert.Equalf(t, tt.want, got, "Permit(ctx, cyclestate, %v, %v)", testPod, testNodeName)
			assert.Equalf(t, tt.want1, got1, "Permit(ctx, cyclestate, %v, %v)", testPod, testNodeName)
		})
	}
}

func Test_wrappedPlugin_Reserve(t *testing.T) {
	t.Parallel()
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "namespace"}}
	testNodeName := "node"

	tests := []struct {
		name           string
		prepareMocksFn func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender)
		want           *framework.Status
		noExtender     bool
	}{
		{
			name: "happy with extnder",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				s.EXPECT().AddSelectedNode("namespace", "pod", "node")
				extender.EXPECT().BeforeReserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Reserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddReserveResult("namespace", "pod", "name", resultstore.SuccessMessage)
				extender.EXPECT().AfterReserve(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Success)).Return(framework.NewStatus(framework.Success))
			},
			want: framework.NewStatus(framework.Success),
		},
		{
			name: "unhappy: BeforeReserve returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				s.EXPECT().AddSelectedNode("namespace", "pod", "node")
				extender.EXPECT().BeforeReserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: AfterReserve returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				s.EXPECT().AddSelectedNode("namespace", "pod", "node")
				extender.EXPECT().BeforeReserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Reserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddReserveResult("namespace", "pod", "name", resultstore.SuccessMessage)
				extender.EXPECT().AfterReserve(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Success)).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: Reserve and AfterReserve return non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				s.EXPECT().AddSelectedNode("namespace", "pod", "node")
				extender.EXPECT().BeforeReserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Reserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddReserveResult("namespace", "pod", "name", "error")
				extender.EXPECT().AfterReserve(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Unschedulable, "error")).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "happy: Reserve returns non-success, but AfterReserve return success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				s.EXPECT().AddSelectedNode("namespace", "pod", "node")
				extender.EXPECT().BeforeReserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Reserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddReserveResult("namespace", "pod", "name", "error")
				extender.EXPECT().AfterReserve(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Unschedulable, "error")).Return(framework.NewStatus(framework.Success))
			},
			want: framework.NewStatus(framework.Success),
		},
		{
			name: "happy without extnder",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				s.EXPECT().AddSelectedNode("namespace", "pod", "node")
				se.EXPECT().Reserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddReserveResult("namespace", "pod", "name", resultstore.SuccessMessage)
			},
			noExtender: true,
			want:       framework.NewStatus(framework.Success),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockReservePlugin(ctrl)
			ex := mock_plugin.NewMockReservePluginExtender(ctrl)
			tt.prepareMocksFn(s, p, ex)

			w := &wrappedPlugin{
				store:                 s,
				originalReservePlugin: p,
			}
			if !tt.noExtender {
				w.reservePluginExtender = ex
			}
			assert.Equalf(t, tt.want, w.Reserve(context.Background(), framework.NewCycleState(), testPod, testNodeName), "Reserve(ctx, cyclestate, %v, %v)", testPod, testNodeName)
		})
	}
}

func Test_wrappedPlugin_Unreserve(t *testing.T) {
	t.Parallel()
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "namespace"}}
	testNodeName := "node"

	tests := []struct {
		name           string
		prepareMocksFn func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender)
		noExtender     bool
	}{
		{
			name: "happy with extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				extender.EXPECT().BeforeUnreserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Unreserve(gomock.Any(), gomock.Any(), testPod, testNodeName)
				extender.EXPECT().AfterUnreserve(gomock.Any(), gomock.Any(), testPod, testNodeName)
			},
		},
		{
			name: "unhappy: BeforeUnreserve returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				extender.EXPECT().BeforeUnreserve(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable))
				se.EXPECT().Name().Return("hoge")
			},
		},
		{
			name: "happy without extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockReservePlugin, extender *mock_plugin.MockReservePluginExtender) {
				se.EXPECT().Unreserve(gomock.Any(), gomock.Any(), testPod, testNodeName)
			},
			noExtender: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockReservePlugin(ctrl)
			ex := mock_plugin.NewMockReservePluginExtender(ctrl)
			tt.prepareMocksFn(s, p, ex)

			w := &wrappedPlugin{
				store:                 s,
				originalReservePlugin: p,
			}
			if !tt.noExtender {
				w.reservePluginExtender = ex
			}
			w.Unreserve(context.Background(), framework.NewCycleState(), testPod, testNodeName)
		})
	}
}

func Test_wrappedPlugin_PreBind(t *testing.T) {
	t.Parallel()
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "namespace"}}
	testNodeName := "node"

	tests := []struct {
		name           string
		prepareMocksFn func(s *mock_plugin.MockStore, se *mock_plugin.MockPreBindPlugin, extender *mock_plugin.MockPreBindPluginExtender)
		want           *framework.Status
		noExtender     bool
	}{
		{
			name: "happy with extnder",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreBindPlugin, extender *mock_plugin.MockPreBindPluginExtender) {
				extender.EXPECT().BeforePreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreBindResult("namespace", "pod", "name", resultstore.SuccessMessage)
				extender.EXPECT().AfterPreBind(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Success)).Return(framework.NewStatus(framework.Success))
			},
			want: framework.NewStatus(framework.Success),
		},
		{
			name: "unhappy: BeforePreBind returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreBindPlugin, extender *mock_plugin.MockPreBindPluginExtender) {
				extender.EXPECT().BeforePreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: AfterPreBind returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreBindPlugin, extender *mock_plugin.MockPreBindPluginExtender) {
				extender.EXPECT().BeforePreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreBindResult("namespace", "pod", "name", resultstore.SuccessMessage)
				extender.EXPECT().AfterPreBind(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Success)).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: PreBind and AfterPreBind return non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreBindPlugin, extender *mock_plugin.MockPreBindPluginExtender) {
				extender.EXPECT().BeforePreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreBindResult("namespace", "pod", "name", "error")
				extender.EXPECT().AfterPreBind(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Unschedulable, "error")).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "happy: PreBind returns non-success, but AfterPreBind return success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreBindPlugin, extender *mock_plugin.MockPreBindPluginExtender) {
				extender.EXPECT().BeforePreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreBindResult("namespace", "pod", "name", "error")
				extender.EXPECT().AfterPreBind(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Unschedulable, "error")).Return(framework.NewStatus(framework.Success))
			},
			want: framework.NewStatus(framework.Success),
		},
		{
			name: "happy without extnder",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPreBindPlugin, extender *mock_plugin.MockPreBindPluginExtender) {
				se.EXPECT().PreBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddPreBindResult("namespace", "pod", "name", resultstore.SuccessMessage)
			},
			noExtender: true,
			want:       framework.NewStatus(framework.Success),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockPreBindPlugin(ctrl)
			ex := mock_plugin.NewMockPreBindPluginExtender(ctrl)
			tt.prepareMocksFn(s, p, ex)

			w := &wrappedPlugin{
				store:                 s,
				originalPreBindPlugin: p,
			}
			if !tt.noExtender {
				w.preBindPluginExtender = ex
			}
			assert.Equalf(t, tt.want, w.PreBind(context.Background(), framework.NewCycleState(), testPod, testNodeName), "PreBind(ctx, cyclestate, %v, %v)", testPod, testNodeName)
		})
	}
}

func Test_wrappedPlugin_Bind(t *testing.T) {
	t.Parallel()
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "namespace"}}
	testNodeName := "node"

	tests := []struct {
		name           string
		prepareMocksFn func(s *mock_plugin.MockStore, se *mock_plugin.MockBindPlugin, extender *mock_plugin.MockBindPluginExtender)
		want           *framework.Status
		noExtender     bool
	}{
		{
			name: "happy with extnder",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockBindPlugin, extender *mock_plugin.MockBindPluginExtender) {
				extender.EXPECT().BeforeBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Bind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddBindResult("namespace", "pod", "name", resultstore.SuccessMessage)
				extender.EXPECT().AfterBind(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Success)).Return(framework.NewStatus(framework.Success))
			},
			want: framework.NewStatus(framework.Success),
		},
		{
			name: "unhappy: BeforeBind returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockBindPlugin, extender *mock_plugin.MockBindPluginExtender) {
				extender.EXPECT().BeforeBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: AfterBind returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockBindPlugin, extender *mock_plugin.MockBindPluginExtender) {
				extender.EXPECT().BeforeBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Bind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddBindResult("namespace", "pod", "name", resultstore.SuccessMessage)
				extender.EXPECT().AfterBind(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Success)).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "unhappy: Bind and AfterBind return non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockBindPlugin, extender *mock_plugin.MockBindPluginExtender) {
				extender.EXPECT().BeforeBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Bind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddBindResult("namespace", "pod", "name", "error")
				extender.EXPECT().AfterBind(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Unschedulable, "error")).Return(framework.NewStatus(framework.Unschedulable))
			},
			want: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "happy: Bind returns non-success, but AfterBind return success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockBindPlugin, extender *mock_plugin.MockBindPluginExtender) {
				extender.EXPECT().BeforeBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Bind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable, "error"))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddBindResult("namespace", "pod", "name", "error")
				extender.EXPECT().AfterBind(gomock.Any(), gomock.Any(), testPod, testNodeName, framework.NewStatus(framework.Unschedulable, "error")).Return(framework.NewStatus(framework.Success))
			},
			want: framework.NewStatus(framework.Success),
		},
		{
			name: "happy without extnder",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockBindPlugin, extender *mock_plugin.MockBindPluginExtender) {
				se.EXPECT().Bind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().Name().Return("name")
				s.EXPECT().AddBindResult("namespace", "pod", "name", resultstore.SuccessMessage)
			},
			noExtender: true,
			want:       framework.NewStatus(framework.Success),
		},
		{
			name:       "unhappy: it is not bind plugin",
			noExtender: true,
			want:       framework.NewStatus(framework.Skip, "called wrapped bind plugin is nil"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			w := &wrappedPlugin{}
			if tt.prepareMocksFn != nil {
				s := mock_plugin.NewMockStore(ctrl)
				p := mock_plugin.NewMockBindPlugin(ctrl)
				ex := mock_plugin.NewMockBindPluginExtender(ctrl)
				tt.prepareMocksFn(s, p, ex)
				w = &wrappedPlugin{
					store:              s,
					originalBindPlugin: p,
				}
				if !tt.noExtender {
					w.bindPluginExtender = ex
				}
			}

			assert.Equalf(t, tt.want, w.Bind(context.Background(), framework.NewCycleState(), testPod, testNodeName), "Bind(ctx, cyclestate, %v, %v)", testPod, testNodeName)
		})
	}
}

func Test_wrappedPlugin_PostBind(t *testing.T) {
	t.Parallel()
	testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "namespace"}}
	testNodeName := "node"

	tests := []struct {
		name           string
		prepareMocksFn func(s *mock_plugin.MockStore, se *mock_plugin.MockPostBindPlugin, extender *mock_plugin.MockPostBindPluginExtender)
		noExtender     bool
	}{
		{
			name: "happy with extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPostBindPlugin, extender *mock_plugin.MockPostBindPluginExtender) {
				extender.EXPECT().BeforePostBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Success))
				se.EXPECT().PostBind(gomock.Any(), gomock.Any(), testPod, testNodeName)
				extender.EXPECT().AfterPostBind(gomock.Any(), gomock.Any(), testPod, testNodeName)
			},
		},
		{
			name: "unhappy: BeforePostBind returns non-success",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPostBindPlugin, extender *mock_plugin.MockPostBindPluginExtender) {
				extender.EXPECT().BeforePostBind(gomock.Any(), gomock.Any(), testPod, testNodeName).Return(framework.NewStatus(framework.Unschedulable))
				se.EXPECT().Name().Return("hoge")
			},
		},
		{
			name: "happy without extender",
			prepareMocksFn: func(s *mock_plugin.MockStore, se *mock_plugin.MockPostBindPlugin, extender *mock_plugin.MockPostBindPluginExtender) {
				se.EXPECT().PostBind(gomock.Any(), gomock.Any(), testPod, testNodeName)
			},
			noExtender: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := mock_plugin.NewMockStore(ctrl)
			p := mock_plugin.NewMockPostBindPlugin(ctrl)
			ex := mock_plugin.NewMockPostBindPluginExtender(ctrl)
			tt.prepareMocksFn(s, p, ex)

			w := &wrappedPlugin{
				store:                  s,
				originalPostBindPlugin: p,
			}
			if !tt.noExtender {
				w.postBindPluginExtender = ex
			}
			w.PostBind(context.Background(), framework.NewCycleState(), testPod, testNodeName)
		})
	}
}

// fake plugins for test

type fakeFilterPlugin struct{}

func (fakeFilterPlugin) Name() string { return "fakeFilterPlugin" }
func (fakeFilterPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ *framework.NodeInfo) *framework.Status {
	return nil
}

type fakePostFilterPlugin struct{}

func (fakePostFilterPlugin) Name() string { return "fakePostFilterPlugin" }
func (fakePostFilterPlugin) PostFilter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	return &framework.PostFilterResult{NominatingInfo: &framework.NominatingInfo{NominatedNodeName: "node1"}}, framework.NewStatus(framework.Success, "postfilter success")
}

type fakeScorePlugin struct{}

func (fakeScorePlugin) Name() string { return "fakeScorePlugin" }
func (pl fakeScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func (fakeScorePlugin) NormalizeScore(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ framework.NodeScoreList) *framework.Status {
	return nil
}

func (fakeScorePlugin) Score(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ string) (int64, *framework.Status) {
	return 1, nil
}

type fakeWrappedPlugin struct{}

func (fakeWrappedPlugin) Name() string { return "fakeWrappedPlugin" }
func (fakeWrappedPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ *framework.NodeInfo) *framework.Status {
	return nil
}

func (fakeWrappedPlugin) PostFilter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	return new(framework.PostFilterResult), nil
}

func (pl fakeWrappedPlugin) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func (fakeWrappedPlugin) NormalizeScore(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ framework.NodeScoreList) *framework.Status {
	return nil
}

func (fakeWrappedPlugin) Score(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ string) (int64, *framework.Status) {
	return 0, nil
}

// all method on this plugin will fail.
type fakeMustFailWrappedPlugin struct{}

func (fakeMustFailWrappedPlugin) Name() string { return "fakeMustFailWrappedPlugin" }
func (fakeMustFailWrappedPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ *framework.NodeInfo) *framework.Status {
	return framework.AsStatus(errors.New("filter failed"))
}

func (fakeMustFailWrappedPlugin) PostFilter(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	return nil, framework.AsStatus(errors.New("postFilter failed"))
}

func (pl fakeMustFailWrappedPlugin) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func (fakeMustFailWrappedPlugin) NormalizeScore(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ framework.NodeScoreList) *framework.Status {
	return framework.AsStatus(errors.New("normalize failed"))
}

func (fakeMustFailWrappedPlugin) Score(_ context.Context, _ *framework.CycleState, _ *v1.Pod, _ string) (int64, *framework.Status) {
	return 0, framework.AsStatus(errors.New("score failed"))
}
