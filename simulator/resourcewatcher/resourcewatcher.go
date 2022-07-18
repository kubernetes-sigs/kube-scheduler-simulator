package resourcewatcher

//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/watchInterface.go -package=mock_resourcewatcher -mock_names Interface=MockWatchInterface k8s.io/apimachinery/pkg/watch Interface
//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/resourceeventproxy.go . ResourceEventProxy
//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/streamWriter.go . StreamWriter

import (
	"context"
	"errors"
	"fmt"
	"io"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
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

	sw "sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter"
)

type resourceKind = sw.ResourceKind

const (
	Pods  resourceKind = "pods"
	Nodes resourceKind = "nodes"
	Pvs   resourceKind = "persistentvolumes"
	Pvcs  resourceKind = "persistentvolumeclaims"
	Scs   resourceKind = "storageclasses"
	Pcs   resourceKind = "priorityclasses"
)

// LastResourceVersions includes each resource's lastResourceVersions.
type LastResourceVersions struct {
	Pods  string `json:"pods"`
	Nodes string `json:"nodes"`
	Pvs   string `json:"pvs"`
	Pvcs  string `json:"pvcs"`
	Scs   string `json:"scs"`
	Pcs   string `json:"pcs"`
}

// StreamWriter is an interface that allows send a received WatchEvent to the frontend.
type StreamWriter interface {
	Write(we *sw.WatchEvent) error
}

// Service watches simulator's resources.
type Service struct {
	client clientset.Interface
}

// NewService initializes Service.
func NewService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

// Watch watches each simulator's resources and send notified events to the frontend continuously.
func (s *Service) Watch(ctx context.Context, stream sw.ResponseStream, lrVersions *LastResourceVersions) error {
	sw := sw.NewStreamWriter(stream)
	proxies := []*resourceEventProxy{
		newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pods, &corev1.Pod{}, lrVersions.Pods),
		newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Nodes, &corev1.Node{}, lrVersions.Nodes),
		newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pvs, &corev1.PersistentVolume{}, lrVersions.Pvs),
		newresourceEventProxy(sw, s.client.CoreV1().RESTClient(), Pvcs, &corev1.PersistentVolumeClaim{}, lrVersions.Pvcs),
		newresourceEventProxy(sw, s.client.StorageV1().RESTClient(), Scs, &storagev1.StorageClass{}, lrVersions.Scs),
		newresourceEventProxy(sw, s.client.SchedulingV1().RESTClient(), Pcs, &schedulingv1.PriorityClass{}, lrVersions.Pcs),
	}
	runctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, p := range proxies {
		go s.Run(p, runctx.Done(), cancel)
	}

	select {
	case <-runctx.Done():
		// runctx monitors s.Run (ListAndWatch) for each resource.
		// If some error occurs in the process before starting the watch,
		// this error is returned.
		return xerrors.Errorf("failed to run ListAndWatch: %w", runctx.Err())
	case <-ctx.Done():
		// This method will return an error and finish to event send when the connection from a client is closed.
		// It includes browser reload, ReadableStream.cancel() calling and so on.
		// This is to allow the front end to handle stream connection disconnections.
		return nil
	}
}

// Run runs ListAndWatch method.
// If an error is returned, call cancel to abort ListAndWatch of other resources being processed in parallel.
func (s *Service) Run(p *resourceEventProxy, stopCh <-chan struct{}, cancel context.CancelFunc) {
	if p.resourceKind() == Pcs {
		cancel()
		return
	}
	if err := s.ListAndWatch(p, stopCh); err != nil {
		cancel()
	}
}

