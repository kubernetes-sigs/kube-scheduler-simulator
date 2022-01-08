package node

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/node/mock_node"
)

func TestService_Delete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		nodeName                string
		preparePodServiceMockFn func(m *mock_node.MockPodService)
		prepareFakeClientSetFn  func() *fake.Clientset
		wantErr                 bool
	}{
		{
			name:     "delete node and pods on node",
			nodeName: "node1",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "this-pod-will-not-be-deleted",
							},
							Spec: corev1.PodSpec{
								NodeName: "other-node",
							},
						},
					},
				}, nil)
				m.EXPECT().Delete(gomock.Any(), "pod1").Return(nil)
				m.EXPECT().Delete(gomock.Any(), "pod2").Return(nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				// add test data.
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			wantErr: false,
		},
		{
			name:     "one of deleting pods fail",
			nodeName: "node1",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "this-pod-will-not-be-deleted",
							},
							Spec: corev1.PodSpec{
								NodeName: "other-node",
							},
						},
					},
				}, nil)
				m.EXPECT().Delete(gomock.Any(), "pod1").Return(nil)
				m.EXPECT().Delete(gomock.Any(), "pod2").Return(errors.New("error"))
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
		{
			name:     "delete node with no pods",
			nodeName: "node1",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				// add test data.
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
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

			mockPodService := mock_node.NewMockPodService(ctrl)
			tt.preparePodServiceMockFn(mockPodService)
			fakeclientset := tt.prepareFakeClientSetFn()

			s := NewNodeService(fakeclientset, mockPodService)
			if err := s.Delete(context.Background(), tt.nodeName); (err != nil) != tt.wantErr {
				t.Fatalf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_DeleteAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		preparePodServiceMockFn func(m *mock_node.MockPodService)
		prepareFakeClientSetFn  func() *fake.Clientset
		wantErr                 bool
	}{
		{
			name: "delete all nodes and pods scheduled on them",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "other-pod1",
							},
							Spec: corev1.PodSpec{
								NodeName: "",
							},
						},
					},
				}, nil)
				m.EXPECT().DeleteAllScheduledPod(gomock.Any()).Return(nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				// add test data.
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			wantErr: false,
		},
		{
			name: "delete nodes with no pods",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
				m.EXPECT().DeleteAllScheduledPod(gomock.Any()).Return(nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				// add test data.
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			wantErr: false,
		},
		{
			name: "fail if delteing all pods returns error",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().List(gomock.Any()).Return(&corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
					},
				}, nil)
				m.EXPECT().DeleteAllScheduledPod(gomock.Any()).Return(errors.New("error"))
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				return c
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockPodService := mock_node.NewMockPodService(ctrl)
			tt.preparePodServiceMockFn(mockPodService)
			fakeclientset := tt.prepareFakeClientSetFn()

			s := NewNodeService(fakeclientset, mockPodService)
			if err := s.DeleteAll(context.Background()); (err != nil) != tt.wantErr {
				t.Fatalf("DeleteAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
