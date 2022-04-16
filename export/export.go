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
	"strings"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	v1beta3config "k8s.io/kube-scheduler/config/v1beta3"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/util"
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
	SchedulerConfig *v1beta3config.KubeSchedulerConfiguration `json:"schedulerConfig"`
}

// ResourcesForImport denotes all resources and scheduler configuration for import.
type ResourcesForImport struct {
	Pods            []v1.PodApplyConfiguration                        `json:"pods"`
	Nodes           []v1.NodeApplyConfiguration                       `json:"nodes"`
	Pvs             []v1.PersistentVolumeApplyConfiguration           `json:"pvs"`
	Pvcs            []v1.PersistentVolumeClaimApplyConfiguration      `json:"pvcs"`
	StorageClasses  []confstoragev1.StorageClassApplyConfiguration    `json:"storageClasses"`
	PriorityClasses []schedulingcfgv1.PriorityClassApplyConfiguration `json:"priorityClasses"`
	SchedulerConfig *v1beta3config.KubeSchedulerConfiguration         `json:"schedulerConfig"`
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
	GetSchedulerConfig() *v1beta3config.KubeSchedulerConfiguration
	RestartScheduler(cfg *v1beta3config.KubeSchedulerConfiguration) error
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

type options struct {
	ignoreErr                    bool
	ignoreSchedulerConfiguration bool
}

type (
	ignoreErrOption                    bool
	ignoreSchedulerConfigurationOption bool
)

type Option interface {
	apply(*options)
}

func (i ignoreErrOption) apply(opts *options) {
	opts.ignoreErr = bool(i)
}

func (i ignoreSchedulerConfigurationOption) apply(opts *options) {
	opts.ignoreSchedulerConfiguration = bool(i)
}

// IgnoreErr is the option to literally ignore errors.
// If it is enabled, the method won't return any errors, but just log errors as error logs.
func (s *Service) IgnoreErr() Option {
	return ignoreErrOption(true)
}

// IgnoreSchedulerConfiguration is the option to ignore the scheduler configuration in the given ResourcesForImport.
// Note: this option is only for Import method.
// If it is enabled, the scheduler will not be restarted in import method.
func (s *Service) IgnoreSchedulerConfiguration() Option {
	return ignoreSchedulerConfigurationOption(true)
}

// Get gets all resources from each service.
func (s *Service) get(ctx context.Context, opts options) (*ResourcesForExport, error) {
	errgrp := util.NewErrGroupWithSemaphore(ctx)
	resources := ResourcesForExport{}

	if err := s.listPods(ctx, &resources, &errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listPods: %w", err)
	}
	if err := s.listNodes(ctx, &resources, &errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listNodes: %w", err)
	}
	if err := s.listPvs(ctx, &resources, &errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listPvs: %w", err)
	}
	if err := s.listPvcs(ctx, &resources, &errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listPvcs: %w", err)
	}
	if err := s.listStorageClasses(ctx, &resources, &errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listStorageClasses: %w", err)
	}
	if err := s.listPcs(ctx, &resources, &errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listPcs: %w", err)
	}
	if err := s.getSchedulerConfig(ctx, &resources, &errgrp); err != nil {
		return nil, xerrors.Errorf("call getSchedulerConfig: %w", err)
	}

	if err := errgrp.Grp.Wait(); err != nil {
		return nil, xerrors.Errorf("get resources all: %w", err)
	}
	return &resources, nil
}

func (s *Service) Export(ctx context.Context, opts ...Option) (*ResourcesForExport, error) {
	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}
	resources, err := s.get(ctx, options)
	if err != nil {
		return nil, xerrors.Errorf("export resources all: %w", err)
	}
	return resources, nil
}

