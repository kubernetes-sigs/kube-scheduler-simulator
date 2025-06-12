package recorder

import (
	"context"
	"encoding/json"
	"os"
	"sync"
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

const defaultPollInterval = 5 * time.Second

type Service struct {
	client       dynamic.Interface
	gvrs         []schema.GroupVersionResource
	path         string
	records      []Record
	recordsMutex sync.Mutex
	pollInterval time.Duration
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
	GVRs          []schema.GroupVersionResource
	RecordFile    string
	FlushInterval *time.Duration
}

func New(client dynamic.Interface, options Options) *Service {
	gvrs := DefaultGVRs
	if options.GVRs != nil {
		gvrs = options.GVRs
	}

	pollInterval := defaultPollInterval
	if options.FlushInterval != nil {
		pollInterval = *options.FlushInterval
	}

	return &Service{
		client:       client,
		gvrs:         gvrs,
		path:         options.RecordFile,
		records:      make([]Record, 0),
		recordsMutex: sync.Mutex{},
		pollInterval: pollInterval,
	}
}

func (s *Service) Run(ctx context.Context) error {
	// create or recreate the file
	f, err := os.Create(s.path)
	if err != nil {
		return xerrors.Errorf("failed to create record file: %w", err)
	}

	go s.record(ctx, f)

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

	s.recordsMutex.Lock()
	s.records = append(s.records, r)
	s.recordsMutex.Unlock()
}

func (s *Service) record(ctx context.Context, file *os.File) {
	defer file.Close()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := s.flushRecords(file); err != nil {
				klog.Errorf("failed to flush records: %v", err)
			}
			return
		case <-ticker.C:
			if err := s.flushRecords(file); err != nil {
				klog.Errorf("failed to flush records: %v", err)
			}
		}
	}
}

func (s *Service) flushRecords(file *os.File) error {
	if len(s.records) == 0 {
		return nil
	}

	s.recordsMutex.Lock()
	records := s.records
	s.records = make([]Record, 0)
	s.recordsMutex.Unlock()

	if err := appendToFile(file, records); err != nil {
		return xerrors.Errorf("failed to append record to file: %w", err)
	}

	return nil
}

func appendToFile(file *os.File, records []Record) error {
	content := make([]byte, 0)
	for _, record := range records {
		b, err := json.Marshal(&record)
		if err != nil {
			return xerrors.Errorf("failed to marshal record: %w", err)
		}

		content = append(content, b...)
		content = append(content, '\n')
	}

	if _, err := file.Write(content); err != nil {
		return xerrors.Errorf("failed to write record: %w", err)
	}

	return nil
}
