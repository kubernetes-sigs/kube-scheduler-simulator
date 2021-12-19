package storageclass

import (
	"context"

	"golang.org/x/xerrors"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Service manages storageClasss.
type Service struct {
	client clientset.Interface
}

// NewStorageClassService initializes Service.
func NewStorageClassService(client clientset.Interface) *Service {
	return &Service{
		client: client,
	}
}

// Get returns the storageClass has given name.
func (s *Service) Get(ctx context.Context, name string) (*storagev1.StorageClass, error) {
	n, err := s.client.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get storageClass: %w", err)
	}
	return n, nil
}

// List list all storageClass.
func (s *Service) List(ctx context.Context) (*storagev1.StorageClassList, error) {
	pl, err := s.client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("list storageClasss: %w", err)
	}
	return pl, nil
}

// Apply applies the storageClass.
func (s *Service) Apply(ctx context.Context, storageClass *v1.StorageClassApplyConfiguration) (*storagev1.StorageClass, error) {
	storageClass.WithKind("StorageClass")
	storageClass.WithAPIVersion("storage.k8s.io/v1")

	newsc, err := s.client.StorageV1().StorageClasses().Apply(ctx, storageClass, metav1.ApplyOptions{Force: true, FieldManager: "simulator"})
	if err != nil {
		return nil, xerrors.Errorf("apply storageClass: %w", err)
	}

	return newsc, nil
}

// Delete deletes the storageClass has given name.
func (s *Service) Delete(ctx context.Context, name string) error {
	err := s.client.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return xerrors.Errorf("delete storageClass: %w", err)
	}

	return nil
}

// Deletes deletes all storageClasss.
func (s *Service) Deletes(ctx context.Context) error {
	if err := s.client.StorageV1().StorageClasses().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{}); err != nil {
		return xerrors.Errorf("delete storageClasss: %w", err)
	}

	return nil
}
