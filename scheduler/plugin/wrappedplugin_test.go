package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	mock_plugin "github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/plugin/mock"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/plugin/resultstore"
)

func Test_NewWrappedPlugin(t *testing.T) {
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
			got := NewWrappedPlugin(tt.args.s, tt.args.p, WithWeightOption(&tt.args.weight))
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
				s.EXPECT().AddFilterResult("default", "pod1", "node1", "BeforefakeFilterPlugin", failure.Message())
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
			e := &Extenders{
				FilterPluginExtender: fe,
			}
			ctx := context.Background()
			tt.prepareEachMockFn(ctx, s, p, fe, tt.args)
			pl, ok := NewWrappedPlugin(s, p, WithExtendersOption(e)).(*wrappedPlugin)
			if !ok { // should never happen
				t.Fatalf("Assert to wrapped plugin: %v", ok)
			}
			gotstatus := pl.Filter(ctx, nil, tt.args.pod, tt.args.nodeInfo)
			assert.Equal(t, tt.wantstatus, gotstatus)
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
			e := &Extenders{
				NormalizeScorePluginExtender: spe,
			}
			ctx := context.Background()
			tt.prepareEachMockFn(ctx, s, se, sp, spe, tt.args)
			pl, ok := NewWrappedPlugin(s, sp, WithExtendersOption(e)).(*wrappedPlugin)
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
				success2 := framework.NewStatus(framework.Error, "Score returned")
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
				s.EXPECT().AddScoreResult("default", "pod1", "node1", "BeforefakeScorePlugin", int64(1111))
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
			e := &Extenders{
				ScorePluginExtender: se,
			}
			ctx := context.Background()
			tt.prepareEachMockFn(ctx, s, p, se, tt.args)
			pl, ok := NewWrappedPlugin(s, p, WithExtendersOption(e)).(*wrappedPlugin)
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
