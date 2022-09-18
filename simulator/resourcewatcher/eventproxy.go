package resourcewatcher

//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/watchInterface.go -package=mock_resourcewatcher -mock_names Interface=MockWatchInterface k8s.io/apimachinery/pkg/watch Interface
//go:generate mockgen --build_flags=--mod=mod -source=eventproxy.go -destination=./mock_eventproxy_test.go -package=resourcewatcher
//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/lister.go -package=mock_resourcewatcher -mock_names Interface=MockListerInterface k8s.io/client-go/tools/cache Lister

import (
	"errors"
	"fmt"
	"io"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	sw "sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter"
)

// eventProxyer is an interface that allows handle events and errors.
type eventProxyer interface {
	listAndHandleItems(lw cache.Lister) error
	watchAndHandleEvent(watcher watch.Interface, stopCh <-chan struct{})
	lastResourceVersion() string
	resourceKind() sw.ResourceKind
	restClient() cache.Getter
}

// eventProxy implements event handler for the specified resource
// and knows where to send the event.
type eventProxy struct {
	// writer knows where to send the event.
	writer StreamWriter
	// The RESTClient to watch the specified.
	c cache.Getter
	// The kind of resource to watch.
	r sw.ResourceKind
	// The Object of resource to watch.
	o runtime.Object
	// The last value of ResourceVersion. This is used to RetryWatcher.
	// First, this value is specified by a user.
	// After that this is updated when received an event everytime.
	//
	// The RetryWatcher will be reconnect to the apiserver if it is disconnected.
	// lrv can be used to ensure that only events
	// that have not yet been received are received when reconnecting.
	lrv string
}

func neweventProxy(sw StreamWriter, c cache.Getter, r sw.ResourceKind, o runtime.Object, lrv string) *eventProxy {
	return &eventProxy{
		writer: sw,
		c:      c,
		r:      r,
		o:      o,
		lrv:    lrv,
	}
}

// listAndHandleItems calls the list for the resource and the results is sent to the client by sendListedItems method.
func (p *eventProxy) listAndHandleItems(lw cache.Lister) error {
	list, err := lw.List(metav1.ListOptions{})
	if err != nil {
		return xerrors.Errorf("failed to list: %w", err)
	}
	items, lrv, err := extractListItem(list)
	if err != nil {
		return xerrors.Errorf("call HandleListItems: %w", err)
	}
	if err := p.sendListedItems(items); err != nil {
		return xerrors.Errorf("call ListItemsHandle: %w", err)
	}
	p.lrv = lrv
	return nil
}

// extractListItem validates whether the object is a list object and returns these items and lastResourceVersion.
func extractListItem(list runtime.Object) ([]runtime.Object, string, error) {
	listMetaInterface, err := meta.ListAccessor(list)
	if err != nil {
		return nil, "", xerrors.Errorf("failed to ListAccessor, unable to understand list result %#v: %w", list, err)
	}
	items, err := meta.ExtractList(list)
	if err != nil {
		return nil, "", xerrors.Errorf("failed to ExtractList, unable to understand list result %#v: %w", list, err)
	}
	return items, listMetaInterface.GetResourceVersion(), nil
}

// watchAndHandleEvent prepares a handler for the wacher and runs the handler
// until the stopCh is closed.
func (p *eventProxy) watchAndHandleEvent(watcher watch.Interface, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	handleFunc := p.watchHandlerFunc(watcher)
	run := func(stopCh <-chan struct{}) {
		if err := handleFunc(stopCh); err != nil {
			p.watchErrorHandler(err)
		}
	}
	var wg wait.Group
	wg.StartWithChannel(stopCh, run)
	wg.Wait()
}

// sendListedItems sends results of list as "ADDED" event to the client.
// This method will be expected to call before starting the watch.
func (p *eventProxy) sendListedItems(items []runtime.Object) error {
	for _, item := range items {
		if err := p.writer.Write(&sw.WatchEvent{Kind: p.r, EventType: watch.Added, Obj: item}); err != nil {
			return xerrors.Errorf("call Write to return list item %#v: %w", item, err)
		}
	}
	return nil
}

// watchHandlerFunc watches the specified resource's event.
// When it receives the event from watcher, it sends the event to stream
// and updates the lastResourceVersion in the eventProxy.
//
//nolint:cyclop // For readability.
func (p *eventProxy) watchHandlerFunc(watcher watch.Interface) func(stopCh <-chan struct{}) error {
	return func(stopCh <-chan struct{}) error {
		for {
			// give the stopCh a chance to stop the loop, even in case of continue statements further down on errors
			select {
			case <-stopCh:
				return nil
			case event, ok := <-watcher.ResultChan():
				// grab the event object
				if !ok {
					return xerrors.New("closed channel")
				}
				obj, ok := event.Object.(metav1.Object)
				if !ok {
					return xerrors.Errorf("failed to cast type from %T to metav1.Object", event.Object)
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
					return xerrors.Errorf("%s: unsupported event type %v, object %#v", p.resourceKind(), event.Type, obj)
				}

				if writingErr != nil {
					return xerrors.Errorf("call Write to watch event: %w", writingErr)
				}
				p.lrv = obj.GetResourceVersion()
			}
		}
	}
}

func (p *eventProxy) resourceKind() sw.ResourceKind {
	return p.r
}

func (p *eventProxy) restClient() cache.Getter {
	return p.c
}

// lastResourceVersion returns the lastResourceVersion value that is kept in the proxy.
func (p *eventProxy) lastResourceVersion() string {
	return p.lrv
}

// WatchErrorHandler handles some errors.
func (p *eventProxy) watchErrorHandler(err error) {
	switch {
	case errors.Is(io.EOF, err):
		// watch closed normally
	case errors.Is(io.ErrUnexpectedEOF, err):
		klog.Infof("watch for %v closed with unexpected EOF: %w", p.r, err)
	default:
		utilruntime.HandleError(fmt.Errorf("failed to watch %v: %w", p.r, err))
	}
}
