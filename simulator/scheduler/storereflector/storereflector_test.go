package storereflector

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/storereflector/mock_storereflector"
)

const (
	ExtenderFilterResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/extender-filter-result"
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
				m.EXPECT().GetStoredResult(gomock.Any()).Return(map[string]string{ExtenderFilterResultAnnotationKey: "some results"})
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
			wantAnnotation: map[string]string{ExtenderFilterResultAnnotationKey: "some results", ResultsHistoryAnnotation: "[{\"kube-scheduler-simulator.sigs.k8s.io/extender-filter-result\":\"some results\"}]"},
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
			original := p.DeepCopy()
			fn(corev1.Pod{}, p)

			// Check that the function doesn't mutate the input object,
			// which is shared with other event handlers.
			assert.Equal(t, original, p)

			updatedPod, _ := c.CoreV1().Pods(tt.podNamespace).Get(context.Background(), tt.podName, metav1.GetOptions{})

			assert.Equal(t, tt.wantAnnotation, updatedPod.Annotations)
		})
	}
}

func Test_updateResultHistory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		p       *corev1.Pod
		m       map[string]string
		wantErr assert.ErrorAssertionFunc
		wantPod *corev1.Pod
	}{
		{
			name: "success: Pod doesn't have annotation yet",
			p: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			},
			m: map[string]string{
				"result1": "fuga",
				"result2": "hoge",
			},
			wantPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ResultsHistoryAnnotation: `[{"result1":"fuga","result2":"hoge"}]`,
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "success: Pod already has annotation",
			p: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ResultsHistoryAnnotation: `[{"result1":"fuga","result2":"hoge"}]`,
					},
				},
			},
			m: map[string]string{
				"result1": "fuga2",
				"result2": "hoge2",
			},
			wantPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ResultsHistoryAnnotation: `[{"result1":"fuga","result2":"hoge"},{"result1":"fuga2","result2":"hoge2"}]`,
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "fail: Pod has broken value on annotation",
			p: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ResultsHistoryAnnotation: `broken`,
					},
				},
			},
			m: map[string]string{
				"result1": "fuga2",
				"result2": "hoge2",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := tt.p
			tt.wantErr(t, updateResultHistory(p, tt.m), fmt.Sprintf("updateResultHistory(%v, %v)", p, tt.m))
			if d := cmp.Diff(p, tt.wantPod); d != "" && tt.wantPod != nil {
				t.Fatalf("unexpected Pod: %v", d)
			}
		})
	}
}
