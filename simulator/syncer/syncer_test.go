package syncer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/restmapper"
	scheduling "k8s.io/kubernetes/pkg/apis/scheduling/v1"
	storage "k8s.io/kubernetes/pkg/apis/storage/v1"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourceapplier"
)

//nolint:gocognit // it is because of huge test cases.
func TestSyncerWithPod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		initialPodsInSrcCluster []*v1.Pod
		podsCreatedInSrcCluster []*v1.Pod
		podsUpdatedInSrcCluster []*v1.Pod
		podsDeletedInSrcCluster []*v1.Pod
		afterPodsInDestCluster  []*v1.Pod
	}{
		{
			name: "unscheduled pod is created in src cluster",
			initialPodsInSrcCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
			},
			podsCreatedInSrcCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-2",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-3",
						Namespace: "default-3",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-3",
							},
						},
					},
				},
			},
			afterPodsInDestCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-2",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-3",
						Namespace: "default-3",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-3",
							},
						},
					},
				},
			},
		},
		{
			name: "pod is created and deleted in src cluster",
			podsCreatedInSrcCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "default-2",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-2",
							},
						},
					},
				},
			},
			podsDeletedInSrcCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
			},
			afterPodsInDestCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "default-2",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-2",
							},
						},
					},
				},
			},
		},
		{
			name: "unscheduled pod is updated in src cluster",
			podsCreatedInSrcCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
			},
			podsUpdatedInSrcCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
			},
			afterPodsInDestCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
			},
		},
		{
			name: "scheduled pod is NOT updated in src cluster",
			podsCreatedInSrcCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
			},
			podsUpdatedInSrcCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},

						NodeName: "node-1", // Got NodeName, so this Pod is scheduled.
					},
				},
			},
			afterPodsInDestCluster: []*v1.Pod{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "container-1",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := runtime.NewScheme()
			v1.AddToScheme(s)
			scheduling.AddToScheme(s)
			storage.AddToScheme(s)
			src := dynamicFake.NewSimpleDynamicClient(s)
			dest := dynamicFake.NewSimpleDynamicClient(s)
			resources := []*restmapper.APIGroupResources{
				{
					Group: metav1.APIGroup{
						Versions: []metav1.GroupVersionForDiscovery{
							{Version: "v1"},
						},
					},
					VersionedResources: map[string][]metav1.APIResource{
						"v1": {
							{Name: "pods", Namespaced: true, Kind: "Pod"},
						},
					},
				},
				{
					Group: metav1.APIGroup{
						Versions: []metav1.GroupVersionForDiscovery{
							{Version: "v1"},
						},
					},
					VersionedResources: map[string][]metav1.APIResource{
						"v1": {
							{Name: "nodes", Namespaced: true, Kind: "Node"},
						},
					},
				},
			}
			mapper := restmapper.NewDiscoveryRESTMapper(resources)
			resourceApplier := resourceapplier.New(dest, mapper, resourceapplier.Options{})
			service := New(src, resourceApplier, Options{})

			ctx, cancel := context.WithCancel(context.Background())

			createdPods := sets.New[podKey]()
			for _, pod := range tt.initialPodsInSrcCluster {
				p, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
				if err != nil {
					t.Fatalf("failed to convert pod to unstructured: %v", err)
				}
				unstructedPod := &unstructured.Unstructured{Object: p}
				_, err = src.Resource(v1.Resource("pods").WithVersion("v1")).Namespace(pod.Namespace).Create(ctx, unstructedPod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("failed to create pod: %v", err)
				}
				createdPods.Insert(podKey{pod.Name, pod.Namespace})
			}

			go service.Run(ctx)
			defer cancel()

			for _, pod := range tt.podsCreatedInSrcCluster {
				p, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
				if err != nil {
					t.Fatalf("failed to convert pod to unstructured: %v", err)
				}
				unstructedPod := &unstructured.Unstructured{Object: p}
				_, err = src.Resource(v1.Resource("pods").WithVersion("v1")).Namespace(pod.Namespace).Create(ctx, unstructedPod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("failed to create pod: %v", err)
				}
				createdPods.Insert(podKey{pod.Name, pod.Namespace})
			}

			for _, pod := range tt.podsUpdatedInSrcCluster {
				p, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
				if err != nil {
					t.Fatalf("failed to convert pod to unstructured: %v", err)
				}
				unstructedPod := &unstructured.Unstructured{Object: p}
				_, err = src.Resource(v1.Resource("pods").WithVersion("v1")).Namespace(pod.Namespace).Update(ctx, unstructedPod, metav1.UpdateOptions{})
				if err != nil {
					t.Fatalf("failed to update pod: %v", err)
				}
			}

			for _, pod := range tt.podsDeletedInSrcCluster {
				err := src.Resource(v1.Resource("pods").WithVersion("v1")).Namespace(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
				if err != nil {
					t.Fatalf("failed to delete pod: %v", err)
				}
			}

			errMessage := ""
			err := wait.PollUntilContextTimeout(ctx, 100*time.Millisecond, 5*time.Second, false, func(context.Context) (done bool, err error) {
				checkedPods := sets.New[podKey]()
				for _, pod := range tt.afterPodsInDestCluster {
					// get Pod from dest cluster
					p, err := dest.Resource(v1.Resource("pods").WithVersion("v1")).Namespace(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
					if err != nil {
						errMessage = fmt.Sprintf("failed to get pod: %v", err)
						return false, nil
					}

					// convert Pod to v1.Pod
					var got v1.Pod
					err = runtime.DefaultUnstructuredConverter.FromUnstructured(p.Object, &got)
					if err != nil {
						errMessage = fmt.Sprintf("failed to convert pod to v1.Pod: %v", err)
						return false, nil
					}

					if diff := cmp.Diff(pod, &got, cmpopts.IgnoreTypes(metav1.Time{})); diff != "" {
						errMessage = fmt.Sprintf("diff: %s", diff)
						return false, nil
					}
					checkedPods.Insert(podKey{pod.Name, pod.Namespace})
				}

				for _, pod := range createdPods.Difference(checkedPods).UnsortedList() {
					// get Pod from dest cluster
					_, err := dest.Resource(v1.Resource("pods").WithVersion("v1")).Namespace(pod.namespace).Get(ctx, pod.name, metav1.GetOptions{})
					if err != nil && !apierrors.IsNotFound(err) {
						errMessage = fmt.Sprintf("failed to get pod: %v", err)
						return false, nil
					}
					if err == nil {
						errMessage = fmt.Sprintf("pod %s/%s should be deleted", pod.namespace, pod.name)
						return false, nil
					}
				}

				return true, nil
			})
			if err != nil {
				t.Fatal(errMessage)
			}
		})
	}
}

type podKey struct{ name, namespace string }
