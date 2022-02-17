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
	IgnoreRestart() export.Option
}

// NewReplicateExistingClusterService initializes Service.
func NewReplicateExistingClusterService(exportService ExportService, existingClusterExportService ExportService) *Service {
	return &Service{
		existingClusterExportService: existingClusterExportService,
		simulatorExportService:       exportService,
	}
}

// ImportFromExistingCluster get resources from existing cluster via existingClusterExportService
// and then apply those to the simulator without restarting it via simulatorExportService.
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
	// Import to the simulator without restarting it.
	if err := s.simulatorExportService.Import(ctx, impRes, s.simulatorExportService.IgnoreErr(), s.simulatorExportService.IgnoreRestart()); err != nil {
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
	plist := []v1.PodApplyConfiguration{}

	for _, p := range pods {
		_p, err := json.Marshal(p)
		if err != nil {
			return nil, xerrors.Errorf("call Marshal to cnvt Pod: %w", err)
		}
		pod := &v1.PodApplyConfiguration{}
		if err := json.Unmarshal(_p, &pod); err != nil {
			return nil, xerrors.Errorf("call Unmarshal to cnvt to PodApplyConfiguration: %w", err)
		}
		plist = append(plist, *pod)
	}
	return plist, nil
}

func cnvtNodeListToApplyConfigurationList(nodes []corev1.Node) ([]v1.NodeApplyConfiguration, error) {
	nlist := []v1.NodeApplyConfiguration{}
	for _, n := range nodes {
		_n, err := json.Marshal(n)
		if err != nil {
			return nil, xerrors.Errorf("call Marshal to cnvt Node: %w", err)
		}
		var node v1.NodeApplyConfiguration
		if err := json.Unmarshal(_n, &node); err != nil {
			return nil, xerrors.Errorf("call Unmarshal to cnvt to NodeApplyConfiguration: %w", err)
		}
		nlist = append(nlist, node)
	}
	return nlist, nil
}

func cnvtPvListToApplyConfigurationList(pvs []corev1.PersistentVolume) ([]v1.PersistentVolumeApplyConfiguration, error) {
	pvlist := []v1.PersistentVolumeApplyConfiguration{}
	for _, p := range pvs {
		_p, err := json.Marshal(p)
		if err != nil {
			return nil, xerrors.Errorf("call Marshal to cnvt PersistentVolume: %w", err)
		}
		var pv v1.PersistentVolumeApplyConfiguration
		if err := json.Unmarshal(_p, &pv); err != nil {
			return nil, xerrors.Errorf("call Unmarshal to cnvt to PersistentVolumeApplyConfiguration: %w", err)
		}
		pvlist = append(pvlist, pv)
	}
	return pvlist, nil
}

func cnvtPvcListToApplyConfigurationList(pvcs []corev1.PersistentVolumeClaim) ([]v1.PersistentVolumeClaimApplyConfiguration, error) {
	pvclist := []v1.PersistentVolumeClaimApplyConfiguration{}
	for _, p := range pvcs {
		_p, err := json.Marshal(p)
		if err != nil {
			return nil, xerrors.Errorf("call Marshal to cnvt PersistentVolumeClaim: %w", err)
		}
		var pvc v1.PersistentVolumeClaimApplyConfiguration
		if err := json.Unmarshal(_p, &pvc); err != nil {
			return nil, xerrors.Errorf("call Unmarshal to cnvt to PersistentVolumeClaimApplyConfiguration: %w", err)
		}
		pvclist = append(pvclist, pvc)
	}
	return pvclist, nil
}

func cnvtStorageClassesListToApplyConfigurationList(scs []storagev1.StorageClass) ([]cfgstoragev1.StorageClassApplyConfiguration, error) {
	sclist := []cfgstoragev1.StorageClassApplyConfiguration{}
	for _, s := range scs {
		_s, err := json.Marshal(s)
		if err != nil {
			return nil, xerrors.Errorf("call Marshal to cnvt StorageClass: %w", err)
		}
		var sc cfgstoragev1.StorageClassApplyConfiguration
		if err := json.Unmarshal(_s, &sc); err != nil {
			return nil, xerrors.Errorf("call Unmarshal to cnvt to StorageClassApplyConfiguration: %w", err)
		}
		sclist = append(sclist, sc)
	}
	return sclist, nil
}

func cnvtPriorityClassesListToApplyConfigurationList(pcs []schedulingv1.PriorityClass) ([]schedulingcfgv1.PriorityClassApplyConfiguration, error) {
	pclist := []schedulingcfgv1.PriorityClassApplyConfiguration{}
	for _, p := range pcs {
		_p, err := json.Marshal(p)
		if err != nil {
			return nil, xerrors.Errorf("call Marshal to cnvt StorageClass: %w", err)
		}
		var pc schedulingcfgv1.PriorityClassApplyConfiguration
		if err := json.Unmarshal(_p, &pc); err != nil {
			return nil, xerrors.Errorf("call Unmarshal to cnvt to StorageClassApplyConfiguration: %w", err)
		}
		pclist = append(pclist, pc)
	}
	return pclist, nil
}
