package resourceapplier

import (
	"context"
	"encoding/json"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// FilteringFunction is a function that filters a resource.
// If it returns false, the resource will not be imported.
type FilteringFunction func(ctx context.Context, resource *unstructured.Unstructured, clients *Clients) (bool, error)

// MutatingFunction is a function that mutates a resource before importing it.
type MutatingFunction func(ctx context.Context, resource *unstructured.Unstructured, clients *Clients) (*unstructured.Unstructured, error)

// Note: Clients and its fields are exposed intentionally so that users can use it in MutatingFunction and FilteringFunction.
type Clients struct {
	// DynamicClient is the dynamic client for the destination cluster, which the resource is supposed to be copied to.
	DynamicClient dynamic.Interface
	RestMapper    meta.RESTMapper
}

type Options struct {
	GVRsToApply          []schema.GroupVersionResource
	FilterBeforeCreating map[schema.GroupVersionResource][]FilteringFunction
	MutateBeforeCreating map[schema.GroupVersionResource][]MutatingFunction
	FilterBeforeUpdating map[schema.GroupVersionResource][]FilteringFunction
	MutateBeforeUpdating map[schema.GroupVersionResource][]MutatingFunction
}

type Service struct {
	clients *Clients

	mutateBeforeCreating map[schema.GroupVersionResource][]MutatingFunction
	filterBeforeCreating map[schema.GroupVersionResource][]FilteringFunction
	mutateBeforeUpdating map[schema.GroupVersionResource][]MutatingFunction
	filterBeforeUpdating map[schema.GroupVersionResource][]FilteringFunction

	GVRsToSync []schema.GroupVersionResource
}

func New(dynamicClient dynamic.Interface, restMapper meta.RESTMapper, options Options) *Service {
	s := &Service{
		clients: &Clients{
			DynamicClient: dynamicClient,
			RestMapper:    restMapper,
		},

		filterBeforeCreating: map[schema.GroupVersionResource][]FilteringFunction{},
		mutateBeforeCreating: map[schema.GroupVersionResource][]MutatingFunction{},
		filterBeforeUpdating: map[schema.GroupVersionResource][]FilteringFunction{},
		mutateBeforeUpdating: map[schema.GroupVersionResource][]MutatingFunction{},

		GVRsToSync: options.GVRsToApply,
	}

	for gvr, fn := range mandatoryFilterForCreating {
		s.addFilterBeforeCreating(gvr, []FilteringFunction{fn})
	}
	for gvr, fn := range mandatoryMutateForCreating {
		s.addMutateBeforeCreating(gvr, []MutatingFunction{fn})
	}
	for gvr, fn := range mandatoryFilterForUpdating {
		s.addFilterBeforeUpdating(gvr, []FilteringFunction{fn})
	}
	for gvr, fn := range mandatoryMutateForUpdating {
		s.addMutateBeforeUpdating(gvr, []MutatingFunction{fn})
	}

	for gvr, fns := range options.FilterBeforeCreating {
		s.addFilterBeforeCreating(gvr, fns)
	}
	for gvr, fns := range options.MutateBeforeCreating {
		s.addMutateBeforeCreating(gvr, fns)
	}
	for gvr, fns := range options.FilterBeforeUpdating {
		s.addFilterBeforeUpdating(gvr, fns)
	}
	for gvr, fns := range options.MutateBeforeUpdating {
		s.addMutateBeforeUpdating(gvr, fns)
	}

	return s
}

func (s *Service) Create(ctx context.Context, resource *unstructured.Unstructured) error {
	// Extract the GroupVersionResource from the Unstructured object
	gvk := resource.GroupVersionKind()
	gvr, err := s.findGVRForGVK(gvk)
	if err != nil {
		return err
	}

	// Namespaces resources should be created within the namespace defined in the Unstructured object
	namespace := resource.GetNamespace()

	// Run the filtering function for the resource.
	if ok, err := s.filterResourceForCreating(ctx, gvr, resource, s.clients); !ok || err != nil {
		return err
	}

	// When creating a resource on the destination cluster, we must remove the metadata such as UID and Generation.
	// It's done for all resources.
	resource = removeUnnecessaryMetadata(resource)

	// Run the mutating function for the resource.
	resource, err = s.mutateResourceForCreating(ctx, gvr, resource, s.clients)
	if err != nil {
		return xerrors.Errorf("failed to mutate resource: %w", err)
	}

	// Create the resource on the destination cluster using the dynamic client
	_, err = s.clients.DynamicClient.Resource(gvr).Namespace(namespace).Create(
		ctx,
		resource,
		metav1.CreateOptions{},
	)
	if err != nil {
		return xerrors.Errorf("failed to create resource: %w", err)
	}

	if gvk.Kind == "Pod" {
		return s.patchPodStatus(ctx, gvr, resource)
	}
	return nil
}

func (s *Service) Update(ctx context.Context, resource *unstructured.Unstructured) error {
	// Extract the GroupVersionResource from the Unstructured object
	gvk := resource.GroupVersionKind()
	gvr, err := s.findGVRForGVK(gvk)
	if err != nil {
		return err
	}

	// Namespaces resources should be created within the namespace defined in the Unstructured object
	namespace := resource.GetNamespace()

	// Run the filtering function for the resource.
	if ok, err := s.filterResourceForUpdating(ctx, gvr, resource, s.clients); !ok || err != nil {
		return err
	}

	// When updating a resource on the destination cluster, we must remove the metadata such as UID and Generation.
	// It's done for all resources.
	resource = removeUnnecessaryMetadata(resource)

	// Run the mutating function for the resource.
	resource, err = s.mutateResourceForUpdating(ctx, gvr, resource, s.clients)
	if err != nil {
		return xerrors.Errorf("failed to mutate resource: %w", err)
	}

	// Update the resource on the destination cluster using the dynamic client
	_, err = s.clients.DynamicClient.Resource(gvr).Namespace(namespace).Update(
		ctx,
		resource,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return xerrors.Errorf("failed to update resource: %w", err)
	}

	return nil
}

func (s *Service) Delete(
	ctx context.Context,
	resource *unstructured.Unstructured,
) error {
	// Extract the GroupVersionResource from the Unstructured object
	gvk := resource.GroupVersionKind()
	gvr, err := s.findGVRForGVK(gvk)
	if err != nil {
		return err
	}

	// Namespaces resources should be created within the namespace defined in the Unstructured object
	namespace := resource.GetNamespace()

	// Create the resource on the destination cluster using the dynamic client
	err = s.clients.DynamicClient.Resource(gvr).Namespace(namespace).Delete(
		ctx,
		resource.GetName(),
		metav1.DeleteOptions{},
	)
	if err != nil {
		return xerrors.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

func (s *Service) filterResourceForCreating(ctx context.Context, gvr schema.GroupVersionResource, resource *unstructured.Unstructured, clients *Clients) (bool, error) {
	filteringFns, ok := s.filterBeforeCreating[gvr]
	if !ok {
		return true, nil
	}

	for _, filteringFn := range filteringFns {
		ok, err := filteringFn(ctx, resource, clients)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

func (s *Service) mutateResourceForCreating(ctx context.Context, gvr schema.GroupVersionResource, resource *unstructured.Unstructured, clients *Clients) (*unstructured.Unstructured, error) {
	mutatingFns, ok := s.mutateBeforeCreating[gvr]
	if !ok {
		return resource, nil
	}

	for _, mutatingFn := range mutatingFns {
		modifiedResource, err := mutatingFn(ctx, resource, clients)
		if err != nil {
			return nil, err
		}
		resource = modifiedResource
	}

	return resource, nil
}

func (s *Service) filterResourceForUpdating(ctx context.Context, gvr schema.GroupVersionResource, resource *unstructured.Unstructured, clients *Clients) (bool, error) {
	filteringFns, ok := s.filterBeforeUpdating[gvr]
	if !ok {
		return true, nil
	}

	for _, filteringFn := range filteringFns {
		ok, err := filteringFn(ctx, resource, clients)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

func (s *Service) mutateResourceForUpdating(ctx context.Context, gvr schema.GroupVersionResource, resource *unstructured.Unstructured, clients *Clients) (*unstructured.Unstructured, error) {
	mutatingFns, ok := s.mutateBeforeUpdating[gvr]
	if !ok {
		return resource, nil
	}

	for _, mutatingFn := range mutatingFns {
		modifiedResource, err := mutatingFn(ctx, resource, clients)
		if err != nil {
			return nil, err
		}
		resource = modifiedResource
	}

	return resource, nil
}

// findGVRForGVK uses the discovery client to get the GroupVersionResource for a given GroupVersionKind.
func (s *Service) findGVRForGVK(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	m, err := s.clients.RestMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return m.Resource, nil
}

// removeUnnecessaryMetadata removes the metadata from the resource.
func removeUnnecessaryMetadata(resource *unstructured.Unstructured) *unstructured.Unstructured {
	resource.SetUID("")
	resource.SetGeneration(0)
	resource.SetResourceVersion("")

	return resource
}

func (s *Service) addFilterBeforeCreating(gvr schema.GroupVersionResource, fn []FilteringFunction) {
	if _, ok := s.filterBeforeCreating[gvr]; !ok {
		s.filterBeforeCreating[gvr] = []FilteringFunction{}
	}

	s.filterBeforeCreating[gvr] = append(s.filterBeforeCreating[gvr], fn...)
}

func (s *Service) addMutateBeforeCreating(gvr schema.GroupVersionResource, fn []MutatingFunction) {
	if _, ok := s.mutateBeforeCreating[gvr]; !ok {
		s.mutateBeforeCreating[gvr] = []MutatingFunction{}
	}

	s.mutateBeforeCreating[gvr] = append(s.mutateBeforeCreating[gvr], fn...)
}

func (s *Service) addFilterBeforeUpdating(gvr schema.GroupVersionResource, fn []FilteringFunction) {
	if _, ok := s.filterBeforeUpdating[gvr]; !ok {
		s.filterBeforeUpdating[gvr] = []FilteringFunction{}
	}

	s.filterBeforeUpdating[gvr] = append(s.filterBeforeUpdating[gvr], fn...)
}

func (s *Service) addMutateBeforeUpdating(gvr schema.GroupVersionResource, fn []MutatingFunction) {
	if _, ok := s.mutateBeforeUpdating[gvr]; !ok {
		s.mutateBeforeUpdating[gvr] = []MutatingFunction{}
	}

	s.mutateBeforeUpdating[gvr] = append(s.mutateBeforeUpdating[gvr], fn...)
}

func (s *Service) patchPodStatus(ctx context.Context, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error {
	namespace := resource.GetNamespace()
	newStatus := resource.Object["status"]
	patchData := map[string]interface{}{
		"status": newStatus,
	}
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return err
	}
	_, err = s.clients.DynamicClient.Resource(gvr).Namespace(namespace).Patch(
		ctx,
		resource.GetName(),
		types.MergePatchType,
		patchBytes, metav1.PatchOptions{},
		"status",
	)
	if err != nil {
		return xerrors.Errorf("failed to patch status: %w, gvr: %v, name: %v", err, gvr, resource.GetName())
	}
	return nil
}
