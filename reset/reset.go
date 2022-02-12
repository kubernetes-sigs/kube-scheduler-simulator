package reset

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type ResetService interface {
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
	resetServices map[string]ResetService
	schedService  SchedulerService
}

// NewResetService initializes Service.
func NewResetService(
	client clientset.Interface,
	resetServices map[string]ResetService,
	schedService SchedulerService,
) *Service {
	return &Service{
		client:        client,
		resetServices: resetServices,
		schedService:  schedService,
	}
}

// Reset cleans up all resources and scheduler configuration.
func (s *Service) Reset(ctx context.Context) error {
	// We need emptyListOpts to satisfy interface.
	emptyListOpts := metav1.ListOptions{}
	for _, rs := range s.resetServices {
		if err := rs.DeleteCollection(ctx, emptyListOpts); err != nil {
			return err
		}
	}
	return s.schedService.ResetScheduler()
}
