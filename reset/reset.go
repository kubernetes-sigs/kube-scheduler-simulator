package reset

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type DeleteService interface {
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

type SchedulerService interface {
	ResetScheduler() error
}

// Service cleans up resources.
type Service struct {
	client clientset.Interface
	// deleteServices has the all services for each resource.
	// key: service name.
	deleteServices map[string]DeleteService
	schedService   SchedulerService
}

// NewResetService initializes Service.
func NewResetService(
	client clientset.Interface,
	deleteServices map[string]DeleteService,
	schedService SchedulerService,
) *Service {
	return &Service{
		client:         client,
		deleteServices: deleteServices,
		schedService:   schedService,
	}
}

// Reset cleans up all resources and scheduler configuration.
func (s *Service) Reset(ctx context.Context) error {
	// We need emptyListOpts to satisfy interface.
	emptyListOpts := metav1.ListOptions{}
	for _, ds := range s.deleteServices {
		if err := ds.DeleteCollection(ctx, emptyListOpts); err != nil {
			return err
		}
	}
	return s.schedService.ResetScheduler()
}
