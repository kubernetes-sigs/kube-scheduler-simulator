package replayer

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/recorder"
)

type Service struct {
	applier    ResourceApplier
	recordFile string
}

type ResourceApplier interface {
	Create(ctx context.Context, resource *unstructured.Unstructured) error
	Update(ctx context.Context, resource *unstructured.Unstructured) error
	Delete(ctx context.Context, resource *unstructured.Unstructured) error
}

type Options struct {
	RecordFile string
}

func New(applier ResourceApplier, options Options) *Service {
	return &Service{applier: applier, recordFile: options.RecordFile}
}

func (s *Service) Replay(ctx context.Context) error {
	file, err := os.Open(s.recordFile)
	if err != nil {
		return xerrors.Errorf("failed to read record directory: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		record, err := s.loadRecordFromLine(reader)
		if err != nil {
			return xerrors.Errorf("failed to load record from line: %w", err)
		}
		if record == nil {
			break
		}

		if err := s.applyEvent(ctx, *record); err != nil {
			return xerrors.Errorf("failed to apply event: %w", err)
		}
	}

	return nil
}

func (s *Service) loadRecordFromLine(reader *bufio.Reader) (*recorder.Record, error) {
	line, err := reader.ReadBytes('\n')
	if len(line) == 0 || err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to read line: %w", err)
	}

	record := &recorder.Record{}
	if err := json.Unmarshal(line, record); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal record: %w", err)
	}

	return record, nil
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
