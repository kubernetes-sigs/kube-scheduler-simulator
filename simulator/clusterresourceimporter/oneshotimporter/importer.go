package oneshotimporter

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
	// Snap will be used to export resources from target cluster.
	Snap(ctx context.Context, opts ...snapshot.Option) (*snapshot.ResourcesForSnap, error)
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
// If you want to use the scheduler configuration along with the imported resources on the simulator,
// you need to set the path of the scheduler configuration file to `kubeSchedulerConfigPath` value in the Simulator Server Configuration.
func (s *Service) ImportClusterResources(ctx context.Context) error {
	expRes, err := s.exportService.Snap(ctx)
	if err != nil {
		return xerrors.Errorf("call Snap of the exportService: %w", err)
	}
	impRes, err := snapshot.ConvertResourcesForSnapToResourcesForLoad(expRes)
	if err != nil {
		return xerrors.Errorf("call ConvertResourcesForSnapToResourcesForLoad: %w", err)
	}
	// Import to the simulator.
	if err := s.importService.Load(ctx, impRes, s.importService.IgnoreErr(), s.importService.IgnoreSchedulerConfiguration()); err != nil {
		return xerrors.Errorf("call Import of the importService: %w", err)
	}
	return nil
}
