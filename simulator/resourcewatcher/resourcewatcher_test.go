package resourcewatcher

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	restfake "k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/cache"

	sw "sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter/mock_streamwriter"
)

func TestEventProxyer_createWatcher(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		prepareFakeRestClientFn func() *restfake.RESTClient
		resourceversion         string
		wantErr                 bool
	}{
		{
			name: "should success",
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			resourceversion: "1",
			wantErr:         false,
		},
		{
			name: "should returns an error when NewRetryWatcher is failed(when resourceversion is 0)",
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			resourceversion: "0",
			wantErr:         true,
		},
		{
			name: "should returns an error when NewRetryWatcher is failed(when resourceversion is empty)",
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			resourceversion: "",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			restclient := tt.prepareFakeRestClientFn()
			mockResponseStream := mock_streamwriter.NewMockResponseStream(ctrl)

			sw := sw.NewStreamWriter(mockResponseStream)
			proxy := neweventProxy(sw, restclient, Pods, &corev1.Pod{}, tt.resourceversion)

			lw := createListWatch(proxy)
			_, err := createWatcher(proxy, lw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("createWatcher %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}

func TestService_doListAndWatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                      string
		prepareFakeClientSetFn    func() *fake.Clientset
		prepareFakeRestClientFn   func() *restfake.RESTClient
		prepareeventProxyerMockFn func(p *MockeventProxyer, getter cache.Getter)
		wantErr                   bool
	}{
		{
			name: "should call watchAndHandleEvent method",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			prepareeventProxyerMockFn: func(p *MockeventProxyer, getter cache.Getter) {
				p.EXPECT().watchAndHandleEvent(gomock.Any(), gomock.Any())
				p.EXPECT().lastResourceVersion().Return("1").Times(2)
				p.EXPECT().restClient().Return(getter)
				p.EXPECT().resourceKind().Return(Pods)
			},
			wantErr: false,
		},
		{
			name: "should return an error when the lastResourceVersion is 0",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			prepareeventProxyerMockFn: func(p *MockeventProxyer, getter cache.Getter) {
				p.EXPECT().watchAndHandleEvent(gomock.Any(), gomock.Any())
				p.EXPECT().lastResourceVersion().Return("0").Times(2)
				p.EXPECT().restClient().Return(getter)
				p.EXPECT().resourceKind().Return(Pods).Times(2)
			},
			wantErr: true,
		},
		{
			name: "should call listAndHandleItems method when the lastResourceVersion is empty",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			prepareeventProxyerMockFn: func(p *MockeventProxyer, getter cache.Getter) {
				p.EXPECT().watchAndHandleEvent(gomock.Any(), gomock.Any())
				p.EXPECT().lastResourceVersion().Do(func() {
					p.EXPECT().lastResourceVersion().Return("1")
				}).Return("")
				p.EXPECT().restClient().Return(getter)
				p.EXPECT().resourceKind().Return(Pods).Times(3)
				p.EXPECT().listAndHandleItems(gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "should return an error when listAndHandleItems method return an error",
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			prepareeventProxyerMockFn: func(p *MockeventProxyer, getter cache.Getter) {
				p.EXPECT().watchAndHandleEvent(gomock.Any(), gomock.Any())
				p.EXPECT().lastResourceVersion().Return("")
				p.EXPECT().restClient().Return(getter)
				p.EXPECT().resourceKind().Return(Pods).Times(2)
				p.EXPECT().listAndHandleItems(gomock.Any()).Return(xerrors.Errorf("failed"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockProxy := NewMockeventProxyer(ctrl)
			fakeClientSet := tt.prepareFakeClientSetFn()
			s := NewService(fakeClientSet)
			fakeRestClient := tt.prepareFakeRestClientFn()
			tt.prepareeventProxyerMockFn(mockProxy, fakeRestClient)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := s.doListAndWatch(mockProxy, ctx.Done()); (err != nil) != tt.wantErr {
				t.Fatalf("doListAndWatch %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}
