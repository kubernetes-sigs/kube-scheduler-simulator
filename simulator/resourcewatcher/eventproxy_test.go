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
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restfake "k8s.io/client-go/rest/fake"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/mock_resourcewatcher"
	sw "sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter"
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

func TestService_listAndHandleItems(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                      string
		prepareFakeClientSetFn    func() *fake.Clientset
		prepareListerMockFn       func(l *mock_resourcewatcher.MockLister, nodes typedcorev1.NodeInterface)
		prepareStreamWriterMockFn func(w *mock_resourcewatcher.MockStreamWriter)
		prepareFakeRestClientFn   func() *restfake.RESTClient
		wantErr                   bool
		wantlastResourceVersion   string
	}{
		{
			name: "should list the resource and update the lastResourceVersion",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			prepareListerMockFn: func(l *mock_resourcewatcher.MockLister, nodes typedcorev1.NodeInterface) {
				list, _ := nodes.List(context.Background(), metav1.ListOptions{})
				list.ResourceVersion = "100"
				l.EXPECT().List(metav1.ListOptions{}).Return(list, nil)
			},
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil)
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			wantErr:                 false,
			wantlastResourceVersion: "100",
		},
		{
			name: "should return an error when HandleListItems return an error",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			prepareListerMockFn: func(l *mock_resourcewatcher.MockLister, nodes typedcorev1.NodeInterface) {
				get, _ := nodes.Get(context.Background(), "node1", metav1.GetOptions{})
				l.EXPECT().List(metav1.ListOptions{}).Return(get, nil)
			},
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			wantErr:                 true,
			wantlastResourceVersion: "1",
		},
		{
			name: "should returns an error and shouldn't changes the lastResourceVersion  when sendListedItems return an error",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			prepareListerMockFn: func(l *mock_resourcewatcher.MockLister, nodes typedcorev1.NodeInterface) {
				list, _ := nodes.List(context.Background(), metav1.ListOptions{})
				list.ResourceVersion = "100"
				l.EXPECT().List(metav1.ListOptions{}).Return(list, nil)
			},
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(xerrors.Errorf("failed"))
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			wantErr:                 true,
			wantlastResourceVersion: "1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			fakeclientset := tt.prepareFakeClientSetFn()
			lister := mock_resourcewatcher.NewMockLister(ctrl)
			tt.prepareListerMockFn(lister, fakeclientset.CoreV1().Nodes())
			sw := mock_resourcewatcher.NewMockStreamWriter(ctrl)
			tt.prepareStreamWriterMockFn(sw)
			fakeRestClient := tt.prepareFakeRestClientFn()
			proxy := neweventProxy(sw, fakeRestClient, Nodes, &corev1.Node{}, "1")

			if err := proxy.listAndHandleItems(lister); (err != nil) != tt.wantErr {
				t.Fatalf("listAndHandleItems %v test, \nerror = %v", tt.name, err)
			}
			v := proxy.lastResourceVersion()
			if v != tt.wantlastResourceVersion {
				t.Fatalf("listAndHandleItems %v test, \nlastResourceVersion = %s, want = %v", tt.name, v, tt.wantlastResourceVersion)
			}
		})
	}
}

