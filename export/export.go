package export

//go:generate mockgen -destination=./mock_$GOPACKAGE/pod.go . PodService
//go:generate mockgen -destination=./mock_$GOPACKAGE/node.go . NodeService
//go:generate mockgen -destination=./mock_$GOPACKAGE/pv.go . PersistentVolumeService
//go:generate mockgen -destination=./mock_$GOPACKAGE/pvc.go . PersistentVolumeClaimService
//go:generate mockgen -destination=./mock_$GOPACKAGE/storageClassc.go . StorageClassService
//go:generate mockgen -destination=./mock_$GOPACKAGE/scheduler.go . SchedulerService
//go:generate mockgen -destination=./mock_$GOPACKAGE/priorityclass.go . PriorityClassService

import (
	"context"
	"runtime"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
)

type Service struct {
	client               clientset.Interface
	podService           PodService
	nodeService          NodeService
	pvService            PersistentVolumeService
	pvcService           PersistentVolumeClaimService
	storageClassService  StorageClassService
	priorityclassService PriorityClassService
	schedulerService     SchedulerService
}

// ResourcesForExport denotes all resources and scheduler configuration for export.
type ResourcesForExport struct {
	Pods            []corev1.Pod                              `json:"pods"`
	Nodes           []corev1.Node                             `json:"nodes"`
	Pvs             []corev1.PersistentVolume                 `json:"pvs"`
	Pvcs            []corev1.PersistentVolumeClaim            `json:"pvcs"`
	StorageClasses  []storagev1.StorageClass                  `json:"storageClasses"`
	PriorityClasses []schedulingv1.PriorityClass              `json:"priorityClasses"`
	SchedulerConfig *v1beta2config.KubeSchedulerConfiguration `json:"schedulerConfig"`
}

// ResourcesForImport denotes all resources and scheduler configuration for import.
type ResourcesForImport struct {
	Pods            []v1.PodApplyConfiguration                        `json:"pods"`
	Nodes           []v1.NodeApplyConfiguration                       `json:"nodes"`
	Pvs             []v1.PersistentVolumeApplyConfiguration           `json:"pvs"`
	Pvcs            []v1.PersistentVolumeClaimApplyConfiguration      `json:"pvcs"`
	StorageClasses  []confstoragev1.StorageClassApplyConfiguration    `json:"storageClasses"`
	PriorityClasses []schedulingcfgv1.PriorityClassApplyConfiguration `json:"priorityClasses"`
	SchedulerConfig *v1beta2config.KubeSchedulerConfiguration         `json:"schedulerConfig"`
}

type PodService interface {
	List(ctx context.Context) (*corev1.PodList, error)
	Apply(ctx context.Context, pod *v1.PodApplyConfiguration) (*corev1.Pod, error)
}

type NodeService interface {
	List(ctx context.Context) (*corev1.NodeList, error)
	Apply(ctx context.Context, nac *v1.NodeApplyConfiguration) (*corev1.Node, error)
}

type PersistentVolumeService interface {
	List(ctx context.Context) (*corev1.PersistentVolumeList, error)
	Apply(ctx context.Context, persistentVolume *v1.PersistentVolumeApplyConfiguration) (*corev1.PersistentVolume, error)
}

type PersistentVolumeClaimService interface {
	Get(ctx context.Context, name string) (*corev1.PersistentVolumeClaim, error)
	List(ctx context.Context) (*corev1.PersistentVolumeClaimList, error)
	Apply(ctx context.Context, persistentVolumeClaime *v1.PersistentVolumeClaimApplyConfiguration) (*corev1.PersistentVolumeClaim, error)
}

type StorageClassService interface {
	List(ctx context.Context) (*storagev1.StorageClassList, error)
	Apply(ctx context.Context, storageClass *confstoragev1.StorageClassApplyConfiguration) (*storagev1.StorageClass, error)
}

type PriorityClassService interface {
	List(ctx context.Context) (*schedulingv1.PriorityClassList, error)
	Apply(ctx context.Context, priorityClass *schedulingcfgv1.PriorityClassApplyConfiguration) (*schedulingv1.PriorityClass, error)
}

