package export

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export/mock_export"
	schedulerCfg "github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/defaultconfig"
)

func TestService_Export(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		wantReturn               *Resources
		wantErr                  bool
	}{
		{
			name: "export all resources",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
				pcs.EXPECT().List(gomock.Any()).Return(&schedulingv1.PriorityClassList{Items: []schedulingv1.PriorityClass{}}, nil)
				schedulers.EXPECT().GetSchedulerConfig().Return(&v1beta2config.KubeSchedulerConfiguration{})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				// add test data.
				return c
			},
			wantReturn: &Resources{
				Pods:            []corev1.Pod{},
				Nodes:           []corev1.Node{},
				Pvs:             []corev1.PersistentVolume{},
				Pvcs:            []corev1.PersistentVolumeClaim{},
				StorageClasses:  []storagev1.StorageClass{},
				PriorityClasses: []schedulingv1.PriorityClass{},
				SchedulerConfig: &v1beta2config.KubeSchedulerConfiguration{},
			},
			wantErr: false,
		},
		{
			name: "export failure on List of PodService",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list pods"))
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
				pcs.EXPECT().List(gomock.Any()).Return(&schedulingv1.PriorityClassList{Items: []schedulingv1.PriorityClass{}}, nil)
				schedulers.EXPECT().GetSchedulerConfig().Return(&v1beta2config.KubeSchedulerConfiguration{})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantReturn: nil,
			wantErr:    true,
		},
		{
			name: "export failure on List of NodeService",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list nodes"))
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
				pcs.EXPECT().List(gomock.Any()).Return(&schedulingv1.PriorityClassList{Items: []schedulingv1.PriorityClass{}}, nil)
				schedulers.EXPECT().GetSchedulerConfig().Return(&v1beta2config.KubeSchedulerConfiguration{})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantReturn: nil,
			wantErr:    true,
		},
		{
			name: "export failure on List of PersistentVolumeService",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list PersistentVolumes"))
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
				pcs.EXPECT().List(gomock.Any()).Return(&schedulingv1.PriorityClassList{Items: []schedulingv1.PriorityClass{}}, nil)
				schedulers.EXPECT().GetSchedulerConfig().Return(&v1beta2config.KubeSchedulerConfiguration{})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantReturn: nil,
			wantErr:    true,
		},
		{
			name: "export failure on List of PersistentVolumeClaims",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list PersistentVolumeClaims"))
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
				pcs.EXPECT().List(gomock.Any()).Return(&schedulingv1.PriorityClassList{Items: []schedulingv1.PriorityClass{}}, nil)
				schedulers.EXPECT().GetSchedulerConfig().Return(&v1beta2config.KubeSchedulerConfiguration{})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantReturn: nil,
			wantErr:    true,
		},
		{
			name: "export failure on List of storageClasses",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list storageClasses"))
				pcs.EXPECT().List(gomock.Any()).Return(&schedulingv1.PriorityClassList{Items: []schedulingv1.PriorityClass{}}, nil)
				schedulers.EXPECT().GetSchedulerConfig().Return(&v1beta2config.KubeSchedulerConfiguration{})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantReturn: nil,
			wantErr:    true,
		},
		{
			name: "export failure on List of priorityClasses",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
				pcs.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list priorityClasses"))
				schedulers.EXPECT().GetSchedulerConfig().Return(&v1beta2config.KubeSchedulerConfiguration{})
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantReturn: nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockSchedulerService := mock_export.NewMockSchedulerService(ctrl)
			mockPriorityClassService := mock_export.NewMockPriorityClassService(ctrl)
			mockStorageClassService := mock_export.NewMockStorageClassService(ctrl)
			mockPVCService := mock_export.NewMockPersistentVolumeClaimService(ctrl)
			mockPVService := mock_export.NewMockPersistentVolumeService(ctrl)
			mockNodeService := mock_export.NewMockNodeService(ctrl)
			mockPodService := mock_export.NewMockPodService(ctrl)
			fakeclientset := tt.prepareFakeClientSetFn()

			s := NewResourcesService(fakeclientset, mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockPriorityClassService, mockSchedulerService)
			tt.prepareEachServiceMockFn(mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockPriorityClassService, mockSchedulerService)
			r, err := s.Export(context.Background())

			diffResponse := cmp.Diff(r, tt.wantReturn)
			if diffResponse != "" || (err != nil) != tt.wantErr {
				t.Fatalf("Export() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
			}
		})
	}
}

