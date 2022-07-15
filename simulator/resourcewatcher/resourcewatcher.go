package resourcewatcher

//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/responseStream.go . ResponseStream
//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/watchInterface.go -package=mock_resourcewatcher -mock_names Interface=MockWatchInterface k8s.io/apimachinery/pkg/watch Interface
//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/resourceeventproxy.go . ResourceEventProxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/klog/v2"
)

// resourceKind represents k8s resource name.
type resourceKind string

const (
	Pods  resourceKind = "pods"
	Nodes resourceKind = "nodes"
	Pvs   resourceKind = "persistentvolumes"
	Pvcs  resourceKind = "persistentvolumeclaims"
	Scs   resourceKind = "storageclasses"
	Pcs   resourceKind = "priorityclasses"
)

// WatchEvent represents an event notified by the watched apiserver.
type WatchEvent struct {
	Kind      resourceKind
	EventType watch.EventType
	// Obj is an object included in the event notified by the watched apiserver.
	Obj interface{}
}

// LastResourceVersions includes each resource's lastResourceVersions.
type LastResourceVersions struct {
	Pods  string `json:"pods"`
	Nodes string `json:"nodes"`
	Pvs   string `json:"pvs"`
	Pvcs  string `json:"pvcs"`
	Scs   string `json:"scs"`
	Pcs   string `json:"pcs"`
}

// ResponseStream is an interface that allows Server Push to a Service.
type ResponseStream interface {
	io.Writer
	http.Flusher
}

// Service watches simulator's resources.
type Service struct {
	client clientset.Interface
}

// NewResourceWatcherService initializes Service.
func NewResourceWatcherService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

// WatchResources watches each simulator's resources and send notified events to the frontend continuously.
func (s *Service) WatchResources(ctx context.Context, stream ResponseStream, lrVersions *LastResourceVersions) error {
	sw := newStreamWriter(stream)
	proxies := []*resourceEventProxy{
		newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pods, &corev1.Pod{}, lrVersions.Pods),
		newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Nodes, &corev1.Node{}, lrVersions.Nodes),
		newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pvs, &corev1.PersistentVolume{}, lrVersions.Pvs),
		newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pvcs, &corev1.PersistentVolumeClaim{}, lrVersions.Pvcs),
		newresourceEventProxy(sw, s.client.StorageV1().RESTClient(), Scs, &storagev1.StorageClass{}, lrVersions.Scs),
		newresourceEventProxy(sw, s.client.SchedulingV1().RESTClient(), Pcs, &schedulingv1.PriorityClass{}, lrVersions.Pcs),
	}
	for _, p := range proxies {
		watcher, err := createWatcher(p)
		if err != nil {
			return xerrors.Errorf("call createWatcher of %s: %w", p.resourceKind(), err)
		}
		go s.WatchAndHandleEvent(p, watcher, ctx.Done())
	}

	// This method will return an error and finish to event send when the connection from a client is closed.
	// It includes browser reload, ReadableStream.cancel() calling and so on.
	// This is to allow the front end to handle stream connection disconnections.
	<-ctx.Done()
	return nil
}

// WatchAndHandleEvent prepares a handler for the wacher and runs the handler
// until the stopCh is closed.
func (s *Service) WatchAndHandleEvent(proxy ResourceEventProxy, watcher watch.Interface, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	whandler := proxy.WatchHandlerFunc(watcher)
	run := func(stopCh <-chan struct{}) {
		if err := whandler(stopCh); err != nil {
			proxy.WatchErrorHandler(err)
		}
	}
	var wg wait.Group
	wg.StartWithChannel(stopCh, run)
	wg.Wait()
}

// createWatcher creates and returns RetryWatcher.
func createWatcher(p *resourceEventProxy) (watch.Interface, error) {
	watcher := cache.NewListWatchFromClient(p.c, string(p.r), corev1.NamespaceAll, fields.Everything())
	rWatcher, err := watchtools.NewRetryWatcher(p.lastResourceVersion, watcher)
	if err != nil {
		return nil, xerrors.Errorf("call NewRetryWatcher: %w", err)
	}
	return rWatcher, nil
}

// ResourceEventProxy is an interface that allows handle events and errors.
type ResourceEventProxy interface {
	WatchHandlerFunc(watcher watch.Interface) func(stopCh <-chan struct{}) error
	WatchErrorHandler(err error)
}