type SchedulerService interface {
	GetSchedulerConfig() *v1beta2config.KubeSchedulerConfiguration
	RestartScheduler(cfg *v1beta2config.KubeSchedulerConfiguration) error
}

func NewExportService(client clientset.Interface, pods PodService, nodes NodeService, pvs PersistentVolumeService, pvcs PersistentVolumeClaimService, storageClasss StorageClassService, priorityClasss PriorityClassService, schedulers SchedulerService) *Service {
	return &Service{
		client:               client,
		podService:           pods,
		nodeService:          nodes,
		pvService:            pvs,
		pvcService:           pvcs,
		storageClassService:  storageClasss,
		priorityclassService: priorityClasss,
		schedulerService:     schedulers,
	}
}

type ErrorGroupWithSemaphore struct {
	g   *errgroup.Group
	sem *semaphore.Weighted
}

func NewErrorGroupWithSemaphore(ctx context.Context) ErrorGroupWithSemaphore {
	g, _ := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))
	return ErrorGroupWithSemaphore{
		g:   g,
		sem: sem,
	}
}

// Get all resources from each service.
func (s *Service) get(ctx context.Context) (*ResourcesForExport, error) {
	errgrp := NewErrorGroupWithSemaphore(ctx)
	resources := ResourcesForExport{}

	errgrp.listPods(ctx, &s.podService, &resources)
	errgrp.listNodes(ctx, &s.nodeService, &resources)
	errgrp.listPvs(ctx, &s.pvService, &resources)
	errgrp.listPvcs(ctx, &s.pvcService, &resources)
	errgrp.listStorageClasses(ctx, &s.storageClassService, &resources)
	errgrp.listPcs(ctx, &s.priorityclassService, &resources)
	errgrp.getSchedulerConfig(ctx, &s.schedulerService, &resources)

	if err := errgrp.g.Wait(); err != nil {
		return nil, xerrors.Errorf("get resources all: %w", err)
	}
	return &resources, nil
}

func (s *Service) Export(ctx context.Context) (*ResourcesForExport, error) {
	resources, err := s.get(ctx)
	if err != nil {
		return nil, xerrors.Errorf("export resources all: %w", err)
	}
	return resources, nil
}

// Import all resources from posted data.
// (1) Restart scheduler based on the data.
// (2) Apply each resource to the scheduler.
//     * If UID is not nil, an error will occur. (try to find existing resource by UID)
// (3) Get all resources. (Separated the get function to unify the struct format.)
func (s *Service) Import(ctx context.Context, resources *ResourcesForImport) error {
	if err := s.schedulerService.RestartScheduler(resources.SchedulerConfig); err != nil {
		return xerrors.Errorf("restart scheduler with imported configuration: %w", err)
	}
	errgrp := NewErrorGroupWithSemaphore(ctx)
	errgrp.applyPcs(ctx, &s.priorityclassService, resources)
	errgrp.applyStorageClasses(ctx, &s.storageClassService, resources)
	errgrp.applyPvcs(ctx, &s.pvcService, resources)
	errgrp.applyPvs(ctx, &s.pvService, &s.pvcService, resources)
	errgrp.applyNodes(ctx, &s.nodeService, resources)
	errgrp.applyPods(ctx, &s.podService, resources)

	if err := errgrp.g.Wait(); err != nil {
		return xerrors.Errorf("apply each resources: %w", err)
	}
	return nil
}

func (eg *ErrorGroupWithSemaphore) listPods(ctx context.Context, s *PodService, r *ResourcesForExport) {
	if err := eg.sem.Acquire(ctx, 1); err != nil {
		klog.Fatalf("failed to acquire semaphore: %v", err)
		return
	}
	eg.g.Go(func() error {
		defer eg.sem.Release(1)
		pods, err := (*s).List(ctx)
		if err != nil {
			return xerrors.Errorf("call list pods: %w", err)
		}
		r.Pods = pods.Items
		return nil
	})
}

