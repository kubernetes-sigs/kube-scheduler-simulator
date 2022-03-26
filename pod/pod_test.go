package pod

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
		wantPodName            string
		wantReturn             corev1.Pod
		wantErr                bool
	}{
		{
			name: "get specifed pod",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName2).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: testDefaultNamespaceName1,
			wantPodName:     "pod1",
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeclientset := tt.prepareFakeClientSetFn()
			s := NewPodService(fakeclientset)
			pod, err := s.Get(context.Background(), tt.wantPodName, tt.targetNamespace)

			if (err != nil) != tt.wantErr || (pod.Name != tt.wantPodName) {
				t.Fatalf("Get() error = %v, wantErr %v\npod name = %s, want %s", err, tt.wantErr, pod.Name, tt.wantPodName)
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
		wantReturn             *corev1.PodList
		wantErr                bool
	}{
		{
			name: "list pods spcified namespace",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName2).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod3",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: testDefaultNamespaceName1,
			wantReturn: &corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: testDefaultNamespaceName1},
						Spec:       corev1.PodSpec{NodeName: "node1"},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: testDefaultNamespaceName1},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "list all pods",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName2).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod3",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: metav1.NamespaceAll,
			wantReturn: &corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: testDefaultNamespaceName1},
						Spec:       corev1.PodSpec{NodeName: "node1"},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: testDefaultNamespaceName1},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod3", Namespace: testDefaultNamespaceName2},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "list empty if there is no pod",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			targetNamespace: metav1.NamespaceAll,
			wantReturn:      &corev1.PodList{},
			wantErr:         false,
		},
		{
			name: "list empty if no pod found",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: testDefaultNamespaceName2,
			wantReturn:      &corev1.PodList{},
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeclientset := tt.prepareFakeClientSetFn()
			s := NewPodService(fakeclientset)
			pods, err := s.List(context.Background(), tt.targetNamespace)
			diffResponse := cmp.Diff(pods, tt.wantReturn)
			if diffResponse != "" || (err != nil) != tt.wantErr {
				t.Fatalf("List() %v test, \nerror = %v, wantErr %v\n%s", tt.name, err, tt.wantErr, diffResponse)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		prepareFakeClientSetFn func() *fake.Clientset
		targetNamespace        string
		targetPodName          string
		wantReturn             *corev1.PodList
		wantErr                bool
	}{
		{
			name: "delete pod specified namespace",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
				}, metav1.CreateOptions{})
				return c
			},
			targetNamespace: testDefaultNamespaceName1,
			targetPodName:   "pod1",
			wantReturn: &corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: testDefaultNamespaceName1},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeclientset := tt.prepareFakeClientSetFn()
			s := NewPodService(fakeclientset)
			err := s.Delete(context.Background(), tt.targetPodName, tt.targetNamespace)
			pods, _ := s.List(context.Background(), tt.targetNamespace)
			diffResponse := cmp.Diff(pods, tt.wantReturn)
			if diffResponse != "" || (err != nil) != tt.wantErr {
				t.Fatalf("Apply() %v test, \nerror = %v, wantErr %v\n%s", tt.name, err, tt.wantErr, diffResponse)
			}
		})
	}
}

func TestService_DeleteCollection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		prepareFakeClientSetFn func() *fake.Clientset
		lopt                   metav1.ListOptions
		wantErr                bool
	}{
		{
			name: "delete all pods",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()

				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName2).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})

				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod3",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods(testDefaultNamespaceName2).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod3",
					},
				}, metav1.CreateOptions{})

				return c
			},
			lopt:    metav1.ListOptions{},
			wantErr: false,
		},
		{
			name: "delete all pods on node",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()

				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})

				c.CoreV1().Pods(testDefaultNamespaceName1).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
				}, metav1.CreateOptions{})

				return c
			},
			lopt: metav1.ListOptions{
				FieldSelector: "spec.nodeName!=",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeclientset := tt.prepareFakeClientSetFn()
			s := NewPodService(fakeclientset)
			if err := s.DeleteCollection(context.Background(), tt.lopt); (err != nil) != tt.wantErr {
				t.Fatalf("DeleteCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
