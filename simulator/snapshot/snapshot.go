package snapshot

//go:generate mockgen -destination=./mock_$GOPACKAGE/scheduler.go . SchedulerService

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	configv1 "k8s.io/kube-scheduler/config/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/util"
)

type Service struct {
	client           clientset.Interface
	schedulerService SchedulerService
}

// ResourcesForSnap indicates all resources and scheduler configuration to be snapped.
type ResourcesForSnap struct {
	Pods            []corev1.Pod                         `json:"pods"`
	Nodes           []corev1.Node                        `json:"nodes"`
	Pvs             []corev1.PersistentVolume            `json:"pvs"`
	Pvcs            []corev1.PersistentVolumeClaim       `json:"pvcs"`
	StorageClasses  []storagev1.StorageClass             `json:"storageClasses"`
	PriorityClasses []schedulingv1.PriorityClass         `json:"priorityClasses"`
	SchedulerConfig *configv1.KubeSchedulerConfiguration `json:"schedulerConfig"`
	Namespaces      []corev1.Namespace                   `json:"namespaces"`
}

// ResourcesForLoad indicates all resources and scheduler configuration to be loaded.
type ResourcesForLoad struct {
	Pods            []v1.PodApplyConfiguration                        `json:"pods"`
	Nodes           []v1.NodeApplyConfiguration                       `json:"nodes"`
	Pvs             []v1.PersistentVolumeApplyConfiguration           `json:"pvs"`
	Pvcs            []v1.PersistentVolumeClaimApplyConfiguration      `json:"pvcs"`
	StorageClasses  []confstoragev1.StorageClassApplyConfiguration    `json:"storageClasses"`
	PriorityClasses []schedulingcfgv1.PriorityClassApplyConfiguration `json:"priorityClasses"`
	SchedulerConfig *configv1.KubeSchedulerConfiguration              `json:"schedulerConfig"`
	Namespaces      []v1.NamespaceApplyConfiguration                  `json:"namespaces"`
}

type SchedulerService interface {
	GetSchedulerConfig() (*configv1.KubeSchedulerConfiguration, error)
	RestartScheduler(cfg *configv1.KubeSchedulerConfiguration) error
}

