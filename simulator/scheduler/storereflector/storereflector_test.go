package storereflector

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/storereflector/mock_storereflector"
)

const (
	ExtenderFilterResultAnnotationKey = "scheduler-simulator/extender-filter-result"
	ResultStoreKey                    = "ExtenderResultStoreKey"
)

func TestReflector_storeAllResultToPodFunc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                        string
		podName                     string
		podNamespace                string
		prepareMockResultStoreSetFn func(m *mock_storereflector.MockResultStore)
		prepareFakeClientSetFn      func() *fake.Clientset
		wantAnnotation              map[string]string
	}{
		{
			name:         "success",
			podName:      "pod1",
			podNamespace: "default",
			prepareMockResultStoreSetFn: func(m *mock_storereflector.MockResultStore) {
				m.EXPECT().AddStoredResultToPod(gomock.Any()).Do(func(pod *corev1.Pod) {
					metav1.SetMetaDataAnnotation(&pod.ObjectMeta, ExtenderFilterResultAnnotationKey, "some results")
				})
				m.EXPECT().DeleteData(gomock.Any())
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods("default").Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
					},
				}, metav1.CreateOptions{})
				return c
			},
			wantAnnotation: map[string]string{ExtenderFilterResultAnnotationKey: "some results"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.prepareFakeClientSetFn()
			ctrl := gomock.NewController(t)
			rs := mock_storereflector.NewMockResultStore(ctrl)
			tt.prepareMockResultStoreSetFn(rs)
			r := &reflector{
				resultStores: map[string]ResultStore{ResultStoreKey: rs},
			}
			fn := r.storeAllResultToPodFunc(c)
			p, _ := c.CoreV1().Pods(tt.podNamespace).Get(context.Background(), tt.podName, metav1.GetOptions{})
			fn(corev1.Pod{}, p)

			assert.Equal(t, tt.wantAnnotation, p.Annotations)
		})
	}
}