// Apply applies all resources from each service.
func (s *Service) apply(ctx context.Context, resources *ResourcesForImport, opts options) error {
	errgrp := util.NewErrGroupWithSemaphore(ctx)

	if err := s.applyPcs(ctx, resources, &errgrp, opts); err != nil {
		return xerrors.Errorf("call applyPcs: %w", err)
	}
	if err := s.applyStorageClasses(ctx, resources, &errgrp, opts); err != nil {
		return xerrors.Errorf("call applyStorageClasses: %w", err)
	}
	if err := s.applyPvcs(ctx, resources, &errgrp, opts); err != nil {
		return xerrors.Errorf("call applyPvcs: %w", err)
	}
	if err := s.applyNodes(ctx, resources, &errgrp, opts); err != nil {
		return xerrors.Errorf("call applyNodes: %w", err)
	}
	if err := s.applyPods(ctx, resources, &errgrp, opts); err != nil {
		return xerrors.Errorf("call applyPods: %w", err)
	}

	if err := errgrp.Grp.Wait(); err != nil {
		return xerrors.Errorf("apply resources: %w", err)
	}

	// `applyPvs` should be called after `applyPvcs` finished,
	// because `applyPvs` look up PersistentVolumeClaim for `Spec.ClaimRef.UID` field.
	if err := s.applyPvs(ctx, resources, &errgrp, opts); err != nil {
		return xerrors.Errorf("call applyPvs: %w", err)
	}
	if err := errgrp.Grp.Wait(); err != nil {
		return xerrors.Errorf("apply PVs: %w", err)
	}
	return nil
}

// Import imports all resources from posted data.
// (1) Restart scheduler based on the data.
// (2) Apply each resource.
//     * If UID is not nil, an error will occur. (This is because the api-server will try to find that in existing resources by UID)
func (s *Service) Import(ctx context.Context, resources *ResourcesForImport, opts ...Option) error {
	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}
	if !options.ignoreSchedulerConfiguration {
		if err := s.schedulerService.RestartScheduler(resources.SchedulerConfig); err != nil {
			return xerrors.Errorf("restart scheduler with imported configuration: %w", err)
		}
	}
	if err := s.apply(ctx, resources, options); err != nil {
		return xerrors.Errorf("import resources all: %w", err)
	}
	return nil
}

func (s *Service) listPods(ctx context.Context, r *ResourcesForExport, eg *util.ErrGroupWithSemaphore, opts options) error {
	if err := eg.Sem.Acquire(ctx, 1); err != nil {
		return xerrors.Errorf("acquire semaphore: %w", err)
	}
	eg.Grp.Go(func() error {
		defer eg.Sem.Release(1)
		pods, err := s.podService.List(ctx)
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list pods: %w", err)
			}
			klog.Errorf("failed to call list pods: %v", err)
			pods = &corev1.PodList{Items: []corev1.Pod{}}
		}
		r.Pods = pods.Items
		return nil
	})
	return nil
}

func (s *Service) listNodes(ctx context.Context, r *ResourcesForExport, eg *util.ErrGroupWithSemaphore, opts options) error {
	if err := eg.Sem.Acquire(ctx, 1); err != nil {
		return xerrors.Errorf("acquire semaphore: %w", err)
	}
	eg.Grp.Go(func() error {
		defer eg.Sem.Release(1)
		nodes, err := s.nodeService.List(ctx)
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list nodes: %w", err)
			}
			klog.Errorf("failed to call list nodes: %v", err)
			nodes = &corev1.NodeList{Items: []corev1.Node{}}
		}
		r.Nodes = nodes.Items
		return nil
	})
	return nil
}

func (s *Service) listPvs(ctx context.Context, r *ResourcesForExport, eg *util.ErrGroupWithSemaphore, opts options) error {
	if err := eg.Sem.Acquire(ctx, 1); err != nil {
		return xerrors.Errorf("acquire semaphore: %w", err)
	}
	eg.Grp.Go(func() error {
		defer eg.Sem.Release(1)
		pvs, err := s.pvService.List(ctx)
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list PersistentVolumes: %w", err)
			}
			klog.Errorf("failed to call list PersistentVolumes: %v", err)
			pvs = &corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}
		}
		r.Pvs = pvs.Items
		return nil
	})
	return nil
}

