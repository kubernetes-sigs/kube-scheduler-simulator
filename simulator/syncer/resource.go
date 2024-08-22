package syncer

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// DefaultGVRs is a list of GroupVersionResource that we sync by default (configurable with Options),
// which is a suitable resource set for the vanilla scheduler.
//
// Note that this order matters - When first importing resources, we want to sync namespaces first, then priorityclasses, storageclasses...
var DefaultGVRs = []schema.GroupVersionResource{
	{Group: "", Version: "v1", Resource: "namespaces"},
	{Group: "scheduling.k8s.io", Version: "v1", Resource: "priorityclasses"},
	{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	{Group: "", Version: "v1", Resource: "nodes"},
	{Group: "", Version: "v1", Resource: "persistentvolumes"},
	{Group: "", Version: "v1", Resource: "pods"},
}

// Event is a type of events that occur in the source cluster.
type Event int

const (
	Add Event = iota
	Update
)

// mandatoryMutatingFunctions is MutatingFunctions that we must register.
// We don't allow users to opt out them.
var mandatoryMutatingFunctions = map[schema.GroupVersionResource]MutatingFunction{
	{Group: "", Version: "v1", Resource: "persistentvolumes"}: mutatePV,
	{Group: "", Version: "v1", Resource: "pods"}:              mutatePods,
}

// mandatoryFilteringFunctions is FilteringFunctions that we must register.
// We don't allow users to opt out them.
var mandatoryFilteringFunctions = map[schema.GroupVersionResource]FilteringFunction{
	{Group: "", Version: "v1", Resource: "pods"}: filterPods,
}

// FilteringFunction is a function that filters a resource.
// If it returns false, the resource will not be imported.
type FilteringFunction func(ctx context.Context, resource *unstructured.Unstructured, clients *Clients, event Event) (bool, error)

// MutatingFunction is a function that mutates a resource before importing it.
type MutatingFunction func(ctx context.Context, resource *unstructured.Unstructured, clients *Clients, event Event) (*unstructured.Unstructured, error)

func mutatePV(ctx context.Context, resource *unstructured.Unstructured, clients *Clients, _ Event) (*unstructured.Unstructured, error) {
	var pv v1.PersistentVolume
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.UnstructuredContent(), &pv)
	if err != nil {
		return nil, err
	}

	if pv.Status.Phase == v1.VolumeBound {
		// PersistentVolumeClaims's UID is changed in a destination cluster when importing from a source cluster,
		// and thus we need to update the PVC UID in the PersistentVolume.
		// Get PVC of pv.Spec.ClaimRef.Name.
		pvc, err := clients.SrcDynamicClient.Resource(schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "persistentvolumeclaims",
		}).Namespace(pv.Spec.ClaimRef.Namespace).Get(ctx, pv.Spec.ClaimRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		pv.Spec.ClaimRef.UID = pvc.GetUID()
	}

	modifiedUnstructed, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pv)
	return &unstructured.Unstructured{Object: modifiedUnstructed}, err
}

func mutatePods(_ context.Context, resource *unstructured.Unstructured, _ *Clients, _ Event) (*unstructured.Unstructured, error) {
	var pod v1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.UnstructuredContent(), &pod)
	if err != nil {
		return nil, err
	}

	// Pods must have the default ServiceAccount because ServiceAccount is not synced.
	pod.Spec.ServiceAccountName = ""
	pod.Spec.DeprecatedServiceAccount = ""

	// If the pod has an owner, it may be deleted because resources such as ReplicaSet are not synced.
	pod.OwnerReferences = nil

	modifiedUnstructed, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	return &unstructured.Unstructured{Object: modifiedUnstructed}, err
}

// filterPods checks if a pod is already scheduled when it's updated.
// We only want to update pods that are not yet scheduled.
func filterPods(_ context.Context, resource *unstructured.Unstructured, _ *Clients, event Event) (bool, error) {
	if event == Add {
		// We always add a Pod, regardless it's scheduled or not.
		return true, nil
	}

	var pod v1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.UnstructuredContent(), &pod)
	if err != nil {
		return false, err
	}

	if pod.Spec.NodeName != "" {
		// We just ignore the not found error because the scheduler may preempt the Pods, or users may remove the resources for debugging.
		klog.Info("Skipped to update resource because we cannot find it in the destination cluster", "resource", klog.KObj(&pod.ObjectMeta))
		return false, nil
	}

	// This Pod should be applied on the destination cluster.
	return true, nil
}