func NewService(client clientset.Interface, schedulers SchedulerService) *Service {
	return &Service{
		client:           client,
		schedulerService: schedulers,
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

// IgnoreSchedulerConfiguration is the option to ignore the scheduler configuration in the given ResourcesForLoad.
// Note: this option is only for Load method.
// If it is enabled, the scheduler will not be restarted in load method.
func (s *Service) IgnoreSchedulerConfiguration() Option {
	return ignoreSchedulerConfigurationOption(true)
}

// get gets all resources from each service.
func (s *Service) get(ctx context.Context, opts options) (*ResourcesForSnap, error) {
	errgrp := util.NewErrGroupWithSemaphore(ctx)
	resources := ResourcesForSnap{}

	if err := s.listPods(ctx, &resources, errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listPods: %w", err)
	}
	if err := s.listNodes(ctx, &resources, errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listNodes: %w", err)
	}
	if err := s.listPvs(ctx, &resources, errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listPvs: %w", err)
	}
	if err := s.listPvcs(ctx, &resources, errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listPvcs: %w", err)
	}
	if err := s.listStorageClasses(ctx, &resources, errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listStorageClasses: %w", err)
	}
	if err := s.listPcs(ctx, &resources, errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listPcs: %w", err)
	}
	if err := s.listNamespaces(ctx, &resources, errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call listNamespaces: %w", err)
	}
	if err := s.getSchedulerConfig(&resources, errgrp, opts); err != nil {
		return nil, xerrors.Errorf("call getSchedulerConfig: %w", err)
	}

	if err := errgrp.Wait(); err != nil {
		return nil, xerrors.Errorf("get resources all: %w", err)
	}
	return &resources, nil
}

// Snap exports all resources as one data.
func (s *Service) Snap(ctx context.Context, opts ...Option) (*ResourcesForSnap, error) {
	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}
	resources, err := s.get(ctx, options)
	if err != nil {
		return nil, xerrors.Errorf("failed to get(): %w", err)
	}
	return resources, nil
}

// Apply applies all resources to each service.
//
//nolint:cyclop // For readability.
func (s *Service) apply(ctx context.Context, resources *ResourcesForLoad, opts options) error {
	errgrp := util.NewErrGroupWithSemaphore(ctx)
	// `applyNamespaces` must be called before calling namespaced resources applying.
	if err := s.applyNamespaces(ctx, resources, errgrp, opts); err != nil {
		return xerrors.Errorf("call applyNamespaces: %w", err)
	}
	if err := errgrp.Wait(); err != nil {
		return xerrors.Errorf("apply resources: %w", err)
	}

	if err := s.applyPcs(ctx, resources, errgrp, opts); err != nil {
		return xerrors.Errorf("call applyPcs: %w", err)
	}
	if err := s.applyStorageClasses(ctx, resources, errgrp, opts); err != nil {
		return xerrors.Errorf("call applyStorageClasses: %w", err)
	}
	if err := s.applyPvcs(ctx, resources, errgrp, opts); err != nil {
		return xerrors.Errorf("call applyPvcs: %w", err)
	}
	if err := s.applyNodes(ctx, resources, errgrp, opts); err != nil {
		return xerrors.Errorf("call applyNodes: %w", err)
	}
	if err := s.applyPods(ctx, resources, errgrp, opts); err != nil {
		return xerrors.Errorf("call applyPods: %w", err)
	}
	if err := errgrp.Wait(); err != nil {
		return xerrors.Errorf("apply resources: %w", err)
	}

	// `applyPvs` should be called after `applyPvcs` finished,
	// because `applyPvs` look up PersistentVolumeClaim for `Spec.ClaimRef.UID` field.
	if err := s.applyPvs(ctx, resources, errgrp, opts); err != nil {
		return xerrors.Errorf("call applyPvs: %w", err)
	}
	if err := errgrp.Wait(); err != nil {
		return xerrors.Errorf("apply PVs: %w", err)
	}
	return nil
}

// Load imports all resources from posted data.
// (1) Restart scheduler based on the data.
// (2) Apply each resource.
//   - If UID is not nil, an error will occur. (This is because the api-server will try to find that from current resources by UID)
func (s *Service) Load(ctx context.Context, resources *ResourcesForLoad, opts ...Option) error {
	options := options{}
	for _, o := range opts {
		o.apply(&options)
	}
	if !options.ignoreSchedulerConfiguration {
		if err := s.schedulerService.RestartScheduler(resources.SchedulerConfig); err != nil {
			if !errors.Is(err, scheduler.ErrServiceDisabled) {
				return xerrors.Errorf("restart scheduler with loaded configuration: %w", err)
			}
			klog.Info("The scheduler configuration hasn't been loaded because of an external scheduler is enabled.")
		}
	}
	if err := s.apply(ctx, resources, options); err != nil {
		return xerrors.Errorf("failed to apply(): %w", err)
	}
	return nil
}

func (s *Service) listPods(ctx context.Context, r *ResourcesForSnap, eg *util.SemaphoredErrGroup, opts options) error {
	if err := eg.Go(func() error {
		pods, err := s.client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list Pod: %w", err)
			}
			klog.Errorf("failed to call list Pod: %v", err)
			pods = &corev1.PodList{Items: []corev1.Pod{}}
		}
		r.Pods = pods.Items
		return nil
	}); err != nil {
		return xerrors.Errorf("start error group: %w", err)
	}
	return nil
}

func (s *Service) listNodes(ctx context.Context, r *ResourcesForSnap, eg *util.SemaphoredErrGroup, opts options) error {
	if err := eg.Go(func() error {
		nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list Node: %w", err)
			}
			klog.Errorf("failed to call list Node: %v", err)
			nodes = &corev1.NodeList{Items: []corev1.Node{}}
		}
		r.Nodes = nodes.Items
		return nil
	}); err != nil {
		return xerrors.Errorf("start error group: %w", err)
	}
	return nil
}

func (s *Service) listPvs(ctx context.Context, r *ResourcesForSnap, eg *util.SemaphoredErrGroup, opts options) error {
	if err := eg.Go(func() error {
		pvs, err := s.client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list PersistentVolume: %w", err)
			}
			klog.Errorf("failed to call list PersistentVolume: %v", err)
			pvs = &corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}
		}
		r.Pvs = pvs.Items
		return nil
	}); err != nil {
		return xerrors.Errorf("start error group: %w", err)
	}
	return nil
}

func (s *Service) listPvcs(ctx context.Context, r *ResourcesForSnap, eg *util.SemaphoredErrGroup, opts options) error {
	if err := eg.Go(func() error {
		pvcs, err := s.client.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list PersistentVolumeClaim: %w", err)
			}
			klog.Errorf("failed to call list PersistentVolumeClaim: %v", err)
			pvcs = &corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}
		}
		r.Pvcs = pvcs.Items
		return nil
	}); err != nil {
		return xerrors.Errorf("start error group: %w", err)
	}
	return nil
}