// resourceEventProxy implements event handler for the specified resource
// and knows where to send the event.
type resourceEventProxy struct {
	// writer knows where to send the event.
	writer StreamWriter
	// The RESTClient to watch the specified.
	c cache.Getter
	// The kind of resource to watch.
	r resourceKind
	// The Object of resource to watch.
	o runtime.Object
	// The last value of ResourceVersion. This is used to RetryWatcher.
	// First, this value is specified by a user.
	// After that this is updated when received an event everytime.
	//
	// The RetryWatcher will be reconnect to the apiserver if it is disconnected.
	// lastResourceVersion can be used to ensure that only events
	// that have not yet been received are received when reconnecting.
	lastResourceVersion string
}

func newresourceEventProxy(sw StreamWriter, c cache.Getter, r resourceKind, o runtime.Object, ilrv string) *resourceEventProxy {
	return &resourceEventProxy{
		writer:              sw,
		c:                   c,
		r:                   r,
		o:                   o,
		lastResourceVersion: ilrv,
	}
}

// WatchHandlerFunc watches the specified resource's event and calls the method to send the event
// and updates lastResourceVersion.
//nolint: cyclop // For readability.
func (p *resourceEventProxy) WatchHandlerFunc(watcher watch.Interface) func(stopCh <-chan struct{}) error {
	return func(stopCh <-chan struct{}) error {
		for {
			// give the stopCh a chance to stop the loop, even in case of continue statements further down on errors
			select {
			case <-stopCh:
				return nil
			case event, ok := <-watcher.ResultChan():
				// grab the event object
				if !ok {
					return xerrors.Errorf("closed channel")
				}
				obj, ok := event.Object.(metav1.Object)
				if !ok {
					return xerrors.Errorf("failed to type cast to metav1.Object: %T", event.Object)
				}
				var writingErr error
				switch event.Type {
				case watch.Added:
					writingErr = p.writer.Write(&WatchEvent{Kind: p.r, EventType: watch.Added, Obj: obj})
				case watch.Modified:
					writingErr = p.writer.Write(&WatchEvent{Kind: p.r, EventType: watch.Modified, Obj: obj})
				case watch.Deleted:
					writingErr = p.writer.Write(&WatchEvent{Kind: p.r, EventType: watch.Deleted, Obj: obj})
				case watch.Bookmark:
					// A `Bookmark` means watch has synced here, just update the resourceVersion
				case watch.Error:
					return xerrors.Errorf("%s: get an error watch event %#v", p.resourceKind(), obj)
				default:
					return xerrors.Errorf("%s: unable to understand watch event: %#v", p.resourceKind(), obj)
				}

				if writingErr != nil {
					return xerrors.Errorf("call Write to watch event: %w", writingErr)
				}
				p.lastResourceVersion = obj.GetResourceVersion()
			}
		}
	}
}

// WatchErrorHandler handles some errors.
func (p *resourceEventProxy) WatchErrorHandler(err error) {
	switch {
	case isExpiredError(err):
		// Don't set LastSyncResourceVersionUnavailable - LIST call with ResourceVersion=RV already
		// has a semantic that it returns data at least as fresh as provided RV.
		// So first try to LIST with setting RV to resource version of last observed object.
		klog.Infof("watch of %v closed with: %w", p.r, err)
	case errors.Is(io.EOF, err):
		// watch closed normally
	case errors.Is(io.ErrUnexpectedEOF, err):
		klog.Infof("watch for %v closed with unexpected EOF: %w", p.r, err)
	default:
		utilruntime.HandleError(fmt.Errorf("failed to watch %v: %w", p.r, err))
	}
}

// resourceKind returns the resourceKind value that was set in initialization.
func (p *resourceEventProxy) resourceKind() resourceKind {
	return p.r
}

func isExpiredError(err error) bool {
	// In Kubernetes 1.17 and earlier, the api server returns both apierrors.StatusReasonExpired and
	// apierrors.StatusReasonGone for HTTP 410 (Gone) status code responses. In 1.18 the kube server is more consistent
	// and always returns apierrors.StatusReasonExpired. For backward compatibility we can only remove the apierrors.IsGone
	// check when we fully drop support for Kubernetes 1.17 servers from reflectors.
	return apierrors.IsResourceExpired(err) || apierrors.IsGone(err)
}
