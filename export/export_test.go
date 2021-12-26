package export

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/export/mock_export"
	schedulerCfg "github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/defaultconfig"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
)

func TestService_Export(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		wantReturn               *Resources
		wantErr                  bool
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
			wantReturn: &Resources{
				Pods:            []corev1.Pod{},
				Nodes:           []corev1.Node{},
				Pvs:             []corev1.PersistentVolume{},
				Pvcs:            []corev1.PersistentVolumeClaim{},
				StorageClasses:  []storagev1.StorageClass{},
				SchedulerConfig: &v1beta2config.KubeSchedulerConfiguration{},
			},
			wantErr: false,
		},
		{
			name: "export failure on List of PodService",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list pods"))
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list nodes"))
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list PersistentVolumes"))
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list PersistentVolumeClaims"))
				storageClasss.EXPECT().List(gomock.Any()).Return(&storagev1.StorageClassList{Items: []storagev1.StorageClass{}}, nil)
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
				pods.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				nodes.EXPECT().List(gomock.Any()).Return(&corev1.NodeList{Items: []corev1.Node{}}, nil)
				pvs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeList{Items: []corev1.PersistentVolume{}}, nil)
				pvcs.EXPECT().List(gomock.Any()).Return(&corev1.PersistentVolumeClaimList{Items: []corev1.PersistentVolumeClaim{}}, nil)
				storageClasss.EXPECT().List(gomock.Any()).Return(nil, xerrors.Errorf("list storageClasses"))
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
			mockStorageClassService := mock_export.NewMockStorageClassService(ctrl)
			mockPVCService := mock_export.NewMockPersistentVolumeClaimService(ctrl)
			mockPVService := mock_export.NewMockPersistentVolumeService(ctrl)
			mockNodeService := mock_export.NewMockNodeService(ctrl)
			mockPodService := mock_export.NewMockPodService(ctrl)
			fakeclientset := tt.prepareFakeClientSetFn()

			s := NewResourcesService(fakeclientset, mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockSchedulerService)
			tt.prepareEachServiceMockFn(mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockSchedulerService)
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
		prepareEachServiceMockFn func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		applyConfiguration       func() *ResourcesApplyConfiguration
		wantErr                  bool
	}{
		{
			name: "import all success",
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("something wrong", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("something wrong"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("something wrong").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("something wrong", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("something wrong"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Bound")).WithSpec(v1.PersistentVolumeSpec().WithClaimRef(v1.ObjectReference().WithName("PVC1"))))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Bound")).WithSpec(v1.PersistentVolumeSpec().WithClaimRef(v1.ObjectReference().WithName("PVC1"))))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
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
			prepareEachServiceMockFn: func(pods *mock_export.MockPodService, nodes *mock_export.MockNodeService, pvs *mock_export.MockPersistentVolumeService, pvcs *mock_export.MockPersistentVolumeClaimService, storageClasss *mock_export.MockStorageClassService, schedulers *mock_export.MockSchedulerService) {
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
				schedulers.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesApplyConfiguration {
				var pods = []v1.PodApplyConfiguration{}
				var nodes = []v1.NodeApplyConfiguration{}
				var pvs = []v1.PersistentVolumeApplyConfiguration{}
				var pvcs = []v1.PersistentVolumeClaimApplyConfiguration{}
				var storageclasses = []confstoragev1.StorageClassApplyConfiguration{}
				pods = append(pods, *v1.Pod("Pod1", "default"))
				nodes = append(nodes, *v1.Node("Node1"))
				pvs = append(pvs, *v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Bound")).WithSpec(v1.PersistentVolumeSpec().WithClaimRef(v1.ObjectReference().WithName("PVC1"))))
				pvcs = append(pvcs, *v1.PersistentVolumeClaim("PVC1", "default"))
				storageclasses = append(storageclasses, *confstoragev1.StorageClass("StorageClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesApplyConfiguration{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					SchedulerConfig: config,
				}
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
			mockStorageClassService := mock_export.NewMockStorageClassService(ctrl)
			mockPVCService := mock_export.NewMockPersistentVolumeClaimService(ctrl)
			mockPVService := mock_export.NewMockPersistentVolumeService(ctrl)
			mockNodeService := mock_export.NewMockNodeService(ctrl)
			mockPodService := mock_export.NewMockPodService(ctrl)
			fakeclientset := tt.prepareFakeClientSetFn()

			s := NewResourcesService(fakeclientset, mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockSchedulerService)
			tt.prepareEachServiceMockFn(mockPodService, mockNodeService, mockPVService, mockPVCService, mockStorageClassService, mockSchedulerService)

			if err := s.Import(context.Background(), tt.applyConfiguration()); (err != nil) != tt.wantErr {
				t.Fatalf("Import() %v test, \nerror = %v", tt.name, err)
			}
		})
	}

}
