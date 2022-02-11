package reset

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type NodeService interface {
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

type PersistentVolumeService interface {
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

type PersistentVolumeClaimService interface {
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

type StorageClassService interface {
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

type PriorityClassService interface {
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

type SchedulerService interface {
	ResetScheduler() error
}

// Service cleans up
type Service struct {
	client       clientset.Interface
	nodeService  NodeService
	pvService    PersistentVolumeService
	pvcService   PersistentVolumeClaimService
	scSerivce    StorageClassService
	pcService    PriorityClassService
	schedService SchedulerService
}

// NewResetService initializes Service.
func NewResetService(
	client clientset.Interface,
	nodeService NodeService,
	pvService PersistentVolumeService,
	pvcService PersistentVolumeClaimService,
	scService StorageClassService,
	pcService PriorityClassService,
	schedService SchedulerService,
) *Service {
	return &Service{
		client:       client,
		nodeService:  nodeService,
		pvService:    pvService,
		pvcService:   pvcService,
		scSerivce:    scService,
		pcService:    pcService,
		schedService: schedService,
	}
}

// Reset cleans up all resources and scheduler configuration.
func (s *Service) Reset(ctx context.Context) error {
	lopts := metav1.ListOptions{
		FieldSelector: "spec.nodeName!=",
	}
	// We need emptyListOpts to satisfy interface.
	emptyListOpts := metav1.ListOptions{}
	if err := s.nodeService.DeleteCollection(ctx, lopts); err != nil {
		return err
	}
	if err := s.pvService.DeleteCollection(ctx, emptyListOpts); err != nil {
		return err
	}
	if err := s.pvcService.DeleteCollection(ctx, emptyListOpts); err != nil {
		return err
	}
	if err := s.scSerivce.DeleteCollection(ctx, emptyListOpts); err != nil {
		return err
	}
	if err := s.pcService.DeleteCollection(ctx, emptyListOpts); err != nil {
		return err
	}
	if err := s.schedService.ResetScheduler(); err != nil {
		return err
	}
	return nil
}
