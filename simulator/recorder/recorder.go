package recorder

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
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

type Service struct {
	client   dynamic.Interface
	gvrs     []schema.GroupVersionResource
	path     string
	recordCh chan Record
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
	GVRs []schema.GroupVersionResource
	Path string
}

func New(client dynamic.Interface, options Options) *Service {
	gvrs := DefaultGVRs
	if options.GVRs != nil {
		gvrs = options.GVRs
	}

	return &Service{
		client:   client,
		gvrs:     gvrs,
		path:     options.Path,
		recordCh: make(chan Record),
	}
}

func (s *Service) Run(ctx context.Context) error {
	infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(s.client, 0, metav1.NamespaceAll, nil)
	for _, gvr := range s.gvrs {
		inf := infFact.ForResource(gvr).Informer()
		_, err := inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { s.RecordEvent(obj, Add) },
			UpdateFunc: func(_, obj interface{}) { s.RecordEvent(obj, Update) },
			DeleteFunc: func(obj interface{}) { s.RecordEvent(obj, Delete) },
		})
		if err != nil {
			return xerrors.Errorf("failed to add event handler: %w", err)
		}
		infFact.Start(ctx.Done())
		infFact.WaitForCacheSync(ctx.Done())
	}

	err := s.backup()
	if err != nil {
		klog.Error("Failed to backup record: ", err)
	}

	go s.record(ctx)

	return nil
}

func (s *Service) RecordEvent(obj interface{}, e Event) {
	unstructObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		klog.Error("Failed to convert runtime.Object to *unstructured.Unstructured")
		return
	}

	r := Record{
		Event:    e,
		Time:     time.Now(),
		Resource: *unstructObj,
	}

	s.recordCh <- r
}

func (s *Service) record(ctx context.Context) {
	records := []Record{}
	for {
		select {
		case r := <-s.recordCh:
			records = append(records, r)

			tempFile, err := os.CreateTemp("", "record_temp.json")
			if err != nil {
				klog.Error("Failed to open file: ", err)
				continue
			}

			b, err := json.Marshal(records)
			if err != nil {
				klog.Error("Failed to marshal record: ", err)
				continue
			}

			_, err = tempFile.Write(b)
			if err != nil {
				klog.Error("Failed to write record: ", err)
				continue
			}

			tempFile.Close()

			if err = moveFile(tempFile.Name(), s.path); err != nil {
				klog.Error("Failed to rename file: ", err)
				continue
			}

		case <-ctx.Done():
			return

		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *Service) backup() error {
	f, err := os.Stat(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return xerrors.Errorf("failed to get file info: %w", err)
	}

	if f.IsDir() {
		return xerrors.New("record file is a directory")
	}

	name := filepath.Base(s.path)
	dir := filepath.Dir(s.path)
	backupPath := filepath.Join(dir, f.ModTime().Format("2006-01-02_150405_")+name)
	return moveFile(s.path, backupPath)
}

func moveFile(srcPath, destPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return os.Remove(srcPath)
}