func (s *Service) listStorageClasses(ctx context.Context, r *ResourcesForSnap, eg *util.SemaphoredErrGroup, opts options) error {
	if err := eg.Go(func() error {
		scs, err := s.client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list StorageClass: %w", err)
			}
			klog.Errorf("failed to call list StorageClass: %v", err)
			scs = &storagev1.StorageClassList{Items: []storagev1.StorageClass{}}
		}
		r.StorageClasses = scs.Items
		return nil
	}); err != nil {
		return xerrors.Errorf("start error group: %w", err)
	}
	return nil
}

func (s *Service) listPcs(ctx context.Context, r *ResourcesForSnap, eg *util.SemaphoredErrGroup, opts options) error {
	if err := eg.Go(func() error {
		pcs, err := s.client.SchedulingV1().PriorityClasses().List(ctx, metav1.ListOptions{})
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list PriorityClass: %w", err)
			}
			klog.Errorf("failed to call list PriorityClass: %v", err)
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
	}); err != nil {
		return xerrors.Errorf("start error group: %w", err)
	}
	return nil
}

func (s *Service) listNamespaces(ctx context.Context, r *ResourcesForSnap, eg *util.SemaphoredErrGroup, opts options) error {
	if err := eg.Go(func() error {
		nss, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			if !opts.ignoreErr {
				return xerrors.Errorf("call list Namespace: %w", err)
			}
			klog.Errorf("failed to call list Namespace: %v", err)
			nss = &corev1.NamespaceList{Items: []corev1.Namespace{}}
		}
		result := []corev1.Namespace{}
		for _, i := range nss.Items {
			if !isIgnoreNamespace(i.GetObjectMeta().GetName()) {
				result = append(result, i)
			}
		}
		r.Namespaces = result
		return nil
	}); err != nil {
		return xerrors.Errorf("start error group: %w", err)
	}
	return nil
}

func (s *Service) getSchedulerConfig(r *ResourcesForSnap, eg *util.SemaphoredErrGroup, opts options) error {
	if err := eg.Go(func() error {
		ss, err := s.schedulerService.GetSchedulerConfig()
		if err != nil && !errors.Is(err, scheduler.ErrServiceDisabled) {
			if !opts.ignoreErr {
				return xerrors.Errorf("get scheduler config: %w", err)
			}
			klog.Errorf("failed to get scheduler config: %v", err)
			return nil
		}
		r.SchedulerConfig = ss
		return nil
	}); err != nil {
		return xerrors.Errorf("start error group: %w", err)
	}
	return nil
}