func TestEventProxyer_sendListedItems(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                      string
		prepareeventProxyerMockFn func(p *MockeventProxyer)
		prepareFakeClientSetFn    func() *fake.Clientset
		prepareStreamWriterMockFn func(w *mock_resourcewatcher.MockStreamWriter)
		prepareFakeRestClientFn   func() *restfake.RESTClient
		prepareItems              func() []runtime.Object
		wantErr                   bool
	}{
		{
			name: "should call Write method",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil)
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			prepareItems: func() []runtime.Object {
				return []runtime.Object{
					fakenode1,
				}
			},
			wantErr: false,
		},
		{
			name: "should call Write method twice",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(nil).Times(2)
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			prepareItems: func() []runtime.Object {
				return []runtime.Object{
					fakenode1,
					fakenode2,
				}
			},
			wantErr: false,
		},
		{
			name: "should return an error when the Write method returns an error",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
				w.EXPECT().Write(gomock.Any()).Return(xerrors.Errorf("failed"))
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			prepareItems: func() []runtime.Object {
				return []runtime.Object{
					fakenode1,
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			sw := mock_resourcewatcher.NewMockStreamWriter(ctrl)
			tt.prepareStreamWriterMockFn(sw)
			fakeRestClient := tt.prepareFakeRestClientFn()
			proxy := neweventProxy(sw, fakeRestClient, Nodes, &corev1.Node{}, "1")
			items := tt.prepareItems()
			if err := proxy.sendListedItems(items); (err != nil) != tt.wantErr {
				t.Fatalf("listAndHandleItems %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}

func TestEventProxyer_watchHandlerFunc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                      string
		prepareStreamWriterMockFn func(sw *mock_resourcewatcher.MockStreamWriter)
		prepareFakeRestClientFn   func() *restfake.RESTClient
		doEvent                   func(fw *watch.FakeWatcher)
		wantErr                   bool
		wantlastResourceVersion   string
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "100",
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
					fw.Add(fakenode2)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "200",
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Modify(fakenode1)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "100",
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Modify(fakenode1)
					fw.Modify(fakenode2)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "200",
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Delete(fakenode1)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "100",
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Delete(fakenode1)
					fw.Delete(fakenode2)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "200",
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
					fw.Modify(fakenode2)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "200",
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
					fw.Modify(fakenode2)
					fw.Modify(fakenode3)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "300",
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
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					fw.Add(fakenode1)
					fw.Modify(fakenode2)
					fw.Delete(fakenode3)
				}()
			},
			wantErr:                 false,
			wantlastResourceVersion: "300",
		},
		{
			name: "should return an error if the passed object is failed to cast to a metav1.Object",
			prepareStreamWriterMockFn: func(w *mock_resourcewatcher.MockStreamWriter) {
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			doEvent: func(fw *watch.FakeWatcher) {
				go func() {
					var obj *fakePod
					fw.Add(obj)
				}()
			},
			wantErr:                 true,
			wantlastResourceVersion: "1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			fakeRestClient := tt.prepareFakeRestClientFn()
			mockStreamWriter := mock_resourcewatcher.NewMockStreamWriter(ctrl)
			tt.prepareStreamWriterMockFn(mockStreamWriter)
			fw := watch.NewFake()

			proxy := neweventProxy(mockStreamWriter, fakeRestClient, Nodes, &corev1.Node{}, "1")

			testFunc := proxy.watchHandlerFunc(fw)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tt.doEvent(fw)
			err := testFunc(ctx.Done())

			if (err != nil) != tt.wantErr {
				t.Fatalf("watchHandlerFunc %v test, \nerror = %v", tt.name, err)
			}
			v := proxy.lastResourceVersion()
			if v != tt.wantlastResourceVersion {
				t.Fatalf("watchHandlerFunc %v test, \nlastResourceVersion = %v, want = %v", tt.name, v, tt.wantlastResourceVersion)
			}
		})
	}
}

func TestEventProxyer_watchHandlerFuncFails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                        string
		prepareWatchInterfaceMockFn func(rs *mock_resourcewatcher.MockWatchInterface)
		prepareFakeRestClientFn     func() *restfake.RESTClient
		wantErr                     bool
	}{
		{
			name: "should return an error if the channel of ResultChan is closed",
			prepareWatchInterfaceMockFn: func(w *mock_resourcewatcher.MockWatchInterface) {
				ch := make(chan watch.Event)
				close(ch)
				w.EXPECT().ResultChan().Return(ch)
			},
			prepareFakeRestClientFn: func() *restfake.RESTClient {
				return &restfake.RESTClient{}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			fakeRestClient := tt.prepareFakeRestClientFn()
			mockStreamWriter := mock_resourcewatcher.NewMockStreamWriter(ctrl)
			mockWatcher := mock_resourcewatcher.NewMockWatchInterface(ctrl)
			tt.prepareWatchInterfaceMockFn(mockWatcher)

			proxy := neweventProxy(mockStreamWriter, fakeRestClient, Nodes, &corev1.Node{}, "1")

			testFunc := proxy.watchHandlerFunc(mockWatcher)

			ctx := context.Background()
			err := testFunc(ctx.Done())

			if (err != nil) != tt.wantErr {
				t.Fatalf("watchHandlerFunc %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}
