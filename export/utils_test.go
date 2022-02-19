package export

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	cfgstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
)

func Test_convertPodListToApplyConfigurationList(t *testing.T) {
	t.Parallel()
	const defaultNamespaceName = "default"
	tests := []struct {
		name       string
		input      func() []corev1.Pod
		wantReturn func() []v1.PodApplyConfiguration
		wantErr    bool
	}{
		{
			name: "convert Pod list to PodApplyConfiguration list",
			input: func() []corev1.Pod {
				return []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod1",
							Namespace: defaultNamespaceName,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod2",
							Namespace: defaultNamespaceName,
						},
					},
				}
			},
			wantReturn: func() []v1.PodApplyConfiguration {
				return []v1.PodApplyConfiguration{
					*new(v1.PodApplyConfiguration).WithName("pod1").WithNamespace(defaultNamespaceName),
					*new(v1.PodApplyConfiguration).WithName("pod2").WithNamespace(defaultNamespaceName),
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := tt.input()
			out, err := convertPodListToApplyConfigurationList(input)
			want := tt.wantReturn()
			for i, o := range out {
				diffResponse := cmp.Diff(o.Name, want[i].Name)
				if diffResponse != "" || (err != nil) != tt.wantErr {
					t.Fatalf("convertToApplyConfiguration() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
				}
			}
		})
	}
}

func Test_convertNodeListToApplyConfigurationList(t *testing.T) {
	t.Parallel()
	const defaultNamespaceName = "default"
	tests := []struct {
		name       string
		input      func() []corev1.Node
		wantReturn func() []v1.NodeApplyConfiguration
		wantErr    bool
	}{
		{
			name: "convert Node list to NodeApplyConfiguration list",
			input: func() []corev1.Node {
				return []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node1",
							Namespace: defaultNamespaceName,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "node2",
							Namespace: defaultNamespaceName,
						},
					},
				}
			},
			wantReturn: func() []v1.NodeApplyConfiguration {
				return []v1.NodeApplyConfiguration{
					*new(v1.NodeApplyConfiguration).WithName("node1").WithNamespace(defaultNamespaceName),
					*new(v1.NodeApplyConfiguration).WithName("node2").WithNamespace(defaultNamespaceName),
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := tt.input()
			out, err := convertNodeListToApplyConfigurationList(input)
			want := tt.wantReturn()
			for i, o := range out {
				diffResponse := cmp.Diff(o.Name, want[i].Name)
				if diffResponse != "" || (err != nil) != tt.wantErr {
					t.Fatalf("convertNodeListToApplyConfigurationList() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
				}
			}
		})
	}
}

func Test_convertPvListToApplyConfigurationList(t *testing.T) {
	t.Parallel()
	const defaultNamespaceName = "default"
	tests := []struct {
		name       string
		input      func() []corev1.PersistentVolume
		wantReturn func() []v1.PersistentVolumeApplyConfiguration
		wantErr    bool
	}{
		{
			name: "convert PersistentVolume list to PersistentVolumeApplyConfiguration list",
			input: func() []corev1.PersistentVolume {
				return []corev1.PersistentVolume{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pv1",
							Namespace: defaultNamespaceName,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pv2",
							Namespace: defaultNamespaceName,
						},
					},
				}
			},
			wantReturn: func() []v1.PersistentVolumeApplyConfiguration {
				return []v1.PersistentVolumeApplyConfiguration{
					*new(v1.PersistentVolumeApplyConfiguration).WithName("pv1").WithNamespace(defaultNamespaceName),
					*new(v1.PersistentVolumeApplyConfiguration).WithName("pv2").WithNamespace(defaultNamespaceName),
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := tt.input()
			out, err := convertPvListToApplyConfigurationList(input)
			want := tt.wantReturn()
			for i, o := range out {
				diffResponse := cmp.Diff(o.Name, want[i].Name)
				if diffResponse != "" || (err != nil) != tt.wantErr {
					t.Fatalf("convertPvListToApplyConfigurationList() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
				}
			}
		})
	}
}

