package replicateexistingcluster

//go:generate mockgen -destination=./mock_$GOPACKAGE/export.go . ExportService

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
)

// Service has two exportServices.
// existingClusterExportService is used to export resources from an existing cluster.
// simulatorExportService is used to import(replicate) the resources to the simulator.
type Service struct {
	existingClusterExportService ExportService
	simulatorExportService       ExportService
}

type ExportService interface {
	Export(ctx context.Context, opts ...export.Option) (*export.ResourcesForExport, error)
	Import(ctx context.Context, resources *export.ResourcesForImport, opts ...export.Option) error
	IgnoreErr() export.Option
	IgnoreSchedulerConfiguration() export.Option
}

// NewReplicateExistingClusterService initializes Service.
func NewReplicateExistingClusterService(exportService ExportService, existingClusterExportService ExportService) *Service {
	return &Service{
		existingClusterExportService: existingClusterExportService,
		simulatorExportService:       exportService,
	}
}

// ImportFromExistingCluster gets resources from existing cluster via existingClusterExportService
// and then apply those resources to the simulator.
// Note: this method doesn't handle scheduler configuration.
// If users want to use their scheduler configuration, they need to use `KUBE_SCHEDULER_CONFIG_PATH` env.
func (s *Service) ImportFromExistingCluster(ctx context.Context) error {
	expRes, err := s.existingClusterExportService.Export(ctx)
	if err != nil {
		return xerrors.Errorf("call Export of existingClusterExportService: %w", err)
	}
	impRes, err := export.ConvertResourcesForImportToResourcesForExport(expRes)
	if err != nil {
		return xerrors.Errorf("call ConvertResourcesForImportToResourcesForExport: %w", err)
	}
	// Import to the simulator.
	if err := s.simulatorExportService.Import(ctx, impRes, s.simulatorExportService.IgnoreErr(), s.simulatorExportService.IgnoreSchedulerConfiguration()); err != nil {
		return xerrors.Errorf("call Import of the simulater export service: %w", err)
	}
	return nil
}
