package clusterresourceimporter

//go:generate mockgen -destination=./mock_$GOPACKAGE/replicate.go . ReplicateService

import (
	"context"

	"golang.org/x/xerrors"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/snapshot"
)

// Service has two ReplicateServices.
// importService is used to import(replicate) these resources to the simulator.
// exportService is used to export resources from a target cluster.
type Service struct {
	importService ReplicateService
	exportService ReplicateService
}

type ReplicateService interface {
	// Save will be used to export resources from target cluster.
	Save(ctx context.Context, opts ...snapshot.Option) (*snapshot.ResourcesForSave, error)
	// Load will be used to import resources the from data which was exported.
	Load(ctx context.Context, resources *snapshot.ResourcesForLoad, opts ...snapshot.Option) error
	IgnoreErr() snapshot.Option
	IgnoreSchedulerConfiguration() snapshot.Option
}

// NewService initializes Service.
func NewService(e ReplicateService, i ReplicateService) *Service {
	return &Service{
		importService: e,
		exportService: i,
	}
}

// ImportClusterResources gets resources from the target cluster via exportService
// and then apply those resources to the simulator.
// Note: this method doesn't handle scheduler configuration.
// If you want to use their scheduler configuration, you need to set config of `KUBE_SCHEDULER_CONFIG_PATH`.
func (s *Service) ImportClusterResources(ctx context.Context) error {
	expRes, err := s.exportService.Save(ctx)
	if err != nil {
		return xerrors.Errorf("call Save of the exportService: %w", err)
	}
	impRes, err := snapshot.ConvertResourcesForSaveToResourcesForLoad(expRes)
	if err != nil {
		return xerrors.Errorf("call ConvertResourcesForSaveToResourcesForLoad: %w", err)
	}
	// Import to the simulator.
	if err := s.importService.Load(ctx, impRes, s.importService.IgnoreErr(), s.importService.IgnoreSchedulerConfiguration()); err != nil {
		return xerrors.Errorf("call Import of the importService: %w", err)
	}
	return nil
}
