package di

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	configcorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	storageconfigv1 "k8s.io/client-go/applyconfigurations/storage/v1"
	configv1 "k8s.io/kube-scheduler/config/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/snapshot"
)

// PodService represents service for manage Pods.
type PodService interface {
	Get(ctx context.Context, name string, namespace string) (*corev1.Pod, error)
	List(ctx context.Context, namespace string) (*corev1.PodList, error)
	Apply(ctx context.Context, namespace string, pod *configcorev1.PodApplyConfiguration) (*corev1.Pod, error)
	Delete(ctx context.Context, name string, namespace string) error
}

// NodeService represents service for manage Nodes.
type NodeService interface {
	Get(ctx context.Context, name string) (*corev1.Node, error)
	List(ctx context.Context) (*corev1.NodeList, error)
	Apply(ctx context.Context, node *configcorev1.NodeApplyConfiguration) (*corev1.Node, error)
	Delete(ctx context.Context, name string) error
}

// PersistentVolumeService represents service for manage Pods.
type PersistentVolumeService interface {
	Get(ctx context.Context, name string) (*corev1.PersistentVolume, error)
	List(ctx context.Context) (*corev1.PersistentVolumeList, error)
	Apply(ctx context.Context, pv *configcorev1.PersistentVolumeApplyConfiguration) (*corev1.PersistentVolume, error)
	Delete(ctx context.Context, name string) error
}

// PersistentVolumeClaimService represents service for manage Nodes.
type PersistentVolumeClaimService interface {
	Get(ctx context.Context, name string, namespace string) (*corev1.PersistentVolumeClaim, error)
	List(ctx context.Context, namespace string) (*corev1.PersistentVolumeClaimList, error)
	Apply(ctx context.Context, namespace string, pvc *configcorev1.PersistentVolumeClaimApplyConfiguration) (*corev1.PersistentVolumeClaim, error)
	Delete(ctx context.Context, name string, namespace string) error
}

// StorageClassService represents service for manage Pods.
type StorageClassService interface {
	Get(ctx context.Context, name string) (*storagev1.StorageClass, error)
	List(ctx context.Context) (*storagev1.StorageClassList, error)
	Apply(ctx context.Context, sc *storageconfigv1.StorageClassApplyConfiguration) (*storagev1.StorageClass, error)
	Delete(ctx context.Context, name string) error
}

// SchedulerService represents service for manage scheduler.
type SchedulerService interface {
	GetSchedulerConfig() (*configv1.KubeSchedulerConfiguration, error)
	RestartScheduler(cfg *configv1.KubeSchedulerConfiguration) error
	StartScheduler(cfg *configv1.KubeSchedulerConfiguration) error
	ResetScheduler() error
	ShutdownScheduler()
	ExtenderService() scheduler.ExtenderService
}

// PriorityClassService represents service for manage scheduler.
type PriorityClassService interface {
	Get(ctx context.Context, name string) (*v1.PriorityClass, error)
	List(ctx context.Context) (*v1.PriorityClassList, error)
	Apply(ctx context.Context, priorityClass *schedulingv1.PriorityClassApplyConfiguration) (*v1.PriorityClass, error)
	Delete(ctx context.Context, name string) error
}

// SnapshotService represents a service for exporting/importing resources on the simulator.
type SnapshotService interface {
	Snap(ctx context.Context, opts ...snapshot.Option) (*snapshot.ResourcesForSnap, error)
	Load(ctx context.Context, resources *snapshot.ResourcesForLoad, opts ...snapshot.Option) error
	IgnoreErr() snapshot.Option
}

type ResetService interface {
	Reset(ctx context.Context) error
}

// ImportClusterResourceService represents a service to import resources from an target cluster.
type ImportClusterResourceService interface {
	ImportClusterResources(ctx context.Context) error
}

// ResourceWatcherService represents service for watch k8s resources.
type ResourceWatcherService interface {
	ListWatch(ctx context.Context, stream streamwriter.ResponseStream, lrVersions *resourcewatcher.LastResourceVersions) error
}

// ExtenderService represents service for the extender of scheduler.
type ExtenderService interface {
	Filter(id int, args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
	Prioritize(id int, args extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error)
	Preempt(id int, args extenderv1.ExtenderPreemptionArgs) (*extenderv1.ExtenderPreemptionResult, error)
	Bind(id int, args extenderv1.ExtenderBindingArgs) (*extenderv1.ExtenderBindingResult, error)
}
