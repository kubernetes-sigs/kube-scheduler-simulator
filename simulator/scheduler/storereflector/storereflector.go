package storereflector

//go:generate mockgen -destination=./mock_$GOPACKAGE/resultstore.go . ResultStore

import (
	"context"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/util"
)

type Reflector interface {
	AddResultStore(store ResultStore, key string)
	ResisterResultSavingToInformer(informerFactory informers.SharedInformerFactory, client clientset.Interface) error
}

// ResultStore represents the store which is stores data and shared with simulator and scheduler.
// Fulfilling this interface will allow the stored results to be saved as data in that Pod
// when the Pod's schedule is complete.
type ResultStore interface {
	// AddStoredResultToPod adds all data corresponding to the pod to the pod.
	AddStoredResultToPod(pod *corev1.Pod)
	// DeleteData deletes all data corresponding to the pod.
	DeleteData(key corev1.Pod)
}

// store manages any ResultStore.
// ResultStore stores any result that should be reflected to the Pod.
type reflector struct {
	resultStores map[string]ResultStore
}

func New() Reflector {
	return &reflector{
		resultStores: map[string]ResultStore{},
	}
}

// AddResultStore adds the ResultStore to the map.
func (s *reflector) AddResultStore(store ResultStore, key string) {
	s.resultStores[key] = store
}

// ResisterResultSavingToInformer registers the event handler to the informerFactory
// to reflects all results on the pod annotation when the scheduling is finished.
func (s *reflector) ResisterResultSavingToInformer(informerFactory informers.SharedInformerFactory, client clientset.Interface) error {
	// Reflector adds scheduling results when pod is updating.
	// This is because Extenders doesn't have any phase to hook scheduling finished. (both successfully and non-successfully)
	_, err := informerFactory.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: s.storeAllResultToPodFunc(client),
		},
	)
	if err != nil {
		return xerrors.Errorf("failed to AddEventHandler of Informer: %w", err)
	}
	return nil
}

// storeAllResultToPodFunc returns the function that reflects all results on the pod annotation when the scheduling is finished.
// It will be used as the even handler of resource updating.
func (s *reflector) storeAllResultToPodFunc(client clientset.Interface) func(interface{}, interface{}) {
	return func(_, newObj interface{}) {
		ctx := context.Background()
		pod, ok := newObj.(*corev1.Pod)
		if !ok {
			klog.ErrorS(nil, "Cannot convert to *v1.Pod", "obj", newObj)
			return
		}

		updateFunc := func() (bool, error) {
			// Fetch the latest Pod object and apply changes to it. Otherwise, our update may be
			// rejected due to our copy being stale. This also ensures we don't modify the copy from
			// the shared informer.
			newPod, err := client.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if err != nil {
				return false, xerrors.Errorf("get pod: %w", err)
			}
			if newPod.UID != pod.UID {
				return false, xerrors.Errorf("pod UID is different: %s != %s", newPod.UID, pod.UID)
			}

			// Call AddStoredResultToPod of all ResultStore which is added to the map
			// to reflects all results on the pod annotation.
			for k := range s.resultStores {
				s.resultStores[k].AddStoredResultToPod(newPod)
			}

			_, err = client.CoreV1().Pods(newPod.Namespace).Update(ctx, newPod, metav1.UpdateOptions{})
			if err != nil {
				// Even though we fetched the latest Pod object, we still might get a conflict
				// because of a concurrent update. Retrying these conflict errors will usually help
				// as long as we re-fetch the latest Pod object each time.
				if apierrors.IsConflict(err) {
					return false, nil
				}
				return false, xerrors.Errorf("update pod: %w", err)
			}
			return true, nil
		}
		if err := util.RetryWithExponentialBackOff(updateFunc); err != nil {
			klog.Errorf("failed to update the pod with retry to record store: %+v", err)
			return
		}

		for k := range s.resultStores {
			// Delete the data from the Reflector only if it is successfully added on the pod's annotations.
			s.resultStores[k].DeleteData(*pod)
		}
	}
}