func (s *Service) listPvcs(ctx context.Context, r *ResourcesForExport, eg *util.ErrGroupWithSemaphore, opts options) error {
	if err := eg.Sem.Acquire(ctx, 1); err != nil {
		return xerrors.Errorf("acquire semaphore: %w", err)
	}
	eg.Grp.Go(func() error {
		defer eg.Sem.Release(1)
		pvcs, err := s.pvcService.List(ctx)
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list PersistentVolumeClaims: %w", err)
			}
			klog.Errorf("failed to call list PersistentVolumeClaims: %v", err)
			pvcs = &corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}
		}
		r.Pvcs = pvcs.Items
		return nil
	})
	return nil
}

func (s *Service) listStorageClasses(ctx context.Context, r *ResourcesForExport, eg *util.ErrGroupWithSemaphore, opts options) error {
	if err := eg.Sem.Acquire(ctx, 1); err != nil {
		return xerrors.Errorf("acquire semaphore: %w", err)
	}
	eg.Grp.Go(func() error {
		defer eg.Sem.Release(1)
		scs, err := s.storageClassService.List(ctx)
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list storageClasses: %w", err)
			}
			klog.Errorf("failed to call list storageClasses: %v", err)
			scs = &storagev1.StorageClassList{Items: []storagev1.StorageClass{}}
		}
		r.StorageClasses = scs.Items
		return nil
	})
	return nil
}

func (s *Service) listPcs(ctx context.Context, r *ResourcesForExport, eg *util.ErrGroupWithSemaphore, opts options) error {
	if err := eg.Sem.Acquire(ctx, 1); err != nil {
		return xerrors.Errorf("acquire semaphore: %w", err)
	}
	eg.Grp.Go(func() error {
		defer eg.Sem.Release(1)
		pcs, err := s.priorityclassService.List(ctx)
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list priorityClasses: %w", err)
			}
			klog.Errorf("failed to call list priorityClasses: %v", err)
			pcs = &schedulingv1.PriorityClassList{Items: []schedulingv1.PriorityClass{}}
		}
		result := []schedulingv1.PriorityClass{}
		for _, i := range pcs.Items {
			if !isSystemPriorityClass(i.GetObjectMeta().GetName()) {
				result = append(result, i)
			}
		}
		r.PriorityClasses = result
		return nil
	})
	return nil
}

func (s *Service) getSchedulerConfig(ctx context.Context, r *ResourcesForExport, eg *util.ErrGroupWithSemaphore) error {
	if err := eg.Sem.Acquire(ctx, 1); err != nil {
		return xerrors.Errorf("acquire semaphore: %w", err)
	}
	eg.Grp.Go(func() error {
		defer eg.Sem.Release(1)
		ss := s.schedulerService.GetSchedulerConfig()
		r.SchedulerConfig = ss
		return nil
	})
	return nil
}

func (s *Service) applyPcs(ctx context.Context, r *ResourcesForImport, eg *util.ErrGroupWithSemaphore, opts options) error {
	for i := range r.PriorityClasses {
		pc := r.PriorityClasses[i]
		if isSystemPriorityClass(*pc.Name) {
			continue
		}
		if err := eg.Sem.Acquire(ctx, 1); err != nil {
			return xerrors.Errorf("acquire semaphore: %w", err)
		}
		eg.Grp.Go(func() error {
			defer eg.Sem.Release(1)
			pc.ObjectMetaApplyConfiguration.UID = nil
			_, err := s.priorityclassService.Apply(ctx, &pc)
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply PriorityClass: %w", err)
				}
				klog.Errorf("failed to apply priorityClasses: %v", err)
			}
			return nil
		})
	}
	return nil
}

func (s *Service) applyStorageClasses(ctx context.Context, r *ResourcesForImport, eg *util.ErrGroupWithSemaphore, opts options) error {
	for i := range r.StorageClasses {
		sc := r.StorageClasses[i]
		if err := eg.Sem.Acquire(ctx, 1); err != nil {
			return xerrors.Errorf("acquire semaphore: %w", err)
		}
		eg.Grp.Go(func() error {
			defer eg.Sem.Release(1)
			sc.ObjectMetaApplyConfiguration.UID = nil
			_, err := s.storageClassService.Apply(ctx, &sc)
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply StorageClass: %w", err)
				}
				klog.Errorf("failed to apply StorageClass: %v", err)
			}
			return nil
		})
	}
	return nil
}

