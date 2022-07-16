package resourcewatcher

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	restfake "k8s.io/client-go/rest/fake"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/mock_resourcewatcher"
	sw "sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter/mock_streamwriter"
)

var (
	fakenode1 = &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "node1",
			ResourceVersion: "100",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Node",
		},
	}
	fakenode2 = &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "node2",
			ResourceVersion: "200",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Node",
		},
	}
	fakenode3 = &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "node3",
			ResourceVersion: "300",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Node",
		},
	}
)

type fakePod struct{}

func (obj *fakePod) GetObjectKind() schema.ObjectKind { return schema.EmptyObjectKind }
func (obj *fakePod) DeepCopyObject() runtime.Object   { panic("DeepCopyObject not supported by fakePod") }

func TestResourceEventProxy_CreateWatcher(t *testing.T) {
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
			proxy := newresourceEventProxy(sw, restclient, Pods, &corev1.Pod{}, tt.resourceversion)

			_, err := createWatcher(proxy)
			if (err != nil) != tt.wantErr {
				t.Fatalf("createWatcher %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}

func TestResourceEventProxy_WatchHandlerFunc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                      string
		prepareStreamWriterMockFn func(sw *mock_resourcewatcher.MockStreamWriter)
		prepareFakeClientSetFn    func() *fake.Clientset
		doEvent                   func(fw *watch.FakeWatcher)
		wantErr                   bool
		wantLastResourceVersion   string
	}{
		{
			name: "should call the Write method (with ADDED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Added, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "100",
		},
		{
			name: "should call the Write method (with twice ADDED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Added, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Added, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node2", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
					fw.Add(fakenode2)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "200",
		},
		{
			name: "should call the Write method (with MODIFIED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Modified, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Modify(fakenode1)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "100",
		},
		{
			name: "should call the Write method (with twice MODIFIED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Modified, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Modified, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node2", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Modify(fakenode1)
					fw.Modify(fakenode2)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "200",
		},
		{
			name: "should call the Write method (with DELETED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Deleted, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Delete(fakenode1)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "100",
		},
		{
			name: "should call the Write method (with twice DELETED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Deleted, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Deleted, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node2", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Delete(fakenode1)
					fw.Delete(fakenode2)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "200",
		},
		{
			name: "should call the Write method (with ADDED and MODIFIED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Added, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Modified, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node2", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
					fw.Modify(fakenode2)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "200",
		},
		{
			name: "should call the Write method (with ADDED and twice MODIFIED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Added, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Modified, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node2", obj.GetName())
				})
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Modified, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node3", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
					fw.Modify(fakenode2)
					fw.Modify(fakenode3)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "300",
		},
		{
			name: "should call the Write method (with ADDED, MODIFIED and DELETED event)",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Added, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node1", obj.GetName())
				})
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Modified, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node2", obj.GetName())
				})
				w.EXPECT().Write(gomock.Any()).Return(nil).Do(func(e *sw.WatchEvent) {
					assert.Equal(t, Nodes, e.Kind)
					assert.Equal(t, watch.Deleted, e.EventType)
					obj, ok := e.Obj.(metav1.Object)
					assert.True(t, ok)
					assert.Equal(t, "node3", obj.GetName())
				})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
					fw.Modify(fakenode2)
					fw.Delete(fakenode3)
				}()
			},
			wantErr:                 false,
			wantLastResourceVersion: "300",
		},
		{
			name: "should return an error if the passed object is failed to cast to a metav1.Object",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					var obj *fakePod
					fw.Add(obj)
				}()
			},
			wantErr:                 true,
			wantLastResourceVersion: "1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			fakeclientset := tt.prepareFakeClientSetFn()
			mockStreamWriter := mock_resourcewatcher.NewMockStreamWriter(ctrl)
			tt.prepareStreamWriterMockFn(mockStreamWriter)
			fw := watch.NewFake()

			proxy := newresourceEventProxy(mockStreamWriter, fakeclientset.CoreV1().RESTClient(), Nodes, &corev1.Node{}, "1")

			testFunc := proxy.WatchHandlerFunc(fw)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tt.doEvent(fw)
			err := testFunc(ctx.Done())

			if (err != nil) != tt.wantErr {
				t.Fatalf("watchHandlerFunc %v test, \nerror = %v", tt.name, err)
			}
			v := proxy.LastResourceVersion()
			if v != tt.wantLastResourceVersion {
				t.Fatalf("watchHandlerFunc %v test, \nlastResourceVersion = %v, want = %v", tt.name, v, tt.wantLastResourceVersion)
			}
		})
	}
}

func TestResourceEventProxy_WatchHandlerFuncFails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                        string
		prepareWatchInterfaceMockFn func(rs *mock_resourcewatcher.MockWatchInterface)
		prepareFakeClientSetFn      func() *fake.Clientset
		wantErr                     bool
	}{
		{
			name: "should return an error if the channel of ResultChan is closed",
			prepareWatchInterfaceMockFn: func(w *mock_resourcewatcher.MockWatchInterface) {
				ch := make(chan watch.Event)
				close(ch)
				w.EXPECT().ResultChan().Return(ch)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			fakeclientset := tt.prepareFakeClientSetFn()
			mockStreamWriter := mock_resourcewatcher.NewMockStreamWriter(ctrl)
			mockWatcher := mock_resourcewatcher.NewMockWatchInterface(ctrl)
			tt.prepareWatchInterfaceMockFn(mockWatcher)

			proxy := newresourceEventProxy(mockStreamWriter, fakeclientset.CoreV1().RESTClient(), Nodes, &corev1.Node{}, "1")

			testFunc := proxy.WatchHandlerFunc(mockWatcher)

			ctx := context.Background()
			err := testFunc(ctx.Done())

			if (err != nil) != tt.wantErr {
				t.Fatalf("watchHandlerFunc %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}

func TestService_WatchAndHandleEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                            string
		prepareResourceEventProxyMockFn func(p *mock_resourcewatcher.MockResourceEventProxy)
		prepareFakeClientSetFn          func() *fake.Clientset
		wantErr                         bool
	}{
		{
			name: "should call WatchHandlerFunc",
			prepareResourceEventProxyMockFn: func(p *mock_resourcewatcher.MockResourceEventProxy) {
				p.EXPECT().WatchHandlerFunc(gomock.Any()).Return(func(stopCh <-chan struct{}) error { return nil })
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
		},
		{
			name: "should call WatchHandlerFunc and WatchErrorHandler if the WatchHandlerFunc's function returns an error",
			prepareResourceEventProxyMockFn: func(p *mock_resourcewatcher.MockResourceEventProxy) {
				expectedErr := xerrors.Errorf("closed channel or something wrong")
				p.EXPECT().WatchHandlerFunc(gomock.Any()).Return(func(stopCh <-chan struct{}) error { return expectedErr })
				p.EXPECT().WatchErrorHandler(expectedErr)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			fakeclientset := tt.prepareFakeClientSetFn()
			mockProxy := mock_resourcewatcher.NewMockResourceEventProxy(ctrl)
			tt.prepareResourceEventProxyMockFn(mockProxy)
			fw := watch.NewFake()

			s := NewService(fakeclientset)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			s.WatchAndHandleEvent(mockProxy, fw, ctx.Done())
		})
	}
}
