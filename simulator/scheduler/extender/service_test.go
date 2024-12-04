package extender

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/xerrors"
	"k8s.io/client-go/kubernetes/fake"
	configv1 "k8s.io/kube-scheduler/config/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/extender/mock_extender"
)

func TestService_Filter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareFakeClientSetFn   func() *fake.Clientset
		prepareMockExtenderSetFn func(m *mock_extender.MockExtender)
		prepareMockStoreSetFn    func(m *mock_extender.MockStore)
		wantErr                  bool
	}{
		{
			name: "success",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareMockExtenderSetFn: func(m *mock_extender.MockExtender) {
				m.EXPECT().Filter(extenderv1.ExtenderArgs{}).Return(&extenderv1.ExtenderFilterResult{}, nil)
				m.EXPECT().Name().Return("ext1")
			},
			prepareMockStoreSetFn: func(m *mock_extender.MockStore) {
				m.EXPECT().AddFilterResult(extenderv1.ExtenderArgs{}, gomock.Any(), "ext1")
			},
			wantErr: false,
		},
		{
			name: "return an error if the extender return an error",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareMockExtenderSetFn: func(m *mock_extender.MockExtender) {
				m.EXPECT().Filter(extenderv1.ExtenderArgs{}).Return(nil, xerrors.New("failed"))
			},
			prepareMockStoreSetFn: func(m *mock_extender.MockStore) {
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.prepareFakeClientSetFn()
			ctrl := gomock.NewController(t)
			mStore := mock_extender.NewMockStore(ctrl)
			mExtender := mock_extender.NewMockExtender(ctrl)
			tt.prepareMockStoreSetFn(mStore)
			tt.prepareMockExtenderSetFn(mExtender)

			s := &Service{
				client:    c,
				extenders: []Extender{mExtender},
				store:     mStore,
			}
			args := extenderv1.ExtenderArgs{}
			_, err := s.Filter(0, args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Prioritize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareFakeClientSetFn   func() *fake.Clientset
		prepareMockExtenderSetFn func(m *mock_extender.MockExtender)
		prepareMockStoreSetFn    func(m *mock_extender.MockStore)
		wantErr                  bool
	}{
		{
			name: "success",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareMockExtenderSetFn: func(m *mock_extender.MockExtender) {
				m.EXPECT().Prioritize(extenderv1.ExtenderArgs{}).Return(&extenderv1.HostPriorityList{}, nil)
				m.EXPECT().Name().Return("ext1")
			},
			prepareMockStoreSetFn: func(m *mock_extender.MockStore) {
				m.EXPECT().AddPrioritizeResult(extenderv1.ExtenderArgs{}, gomock.Any(), "ext1")
			},
			wantErr: false,
		},
		{
			name: "return an error if the extender return an error",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareMockExtenderSetFn: func(m *mock_extender.MockExtender) {
				m.EXPECT().Prioritize(extenderv1.ExtenderArgs{}).Return(nil, xerrors.New("failed"))
			},
			prepareMockStoreSetFn: func(m *mock_extender.MockStore) {
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.prepareFakeClientSetFn()
			ctrl := gomock.NewController(t)
			mStore := mock_extender.NewMockStore(ctrl)
			mExtender := mock_extender.NewMockExtender(ctrl)
			tt.prepareMockStoreSetFn(mStore)
			tt.prepareMockExtenderSetFn(mExtender)

			s := &Service{
				client:    c,
				extenders: []Extender{mExtender},
				store:     mStore,
			}
			args := extenderv1.ExtenderArgs{}
			_, err := s.Prioritize(0, args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Preempt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareFakeClientSetFn   func() *fake.Clientset
		prepareMockExtenderSetFn func(m *mock_extender.MockExtender)
		prepareMockStoreSetFn    func(m *mock_extender.MockStore)
		wantErr                  bool
	}{
		{
			name: "success",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareMockExtenderSetFn: func(m *mock_extender.MockExtender) {
				m.EXPECT().Preempt(extenderv1.ExtenderPreemptionArgs{}).Return(&extenderv1.ExtenderPreemptionResult{}, nil)
				m.EXPECT().Name().Return("ext1")
			},
			prepareMockStoreSetFn: func(m *mock_extender.MockStore) {
				m.EXPECT().AddPreemptResult(extenderv1.ExtenderPreemptionArgs{}, gomock.Any(), "ext1")
			},
			wantErr: false,
		},
		{
			name: "return an error if the extender return an error",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareMockExtenderSetFn: func(m *mock_extender.MockExtender) {
				m.EXPECT().Preempt(extenderv1.ExtenderPreemptionArgs{}).Return(nil, xerrors.New("failed"))
			},
			prepareMockStoreSetFn: func(m *mock_extender.MockStore) {
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.prepareFakeClientSetFn()
			ctrl := gomock.NewController(t)
			mStore := mock_extender.NewMockStore(ctrl)
			mExtender := mock_extender.NewMockExtender(ctrl)
			tt.prepareMockStoreSetFn(mStore)
			tt.prepareMockExtenderSetFn(mExtender)

			s := &Service{
				client:    c,
				extenders: []Extender{mExtender},
				store:     mStore,
			}
			args := extenderv1.ExtenderPreemptionArgs{}
			_, err := s.Preempt(0, args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Bind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareFakeClientSetFn   func() *fake.Clientset
		prepareMockExtenderSetFn func(m *mock_extender.MockExtender)
		prepareMockStoreSetFn    func(m *mock_extender.MockStore)
		wantErr                  bool
	}{
		{
			name: "success",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareMockExtenderSetFn: func(m *mock_extender.MockExtender) {
				m.EXPECT().Bind(extenderv1.ExtenderBindingArgs{}).Return(&extenderv1.ExtenderBindingResult{}, nil)
				m.EXPECT().Name().Return("ext1")
			},
			prepareMockStoreSetFn: func(m *mock_extender.MockStore) {
				m.EXPECT().AddBindResult(extenderv1.ExtenderBindingArgs{}, gomock.Any(), "ext1")
			},
			wantErr: false,
		},
		{
			name: "return an error if the extender return an error",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareMockExtenderSetFn: func(m *mock_extender.MockExtender) {
				m.EXPECT().Bind(extenderv1.ExtenderBindingArgs{}).Return(nil, xerrors.New("failed"))
			},
			prepareMockStoreSetFn: func(m *mock_extender.MockStore) {
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.prepareFakeClientSetFn()
			ctrl := gomock.NewController(t)
			mStore := mock_extender.NewMockStore(ctrl)
			mExtender := mock_extender.NewMockExtender(ctrl)
			tt.prepareMockStoreSetFn(mStore)
			tt.prepareMockExtenderSetFn(mExtender)

			s := &Service{
				client:    c,
				extenders: []Extender{mExtender},
				store:     mStore,
			}
			args := extenderv1.ExtenderBindingArgs{}
			_, err := s.Bind(0, args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_OverrideExtendersCfgToSimulator(t *testing.T) {
	t.Parallel()
	target := configv1.KubeSchedulerConfiguration{}
	es := make([]configv1.Extender, 2)
	port := 80
	for i := range es {
		es[i].EnableHTTPS = true
		es[i].TLSConfig = new(configv1.ExtenderTLSConfig)
		es[i].URLPrefix = "http://example.com/"
		es[i].FilterVerb = "f"
		es[i].PrioritizeVerb = "p"
		es[i].PreemptVerb = "pr"
		es[i].BindVerb = "b"
	}
	target.Extenders = es

	OverrideExtendersCfgToSimulator(&target, port)

	// OverrideExtendersCfgToSimulator changes all extender config included in KubeSchedulerConfiguration.
	for i, e := range target.Extenders {
		s := strconv.Itoa(i)
		// Replaced with false.
		assert.Equal(t, false, e.EnableHTTPS)
		// Replaced with nil.
		if e.TLSConfig != nil {
			t.Fatalf("TLSConfig = %v, expected = nil", e.TLSConfig)
		}
		// Rewrite settings for simulator.
		assert.Equal(t, "http://localhost:80/api/v1/extender/", e.URLPrefix)
		assert.Equal(t, "filter/"+s, e.FilterVerb)
		assert.Equal(t, "prioritize/"+s, e.PrioritizeVerb)
		assert.Equal(t, "preempt/"+s, e.PreemptVerb)
		assert.Equal(t, "bind/"+s, e.BindVerb)
	}
}