// ListAndWatch runs list and watch on the target resource. The list is not always ran
// This method returns error unless an error occurs in the watch. If an error occurs in the watch,
// it outputs a log and re-run the watch.
func (s *Service) ListAndWatch(p *resourceEventProxy, stopCh <-chan struct{}) error {
	lw := createListWatch(p)
	// If the lastResourceVersion isn't specified by client, call the list and return the result as ADDED event first.
	if p.lastResourceVersion == "" {
		if err := p.ListAndHandleItems(*lw); err != nil {
			return xerrors.Errorf("call to ListAndHandleItems of %s: %w", p.resourceKind(), err)
		}
	}
	watcher, err := createWatcher(p, lw)
	if err != nil {
		return xerrors.Errorf("call to createWatcher of %s: %w", p.resourceKind(), err)
	}
	s.WatchAndHandleEvent(p, watcher, stopCh)
	return nil
}

// ListAndHandleItems call the list for the resource and the results is send to the client by ListItemsHandler method.
func (p *resourceEventProxy) ListAndHandleItems(lw cache.ListWatch) error {
	list, err := lw.List(metav1.ListOptions{})
	if err != nil {
		return xerrors.Errorf("failed to list: %w", err)
	}
	listMetaInterface, err := meta.ListAccessor(list)
	if err != nil {
		return xerrors.Errorf("unable to understand list result %#v: %w", list, err)
	}
	p.lastResourceVersion = listMetaInterface.GetResourceVersion()
	items, err := meta.ExtractList(list)
	if err != nil {
		return xerrors.Errorf("unable to understand list result %#v: %w", list, err)
	}
	if err := p.ListItemsHandler(items); err != nil {
		return xerrors.Errorf("call ListItemsHandle: %w", err)
	}
	return nil
}

// WatchAndHandleEvent prepares a handler for the wacher and runs the handler
// until the stopCh is closed.
func (s *Service) WatchAndHandleEvent(p ResourceEventProxy, watcher watch.Interface, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	whandler := p.WatchHandlerFunc(watcher)
	run := func(stopCh <-chan struct{}) {
		if err := whandler(stopCh); err != nil {
			p.WatchErrorHandler(err)
		}
	}
	var wg wait.Group
	wg.StartWithChannel(stopCh, run)
	wg.Wait()
}

// createListWatch creates and returns ListWatch.
func createListWatch(p *resourceEventProxy) *cache.ListWatch {
	return cache.NewListWatchFromClient(p.c, string(p.r), corev1.NamespaceAll, fields.Everything())
}

// createWatcher creates and returns RetryWatcher.
func createWatcher(p *resourceEventProxy, lw *cache.ListWatch) (watch.Interface, error) {
	rWatcher, err := watchtools.NewRetryWatcher(p.lastResourceVersion, lw)
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

// ListItemsHandler sends results of list as "ADDED" event to the client.
// This method will be expected to call before starting the watch.
func (p *resourceEventProxy) ListItemsHandler(items []runtime.Object) error {
	for _, item := range items {
		if err := p.writer.Write(&sw.WatchEvent{Kind: p.r, EventType: watch.Added, Obj: item}); err != nil {
			return xerrors.Errorf("call Write to return list item %#v: %w", item, err)
		}
	}
	return nil
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
					writingErr = p.writer.Write(&sw.WatchEvent{Kind: p.r, EventType: watch.Added, Obj: obj})
				case watch.Modified:
					writingErr = p.writer.Write(&sw.WatchEvent{Kind: p.r, EventType: watch.Modified, Obj: obj})
				case watch.Deleted:
					writingErr = p.writer.Write(&sw.WatchEvent{Kind: p.r, EventType: watch.Deleted, Obj: obj})
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

// LastResourceVersion returns the lastResourceVersion value that is kept in the proxy.
func (p *resourceEventProxy) LastResourceVersion() string {
	return p.lastResourceVersion
}

func isExpiredError(err error) bool {
	// In Kubernetes 1.17 and earlier, the api server returns both apierrors.StatusReasonExpired and
	// apierrors.StatusReasonGone for HTTP 410 (Gone) status code responses. In 1.18 the kube server is more consistent
	// and always returns apierrors.StatusReasonExpired. For backward compatibility we can only remove the apierrors.IsGone
	// check when we fully drop support for Kubernetes 1.17 servers from reflectors.
	return apierrors.IsResourceExpired(err) || apierrors.IsGone(err)
}
