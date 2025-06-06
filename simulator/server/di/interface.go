package di

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	configv1 "k8s.io/kube-scheduler/config/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/snapshot"
)

// SchedulerService represents service for manage scheduler.
type SchedulerService interface {
	GetSchedulerConfig() (*configv1.KubeSchedulerConfiguration, error)
	SetSchedulerConfig(cfg *configv1.KubeSchedulerConfiguration)
	RestartScheduler(cfg *configv1.KubeSchedulerConfiguration) error
	ResetScheduler() error
	ShutdownScheduler()
	ExtenderService() scheduler.ExtenderService
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

// OneShotClusterResourceImporter represents a service to import resources from a target cluster when starting the simulator.
type OneShotClusterResourceImporter interface {
	ImportClusterResources(ctx context.Context, labelSelector metav1.LabelSelector) error
}

// ResourceSyncer represents a service to constantly sync resources from a target cluster.
type ResourceSyncer interface {
	// Run starts the resource syncer.
	// It should be run until the context is canceled.
	Run(ctx context.Context) error
}

// RecorderService represents a service to record events in a target cluster.
type RecorderService interface {
	// Run starts the recorder.
	// It should be run until the context is canceled.
	Run(ctx context.Context) error
}

// ReplayService represents a service to replay events recorded in a file.
type ReplayService interface {
	// Replay replays the recorded events.
	// It should be run until the context is canceled.
	Replay(ctx context.Context) error
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