func Test_convertPvcListToApplyConfigurationList(t *testing.T) {
	t.Parallel()
	const defaultNamespaceName = "default"
	tests := []struct {
		name       string
		input      func() []corev1.PersistentVolumeClaim
		wantReturn func() []v1.PersistentVolumeClaimApplyConfiguration
		wantErr    bool
	}{
		{
			name: "convert PersistentVolumeClaim list to PersistentVolumeClaimApplyConfiguration list",
			input: func() []corev1.PersistentVolumeClaim {
				return []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc1",
							Namespace: defaultNamespaceName,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc2",
							Namespace: defaultNamespaceName,
						},
					},
				}
			},
			wantReturn: func() []v1.PersistentVolumeClaimApplyConfiguration {
				return []v1.PersistentVolumeClaimApplyConfiguration{
					*new(v1.PersistentVolumeClaimApplyConfiguration).WithName("pvc1").WithNamespace(defaultNamespaceName),
					*new(v1.PersistentVolumeClaimApplyConfiguration).WithName("pvc2").WithNamespace(defaultNamespaceName),
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := tt.input()
			out, err := convertPvcListToApplyConfigurationList(input)
			want := tt.wantReturn()
			for i, o := range out {
				diffResponse := cmp.Diff(o.Name, want[i].Name)
				if diffResponse != "" || (err != nil) != tt.wantErr {
					t.Fatalf("convertPvcListToApplyConfigurationList() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
				}
			}
		})
	}
}

func Test_convertStorageClassesListToApplyConfigurationList(t *testing.T) {
	t.Parallel()
	const defaultNamespaceName = "default"
	tests := []struct {
		name       string
		input      func() []storagev1.StorageClass
		wantReturn func() []cfgstoragev1.StorageClassApplyConfiguration
		wantErr    bool
	}{
		{
			name: "convert StorageClass list to StorageClassApplyConfiguration list",
			input: func() []storagev1.StorageClass {
				return []storagev1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "sc1",
							Namespace: defaultNamespaceName,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "sc2",
							Namespace: defaultNamespaceName,
						},
					},
				}
			},
			wantReturn: func() []cfgstoragev1.StorageClassApplyConfiguration {
				return []cfgstoragev1.StorageClassApplyConfiguration{
					*new(cfgstoragev1.StorageClassApplyConfiguration).WithName("sc1").WithNamespace(defaultNamespaceName),
					*new(cfgstoragev1.StorageClassApplyConfiguration).WithName("sc2").WithNamespace(defaultNamespaceName),
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := tt.input()
			out, err := convertStorageClassesListToApplyConfigurationList(input)
			want := tt.wantReturn()
			for i, o := range out {
				diffResponse := cmp.Diff(o.Name, want[i].Name)
				if diffResponse != "" || (err != nil) != tt.wantErr {
					t.Fatalf("convertStorageClassesListToApplyConfigurationList() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
				}
			}
		})
	}
}

func Test_convertPriorityClassesListToApplyConfigurationList(t *testing.T) {
	t.Parallel()
	const defaultNamespaceName = "default"
	tests := []struct {
		name       string
		input      func() []schedulingv1.PriorityClass
		wantReturn func() []schedulingcfgv1.PriorityClassApplyConfiguration
		wantErr    bool
	}{
		{
			name: "convert PriorityClass list to PriorityClassApplyConfiguration list",
			input: func() []schedulingv1.PriorityClass {
				return []schedulingv1.PriorityClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pc1",
							Namespace: defaultNamespaceName,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pc2",
							Namespace: defaultNamespaceName,
						},
					},
				}
			},
			wantReturn: func() []schedulingcfgv1.PriorityClassApplyConfiguration {
				return []schedulingcfgv1.PriorityClassApplyConfiguration{
					*new(schedulingcfgv1.PriorityClassApplyConfiguration).WithName("pc1").WithNamespace(defaultNamespaceName),
					*new(schedulingcfgv1.PriorityClassApplyConfiguration).WithName("pc2").WithNamespace(defaultNamespaceName),
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := tt.input()
			out, err := convertPriorityClassesListToApplyConfigurationList(input)
			want := tt.wantReturn()
			for i, o := range out {
				diffResponse := cmp.Diff(o.Name, want[i].Name)
				if diffResponse != "" || (err != nil) != tt.wantErr {
					t.Fatalf("convertPriorityClassesListToApplyConfigurationList() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
				}
			}
		})
	}
}