func TestService_Import(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		applyConfiguration       func() *ResourcesApplyConfiguration
		wantErr                  bool
	}{
		{
			name: "import all success",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil).Do(func(_ context.Context, cfg *v1.PodApplyConfiguration) {
					assert.Equal(t, "Pod1", *cfg.Name)
				})
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil).Do(func(_ context.Context, cfg *v1.NodeApplyConfiguration) {
					assert.Equal(t, "Node1", *cfg.Name)
				})
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "PVC1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil).Do(func(_ context.Context, cfg *confstoragev1.StorageClassApplyConfiguration) {
					assert.Equal(t, "StorageClass1", *cfg.Name)
				})
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil).Do(func(_ context.Context, cfg *schedulingcfgv1.PriorityClassApplyConfiguration) {
					assert.Equal(t, "PriorityClass1", *cfg.Name)
				})
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: false,
		},
		{
			name: "import failure on Pod Apply",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(nil, xerrors.Errorf("apply pod"))
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil).Do(func(_ context.Context, cfg *v1.NodeApplyConfiguration) {
					assert.Equal(t, "Node1", *cfg.Name)
				})
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "PVC1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil).Do(func(_ context.Context, cfg *confstoragev1.StorageClassApplyConfiguration) {
					assert.Equal(t, "StorageClass1", *cfg.Name)
				})
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil).Do(func(_ context.Context, cfg *schedulingcfgv1.PriorityClassApplyConfiguration) {
					assert.Equal(t, "PriorityClass1", *cfg.Name)
				})
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("something wrong", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
		{
			name: "import failure on Node Apply",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil).Do(func(_ context.Context, cfg *v1.PodApplyConfiguration) {
					assert.Equal(t, "Pod1", *cfg.Name)
				})
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(nil, xerrors.Errorf("apply node"))
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "PVC1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil).Do(func(_ context.Context, cfg *confstoragev1.StorageClassApplyConfiguration) {
					assert.Equal(t, "StorageClass1", *cfg.Name)
				})
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil).Do(func(_ context.Context, cfg *schedulingcfgv1.PriorityClassApplyConfiguration) {
					assert.Equal(t, "PriorityClass1", *cfg.Name)
				})
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("something wrong"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
		{
			name: "import failure on PersistentVolume Apply",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil).Do(func(_ context.Context, cfg *v1.PodApplyConfiguration) {
					assert.Equal(t, "Pod1", *cfg.Name)
				})
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil).Do(func(_ context.Context, cfg *v1.NodeApplyConfiguration) {
					assert.Equal(t, "Node1", *cfg.Name)
				})
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(nil, xerrors.Errorf("apply PersistentVolume"))
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "PVC1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil).Do(func(_ context.Context, cfg *confstoragev1.StorageClassApplyConfiguration) {
					assert.Equal(t, "StorageClass1", *cfg.Name)
				})
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil).Do(func(_ context.Context, cfg *schedulingcfgv1.PriorityClassApplyConfiguration) {
					assert.Equal(t, "PriorityClass1", *cfg.Name)
				})
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("something wrong").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
		{
			name: "import failure on PersistentVolumeClaim Apply",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil).Do(func(_ context.Context, cfg *v1.PodApplyConfiguration) {
					assert.Equal(t, "Pod1", *cfg.Name)
				})
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil).Do(func(_ context.Context, cfg *v1.NodeApplyConfiguration) {
					assert.Equal(t, "Node1", *cfg.Name)
				})
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(nil, xerrors.Errorf("apply PersistentVolumeClaim"))
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil).Do(func(_ context.Context, cfg *confstoragev1.StorageClassApplyConfiguration) {
					assert.Equal(t, "StorageClass1", *cfg.Name)
				})
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil).Do(func(_ context.Context, cfg *schedulingcfgv1.PriorityClassApplyConfiguration) {
					assert.Equal(t, "PriorityClass1", *cfg.Name)
				})
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("something wrong", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
		{
			name: "import failure on StorageClass Apply",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil).Do(func(_ context.Context, cfg *v1.PodApplyConfiguration) {
					assert.Equal(t, "Pod1", *cfg.Name)
				})
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil).Do(func(_ context.Context, cfg *v1.NodeApplyConfiguration) {
					assert.Equal(t, "Node1", *cfg.Name)
				})
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "PVC1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(nil, xerrors.Errorf("apply StorageClass"))
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil).Do(func(_ context.Context, cfg *schedulingcfgv1.PriorityClassApplyConfiguration) {
					assert.Equal(t, "PriorityClass1", *cfg.Name)
				})
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("something wrong"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
		{
			name: "import failure on PriorityClass Apply",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil).Do(func(_ context.Context, cfg *v1.PodApplyConfiguration) {
					assert.Equal(t, "Pod1", *cfg.Name)
				})
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil).Do(func(_ context.Context, cfg *v1.NodeApplyConfiguration) {
					assert.Equal(t, "Node1", *cfg.Name)
				})
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "PVC1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil).Do(func(_ context.Context, cfg *confstoragev1.StorageClassApplyConfiguration) {
					assert.Equal(t, "StorageClass1", *cfg.Name)
				})
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(nil, xerrors.Errorf("apply PriorityClass"))
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("something wrong"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
		{
			name: "import success when PersistentVolumeClaim was not found (Get() return err)",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil).Do(func(_ context.Context, cfg *v1.PodApplyConfiguration) {
					assert.Equal(t, "Pod1", *cfg.Name)
				})
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil).Do(func(_ context.Context, cfg *v1.NodeApplyConfiguration) {
					assert.Equal(t, "Node1", *cfg.Name)
				})
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
				})
				pvcs.EXPECT().Get(gomock.Any(), "PVC1").Return(nil, xerrors.Errorf("get persistentVolumeClaim"))
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "PVC1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil).Do(func(_ context.Context, cfg *confstoragev1.StorageClassApplyConfiguration) {
					assert.Equal(t, "StorageClass1", *cfg.Name)
				})
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil).Do(func(_ context.Context, cfg *schedulingcfgv1.PriorityClassApplyConfiguration) {
					assert.Equal(t, "PriorityClass1", *cfg.Name)
				})
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Bound")).WithSpec(v1.PersistentVolumeSpec().WithClaimRef(v1.ObjectReference().WithName("PVC1"))))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: false,
		},
		{
			name: "import success when PersistentVolumeClaim was found (Get() return err)",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil).Do(func(_ context.Context, cfg *v1.PodApplyConfiguration) {
					assert.Equal(t, "Pod1", *cfg.Name)
				})
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil).Do(func(_ context.Context, cfg *v1.NodeApplyConfiguration) {
					assert.Equal(t, "Node1", *cfg.Name)
				})
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "PVC1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil).Do(func(_ context.Context, cfg *confstoragev1.StorageClassApplyConfiguration) {
					assert.Equal(t, "StorageClass1", *cfg.Name)
				})
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil).Do(func(_ context.Context, cfg *schedulingcfgv1.PriorityClassApplyConfiguration) {
					assert.Equal(t, "PriorityClass1", *cfg.Name)
				})
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Bound")).WithSpec(v1.PersistentVolumeSpec().WithClaimRef(v1.ObjectReference().WithName("PVC1"))))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: false,
		},
		{
			name: "PV be related with PVC and assign UID",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil)
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil)
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "PV1", *cfg.Name)
					assert.Equal(t, "PVC1", *cfg.Spec.ClaimRef.Name)
					assert.Equal(t, "testUID", string(*cfg.Spec.ClaimRef.UID))
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						UID: "testUID",
					},
				}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil)
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil)
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Bound")).WithSpec(v1.PersistentVolumeSpec().WithClaimRef(v1.ObjectReference().WithName("PVC1"))))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: false,
		},
		{
			name: "success import pv and pvc from json string (UID be set to nil)",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil)
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil)
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "pv1", *cfg.Name)
					assert.Empty(t, cfg.ObjectMetaApplyConfiguration.UID)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "pvc1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil)
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil)
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				j := `{"pods":[],"nodes":[],`
				j += `"pvs":[{"metadata":{"name":"pv1","uid":"b0184e68-5ba6-4533-b3fd-bde9416ad03d","resourceVersion":"565","creationTimestamp":"2021-12-28T01:06:35Z","annotations":{"pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:capacity":{"f:storage":{}},"f:hostPath":{"f:path":{},"f:type":{}},"f:persistentVolumeReclaimPolicy":{},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:phase":{}}},"subresource":"status"},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:claimRef":{}}}}]},"spec":{"capacity":{"storage":"1Gi"},"hostPath":{"path":"/tmp/data","type":"DirectoryOrCreate"},"accessModes":["ReadWriteOnce"],"claimRef":{"kind":"PersistentVolumeClaim","namespace":"default","name":"pvc1","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","apiVersion":"v1","resourceVersion":"557"},"persistentVolumeReclaimPolicy":"Delete","volumeMode":"Filesystem"},"status":{"phase":"Bound"}}]`
				j += `,"pvcs":[{"metadata":{"name":"pvc1","namespace":"default","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","resourceVersion":"567","creationTimestamp":"2021-12-28T01:06:32Z","annotations":{"pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:32Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{"f:storage":{}}},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:volumeName":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{".":{},"f:storage":{}},"f:phase":{}}},"subresource":"status"}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"volumeName":"pv1","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"1Gi"}}}],"storageClasses":[],"priorityClasses":[],"schedulerConfig":{"parallelism":16,"leaderElection":{"leaderElect":true,"leaseDuration":"15s","renewDeadline":"10s","retryPeriod":"2s","resourceLock":"leases","resourceName":"kube-scheduler","resourceNamespace":"kube-system"},"clientConnection":{"kubeconfig":"","acceptContentTypes":"","contentType":"application/vnd.kubernetes.protobuf","qps":50,"burst":100},"healthzBindAddress":"0.0.0.0:10251","metricsBindAddress":"0.0.0.0:10251","enableProfiling":true,"enableContentionProfiling":true,"percentageOfNodesToScore":0,"podInitialBackoffSeconds":1,"podMaxBackoffSeconds":10,"profiles":[{"schedulerName":"default-scheduler","plugins":{"queueSort":{"enabled":[{"name":"PrioritySort"}]},"preFilter":{"enabled":[{"name":"NodeResourcesFit"},{"name":"NodePorts"},{"name":"VolumeRestrictions"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"},{"name":"VolumeBinding"},{"name":"NodeAffinity"}]},"filter":{"enabled":[{"name":"NodeUnschedulable"},{"name":"NodeName"},{"name":"TaintToleration"},{"name":"NodeAffinity"},{"name":"NodePorts"},{"name":"NodeResourcesFit"},{"name":"VolumeRestrictions"},{"name":"EBSLimits"},{"name":"GCEPDLimits"},{"name":"NodeVolumeLimits"},{"name":"AzureDiskLimits"},{"name":"VolumeBinding"},{"name":"VolumeZone"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"}]},"postFilter":{"enabled":[{"name":"DefaultPreemption"}]},"preScore":{"enabled":[{"name":"InterPodAffinity"},{"name":"PodTopologySpread"},{"name":"TaintToleration"},{"name":"NodeAffinity"}]},"score":{"enabled":[{"name":"NodeResourcesBalancedAllocation","weight":1},{"name":"ImageLocality","weight":1},{"name":"InterPodAffinity","weight":1},{"name":"NodeResourcesFit","weight":1},{"name":"NodeAffinity","weight":1},{"name":"PodTopologySpread","weight":2},{"name":"TaintToleration","weight":1}]},"reserve":{"enabled":[{"name":"VolumeBinding"}]},"permit":{},"preBind":{"enabled":[{"name":"VolumeBinding"}]},"bind":{"enabled":[{"name":"DefaultBinder"}]},"postBind":{}},"pluginConfig":[{"name":"DefaultPreemption","args":{"kind":"DefaultPreemptionArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","minCandidateNodesPercentage":10,"minCandidateNodesAbsolute":100}},{"name":"InterPodAffinity","args":{"kind":"InterPodAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","hardPodAffinityWeight":1}},{"name":"NodeAffinity","args":{"kind":"NodeAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2"}},{"name":"NodeResourcesBalancedAllocation","args":{"kind":"NodeResourcesBalancedAllocationArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}},{"name":"NodeResourcesFit","args":{"kind":"NodeResourcesFitArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","scoringStrategy":{"type":"LeastAllocated","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}}},{"name":"PodTopologySpread","args":{"kind":"PodTopologySpreadArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","defaultingType":"System"}},{"name":"VolumeBinding","args":{"kind":"VolumeBindingArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","bindTimeoutSeconds":600}}]}]}}`
				b := []byte(j)
				r := ResourcesApplyConfiguration{}
				if err := json.Unmarshal(b, &r); err != nil {
					panic(err)
				}
				return &r
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: false,
		},
		{
			name: "success import when pv.Status is not exist",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil)
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil)
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "pv1", *cfg.Name)
					assert.Empty(t, cfg.Status)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "pvc1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil)
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil)
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				j := `{"pods":[],"nodes":[],`
				// delete status object
				j += `"pvs":[{"metadata":{"name":"pv1","uid":"b0184e68-5ba6-4533-b3fd-bde9416ad03d","resourceVersion":"565","creationTimestamp":"2021-12-28T01:06:35Z","annotations":{"pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:capacity":{"f:storage":{}},"f:hostPath":{"f:path":{},"f:type":{}},"f:persistentVolumeReclaimPolicy":{},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:phase":{}}},"subresource":"status"},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:claimRef":{}}}}]},"spec":{"capacity":{"storage":"1Gi"},"hostPath":{"path":"/tmp/data","type":"DirectoryOrCreate"},"accessModes":["ReadWriteOnce"],"claimRef":{"kind":"PersistentVolumeClaim","namespace":"default","name":"pvc1","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","apiVersion":"v1","resourceVersion":"557"},"persistentVolumeReclaimPolicy":"Delete","volumeMode":"Filesystem"}}]`
				j += `,"pvcs":[{"metadata":{"name":"pvc1","namespace":"default","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","resourceVersion":"567","creationTimestamp":"2021-12-28T01:06:32Z","annotations":{"pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:32Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{"f:storage":{}}},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:volumeName":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{".":{},"f:storage":{}},"f:phase":{}}},"subresource":"status"}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"volumeName":"pv1","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"1Gi"}}}],"storageClasses":[],"priorityClasses":[],"schedulerConfig":{"parallelism":16,"leaderElection":{"leaderElect":true,"leaseDuration":"15s","renewDeadline":"10s","retryPeriod":"2s","resourceLock":"leases","resourceName":"kube-scheduler","resourceNamespace":"kube-system"},"clientConnection":{"kubeconfig":"","acceptContentTypes":"","contentType":"application/vnd.kubernetes.protobuf","qps":50,"burst":100},"healthzBindAddress":"0.0.0.0:10251","metricsBindAddress":"0.0.0.0:10251","enableProfiling":true,"enableContentionProfiling":true,"percentageOfNodesToScore":0,"podInitialBackoffSeconds":1,"podMaxBackoffSeconds":10,"profiles":[{"schedulerName":"default-scheduler","plugins":{"queueSort":{"enabled":[{"name":"PrioritySort"}]},"preFilter":{"enabled":[{"name":"NodeResourcesFit"},{"name":"NodePorts"},{"name":"VolumeRestrictions"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"},{"name":"VolumeBinding"},{"name":"NodeAffinity"}]},"filter":{"enabled":[{"name":"NodeUnschedulable"},{"name":"NodeName"},{"name":"TaintToleration"},{"name":"NodeAffinity"},{"name":"NodePorts"},{"name":"NodeResourcesFit"},{"name":"VolumeRestrictions"},{"name":"EBSLimits"},{"name":"GCEPDLimits"},{"name":"NodeVolumeLimits"},{"name":"AzureDiskLimits"},{"name":"VolumeBinding"},{"name":"VolumeZone"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"}]},"postFilter":{"enabled":[{"name":"DefaultPreemption"}]},"preScore":{"enabled":[{"name":"InterPodAffinity"},{"name":"PodTopologySpread"},{"name":"TaintToleration"},{"name":"NodeAffinity"}]},"score":{"enabled":[{"name":"NodeResourcesBalancedAllocation","weight":1},{"name":"ImageLocality","weight":1},{"name":"InterPodAffinity","weight":1},{"name":"NodeResourcesFit","weight":1},{"name":"NodeAffinity","weight":1},{"name":"PodTopologySpread","weight":2},{"name":"TaintToleration","weight":1}]},"reserve":{"enabled":[{"name":"VolumeBinding"}]},"permit":{},"preBind":{"enabled":[{"name":"VolumeBinding"}]},"bind":{"enabled":[{"name":"DefaultBinder"}]},"postBind":{}},"pluginConfig":[{"name":"DefaultPreemption","args":{"kind":"DefaultPreemptionArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","minCandidateNodesPercentage":10,"minCandidateNodesAbsolute":100}},{"name":"InterPodAffinity","args":{"kind":"InterPodAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","hardPodAffinityWeight":1}},{"name":"NodeAffinity","args":{"kind":"NodeAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2"}},{"name":"NodeResourcesBalancedAllocation","args":{"kind":"NodeResourcesBalancedAllocationArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}},{"name":"NodeResourcesFit","args":{"kind":"NodeResourcesFitArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","scoringStrategy":{"type":"LeastAllocated","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}}},{"name":"PodTopologySpread","args":{"kind":"PodTopologySpreadArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","defaultingType":"System"}},{"name":"VolumeBinding","args":{"kind":"VolumeBindingArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","bindTimeoutSeconds":600}}]}]}}`
				b := []byte(j)
				r := ResourcesApplyConfiguration{}
				if err := json.Unmarshal(b, &r); err != nil {
					panic(err)
				}
				return &r
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: false,
		},
		{
			name: "success import when pv.Status.Phase is not exist",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, pcs *mock_export.MockPriorityClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Pod{}, nil)
				nodes.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.Node{}, nil)
				pvs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolume{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeApplyConfiguration) {
					assert.Equal(t, "pv1", *cfg.Name)
					assert.Empty(t, cfg.Status.Phase)
				})
				pvcs.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil)
				pvcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&corev1.PersistentVolumeClaim{}, nil).Do(func(_ context.Context, cfg *v1.PersistentVolumeClaimApplyConfiguration) {
					assert.Equal(t, "pvc1", *cfg.Name)
				})
				storageClasss.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&storagev1.StorageClass{}, nil)
				pcs.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&schedulingv1.PriorityClass{}, nil)
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				j := `{"pods":[],"nodes":[],`
				// delete Phase key&value
				j += `"pvs":[{"metadata":{"name":"pv1","uid":"b0184e68-5ba6-4533-b3fd-bde9416ad03d","resourceVersion":"565","creationTimestamp":"2021-12-28T01:06:35Z","annotations":{"pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:capacity":{"f:storage":{}},"f:hostPath":{"f:path":{},"f:type":{}},"f:persistentVolumeReclaimPolicy":{},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:phase":{}}},"subresource":"status"},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:claimRef":{}}}}]},"spec":{"capacity":{"storage":"1Gi"},"hostPath":{"path":"/tmp/data","type":"DirectoryOrCreate"},"accessModes":["ReadWriteOnce"],"claimRef":{"kind":"PersistentVolumeClaim","namespace":"default","name":"pvc1","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","apiVersion":"v1","resourceVersion":"557"},"persistentVolumeReclaimPolicy":"Delete","volumeMode":"Filesystem"},"status":{}}]`
				j += `,"pvcs":[{"metadata":{"name":"pvc1","namespace":"default","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","resourceVersion":"567","creationTimestamp":"2021-12-28T01:06:32Z","annotations":{"pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:32Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{"f:storage":{}}},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:volumeName":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{".":{},"f:storage":{}},"f:phase":{}}},"subresource":"status"}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"volumeName":"pv1","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"1Gi"}}}],"storageClasses":[],"priorityClasses":[],"schedulerConfig":{"parallelism":16,"leaderElection":{"leaderElect":true,"leaseDuration":"15s","renewDeadline":"10s","retryPeriod":"2s","resourceLock":"leases","resourceName":"kube-scheduler","resourceNamespace":"kube-system"},"clientConnection":{"kubeconfig":"","acceptContentTypes":"","contentType":"application/vnd.kubernetes.protobuf","qps":50,"burst":100},"healthzBindAddress":"0.0.0.0:10251","metricsBindAddress":"0.0.0.0:10251","enableProfiling":true,"enableContentionProfiling":true,"percentageOfNodesToScore":0,"podInitialBackoffSeconds":1,"podMaxBackoffSeconds":10,"profiles":[{"schedulerName":"default-scheduler","plugins":{"queueSort":{"enabled":[{"name":"PrioritySort"}]},"preFilter":{"enabled":[{"name":"NodeResourcesFit"},{"name":"NodePorts"},{"name":"VolumeRestrictions"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"},{"name":"VolumeBinding"},{"name":"NodeAffinity"}]},"filter":{"enabled":[{"name":"NodeUnschedulable"},{"name":"NodeName"},{"name":"TaintToleration"},{"name":"NodeAffinity"},{"name":"NodePorts"},{"name":"NodeResourcesFit"},{"name":"VolumeRestrictions"},{"name":"EBSLimits"},{"name":"GCEPDLimits"},{"name":"NodeVolumeLimits"},{"name":"AzureDiskLimits"},{"name":"VolumeBinding"},{"name":"VolumeZone"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"}]},"postFilter":{"enabled":[{"name":"DefaultPreemption"}]},"preScore":{"enabled":[{"name":"InterPodAffinity"},{"name":"PodTopologySpread"},{"name":"TaintToleration"},{"name":"NodeAffinity"}]},"score":{"enabled":[{"name":"NodeResourcesBalancedAllocation","weight":1},{"name":"ImageLocality","weight":1},{"name":"InterPodAffinity","weight":1},{"name":"NodeResourcesFit","weight":1},{"name":"NodeAffinity","weight":1},{"name":"PodTopologySpread","weight":2},{"name":"TaintToleration","weight":1}]},"reserve":{"enabled":[{"name":"VolumeBinding"}]},"permit":{},"preBind":{"enabled":[{"name":"VolumeBinding"}]},"bind":{"enabled":[{"name":"DefaultBinder"}]},"postBind":{}},"pluginConfig":[{"name":"DefaultPreemption","args":{"kind":"DefaultPreemptionArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","minCandidateNodesPercentage":10,"minCandidateNodesAbsolute":100}},{"name":"InterPodAffinity","args":{"kind":"InterPodAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","hardPodAffinityWeight":1}},{"name":"NodeAffinity","args":{"kind":"NodeAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2"}},{"name":"NodeResourcesBalancedAllocation","args":{"kind":"NodeResourcesBalancedAllocationArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}},{"name":"NodeResourcesFit","args":{"kind":"NodeResourcesFitArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","scoringStrategy":{"type":"LeastAllocated","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}}},{"name":"PodTopologySpread","args":{"kind":"PodTopologySpreadArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","defaultingType":"System"}},{"name":"VolumeBinding","args":{"kind":"VolumeBindingArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","bindTimeoutSeconds":600}}]}]}}`
				b := []byte(j)
				r := ResourcesApplyConfiguration{}
				if err := json.Unmarshal(b, &r); err != nil {
					panic(err)
				}
				return &r
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockSchedulerService := mock_export.NewMockSchedulerService(ctrl)
			mockPriorityClassService := mock_export.NewMockPriorityClassService(ctrl)
			mockStorageClassService := mock_export.NewMockStorageClassService(ctrl)
			mockPVCService := mock_export.NewMockPersistentVolumeClaimService(ctrl)
			mockPVService := mock_export.NewMockPersistentVolumeService(ctrl)
			mockNodeService := mock_export.NewMockNodeService(ctrl)
			mockPodService := mock_export.NewMockPodService(ctrl)
			fakeclientset := tt.prepareFakeClientSetFn()

			s := NewResourcesService(fakeclientset, mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockPriorityClassService, mockSchedulerService)
			tt.prepareEachServiceMockFn(mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockPriorityClassService, mockSchedulerService)

			if err := s.Import(context.Background(), tt.applyConfiguration()); (err != nil) != tt.wantErr {
				t.Fatalf("Import() %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}
