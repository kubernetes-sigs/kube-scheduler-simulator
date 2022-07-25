package resourcewatcher

//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/streamWriter.go . StreamWriter

import (
	"context"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/klog/v2"

	sw "sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter"
)

const (
	Pods  sw.ResourceKind = "pods"
	Nodes sw.ResourceKind = "nodes"
	Pvs   sw.ResourceKind = "persistentvolumes"
	Pvcs  sw.ResourceKind = "persistentvolumeclaims"
	Scs   sw.ResourceKind = "storageclasses"
	Pcs   sw.ResourceKind = "priorityclasses"
)

// LastResourceVersions includes each resource's LastResourceVersions.
type LastResourceVersions struct {
	Pods  string
	Nodes string
	Pvs   string
	Pvcs  string
	Scs   string
	Pcs   string
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

// ListWatch watches each simulator's resources and send notified events to the frontend continuously.
func (s *Service) ListWatch(ctx context.Context, stream sw.ResponseStream, lrVersions *LastResourceVersions) error {
	sw := sw.NewStreamWriter(stream)
	proxies := []*eventProxy{
		neweventProxy(sw, s.client.CoreV1().RESTClient(), Pods, &corev1.Pod{}, lrVersions.Pods),
		neweventProxy(sw, s.client.CoreV1().RESTClient(), Nodes, &corev1.Node{}, lrVersions.Nodes),
		neweventProxy(sw, s.client.CoreV1().RESTClient(), Pvs, &corev1.PersistentVolume{}, lrVersions.Pvs),
		neweventProxy(sw, s.client.CoreV1().RESTClient(), Pvcs, &corev1.PersistentVolumeClaim{}, lrVersions.Pvcs),
		neweventProxy(sw, s.client.StorageV1().RESTClient(), Scs, &storagev1.StorageClass{}, lrVersions.Scs),
		neweventProxy(sw, s.client.SchedulingV1().RESTClient(), Pcs, &schedulingv1.PriorityClass{}, lrVersions.Pcs),
	}
	runctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, p := range proxies {
		go s.run(p, runctx.Done(), cancel)
	}

	select {
	case <-runctx.Done():
		// ruuctx monitors s.Run (ListAndWatch) for each resource.
		// If some error occurs in the process before starting the watch,
		// ths error is returned.
		return xerrors.Errorf("failed to run ListAndWatch: %w", runctx.Err())
	case <-ctx.Done():
		// This method will return an error and finish to event send when the connection from a client is closed.
		// It includes browser reload, ReadableStream.cancel() calling and so on.
		// This is to allow the front end to handle stream connection disconnections.
		return nil
	}
}

// run runs doListAndWatch method.
// If an error is returned, call cancel to abort ListAndWatch of other resources being processed in parallel.
func (s *Service) run(p *eventProxy, stopCh <-chan struct{}, cancel context.CancelFunc) {
	defer cancel()
	// ListAndWatch usually continues to wait for WATCH to end and does not return any value.
	if err := s.doListAndWatch(p, stopCh); err != nil {
		cancel()
		klog.Errorf("call ListAndWatch: %w", err)
	}
}

// ListAndWatch runs list and watch on the target resource. The list is not always ran
// This method returns error unless an error occurs in the watch. If an error occurs in the watch,
// it outputs a log and re-run the watch.
func (s *Service) doListAndWatch(p eventProxyer, stopCh <-chan struct{}) error {
	lw := createListWatch(p)
	// If the lastResourceVersion isn't specified by client, call the list and return the result as ADDED event first.
	if p.lastResourceVersion() == "" {
		if err := p.listAndHandleItems(lw); err != nil {
			return xerrors.Errorf("call listAndHandleItems for %s: %w", p.resourceKind(), err)
		}
	}
	watcher, err := createWatcher(p, lw)
	if err != nil {
		return xerrors.Errorf("call createWatcher for %s: %w", p.resourceKind(), err)
	}
	p.watchAndHandleEvent(watcher, stopCh)
	return nil
}

// createListWatch creates and returns ListWatch.
func createListWatch(p eventProxyer) cache.ListerWatcher {
	return cache.NewListWatchFromClient(p.restClient(), string(p.resourceKind()), corev1.NamespaceAll, fields.Everything())
}

// createWatcher creates and returns RetryWatcher.
func createWatcher(p eventProxyer, lw cache.ListerWatcher) (watch.Interface, error) {
	rWatcher, err := watchtools.NewRetryWatcher(p.lastResourceVersion(), lw)
	if err != nil {
		return nil, xerrors.Errorf("call NewRetryWatcher: %w", err)
	}
	return rWatcher, nil
}
