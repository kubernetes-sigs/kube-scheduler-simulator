package pod

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestService_DeleteCollection(t *testing.T) {
	t.Parallel()
	const defaultNamespaceName = "default"
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

				c.CoreV1().Pods(defaultNamespaceName).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})

				c.CoreV1().Pods(defaultNamespaceName).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
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

				c.CoreV1().Pods(defaultNamespaceName).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})

				c.CoreV1().Pods(defaultNamespaceName).Create(context.Background(), &corev1.Pod{
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