func (eg *ErrorGroupWithSemaphore) listNodes(ctx context.Context, s *NodeService, r *ResourcesForExport) {
	if err := eg.sem.Acquire(ctx, 1); err != nil {
		klog.Fatalf("failed to acquire semaphore: %v", err)
		return
	}
	eg.g.Go(func() error {
		defer eg.sem.Release(1)
		nodes, err := (*s).List(ctx)
		if err != nil {
			return xerrors.Errorf("call list nodes: %w", err)
		}
		r.Nodes = nodes.Items
		return nil
	})
}

func (eg *ErrorGroupWithSemaphore) listPvs(ctx context.Context, s *PersistentVolumeService, r *ResourcesForExport) {
	if err := eg.sem.Acquire(ctx, 1); err != nil {
		klog.Fatalf("failed to acquire semaphore: %v", err)
		return
	}
	eg.g.Go(func() error {
		defer eg.sem.Release(1)
		pvs, err := (*s).List(ctx)
		if err != nil {
			return xerrors.Errorf("call list PersistentVolumes: %w", err)
		}
		r.Pvs = pvs.Items
		return nil
	})
}

func (eg *ErrorGroupWithSemaphore) listPvcs(ctx context.Context, s *PersistentVolumeClaimService, r *ResourcesForExport) {
	if err := eg.sem.Acquire(ctx, 1); err != nil {
		klog.Fatalf("failed to acquire semaphore: %v", err)
		return
	}
	eg.g.Go(func() error {
		defer eg.sem.Release(1)
		pvcs, err := (*s).List(ctx)
		if err != nil {
			return xerrors.Errorf("call list PersistentVolumeClaims: %w", err)
		}
		r.Pvcs = pvcs.Items
		return nil
	})
}

func (eg *ErrorGroupWithSemaphore) listStorageClasses(ctx context.Context, s *StorageClassService, r *ResourcesForExport) {
	if err := eg.sem.Acquire(ctx, 1); err != nil {
		klog.Fatalf("failed to acquire semaphore: %v", err)
		return
	}
	eg.g.Go(func() error {
		defer eg.sem.Release(1)
		scs, err := (*s).List(ctx)
		if err != nil {
			return xerrors.Errorf("to call list storageClasses: %w", err)
		}
		r.StorageClasses = scs.Items
		return nil
	})
}

func (eg *ErrorGroupWithSemaphore) listPcs(ctx context.Context, s *PriorityClassService, r *ResourcesForExport) {
	if err := eg.sem.Acquire(ctx, 1); err != nil {
		klog.Fatalf("failed to acquire semaphore: %v", err)
		return
	}
	eg.g.Go(func() error {
		defer eg.sem.Release(1)
		pcs, err := (*s).List(ctx)
		if err != nil {
			return xerrors.Errorf("to call list priorityClasses: %w", err)
		}
		r.PriorityClasses = pcs.Items
		return nil
	})
}

func (eg *ErrorGroupWithSemaphore) getSchedulerConfig(ctx context.Context, s *SchedulerService, r *ResourcesForExport) {
	if err := eg.sem.Acquire(ctx, 1); err != nil {
		klog.Fatalf("failed to acquire semaphore: %v", err)
		return
	}
	eg.g.Go(func() error {
		defer eg.sem.Release(1)
		ss := (*s).GetSchedulerConfig()
		r.SchedulerConfig = ss
		return nil
	})
}

func (eg *ErrorGroupWithSemaphore) applyPcs(ctx context.Context, s *PriorityClassService, r *ResourcesForImport) {
	for i := range r.PriorityClasses {
		pc := r.PriorityClasses[i]
		if err := eg.sem.Acquire(ctx, 1); err != nil {
			klog.Fatalf("failed to acquire semaphore: %v", err)
			return
		}
		eg.g.Go(func() error {
			defer eg.sem.Release(1)
			pc.ObjectMetaApplyConfiguration.UID = nil
			_, err := (*s).Apply(ctx, &pc)
			if err != nil {
				return xerrors.Errorf("apply PriorityClass: %w", err)
			}
			return nil
		})
	}
}

