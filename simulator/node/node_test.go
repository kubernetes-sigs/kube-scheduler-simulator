package node

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/simulator/node/mock_node"
)

const (
	testDefaultNamespaceName1 = "default1"
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
				m.EXPECT().List(gomock.Any(), metav1.NamespaceAll).Return(&corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "pod1",
								Namespace: testDefaultNamespaceName1,
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "pod2",
								Namespace: testDefaultNamespaceName1,
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "this-pod-will-not-be-deleted",
								Namespace: testDefaultNamespaceName1,
							},
							Spec: corev1.PodSpec{
								NodeName: "other-node",
							},
						},
					},
				}, nil)
				m.EXPECT().Delete(gomock.Any(), "pod1", testDefaultNamespaceName1).Return(nil)
				m.EXPECT().Delete(gomock.Any(), "pod2", testDefaultNamespaceName1).Return(nil)
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
				m.EXPECT().List(gomock.Any(), metav1.NamespaceAll).Return(&corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "pod1",
								Namespace: testDefaultNamespaceName1,
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "pod2",
								Namespace: testDefaultNamespaceName1,
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "this-pod-will-not-be-deleted",
								Namespace: testDefaultNamespaceName1,
							},
							Spec: corev1.PodSpec{
								NodeName: "other-node",
							},
						},
					},
				}, nil)
				m.EXPECT().Delete(gomock.Any(), "pod1", testDefaultNamespaceName1).Return(nil)
				m.EXPECT().Delete(gomock.Any(), "pod2", testDefaultNamespaceName1).Return(errors.New("error"))
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
				m.EXPECT().List(gomock.Any(), metav1.NamespaceAll).Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)
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

func TestService_DeleteCollection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		preparePodServiceMockFn func(m *mock_node.MockPodService)
		prepareFakeClientSetFn  func() *fake.Clientset
		lopts                   metav1.ListOptions
		wantErr                 bool
	}{
		{
			name: "delete all nodes and pods scheduled on them",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().DeleteCollection(gomock.Any(), gomock.Any(), metav1.ListOptions{
					FieldSelector: "spec.nodeName=node1",
				}).Return(nil)
				m.EXPECT().DeleteCollection(gomock.Any(), gomock.Any(), metav1.ListOptions{
					FieldSelector: "spec.nodeName=node2",
				}).Return(nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
					},
				}, metav1.CreateOptions{})
				return c
			},
			lopts: metav1.ListOptions{
				FieldSelector: "spec.nodeName!=",
			},
			wantErr: false,
		},
		{
			name: "delete nodes with no pods",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().DeleteCollection(gomock.Any(), gomock.Any(), metav1.ListOptions{
					FieldSelector: "spec.nodeName=node1",
				}).Return(nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			lopts: metav1.ListOptions{
				FieldSelector: "spec.nodeName!=",
			},
			wantErr: false,
		},
		{
			name: "fail if deleteing all pods returns error",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().DeleteCollection(gomock.Any(), gomock.Any(), metav1.ListOptions{
					FieldSelector: "spec.nodeName=node1",
				}).Return(errors.New("error"))
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods("default").Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			lopts: metav1.ListOptions{
				FieldSelector: "spec.nodeName!=",
			},
			wantErr: true,
		},
		{
			name: "delete nodes with multiple existing namespaced pods",
			preparePodServiceMockFn: func(m *mock_node.MockPodService) {
				m.EXPECT().DeleteCollection(gomock.Any(), "default1", metav1.ListOptions{
					FieldSelector: "spec.nodeName=node1",
				}).Return(nil)
				m.EXPECT().DeleteCollection(gomock.Any(), "default2", metav1.ListOptions{
					FieldSelector: "spec.nodeName=node1",
				}).Return(nil)
				m.EXPECT().DeleteCollection(gomock.Any(), "default3", metav1.ListOptions{
					FieldSelector: "spec.nodeName=node1",
				}).Return(errors.New("error"))
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default2",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods("default1").Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Pods("default2").Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod2",
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}, metav1.CreateOptions{})
				c.CoreV1().Nodes().Create(context.Background(), &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}, metav1.CreateOptions{})
				return c
			},
			lopts: metav1.ListOptions{
				FieldSelector: "spec.nodeName!=",
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
			if err := s.DeleteCollection(context.Background(), tt.lopts); (err != nil) != tt.wantErr {
				t.Fatalf("DeleteCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
