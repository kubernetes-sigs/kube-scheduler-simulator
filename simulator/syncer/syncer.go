package syncer

import (
	"context"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourceapplier"
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

type Service struct {
	gvrs                   []schema.GroupVersionResource
	srcDynamicClient       dynamic.Interface
	resourceApplierService *resourceapplier.Service
}

type Options struct {
	// GVRsToSync is a list of GroupVersionResource that will be synced.
	// If GVRsToSync is nil, defaultGVRs are used.
	GVRsToSync []schema.GroupVersionResource
}

func New(srcDynamicClient dynamic.Interface, resourceApplierService *resourceapplier.Service, options Options) *Service {
	s := &Service{
		gvrs:                   DefaultGVRs,
		srcDynamicClient:       srcDynamicClient,
		resourceApplierService: resourceApplierService,
	}

	if options.GVRsToSync != nil {
		s.gvrs = options.GVRsToSync
	}

	return s
}

func (s *Service) Run(ctx context.Context) error {
	klog.Info("Starting the cluster resource importer")

	infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(s.srcDynamicClient, 0, metav1.NamespaceAll, nil)
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

func (s *Service) addFunc(obj interface{}) {
	ctx := context.Background()
	unstructObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		klog.Error("Failed to convert runtime.Object to *unstructured.Unstructured")
		return
	}

	err := s.resourceApplierService.Create(ctx, unstructObj)
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

	err := s.resourceApplierService.Update(ctx, unstructObj)
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

	err := s.resourceApplierService.Delete(ctx, unstructObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// We just ignore the not found error because the scheduler may preempt the Pods, or users may remove the resources for debugging.
			klog.Info("Skipped to delete resource on destination: ", err)
		} else {
			klog.ErrorS(err, "Failed to delete resource on destination cluster")
		}
	}
}
