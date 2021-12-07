package priorityclass

import (
	"context"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schedulingv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages priorityClass.
type Service struct {
	client clientset.Interface
}

// NewPriorityClassService initializes Service.
func NewPriorityClassService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

func (s Service) Get(ctx context.Context, name string) (*v1.PriorityClass, error) {
	n, err := s.client.SchedulingV1().PriorityClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get priorityClass: %w", err)
	}
	return n, nil
}

func (s Service) List(ctx context.Context) (*v1.PriorityClassList, error) {
	pl, err := s.client.SchedulingV1().PriorityClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list priorityClass: %w", err)
	}
	return pl, nil
}

func (s Service) Apply(ctx context.Context, priorityClass *schedulingv1.PriorityClassApplyConfiguration) (*v1.PriorityClass, error) {
	priorityClass.WithKind("PriorityClass")
	priorityClass.WithAPIVersion("scheduling.k8s.io/v1")

	newsc, err := s.client.SchedulingV1().PriorityClasses().Apply(ctx, priorityClass, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return nil, xerrors.Errorf("apply priorityClass: %w", err)
	}

	return newsc, nil
}

func (s Service) Delete(ctx context.Context, name string) error {
	err := s.client.SchedulingV1().PriorityClasses().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return xerrors.Errorf("delete priorityClass: %w", err)
	}

	return nil
}
