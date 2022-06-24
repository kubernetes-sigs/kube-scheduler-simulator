package watcher

//go:generate mockgen -destination=./mock_$GOPACKAGE/responseStream.go . ResponseStream
//go:generate mockgen -destination=./mock_$GOPACKAGE/watchInterface.go -package=mock_watcher -mock_names Interface=MockWatchInterface k8s.io/apimachinery/pkg/watch Interface
//go:generate mockgen -destination=./mock_$GOPACKAGE/resourceeventproxy.go . ResourceEventProxy

import (
	"context"
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
// These values are specified from frontend.
type LastResourceVersions struct {
	Pods  string `json:"pods"`
	Nodes string `json:"nodes"`
	Pvs   string `json:"pvs"`
	Pvcs  string `json:"pvcs"`
	Scs   string `json:"storageClasses"`
	Pcs   string `json:"priorityClasses"`
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

// NewWatcherService initializes Service.
func NewWatcherService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

// WatchResources watches each simulator's resources and pushes notified events to the frontend continuously.
func (s *Service) WatchResources(ctx context.Context, stream ResponseStream, lrVersions *LastResourceVersions) error {
	sw := newStreamWriter(stream)
	podsEventProxy := newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pods, &corev1.Pod{}, lrVersions.Pods)
	nodesEventProxy := newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Nodes, &corev1.Node{}, lrVersions.Nodes)
	pvsEventProxy := newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pvs, &corev1.PersistentVolume{}, lrVersions.Pvs)
	pvcsEventProxy := newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pvcs, &corev1.PersistentVolumeClaim{}, lrVersions.Pvcs)
	scsEventProxy := newresourceEventProxy(sw, s.client.StorageV1().RESTClient(), Scs, &storagev1.StorageClass{}, lrVersions.Scs)
	pcsEventProxy := newresourceEventProxy(sw, s.client.SchedulingV1().RESTClient(), Pcs, &schedulingv1.PriorityClass{}, lrVersions.Pcs)

	podsWatcher, err := createWatcher(podsEventProxy)
	if err != nil {
		return xerrors.Errorf("call createWatcher of pod: %w", err)
	}
	nodesWatcher, err := createWatcher(nodesEventProxy)
	if err != nil {
		return xerrors.Errorf("call createWatcher of node: %w", err)
	}
	pvsWatcher, err := createWatcher(pvsEventProxy)
	if err != nil {
		return xerrors.Errorf("call createWatcher of pv: %w", err)
	}
	pvcsWatcher, err := createWatcher(pvcsEventProxy)
	if err != nil {
		return xerrors.Errorf("call createWatcher of pvc: %w", err)
	}
	scsWatcher, err := createWatcher(scsEventProxy)
	if err != nil {
		return xerrors.Errorf("call createWatcher of scs: %w", err)
	}
	pcsWatcher, err := createWatcher(pcsEventProxy)
	if err != nil {
		return xerrors.Errorf("call createWatcher of pcs: %w", err)
	}
	go s.WatchAndHandleEvent(podsEventProxy, podsWatcher, ctx.Done())
	go s.WatchAndHandleEvent(nodesEventProxy, nodesWatcher, ctx.Done())
	go s.WatchAndHandleEvent(pvsEventProxy, pvsWatcher, ctx.Done())
	go s.WatchAndHandleEvent(pvcsEventProxy, pvcsWatcher, ctx.Done())
	go s.WatchAndHandleEvent(scsEventProxy, scsWatcher, ctx.Done())
	go s.WatchAndHandleEvent(pcsEventProxy, pcsWatcher, ctx.Done())

	<-ctx.Done()
	return xerrors.Errorf("terminated to watch: %w", ctx.Err())
}

// WatchAndHandleEvent prepares an handler for the wacher and runs the handler
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
				switch event.Type {
				case watch.Added:
					p.writer.Write(&WatchEvent{Kind: p.r, EventType: watch.Added, Obj: obj})
					break
				case watch.Modified:
					p.writer.Write(&WatchEvent{Kind: p.r, EventType: watch.Modified, Obj: obj})
					break
				case watch.Deleted:
					p.writer.Write(&WatchEvent{Kind: p.r, EventType: watch.Deleted, Obj: obj})
					break
				default:
					break
				}
				p.lastResourceVersion = obj.GetResourceVersion()
			default:
			}
		}
	}
}

// WatchErrorHandler hanldes some errors.
func (p *resourceEventProxy) WatchErrorHandler(err error) {
	switch {
	case isExpiredError(err):
		// Don't set LastSyncResourceVersionUnavailable - LIST call with ResourceVersion=RV already
		// has a semantic that it returns data at least as fresh as provided RV.
		// So first try to LIST with setting RV to resource version of last observed object.
		klog.Infof("watch of %v closed with: %w", p.r, err)
	case err == io.EOF:
		// watch closed normally
	case err == io.ErrUnexpectedEOF:
		klog.Infof("watch for %v closed with unexpected EOF: %w", p.r, err)
	default:
		utilruntime.HandleError(fmt.Errorf("failed to watch %v: %w", p.r, err))
	}
}

func isExpiredError(err error) bool {
	// In Kubernetes 1.17 and earlier, the api server returns both apierrors.StatusReasonExpired and
	// apierrors.StatusReasonGone for HTTP 410 (Gone) status code responses. In 1.18 the kube server is more consistent
	// and always returns apierrors.StatusReasonExpired. For backward compatibility we can only remove the apierrors.IsGone
	// check when we fully drop support for Kubernetes 1.17 servers from reflectors.
	return apierrors.IsResourceExpired(err) || apierrors.IsGone(err)
}
