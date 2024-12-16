package replayer

import (
	"context"
	"encoding/json"
	"os"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/recorder"
)

type Service struct {
	applier ResourceApplier
	path    string
}

type ResourceApplier interface {
	Create(ctx context.Context, resource *unstructured.Unstructured) error
	Update(ctx context.Context, resource *unstructured.Unstructured) error
	Delete(ctx context.Context, resource *unstructured.Unstructured) error
}

type Options struct {
	Path string
}

func New(applier ResourceApplier, options Options) *Service {
	return &Service{applier: applier, path: options.Path}
}

func (s *Service) Replay(ctx context.Context) error {
	records := []recorder.Record{}

	b, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &records); err != nil {
		return err
	}

	for _, record := range records {
		if err := s.applyEvent(ctx, record); err != nil {
			return xerrors.Errorf("failed to replay event: %w", err)
		}
	}

	return nil
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
