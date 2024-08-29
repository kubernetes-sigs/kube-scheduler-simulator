package syncer

import (
	"context"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type Service struct {
	clients *Clients

	gvrs               []schema.GroupVersionResource
	mutatingFunctions  map[schema.GroupVersionResource][]MutatingFunction
	filteringFunctions map[schema.GroupVersionResource][]FilteringFunction
}

// Note: Clients and its fields are exposed intentionally so that users can use it in MutatingFunction and FilteringFunction.
type Clients struct {
	// SrcDynamicClient is the dynamic client for the source cluster, which the resource is supposed to be copied from.
	SrcDynamicClient dynamic.Interface
	// DestDynamicClient is the dynamic client for the destination cluster, which the resource is supposed to be copied to.
	DestDynamicClient dynamic.Interface
	RestMapper        meta.RESTMapper
}

type Options struct {
	// GVRsToSync is a list of GroupVersionResource that will be synced.
	// If GVRsToSync is nil, defaultGVRs are used.
	GVRsToSync []schema.GroupVersionResource
	// AdditionalMutatingFunctions is a list of mutating functions that users add.
	AdditionalMutatingFunctions map[schema.GroupVersionResource]MutatingFunction
	// AdditionalFilteringFunctions is a list of filtering functions that users add.
	AdditionalFilteringFunctions map[schema.GroupVersionResource]FilteringFunction
}

func New(srcDynamicClient, destDynamicClient dynamic.Interface, restMapper meta.RESTMapper, options Options) *Service {
	s := &Service{
		clients: &Clients{
			SrcDynamicClient:  srcDynamicClient,
			DestDynamicClient: destDynamicClient,
			RestMapper:        restMapper,
		},
		gvrs:               DefaultGVRs,
		mutatingFunctions:  map[schema.GroupVersionResource][]MutatingFunction{},
		filteringFunctions: map[schema.GroupVersionResource][]FilteringFunction{},
	}

	if options.GVRsToSync != nil {
		s.gvrs = options.GVRsToSync
	}

	s.addMutatingFunctoins(mandatoryMutatingFunctions)
	s.addMutatingFunctoins(options.AdditionalMutatingFunctions)

	s.addFilteringFunctoins(mandatoryFilteringFunctions)
	s.addFilteringFunctoins(options.AdditionalFilteringFunctions)

	return s
}

func (s *Service) Run(ctx context.Context) error {
	klog.Info("Starting the cluster resource importer")

	infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(s.clients.SrcDynamicClient, 0, metav1.NamespaceAll, nil)
	for _, gvr := range s.gvrs {
		inf := infFact.ForResource(gvr).Informer()
		_, err := inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    s.addFunc,
			UpdateFunc: s.updateFunc,
			DeleteFunc: s.deleteFunc,
		})
		if err != nil {
			return xerrors.Errorf("failed to add event handler: %w", err)
		}
		go inf.Run(ctx.Done())
		infFact.WaitForCacheSync(ctx.Done())
	}

	klog.Info("Cluster resource syncer started")

	return nil
}

// createResourceOnDestinationCluster creates the resource on the destination cluster.
func (s *Service) createResourceOnDestinationCluster(
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

	// Run the filtering function for the resource.
	if ok, err := s.filterResource(ctx, gvr, resource, Add); !ok || err != nil {
		return err
	}

	// When creating a resource on the destination cluster, we must remove the metadata such as UID and Generation.
	// It's done for all resources.
	resource = removeUnnecessaryMetadata(resource)

	// Run the mutating function for the resource.
	resource, err = s.mutateResource(ctx, gvr, resource, Add)
	if err != nil {
		return xerrors.Errorf("failed to mutate resource: %w", err)
	}

	// Create the resource on the destination cluster using the dynamic client
	_, err = s.clients.DestDynamicClient.Resource(gvr).Namespace(namespace).Create(
		ctx,
		resource,
		metav1.CreateOptions{},
	)
	if err != nil {
		return xerrors.Errorf("failed to create resource: %w", err)
	}

	return nil
}

