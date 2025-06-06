package resourceapplier

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// mandatoryFilterForCreating is FilteringFunctions that we must register for creating.
// We don't allow users to opt out them.
var mandatoryFilterForCreating = map[schema.GroupVersionResource]FilteringFunction{}

// mandatoryMutateForCreating is MutatingFunctions that we must register for creating.
// We don't allow users to opt out them.
var mandatoryMutateForCreating = map[schema.GroupVersionResource]MutatingFunction{
	{Group: "", Version: "v1", Resource: "persistentvolumes"}: mutatePV,
	{Group: "", Version: "v1", Resource: "pods"}:              mutatePods,
}

// mandatoryFilterForUpdating is FilteringFunctions that we must register.
// We don't allow users to opt out them.
var mandatoryFilterForUpdating = map[schema.GroupVersionResource]FilteringFunction{
	{Group: "", Version: "v1", Resource: "pods"}: filterPodsForUpdating,
}

// mandatoryMutateForUpdating is MutatingFunctions that we must register for updating.
// We don't allow users to opt out them.
var mandatoryMutateForUpdating = map[schema.GroupVersionResource]MutatingFunction{
	{Group: "", Version: "v1", Resource: "persistentvolumes"}: mutatePV,
	{Group: "", Version: "v1", Resource: "pods"}:              mutatePods,
}

func mutatePV(ctx context.Context, resource *unstructured.Unstructured, clients *Clients) (*unstructured.Unstructured, error) {
	var pv v1.PersistentVolume
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.UnstructuredContent(), &pv)
	if err != nil {
		return nil, err
	}

	if pv.Status.Phase == v1.VolumeBound {
		// PersistentVolumeClaims's UID is changed in a destination cluster when importing from a source cluster,
		// and thus we need to update the PVC UID in the PersistentVolume.
		// Get PVC of pv.Spec.ClaimRef.Name.
		pvc, err := clients.DynamicClient.Resource(schema.GroupVersionResource{
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

func mutatePods(_ context.Context, resource *unstructured.Unstructured, _ *Clients) (*unstructured.Unstructured, error) {
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

	if len(pod.Finalizers) != 0 {
		pod.Finalizers = make([]string, 0)
	}

	modifiedUnstructed, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	return &unstructured.Unstructured{Object: modifiedUnstructed}, err
}

// filterPods checks if a pod is already scheduled when it's updated.
// We only want to update pods that are not yet scheduled.
func filterPodsForUpdating(_ context.Context, resource *unstructured.Unstructured, _ *Clients) (bool, error) {
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