func (s *Service) applyPcs(ctx context.Context, r *ResourcesForLoad, eg *util.SemaphoredErrGroup, opts options) error {
	for i := range r.PriorityClasses {
		pc := r.PriorityClasses[i]
		if isSystemPriorityClass(*pc.Name) {
			continue
		}
		if err := eg.Go(func() error {
			pc.ObjectMetaApplyConfiguration.UID = nil
			pc.WithAPIVersion("scheduling.k8s.io/v1").WithKind("PriorityClass")
			_, err := s.client.SchedulingV1().PriorityClasses().Apply(ctx, &pc, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply PriorityClass: %w", err)
				}
				klog.Errorf("failed to apply priorityClasses: %v", err)
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("start error group: %w", err)
		}
	}
	return nil
}

func (s *Service) applyStorageClasses(ctx context.Context, r *ResourcesForLoad, eg *util.SemaphoredErrGroup, opts options) error {
	for i := range r.StorageClasses {
		sc := r.StorageClasses[i]
		if err := eg.Go(func() error {
			sc.ObjectMetaApplyConfiguration.UID = nil
			sc.WithAPIVersion("storage.k8s.io/v1").WithKind("StorageClass")
			_, err := s.client.StorageV1().StorageClasses().Apply(ctx, &sc, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply StorageClass: %w", err)
				}
				klog.Errorf("failed to apply StorageClass: %v", err)
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("start error group: %w", err)
		}
	}
	return nil
}

func (s *Service) applyPvcs(ctx context.Context, r *ResourcesForLoad, eg *util.SemaphoredErrGroup, opts options) error {
	for i := range r.Pvcs {
		pvc := r.Pvcs[i]
		if err := eg.Go(func() error {
			pvc.ObjectMetaApplyConfiguration.UID = nil
			pvc.WithAPIVersion("v1").WithKind("PersistentVolumeClaim")
			_, err := s.client.CoreV1().PersistentVolumeClaims(*pvc.Namespace).Apply(ctx, &pvc, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply PersistentVolumeClaims: %w", err)
				}
				klog.Errorf("failed to apply PersistentVolumeClaims: %v", err)
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("start error group: %w", err)
		}
	}
	return nil
}

func (s *Service) applyPvs(ctx context.Context, r *ResourcesForLoad, eg *util.SemaphoredErrGroup, opts options) error {
	for i := range r.Pvs {
		pv := r.Pvs[i]
		if err := eg.Go(func() error {
			pv.ObjectMetaApplyConfiguration.UID = nil
			pv.WithAPIVersion("v1").WithKind("PersistentVolume")
			if pv.Status != nil && pv.Status.Phase != nil {
				if *pv.Status.Phase == "Bound" {
					// PersistentVolumeClaims's UID has been changed to a new value.
					pvc, err := s.client.CoreV1().PersistentVolumeClaims(*pv.Spec.ClaimRef.Namespace).Get(ctx, *pv.Spec.ClaimRef.Name, metav1.GetOptions{})
					if err == nil {
						pv.Spec.ClaimRef.UID = &pvc.UID
					} else {
						klog.Errorf("failed to Get PersistentVolumeClaims from the specified name: %v", err)
						pv.Spec.ClaimRef.UID = nil
					}
				}
			}
			_, err := s.client.CoreV1().PersistentVolumes().Apply(ctx, &pv, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply PersistentVolume: %w", err)
				}
				klog.Errorf("failed to apply PersistentVolume: %v", err)
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("start error group: %w", err)
		}
	}
	return nil
}

func (s *Service) applyNodes(ctx context.Context, r *ResourcesForLoad, eg *util.SemaphoredErrGroup, opts options) error {
	for i := range r.Nodes {
		node := r.Nodes[i]
		if err := eg.Go(func() error {
			node.ObjectMetaApplyConfiguration.UID = nil
			node.WithAPIVersion("v1").WithKind("Node")
			_, err := s.client.CoreV1().Nodes().Apply(ctx, &node, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply Node: %w", err)
				}
				klog.Errorf("failed to apply Node: %v", err)
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("start error group: %w", err)
		}
	}
	return nil
}

func (s *Service) applyPods(ctx context.Context, r *ResourcesForLoad, eg *util.SemaphoredErrGroup, opts options) error {
	for i := range r.Pods {
		pod := r.Pods[i]
		if err := eg.Go(func() error {
			pod.ObjectMetaApplyConfiguration.UID = nil
			pod.WithAPIVersion("v1").WithKind("Pod")
			_, err := s.client.CoreV1().Pods(*pod.Namespace).Apply(ctx, &pod, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply Pod: %w", err)
				}
				klog.Errorf("failed to apply Pod: %v", err)
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("start error group: %w", err)
		}
	}
	return nil
}

func (s *Service) applyNamespaces(ctx context.Context, r *ResourcesForLoad, eg *util.SemaphoredErrGroup, opts options) error {
	for i := range r.Namespaces {
		ns := r.Namespaces[i]
		if isIgnoreNamespace(*ns.Name) {
			continue
		}
		if err := eg.Go(func() error {
			ns.ObjectMetaApplyConfiguration.UID = nil
			ns.WithAPIVersion("v1").WithKind("Namespace")
			_, err := s.client.CoreV1().Namespaces().Apply(ctx, &ns, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
			if err != nil {
				if !opts.ignoreErr {
					return xerrors.Errorf("apply Namespace: %w", err)
				}
				klog.Errorf("failed to apply Namespace: %v", err)
			}
			return nil
		}); err != nil {
			return xerrors.Errorf("start error group: %w", err)
		}
	}
	return nil
}

// isSystemPriorityClass returns whether the given name of PriorityClass is prefixed with `system-` or not.
// The `system-` prefix is reserved by Kubernetes, and users cannot create a PriorityClass with such a name.
// See: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass
//
// So, we need to exclude these PriorityClasses when saving/loading any PriorityClasses.
func isSystemPriorityClass(name string) bool {
	return strings.HasPrefix(name, "system-")
}

// isSystemNamespace returns whether the given name of Namespace is prefixed with `kube-` or not.
// The `kube-` prefix is reserved by Kubernetes, and users cannot create a Namespace with such a name.
// See: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/#working-with-namespaces
//
// So, we need to exclude these Namespaces when saving/loading any Namespaces.
func isSystemNamespace(name string) bool {
	return strings.HasPrefix(name, "kube-")
}

// isIgnoreNamespace returns whether the given name of Namespace is ignored namespace or not.
// It's system reserved one and default namespace.
func isIgnoreNamespace(name string) bool {
	return isSystemNamespace(name) || name == "default"
}
