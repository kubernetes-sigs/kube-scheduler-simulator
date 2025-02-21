package replayer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/recorder"
)

type Service struct {
	applier   ResourceApplier
	recordDir string
}

type ResourceApplier interface {
	Create(ctx context.Context, resource *unstructured.Unstructured) error
	Update(ctx context.Context, resource *unstructured.Unstructured) error
	Delete(ctx context.Context, resource *unstructured.Unstructured) error
}

type Options struct {
	RecordDir string
}

func New(applier ResourceApplier, options Options) *Service {
	return &Service{applier: applier, recordDir: options.RecordDir}
}

func (s *Service) Replay(ctx context.Context) error {
	files, err := os.ReadDir(s.recordDir)
	if err != nil {
		return xerrors.Errorf("failed to read record directory: %w", err)
	}

	records := []recorder.Record{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		recordsToAppend, err := s.loadRecordFromFile(file.Name())
		if err != nil {
			return xerrors.Errorf("failed to load record from file: %w", err)
		}

		records = append(records, recordsToAppend...)
	}

	for _, record := range records {
		if err := s.applyEvent(ctx, record); err != nil {
			return xerrors.Errorf("failed to replay event: %w", err)
		}
	}

	return nil
}

func (s *Service) loadRecordFromFile(path string) ([]recorder.Record, error) {
	filePath := filepath.Join(s.recordDir, path)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, xerrors.Errorf("failed to read file %s: %w", filePath, err)
	}

	var records []recorder.Record
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal records from %s: %w", filePath, err)
	}

	return records, nil
}

func (s *Service) applyEvent(ctx context.Context, record recorder.Record) error {
	switch record.Event {
	case recorder.Add:
		if err := s.applier.Create(ctx, &record.Resource); err != nil {
			if errors.IsAlreadyExists(err) {
				klog.Warningf("resource already exists: %v", err)
			} else {
				return xerrors.Errorf("failed to create resource: %w", err)
			}
		}
	case recorder.Update:
		if err := s.applier.Update(ctx, &record.Resource); err != nil {
			return xerrors.Errorf("failed to update resource: %w", err)
		}
	case recorder.Delete:
		if err := s.applier.Delete(ctx, &record.Resource); err != nil {
			return xerrors.Errorf("failed to delete resource: %w", err)
		}
	default:
		return xerrors.Errorf("unknown event: %v", record.Event)
	}

	return nil
}
