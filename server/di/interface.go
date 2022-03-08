package di

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	configv1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	storageconfigv1 "k8s.io/client-go/applyconfigurations/storage/v1"
	v1beta3config "k8s.io/kube-scheduler/config/v1beta3"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
)

// PodService represents service for manage Pods.
type PodService interface {
	Get(ctx context.Context, name string) (*corev1.Pod, error)
	List(ctx context.Context) (*corev1.PodList, error)
	Apply(ctx context.Context, pod *configv1.PodApplyConfiguration) (*corev1.Pod, error)
	Delete(ctx context.Context, name string) error
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

// NodeService represents service for manage Nodes.
type NodeService interface {
	Get(ctx context.Context, name string) (*corev1.Node, error)
	List(ctx context.Context) (*corev1.NodeList, error)
	Apply(ctx context.Context, node *configv1.NodeApplyConfiguration) (*corev1.Node, error)
	Delete(ctx context.Context, name string) error
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

// PersistentVolumeService represents service for manage Pods.
type PersistentVolumeService interface {
	Get(ctx context.Context, name string) (*corev1.PersistentVolume, error)
	List(ctx context.Context) (*corev1.PersistentVolumeList, error)
	Apply(ctx context.Context, pv *configv1.PersistentVolumeApplyConfiguration) (*corev1.PersistentVolume, error)
	Delete(ctx context.Context, name string) error
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

// PersistentVolumeClaimService represents service for manage Nodes.
type PersistentVolumeClaimService interface {
	Get(ctx context.Context, name string) (*corev1.PersistentVolumeClaim, error)
	List(ctx context.Context) (*corev1.PersistentVolumeClaimList, error)
	Apply(ctx context.Context, pvc *configv1.PersistentVolumeClaimApplyConfiguration) (*corev1.PersistentVolumeClaim, error)
	Delete(ctx context.Context, name string) error
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

// StorageClassService represents service for manage Pods.
type StorageClassService interface {
	Get(ctx context.Context, name string) (*storagev1.StorageClass, error)
	List(ctx context.Context) (*storagev1.StorageClassList, error)
	Apply(ctx context.Context, sc *storageconfigv1.StorageClassApplyConfiguration) (*storagev1.StorageClass, error)
	Delete(ctx context.Context, name string) error
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

// SchedulerService represents service for manage scheduler.
type SchedulerService interface {
	GetSchedulerConfig() *v1beta3config.KubeSchedulerConfiguration
	RestartScheduler(cfg *v1beta3config.KubeSchedulerConfiguration) error
	StartScheduler(cfg *v1beta3config.KubeSchedulerConfiguration) error
	ResetScheduler() error
	ShutdownScheduler()
}

// PriorityClassService represents service for manage scheduler.
type PriorityClassService interface {
	Get(ctx context.Context, name string) (*v1.PriorityClass, error)
	List(ctx context.Context) (*v1.PriorityClassList, error)
	Apply(ctx context.Context, priorityClass *schedulingv1.PriorityClassApplyConfiguration) (*v1.PriorityClass, error)
	Delete(ctx context.Context, name string) error
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

type ExportService interface {
	Export(ctx context.Context, opts ...export.Option) (*export.ResourcesForExport, error)
	Import(ctx context.Context, resources *export.ResourcesForImport, opts ...export.Option) error
	IgnoreErr() export.Option
}

type ResetService interface {
	Reset(ctx context.Context) error
}

// ReplicateExistingClusterService represents a service to import resources from the existing cluster.
type ReplicateExistingClusterService interface {
	ImportFromExistingCluster(ctx context.Context) error
}