func (eg *ErrorGroupWithSemaphore) applyStorageClasses(ctx context.Context, s *StorageClassService, r *ResourcesForImport) {
	for i := range r.StorageClasses {
		sc := r.StorageClasses[i]
		if err := eg.sem.Acquire(ctx, 1); err != nil {
			klog.Fatalf("failed to acquire semaphore: %v", err)
			return
		}
		eg.g.Go(func() error {
			defer eg.sem.Release(1)
			sc.ObjectMetaApplyConfiguration.UID = nil
			_, err := (*s).Apply(ctx, &sc)
			if err != nil {
				return xerrors.Errorf("apply StorageClass: %w", err)
			}
			return nil
		})
	}
}

func (eg *ErrorGroupWithSemaphore) applyPvcs(ctx context.Context, s *PersistentVolumeClaimService, r *ResourcesForImport) {
	for i := range r.Pvcs {
		pvc := r.Pvcs[i]
		if err := eg.sem.Acquire(ctx, 1); err != nil {
			klog.Fatalf("failed to acquire semaphore: %v", err)
			return
		}
		eg.g.Go(func() error {
			defer eg.sem.Release(1)
			pvc.ObjectMetaApplyConfiguration.UID = nil
			_, err := (*s).Apply(ctx, &pvc)
			if err != nil {
				return xerrors.Errorf("apply PersistentVolumeClaims: %w", err)
			}
			return nil
		})
	}
}

func (eg *ErrorGroupWithSemaphore) applyPvs(ctx context.Context, s *PersistentVolumeService, pvcs *PersistentVolumeClaimService, r *ResourcesForImport) {
	for i := range r.Pvs {
		pv := r.Pvs[i]
		if err := eg.sem.Acquire(ctx, 1); err != nil {
			klog.Fatalf("failed to acquire semaphore: %v", err)
			return
		}
		eg.g.Go(func() error {
			defer eg.sem.Release(1)
			pv.ObjectMetaApplyConfiguration.UID = nil
			if pv.Status != nil && pv.Status.Phase != nil {
				if *pv.Status.Phase == "Bound" {
					// PersistentVolumeClaims's UID has been changed to a new value.
					pvc, err := (*pvcs).Get(ctx, *pv.Spec.ClaimRef.Name)
					if err == nil {
						pv.Spec.ClaimRef.UID = &pvc.UID
					} else {
						klog.Warningf("failed to Get PersistentVolumeClaims from the specified name: %v", err)
						pv.Spec.ClaimRef.UID = nil
					}
				}
			}
			_, err := (*s).Apply(ctx, &pv)
			if err != nil {
				return xerrors.Errorf("apply PersistentVolume: %w", err)
			}
			return nil
		})
	}
}

func (eg *ErrorGroupWithSemaphore) applyNodes(ctx context.Context, s *NodeService, r *ResourcesForImport) {
	for i := range r.Nodes {
		node := r.Nodes[i]
		if err := eg.sem.Acquire(ctx, 1); err != nil {
			klog.Fatalf("failed to acquire semaphore: %v", err)
			return
		}
		eg.g.Go(func() error {
			defer eg.sem.Release(1)
			node.ObjectMetaApplyConfiguration.UID = nil
			_, err := (*s).Apply(ctx, &node)
			if err != nil {
				return xerrors.Errorf("apply Node: %w", err)
			}
			return nil
		})
	}
}

func (eg *ErrorGroupWithSemaphore) applyPods(ctx context.Context, s *PodService, r *ResourcesForImport) {
	for i := range r.Pods {
		pod := r.Pods[i]
		if err := eg.sem.Acquire(ctx, 1); err != nil {
			klog.Fatalf("failed to acquire semaphore: %v", err)
			return
		}
		eg.g.Go(func() error {
			defer eg.sem.Release(1)
			pod.ObjectMetaApplyConfiguration.UID = nil
			_, err := (*s).Apply(ctx, &pod)
			if err != nil {
				return xerrors.Errorf("apply Pod: %w", err)
			}
			return nil
		})
	}
}
