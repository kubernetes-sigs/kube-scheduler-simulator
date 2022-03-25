package persistentvolumeclaim

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testDefaultNamespaceName1 = "default1"
	testDefaultNamespaceName2 = "default2"
)

func TestService_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		prepareFakeClientSetFn func() *fake.Clientset
		targetNamespace        string
		wantPVName             string
		wantReturn             corev1.PersistentVolumeClaim
		wantErr                bool
	}{
		{
			name: "get specifed pvc",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName1).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc1",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName1).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc2",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName2).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc3",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: testDefaultNamespaceName1,
			wantPVName:      "pvc1",
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeclientset := tt.prepareFakeClientSetFn()
			s := NewPersistentVolumeClaimService(fakeclientset)
			pod, err := s.Get(context.Background(), tt.wantPVName, tt.targetNamespace)

			if (err != nil) != tt.wantErr || (pod.Name != tt.wantPVName) {
				t.Fatalf("Get() error = %v, wantErr %v\npod name = %s, want %s", err, tt.wantErr, pod.Name, tt.wantPVName)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		prepareFakeClientSetFn func() *fake.Clientset
		targetNamespace        string
		wantReturn             *corev1.PersistentVolumeClaimList
		wantErr                bool
	}{
		{
			name: "list pvcs spcified namespace",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName1).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc1",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName1).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc2",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName2).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc3",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: testDefaultNamespaceName1,
			wantReturn: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc1",
							Namespace: testDefaultNamespaceName1,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							VolumeName: "pv1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc2",
							Namespace: testDefaultNamespaceName1,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							VolumeName: "pv1",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "list all pvcs",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName1).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc1",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName1).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc2",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName2).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc3",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: corev1.NamespaceAll,
			wantReturn: &corev1.PersistentVolumeClaimList{
				Items: []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc1",
							Namespace: testDefaultNamespaceName1,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							VolumeName: "pv1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc2",
							Namespace: testDefaultNamespaceName1,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							VolumeName: "pv1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc3",
							Namespace: testDefaultNamespaceName2,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							VolumeName: "pv1",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "list empty if there is no pvc",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			targetNamespace: testDefaultNamespaceName1,
			wantReturn:      &corev1.PersistentVolumeClaimList{},
			wantErr:         false,
		},
		{
			name: "list empty if no pvc found",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().PersistentVolumeClaims(testDefaultNamespaceName1).Create(context.Background(), &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc1",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: testDefaultNamespaceName2,
			wantReturn:      &corev1.PersistentVolumeClaimList{},
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeclientset := tt.prepareFakeClientSetFn()
			s := NewPersistentVolumeClaimService(fakeclientset)
			pods, err := s.List(context.Background(), tt.targetNamespace)
			diffResponse := cmp.Diff(pods, tt.wantReturn)
			if diffResponse != "" || (err != nil) != tt.wantErr {
				t.Fatalf("List() %v test, \nerror = %v, wantErr %v\n%s", tt.name, err, tt.wantErr, diffResponse)
			}
		})
	}
}
