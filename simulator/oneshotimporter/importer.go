package oneshotimporter

//go:generate mockgen -destination=./mock_$GOPACKAGE/replicate.go . ReplicateService

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourceapplier"
)

// Service has two ReplicateServices.
// importService is used to import(replicate) these resources to the simulator.
// exportService is used to export resources from a target cluster.
type Service struct {
	srcDynamicClient      dynamic.Interface
	resouceApplierService *resourceapplier.Service
	gvrs                  []schema.GroupVersionResource
}

// DefaultGVRs is a list of GroupVersionResource that we import.
// Note that this order matters - When first importing resources, we want to import namespaces first, then priorityclasses, storageclasses...
var DefaultGVRs = []schema.GroupVersionResource{
	{Group: "", Version: "v1", Resource: "namespaces"},
	{Group: "scheduling.k8s.io", Version: "v1", Resource: "priorityclasses"},
	{Group: "", Version: "v1", Resource: "serviceaccounts"},
	{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	{Group: "", Version: "v1", Resource: "nodes"},
	{Group: "", Version: "v1", Resource: "persistentvolumes"},
	{Group: "", Version: "v1", Resource: "pods"},
}

// NewService initializes Service.
func NewService(srcClient dynamic.Interface, resourceApplier *resourceapplier.Service) *Service {
	gvrs := DefaultGVRs
	if resourceApplier.GVRsToSync != nil {
		gvrs = resourceApplier.GVRsToSync
	}

	return &Service{
		srcDynamicClient:      srcClient,
		resouceApplierService: resourceApplier,
		gvrs:                  gvrs,
	}
}

// ImportClusterResources gets resources from the target cluster via exportService
// and then apply those resources to the simulator.
// Note: this method doesn't handle scheduler configuration.
// If you want to use the scheduler configuration along with the imported resources on the simulator,
// you need to set the path of the scheduler configuration file to `kubeSchedulerConfigPath` value in the Simulator Server Configuration.
func (s *Service) ImportClusterResources(ctx context.Context, labelSelector metav1.LabelSelector) error {
	for _, gvr := range s.gvrs {
		if err := s.importResource(ctx, gvr, labelSelector); err != nil {
			return xerrors.Errorf("import resource %s: %w", gvr.String(), err)
		}
	}

	return nil
}

func (s *Service) importResource(ctx context.Context, gvr schema.GroupVersionResource, labelSelector metav1.LabelSelector) error {
	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return xerrors.Errorf("convert label selector: %w", err)
	}

	resources, err := s.srcDynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return xerrors.Errorf("list resources: %w", err)
	}

	var wg sync.WaitGroup
	for _, resource := range resources.Items {
		wg.Add(1)
		fmt.Printf("importing resource: %s\n", resource.GetName())
		go func(r *unstructured.Unstructured) {
			defer wg.Done()
			if err := s.resouceApplierService.Create(ctx, r); err != nil {
				klog.Warningf("failed to import resource: %v", err)
			}
		}(&resource)
	}
	wg.Wait()

	return nil
}
