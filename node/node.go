package node

//go:generate mockgen -destination=./mock_$GOPACKAGE/$GOFILE . PodService

import (
	"context"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages node.
type Service struct {
	client     clientset.Interface
	podService PodService
}

// PodService represents service for manage Pods.
type PodService interface {
	List(ctx context.Context) (*corev1.PodList, error)
	Delete(ctx context.Context, name string) error
}

// NewNodeService initializes Service.
func NewNodeService(client clientset.Interface, ps PodService) *Service {
	return &Service{
		client:     client,
		podService: ps,
	}
}

// Get returns the node has given name.
func (s *Service) Get(ctx context.Context, name string) (*corev1.Node, error) {
	n, err := s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get nodes: %w", err)
	}

	return n, nil
}

// List lists all nodes.
func (s *Service) List(ctx context.Context) (*corev1.NodeList, error) {
	nl, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list nodes: %w", err)
	}

	return nl, nil
}

// Apply a unique node by using the simulator ID.
func (s *Service) Apply(ctx context.Context, nac *v1.NodeApplyConfiguration) (*corev1.Node, error) {
	nac.WithAPIVersion("v1")
	nac.WithKind("Node")

	newnode, err := s.client.CoreV1().Nodes().Apply(ctx, nac, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return nil, xerrors.Errorf("apply node: %w", err)
	}

	return newnode, nil
}

// Delete deletes the node has given name.
func (s *Service) Delete(ctx context.Context, name string) error {
	pl, err := s.podService.List(ctx)
	if err != nil {
		return xerrors.Errorf("list pods: %w", err)
	}

	// delete pods on node
	for i := range pl.Items {
		pod := pl.Items[i]
		if name != pod.Spec.NodeName {
			continue
		}

		if err := s.podService.Delete(ctx, pod.Name); err != nil {
			return xerrors.Errorf("delete pod: %w", err)
		}
	}

	// delete node
	if err := s.client.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return xerrors.Errorf("delete node: %w", err)
	}
	return nil
}

// Deletes deletes all nodes.
func (s *Service) Deletes(ctx context.Context) error {
	pl, err := s.podService.List(ctx)
	if err != nil {
		return xerrors.Errorf("list pods: %w", err)
	}

	// delete all pods
	for i := range pl.Items {
		pod := pl.Items[i]
		if err := s.podService.Delete(ctx, pod.Name); err != nil {
			return xerrors.Errorf("delete pod: %w", err)
		}
	}

	// delete all nodes
	if err := s.client.CoreV1().Nodes().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		return xerrors.Errorf("delete nodes: %w", err)
	}

	return nil
}