func (s *Service) applyPvcs(ctx context.Context, r *ResourcesForImport, eg *util.ErrGroupWithSemaphore, opts options) error {
	for i := range r.Pvcs {
		pvc := r.Pvcs[i]
		if err := eg.Sem.Acquire(ctx, 1); err != nil {
			return xerrors.Errorf("acquire semaphore: %w", err)
		}
		eg.Grp.Go(func() error {
			defer eg.Sem.Release(1)
			pvc.ObjectMetaApplyConfiguration.UID = nil
			_, err := s.pvcService.Apply(ctx, &pvc)
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply PersistentVolumeClaims: %w", err)
				}
				klog.Errorf("failed to apply PersistentVolumeClaims: %v", err)
			}
			return nil
		})
	}
	return nil
}

func (s *Service) applyPvs(ctx context.Context, r *ResourcesForImport, eg *util.ErrGroupWithSemaphore, opts options) error {
	for i := range r.Pvs {
		pv := r.Pvs[i]
		if err := eg.Sem.Acquire(ctx, 1); err != nil {
			return xerrors.Errorf("acquire semaphore: %w", err)
		}
		eg.Grp.Go(func() error {
			defer eg.Sem.Release(1)
			pv.ObjectMetaApplyConfiguration.UID = nil
			if pv.Status != nil && pv.Status.Phase != nil {
				if *pv.Status.Phase == "Bound" {
					// PersistentVolumeClaims's UID has been changed to a new value.
					pvc, err := s.pvcService.Get(ctx, *pv.Spec.ClaimRef.Name)
					if err == nil {
						pv.Spec.ClaimRef.UID = &pvc.UID
					} else {
						klog.Errorf("failed to Get PersistentVolumeClaims from the specified name: %v", err)
						pv.Spec.ClaimRef.UID = nil
					}
				}
			}
			_, err := s.pvService.Apply(ctx, &pv)
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply PersistentVolume: %w", err)
				}
				klog.Errorf("failed to apply PersistentVolume: %v", err)
			}
			return nil
		})
	}
	return nil
}

func (s *Service) applyNodes(ctx context.Context, r *ResourcesForImport, eg *util.ErrGroupWithSemaphore, opts options) error {
	for i := range r.Nodes {
		node := r.Nodes[i]
		if err := eg.Sem.Acquire(ctx, 1); err != nil {
			return xerrors.Errorf("acquire semaphore: %w", err)
		}
		eg.Grp.Go(func() error {
			defer eg.Sem.Release(1)
			node.ObjectMetaApplyConfiguration.UID = nil
			_, err := s.nodeService.Apply(ctx, &node)
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply Node: %w", err)
				}
				klog.Errorf("failed to apply Node: %v", err)
			}
			return nil
		})
	}
	return nil
}

func (s *Service) applyPods(ctx context.Context, r *ResourcesForImport, eg *util.ErrGroupWithSemaphore, opts options) error {
	for i := range r.Pods {
		pod := r.Pods[i]
		if err := eg.Sem.Acquire(ctx, 1); err != nil {
			return xerrors.Errorf("acquire semaphore: %w", err)
		}
		eg.Grp.Go(func() error {
			defer eg.Sem.Release(1)
			pod.ObjectMetaApplyConfiguration.UID = nil
			_, err := s.podService.Apply(ctx, &pod)
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply Pod: %w", err)
				}
				klog.Errorf("failed to apply Pod: %v", err)
			}
			return nil
		})
	}
	return nil
}

// isSystemPriorityClass returns whether the given name of PriorityClass is prefixed with `system-` or not.
// The `system-` prefix is reserved by Kubernetes, and users cannot create a PriorityClass with such a name.
// See: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass
//
// So, we need to exclude these PriorityClasses when import/export PriorityClasses.
func isSystemPriorityClass(name string) bool {
	return strings.HasPrefix(name, "system-")
}
