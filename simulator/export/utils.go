package export

import (
	"encoding/json"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	cfgstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
)

func ConvertResourcesForImportToResourcesForExport(expRes *ResourcesForExport) (*ResourcesForImport, error) {
	pods, err := convertPodListToApplyConfigurationList(expRes.Pods)
	if err != nil {
		return nil, xerrors.Errorf("call convertPodListToApplyConfigurationList: %w", err)
	}
	nodes, err := convertNodeListToApplyConfigurationList(expRes.Nodes)
	if err != nil {
		return nil, xerrors.Errorf("call convertNodeListToApplyConfigurationList: %w", err)
	}
	pvs, err := convertPvListToApplyConfigurationList(expRes.Pvs)
	if err != nil {
		return nil, xerrors.Errorf("call convertPvListToApplyConfigurationList: %w", err)
	}
	pvcs, err := convertPvcListToApplyConfigurationList(expRes.Pvcs)
	if err != nil {
		return nil, xerrors.Errorf("call convertPvcListToApplyConfigurationList: %w", err)
	}
	scs, err := convertStorageClassesListToApplyConfigurationList(expRes.StorageClasses)
	if err != nil {
		return nil, xerrors.Errorf("call convertStorageClassesListToApplyConfigurationList: %w", err)
	}
	pcs, err := convertPriorityClassesListToApplyConfigurationList(expRes.PriorityClasses)
	if err != nil {
		return nil, xerrors.Errorf("call convertPriorityClassesListToApplyConfigurationList: %w", err)
	}
	return &ResourcesForImport{
		Pods:            pods,
		Nodes:           nodes,
		Pvs:             pvs,
		Pvcs:            pvcs,
		StorageClasses:  scs,
		PriorityClasses: pcs,
		SchedulerConfig: expRes.SchedulerConfig,
	}, nil
}

func convertPodListToApplyConfigurationList(pods []corev1.Pod) ([]v1.PodApplyConfiguration, error) {
	rto := make([]v1.PodApplyConfiguration, len(pods))
	for i, p := range pods {
		if err := convertToApplyConfiguration(p, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert Pod to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func convertNodeListToApplyConfigurationList(nodes []corev1.Node) ([]v1.NodeApplyConfiguration, error) {
	rto := make([]v1.NodeApplyConfiguration, len(nodes))
	for i, n := range nodes {
		if err := convertToApplyConfiguration(n, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert Node to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func convertPvListToApplyConfigurationList(pvs []corev1.PersistentVolume) ([]v1.PersistentVolumeApplyConfiguration, error) {
	rto := make([]v1.PersistentVolumeApplyConfiguration, len(pvs))
	for i, p := range pvs {
		if err := convertToApplyConfiguration(p, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert PersistentVolume to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func convertPvcListToApplyConfigurationList(pvcs []corev1.PersistentVolumeClaim) ([]v1.PersistentVolumeClaimApplyConfiguration, error) {
	rto := make([]v1.PersistentVolumeClaimApplyConfiguration, len(pvcs))
	for i, p := range pvcs {
		if err := convertToApplyConfiguration(p, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert PersistentVolumeClaim to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func convertStorageClassesListToApplyConfigurationList(scs []storagev1.StorageClass) ([]cfgstoragev1.StorageClassApplyConfiguration, error) {
	rto := make([]cfgstoragev1.StorageClassApplyConfiguration, len(scs))
	for i, s := range scs {
		if err := convertToApplyConfiguration(s, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert StorageClass to apply configuration: %w", err)
		}
	}
	return rto, nil
}

func convertPriorityClassesListToApplyConfigurationList(pcs []schedulingv1.PriorityClass) ([]schedulingcfgv1.PriorityClassApplyConfiguration, error) {
	rto := make([]schedulingcfgv1.PriorityClassApplyConfiguration, len(pcs))
	for i, p := range pcs {
		if err := convertToApplyConfiguration(p, &rto[i]); err != nil {
			return nil, xerrors.Errorf("convert PriorityClasses to apply configuration: %w", err)
		}
	}
	return rto, nil
}

// convertToApplyConfiguration is convert some object to XXXXApplyConfiguration.
// out should be the pointer of XXXXApplyConfiguration, otherwise, you can not get the result of conversion.
//nolint:funlen,cyclop // For readability.
func convertToApplyConfiguration(in interface{}, out interface{}) error {
	_in, err := json.Marshal(in)
	if err != nil {
		return xerrors.Errorf("call Marshal to convert object: %w", err)
	}
	switch in.(type) {
	case corev1.Pod:
		typedout, ok := out.(*v1.PodApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to convert Pod: %w", err)
		}
		return nil
	case corev1.Node:
		typedout, ok := out.(*v1.NodeApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to convert Node: %w", err)
		}
		return nil
	case corev1.PersistentVolume:
		typedout, ok := out.(*v1.PersistentVolumeApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to convert PersistentVolume: %w", err)
		}
		return nil
	case corev1.PersistentVolumeClaim:
		typedout, ok := out.(*v1.PersistentVolumeClaimApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to convert PersistentVolumeClaim: %w", err)
		}
		return nil
	case storagev1.StorageClass:
		typedout, ok := out.(*cfgstoragev1.StorageClassApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to convert StorageClass: %w", err)
		}
		return nil
	case schedulingv1.PriorityClass:
		typedout, ok := out.(*schedulingcfgv1.PriorityClassApplyConfiguration)
		if !ok {
			return xerrors.New("unexpected type was given as out")
		}
		if err := json.Unmarshal(_in, &typedout); err != nil {
			return xerrors.Errorf("call Unmarshal to convert PriorityClass: %w", err)
		}
		return nil
	default:
		return xerrors.Errorf("unknown type")
	}
}
