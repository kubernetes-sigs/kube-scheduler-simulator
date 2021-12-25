package export

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/export/mock_export"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
)

func TestService_Export(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		expectReturn             *Resources
	}{
		{
			name: "export all resources",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
				schedulers.EXPECT().GetSchedulerConfig().Return(&v1beta2config.KubeSchedulerConfiguration{})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				// add test data.
				return c
			},
			expectReturn: &Resources{
				Pods:            []corev1.Pod{},
				Nodes:           []corev1.Node{},
				Pvs:             []corev1.PersistentVolume{},
				Pvcs:            []corev1.PersistentVolumeClaim{},
				StorageClasses:  []storagev1.StorageClass{},
				SchedulerConfig: &v1beta2config.KubeSchedulerConfiguration{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockSchedulerService := mock_export.NewMockSchedulerService(ctrl)
			mockStorageClassService := mock_export.NewMockStorageClassService(ctrl)
			mockPVCService := mock_export.NewMockPersistentVolumeClaimService(ctrl)
			mockPVService := mock_export.NewMockPersistentVolumeService(ctrl)
			mockNodeService := mock_export.NewMockNodeService(ctrl)
			mockPodService := mock_export.NewMockPodService(ctrl)
			fakeclientset := tt.prepareFakeClientSetFn()

			s := NewResourcesService(fakeclientset, mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockSchedulerService)
			tt.prepareEachServiceMockFn(mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockSchedulerService)
			r, err := s.Export(context.Background())
			if diff := cmp.Diff(r, tt.expectReturn); diff != "" {
				t.Fatalf("Export() %v test, error = %v,\n%s", tt.name, err, diff)
			}
		})
	}
}