func (s *Service) updateResourceOnDestinationCluster(
	ctx context.Context,
	resource *unstructured.Unstructured,
) error {
	// Extract the GroupVersionResource from the Unstructured object.
	gvk := resource.GroupVersionKind()
	gvr, err := s.findGVRForGVK(gvk)
	if err != nil {
		return err
	}

	// Namespaces resources should be created within the namespace defined in the Unstructured object.
	namespace := resource.GetNamespace()

	// Run the filtering function for the resource.
	if ok, err := s.filterResource(ctx, gvr, resource, Update); !ok || err != nil {
		return err
	}

	// Run the mutating function for the resource.
	resource, err = s.mutateResource(ctx, gvr, resource, Update)
	if err != nil {
		return xerrors.Errorf("failed to mutate resource: %w", err)
	}

	// Create the resource on the destination cluster using the dynamic client
	_, err = s.clients.DestDynamicClient.Resource(gvr).Namespace(namespace).Update(
		ctx,
		resource,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return xerrors.Errorf("failed to create resource: %w", err)
	}

	return nil
}

// removeUnnecessaryMetadata removes the metadata from the resource.
func removeUnnecessaryMetadata(resource *unstructured.Unstructured) *unstructured.Unstructured {
	resource.SetUID("")
	resource.SetGeneration(0)
	resource.SetResourceVersion("")

	return resource
}

func (s *Service) deleteResourceOnDestinationCluster(
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
	err = s.clients.DestDynamicClient.Resource(gvr).Namespace(namespace).Delete(
		ctx,
		resource.GetName(),
		metav1.DeleteOptions{},
	)
	if err != nil {
		return xerrors.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

// findGVRForGVK uses the discovery client to get the GroupVersionResource for a given GroupVersionKind.
func (s *Service) findGVRForGVK(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	m, err := s.clients.RestMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return m.Resource, nil
}

func (s *Service) addFunc(obj interface{}) {
	ctx := context.Background()
	unstructObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		klog.Error("Failed to convert runtime.Object to *unstructured.Unstructured")
		return
	}

	err := s.createResourceOnDestinationCluster(ctx, unstructObj)
	if err != nil {
		klog.ErrorS(err, "Failed to create resource on destination cluster")
	}
}

func (s *Service) updateFunc(_, newObj interface{}) {
	ctx := context.Background()
	unstructObj, ok := newObj.(*unstructured.Unstructured)
	if !ok {
		klog.Error("Failed to convert runtime.Object to *unstructured.Unstructured")
		return
	}

	err := s.updateResourceOnDestinationCluster(ctx, unstructObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// We just ignore the not found error because the scheduler may preempt the Pods, or users may remove the resources for debugging.
			klog.Info("Skipped to update resource on destination: ", err)
		} else {
			klog.ErrorS(err, "Failed to update resource on destination cluster")
		}
	}
}

func (s *Service) deleteFunc(obj interface{}) {
	ctx := context.Background()
	unstructObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		klog.Error("Failed to convert runtime.Object to *unstructured.Unstructured")
		return
	}

	err := s.deleteResourceOnDestinationCluster(ctx, unstructObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// We just ignore the not found error because the scheduler may preempt the Pods, or users may remove the resources for debugging.
			klog.Info("Skipped to delete resource on destination: ", err)
		} else {
			klog.ErrorS(err, "Failed to delete resource on destination cluster")
		}
	}
}

func (s *Service) addMutatingFunctoins(m map[schema.GroupVersionResource]MutatingFunction) {
	for k, v := range m {
		if s.mutatingFunctions[k] == nil {
			s.mutatingFunctions[k] = []MutatingFunction{v}
		} else {
			s.mutatingFunctions[k] = append(s.mutatingFunctions[k], v)
		}
	}
}

func (s *Service) addFilteringFunctoins(m map[schema.GroupVersionResource]FilteringFunction) {
	for k, v := range m {
		if s.filteringFunctions[k] == nil {
			s.filteringFunctions[k] = []FilteringFunction{v}
		} else {
			s.filteringFunctions[k] = append(s.filteringFunctions[k], v)
		}
	}
}

func (s *Service) mutateResource(ctx context.Context, gvr schema.GroupVersionResource, resource *unstructured.Unstructured, event Event) (*unstructured.Unstructured, error) {
	mutatingFns, ok := s.mutatingFunctions[gvr]
	if !ok {
		return resource, nil
	}

	for _, mutatingFn := range mutatingFns {
		var err error
		resource, err = mutatingFn(ctx, resource, s.clients, event)
		if err != nil {
			return resource, err
		}
	}

	return resource, nil
}

func (s *Service) filterResource(ctx context.Context, gvr schema.GroupVersionResource, resource *unstructured.Unstructured, event Event) (bool, error) {
	filteringFns, ok := s.filteringFunctions[gvr]
	if !ok {
		return true, nil
	}

	for _, filteringFn := range filteringFns {
		ok, err := filteringFn(ctx, resource, s.clients, event)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}
