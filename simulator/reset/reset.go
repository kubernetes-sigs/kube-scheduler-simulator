package reset

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type DeleteService interface {
	DeleteCollection(ctx context.Context, lopts metav1.ListOptions) error
}

type DeleteServicesForNamespacedResources interface {
	DeleteCollection(ctx context.Context, namespace string, lopts metav1.ListOptions) error
}

type SchedulerService interface {
	ResetScheduler() error
}

// Service cleans up resources.
type Service struct {
	client clientset.Interface
	// deleteServices has the all services for each non-namespaced resource.
	// key: service name.
	deleteServices map[string]DeleteService
	// deleteNamespacedServices has the all services for each namespaced resource.
	// key: service name.
	deleteServicesForNamespacedResources map[string]DeleteServicesForNamespacedResources
	schedService                         SchedulerService
}

// NewResetService initializes Service.
func NewResetService(
	client clientset.Interface,
	deleteServices map[string]DeleteService,
	deleteServicesForNamespacedResources map[string]DeleteServicesForNamespacedResources,
	schedService SchedulerService,
) *Service {
	return &Service{
		client:                               client,
		deleteServices:                       deleteServices,
		deleteServicesForNamespacedResources: deleteServicesForNamespacedResources,
		schedService:                         schedService,
	}
}

// Reset cleans up all resources and scheduler configuration.
func (s *Service) Reset(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	nsList, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return xerrors.Errorf("list namespaces: %w", err)
	}
	for k, ds := range s.deleteServicesForNamespacedResources {
		ds := ds
		k := k
		for _, ns := range nsList.Items {
			ns := ns
			eg.Go(func() error {
				// this method deletes all resources on specified namespace.
				if err := ds.DeleteCollection(ctx, ns.GetName(), metav1.ListOptions{}); err != nil {
					return xerrors.Errorf("delete collecton of %s service: %w", k, err)
				}
				return nil
			})
		}
	}
	for k, ds := range s.deleteServices {
		ds := ds
		k := k
		eg.Go(func() error {
			if err := ds.DeleteCollection(ctx, metav1.ListOptions{}); err != nil {
				return xerrors.Errorf("delete collecton of %s service: %w", k, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	if err := s.schedService.ResetScheduler(); err != nil {
		return xerrors.Errorf("reset scheduler: %w", err)
	}
	return nil
}
