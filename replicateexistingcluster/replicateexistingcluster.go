package replicateexistingcluster

//go:generate mockgen -destination=./mock_$GOPACKAGE/export.go . ExportService

import (
	"context"
	"encoding/json"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	cfgstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"

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
func (s *Service) ImportFromExistingCluster() error {
	ctx := context.Background()
	expRes, err := s.existingClusterExportService.Export(ctx)
	if err != nil {
		return xerrors.Errorf("call Export of existingClusterExportService: %w", err)
	}
	impRes, err := cnvtToResourcesForImportFromResourcesForExport(expRes)
	if err != nil {
		return xerrors.Errorf("call cnvtToResourcesForImportFromResourcesForExport: %w", err)
	}
	// Import to the simulator.
	if err := s.simulatorExportService.Import(ctx, impRes, s.simulatorExportService.IgnoreErr(), s.simulatorExportService.IgnoreSchedulerConfiguration()); err != nil {
		return xerrors.Errorf("call Import of the simulater export service: %w", err)
	}
	return nil
}

func cnvtToResourcesForImportFromResourcesForExport(expRes *export.ResourcesForExport) (*export.ResourcesForImport, error) {
	pods, err := cnvtPodListToApplyConfigurationList(expRes.Pods)
	if err != nil {
		return nil, xerrors.Errorf("call cnvtPodListToApplyConfigurationList: %w", err)
	}
	nodes, err := cnvtNodeListToApplyConfigurationList(expRes.Nodes)
	if err != nil {
		return nil, xerrors.Errorf("call cnvtNodeListToApplyConfigurationList: %w", err)
	}
	pvs, err := cnvtPvListToApplyConfigurationList(expRes.Pvs)
	if err != nil {
		return nil, xerrors.Errorf("call cnvtPvListToApplyConfigurationList: %w", err)
	}
	pvcs, err := cnvtPvcListToApplyConfigurationList(expRes.Pvcs)
	if err != nil {
		return nil, xerrors.Errorf("call cnvtPvcListToApplyConfigurationList: %w", err)
	}
	scs, err := cnvtStorageClassesListToApplyConfigurationList(expRes.StorageClasses)
	if err != nil {
		return nil, xerrors.Errorf("call cnvtStorageClassesListToApplyConfigurationList: %w", err)
	}
	pcs, err := cnvtPriorityClassesListToApplyConfigurationList(expRes.PriorityClasses)
	if err != nil {
		return nil, xerrors.Errorf("call cnvtPriorityClassesListToApplyConfigurationList: %w", err)
	}
	return &export.ResourcesForImport{
		Pods:            pods,
		Nodes:           nodes,
		Pvs:             pvs,
		Pvcs:            pvcs,
		StorageClasses:  scs,
		PriorityClasses: pcs,
		// existingClusterExportService can't export the SchedulerConfig.
		SchedulerConfig: nil,
	}, nil
}

func cnvtPodListToApplyConfigurationList(pods []corev1.Pod) ([]v1.PodApplyConfiguration, error) {
	rto := make([]v1.PodApplyConfiguration, len(pods))
	for i, p := range pods {
		if err := convertToApplyConfiguration(p, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert Pod to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func cnvtNodeListToApplyConfigurationList(nodes []corev1.Node) ([]v1.NodeApplyConfiguration, error) {
	rto := make([]v1.NodeApplyConfiguration, len(nodes))
	for i, n := range nodes {
		if err := convertToApplyConfiguration(n, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert Node to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func cnvtPvListToApplyConfigurationList(pvs []corev1.PersistentVolume) ([]v1.PersistentVolumeApplyConfiguration, error) {
	rto := make([]v1.PersistentVolumeApplyConfiguration, len(pvs))
	for i, p := range pvs {
		if err := convertToApplyConfiguration(p, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert PersistentVolume to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func cnvtPvcListToApplyConfigurationList(pvcs []corev1.PersistentVolumeClaim) ([]v1.PersistentVolumeClaimApplyConfiguration, error) {
	rto := make([]v1.PersistentVolumeClaimApplyConfiguration, len(pvcs))
	for i, p := range pvcs {
		if err := convertToApplyConfiguration(p, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert PersistentVolumeClaim to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func cnvtStorageClassesListToApplyConfigurationList(scs []storagev1.StorageClass) ([]cfgstoragev1.StorageClassApplyConfiguration, error) {
	rto := make([]cfgstoragev1.StorageClassApplyConfiguration, len(scs))
	for i, s := range scs {
		if err := convertToApplyConfiguration(s, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert StorageClass to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func cnvtPriorityClassesListToApplyConfigurationList(pcs []schedulingv1.PriorityClass) ([]schedulingcfgv1.PriorityClassApplyConfiguration, error) {
	rto := make([]schedulingcfgv1.PriorityClassApplyConfiguration, len(pcs))
	for i, p := range pcs {
		if err := convertToApplyConfiguration(p, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert PriorityClasses to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func convertToApplyConfiguration(in interface{}, out interface{}) error {
	_in, err := json.Marshal(in)
	if err != nil {
		return xerrors.Errorf("call Marshal to cnvt object: %w", err)
	}
	switch in.(type) {
	case corev1.Pod:
		typedout, ok := out.(*v1.PodApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to cnvt Pod: %w", err)
		}
		return nil
	case corev1.Node:
		typedout, ok := out.(*v1.NodeApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to cnvt Node: %w", err)
		}
		return nil
	case corev1.PersistentVolume:
		typedout, ok := out.(*v1.PersistentVolumeApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to cnvt PersistentVolume: %w", err)
		}
		return nil
	case corev1.PersistentVolumeClaim:
		typedout, ok := out.(*v1.PersistentVolumeClaimApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to cnvt PersistentVolumeClaim: %w", err)
		}
		return nil
	case storagev1.StorageClass:
		typedout, ok := out.(*cfgstoragev1.StorageClassApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to cnvt StorageClass: %w", err)
		}
		return nil
	case schedulingv1.PriorityClass:
		typedout, ok := out.(*schedulingcfgv1.PriorityClassApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to cnvt PriorityClass: %w", err)
		}
		return nil
	default:
		return xerrors.Errorf("unknown type")
	}
}
