package recorder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
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
	RecordDir           string
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
		path:                options.RecordDir,
		recordCh:            make(chan Record, recordBatchCapacity),
		recordBatchCapacity: recordBatchCapacity,
	}
}

func (s *Service) Run(ctx context.Context) error {
	err := os.MkdirAll(s.path, 0o755)
	if err != nil {
		return xerrors.Errorf("failed to create record directory: %w", err)
	}

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
	records := make([]Record, 0, s.recordBatchCapacity)
	count := 0
	writeRecord := func() {
		filePath := path.Join(s.path, fmt.Sprintf("record-%018d.json", count))
		err := writeToFile(filePath, records)
		if err != nil {
			klog.Errorf("failed to write records to file: %v", err)
			return
		}

		count++
		records = make([]Record, 0, s.recordBatchCapacity)
	}

	for {
		select {
		case r := <-s.recordCh:
			records = append(records, r)
			if len(records) == s.recordBatchCapacity {
				writeRecord()
			}

		case <-ctx.Done():
			if len(records) > 0 {
				writeRecord()
			}
			return

		default:
			// flush the buffer
			if len(records) > 0 {
				writeRecord()
			}
		}

	}
}

func writeToFile(filePath string, records []Record) error {
	file, err := os.Create(filePath)
	if err != nil {
		return xerrors.Errorf("failed to create record file: %w", err)
	}
	defer file.Close()

	b, err := json.Marshal(records)
	if err != nil {
		return xerrors.Errorf("failed to marshal records: %w", err)
	}

	_, err = file.Write(b)
	if err != nil {
		return xerrors.Errorf("failed to write records: %w", err)
	}

	return nil
}
