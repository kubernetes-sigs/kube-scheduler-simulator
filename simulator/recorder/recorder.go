package recorder

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type Event string

var (
	Add    Event = "Add"
	Update Event = "Update"
	Delete Event = "Delete"
)

const defaultRecordBatchCapacity = 1000

type Service struct {
	client              dynamic.Interface
	gvrs                []schema.GroupVersionResource
	path                string
	recordCh            chan Record
	recordBatchCapacity int
}

type Record struct {
	Time     time.Time                 `json:"time"`
	Event    Event                     `json:"event"`
	Resource unstructured.Unstructured `json:"resource"`
}

var DefaultGVRs = []schema.GroupVersionResource{
	{Group: "", Version: "v1", Resource: "namespaces"},
	{Group: "scheduling.k8s.io", Version: "v1", Resource: "priorityclasses"},
	{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	{Group: "", Version: "v1", Resource: "nodes"},
	{Group: "", Version: "v1", Resource: "persistentvolumes"},
	{Group: "", Version: "v1", Resource: "pods"},
}

type Options struct {
	GVRs                []schema.GroupVersionResource
	RecordFile          string
	RecordBatchCapacity *int
}

func New(client dynamic.Interface, options Options) *Service {
	gvrs := DefaultGVRs
	if options.GVRs != nil {
		gvrs = options.GVRs
	}

	recordBatchCapacity := defaultRecordBatchCapacity
	if options.RecordBatchCapacity != nil {
		recordBatchCapacity = *options.RecordBatchCapacity
	}

	return &Service{
		client:              client,
		gvrs:                gvrs,
		path:                options.RecordFile,
		recordCh:            make(chan Record, recordBatchCapacity),
		recordBatchCapacity: recordBatchCapacity,
	}
}

func (s *Service) Run(ctx context.Context) error {
	go s.record(ctx)

	infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(s.client, 0, metav1.NamespaceAll, nil)
	for _, gvr := range s.gvrs {
		inf := infFact.ForResource(gvr).Informer()
		_, err := inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { s.recordEvent(obj, Add) },
			UpdateFunc: func(_, obj interface{}) { s.recordEvent(obj, Update) },
			DeleteFunc: func(obj interface{}) { s.recordEvent(obj, Delete) },
		})
		if err != nil {
			return xerrors.Errorf("failed to add event handler: %w", err)
		}
		infFact.Start(ctx.Done())
		infFact.WaitForCacheSync(ctx.Done())
	}

	return nil
}

func (s *Service) recordEvent(obj interface{}, e Event) {
	unstructObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		klog.Error("Failed to convert runtime.Object to *unstructured.Unstructured")
		return
	}

	// we only need name and namespace for DELETE events
	if e == Delete {
		unstructObj = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": unstructObj.GetAPIVersion(),
				"kind":       unstructObj.GetKind(),
				"metadata": map[string]interface{}{
					"name":      unstructObj.GetName(),
					"namespace": unstructObj.GetNamespace(),
				},
			},
		}
	}

	r := Record{
		Event:    e,
		Time:     time.Now(),
		Resource: *unstructObj,
	}

	s.recordCh <- r
}

func (s *Service) record(ctx context.Context) {
	for {
		select {
		case r := <-s.recordCh:
			if err := appendToFile(s.path, r); err != nil {
				klog.Errorf("failed to append record to file: %v", err)
			}

		case <-ctx.Done():
			// flush the buffer
			for len(s.recordCh) > 0 {
				r := <-s.recordCh
				if err := appendToFile(s.path, r); err != nil {
					klog.Errorf("failed to append record to file: %v", err)
				}
			}

			return
		}
	}
}

func appendToFile(filePath string, record Record) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return xerrors.Errorf("failed to create record file: %w", err)
	}
	defer file.Close()

	b, err := json.Marshal(&record)
	if err != nil {
		return xerrors.Errorf("failed to marshal record: %w", err)
	}

	if _, err := file.Write(append(b, '\n')); err != nil {
		return xerrors.Errorf("failed to write record: %w", err)
	}

	return nil
}
