package pod

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages pods.
type Service struct {
	client clientset.Interface
}

// NewPodService initializes Service.
func NewPodService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

// Get returns the pod has given name.
func (s *Service) Get(ctx context.Context, name string, namespace string) (*corev1.Pod, error) {
	n, err := s.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get pod: %w", err)
	}
	return n, nil
}

// List list all pods.
func (s *Service) List(ctx context.Context, namespace string) (*corev1.PodList, error) {
	pl, err := s.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list pods: %w", err)
	}
	return pl, nil
}

// Apply applies the pod.
func (s *Service) Apply(ctx context.Context, namespace string, pod *v1.PodApplyConfiguration) (*corev1.Pod, error) {
	pod.WithKind("Pod")
	pod.WithAPIVersion("v1")

	newpod, err := s.client.CoreV1().Pods(namespace).Apply(ctx, pod, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return nil, xerrors.Errorf("apply pod: %w", err)
	}
	return newpod, nil
}

// Delete deletes the pod has given name.
func (s *Service) Delete(ctx context.Context, name string, namespace string) error {
	noGrace := int64(0)
	err := s.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{
		// need to use noGrace to avoid waiting kubelet checking.
		// > When a force deletion is performed, the API server does not wait for confirmation from the kubelet that
		//   the Pod has been terminated on the node it was running on.
		// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination-forced
		GracePeriodSeconds: &noGrace,
	})
	if err != nil {
		return fmt.Errorf("delete pod: %w", err)
	}

	return nil
}
