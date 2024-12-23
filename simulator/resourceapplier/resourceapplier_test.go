package resourceapplier

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/restmapper"
	scheduling "k8s.io/kubernetes/pkg/apis/scheduling/v1"
	storage "k8s.io/kubernetes/pkg/apis/storage/v1"
)

func TestResourceApplier_createPods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		podToApply    *corev1.Pod
		podAfterApply *corev1.Pod
		wantErr       bool
	}{
		{
			name: "create a Pod",
			podToApply: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
					},
				},
			},
			podAfterApply: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
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

			client, mapper := prepare()
			service := New(client, mapper, Options{})

			p, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tt.podToApply)
			if err != nil {
				t.Fatalf("failed to convert pod to unstructured: %v", err)
			}
			unstructedPod := &unstructured.Unstructured{Object: p}
			err = service.Create(context.Background(), unstructedPod)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := getResource(tt.podToApply.GroupVersionKind(), tt.podToApply.Name, tt.podToApply.Namespace, mapper, client)
			if err != nil {
				t.Fatalf("failed to get pod when comparing: %v", err)
			}
			var gotPod corev1.Pod
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(got.UnstructuredContent(), &gotPod)
			if err != nil {
				t.Fatalf("failed to convert got unstructured to pod: %v", err)
			}

			if diff := cmp.Diff(*tt.podAfterApply, gotPod); diff != "" {
				t.Errorf("createPods() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResourceApplier_createPodsWithFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		podToCreate *corev1.Pod
		filter      FilteringFunction
		filtered    bool
		wantErr     bool
	}{
		{
			name: "create a Pod but it should not be created because of the filter",
			podToCreate: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
					Labels: map[string]string{
						"ignore": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
					},
				},
			},
			filter: func(_ context.Context, resource *unstructured.Unstructured, _ *Clients) (bool, error) {
				if resource.GetLabels()["ignore"] == "true" {
					return false, nil
				}
				return true, nil
			},
			filtered: true,
			wantErr:  false,
		},
		{
			name: "create a Pod and it should be pass the filter",
			podToCreate: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
					},
				},
			},
			filter: func(_ context.Context, resource *unstructured.Unstructured, _ *Clients) (bool, error) {
				if resource.GetLabels()["ignore"] == "true" {
					return false, nil
				}
				return true, nil
			},
			filtered: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, mapper := prepare()

			options := setFilter(tt.podToCreate.GroupVersionKind(), tt.filter, mapper)
			service := New(client, mapper, options)

			p, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tt.podToCreate)
			if err != nil {
				t.Fatalf("failed to convert pod to unstructured: %v", err)
			}
			unstructedPod := &unstructured.Unstructured{Object: p}
			err = service.Create(context.Background(), unstructedPod)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := getResource(tt.podToCreate.GroupVersionKind(), tt.podToCreate.Name, tt.podToCreate.Namespace, mapper, client)
			if err != nil {
				if tt.filtered && errors.IsNotFound(err) || tt.wantErr {
					return
				}
				t.Fatalf("failed to get pod when comparing: %v", err)
			} else if tt.filtered || tt.wantErr {
				t.Fatalf("pod should not be created but it exists")
			}

			var gotPod corev1.Pod
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(got.UnstructuredContent(), &gotPod)
			if err != nil {
				t.Fatalf("failed to convert got unstructured to pod: %v", err)
			}

			if diff := cmp.Diff(*tt.podToCreate, gotPod); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResourceApplier_updatePods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		originalPod    *corev1.Pod
		updatePod      func(pod *corev1.Pod)
		podAfterUpdate *corev1.Pod
		wantErr        bool
	}{
		{
			name: "update an unscheduled Pod",
			originalPod: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
					},
				},
			},
			updatePod: func(pod *corev1.Pod) {
				pod.Spec.Containers[0].Image = "image-2"
			},
			podAfterUpdate: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-2",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "update an unscheduled Pod to be scheduled but it should not be updated",
			originalPod: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
					},
				},
			},
			updatePod: func(pod *corev1.Pod) {
				pod.Spec.NodeName = "node-1"
			},
			podAfterUpdate: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
					},
				},
			},
		},
		{
			name: "update a scheduled Pod but it should not be updated",
			originalPod: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					NodeName: "node-1",
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
					},
				},
			},
			updatePod: func(pod *corev1.Pod) {
				pod.Spec.Containers[0].Image = "image-2"
			},
			podAfterUpdate: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					NodeName: "node-1",
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
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

			client, mapper := prepare()
			service := New(client, mapper, Options{})

			p, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tt.originalPod)
			if err != nil {
				t.Fatalf("failed to convert pod to unstructured: %v", err)
			}
			unstructedPod := &unstructured.Unstructured{Object: p}
			err = service.Create(context.Background(), unstructedPod)
			if (err != nil) != tt.wantErr {
				t.Errorf("updatePods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.updatePod(tt.originalPod)
			p, err = runtime.DefaultUnstructuredConverter.ToUnstructured(tt.originalPod)
			if err != nil {
				t.Fatalf("failed to convert pod to unstructured: %v", err)
			}
			unstructedPod = &unstructured.Unstructured{Object: p}

			err = service.Update(context.Background(), unstructedPod)
			if (err != nil) != tt.wantErr {
				t.Errorf("updatePods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := getResource(tt.originalPod.GroupVersionKind(), tt.originalPod.Name, tt.originalPod.Namespace, mapper, client)
			if err != nil {
				t.Fatalf("failed to get pod when comparing: %v", err)
			}
			var gotPod corev1.Pod
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(got.UnstructuredContent(), &gotPod)
			if err != nil {
				t.Fatalf("failed to convert got unstructured to pod: %v", err)
			}

			if diff := cmp.Diff(*tt.podAfterUpdate, gotPod); diff != "" {
				t.Errorf("updatePods() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResourceApplier_deletePods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pod     *corev1.Pod
		wantErr bool
	}{
		{
			name: "delete a Pod",
			pod: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "image-1",
						},
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

			client, mapper := prepare()
			service := New(client, mapper, Options{})

			p, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tt.pod)
			if err != nil {
				t.Fatalf("failed to convert pod to unstructured: %v", err)
			}
			unstructedPod := &unstructured.Unstructured{Object: p}
			err = service.Create(context.Background(), unstructedPod)
			if (err != nil) != tt.wantErr {
				t.Errorf("deletePods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = service.Delete(context.Background(), unstructedPod)
			if (err != nil) != tt.wantErr {
				t.Errorf("deletePods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, err = getResource(tt.pod.GroupVersionKind(), tt.pod.Name, tt.pod.Namespace, mapper, client)
			if err == nil {
				t.Fatalf("pod should be deleted but it still exists")
			}

			if !errors.IsNotFound(err) && !tt.wantErr {
				t.Fatalf("failed to check if pod is deleted: %v", err)
			}
		})
	}
}

func TestResourceApplier_createNodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		nodeToApply    *corev1.Node
		nodeAfterApply *corev1.Node
		wantErr        bool
	}{
		{
			name: "create a Node",
			nodeToApply: &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
			},
			nodeAfterApply: &corev1.Node{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, mapper := prepare()
			service := New(client, mapper, Options{})

			n, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tt.nodeToApply)
			if err != nil {
				t.Fatalf("failed to convert node to unstructured: %v", err)
			}
			unstructedNode := &unstructured.Unstructured{Object: n}
			err = service.Create(context.Background(), unstructedNode)
			if (err != nil) != tt.wantErr {
				t.Errorf("createNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := getResource(tt.nodeToApply.GroupVersionKind(), tt.nodeToApply.Name, tt.nodeToApply.Namespace, mapper, client)
			if err != nil {
				t.Fatalf("failed to get node when comparing: %v", err)
			}
			var gotNode corev1.Node
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(got.UnstructuredContent(), &gotNode)
			if err != nil {
				t.Fatalf("failed to convert got unstructured to node: %v", err)
			}

			if diff := cmp.Diff(*tt.nodeAfterApply, gotNode); diff != "" {
				t.Errorf("createNode() mismatch (-want +got):\n %s", diff)
				return
			}
		})
	}
}

func prepare() (*dynamicFake.FakeDynamicClient, meta.RESTMapper) {
	s := runtime.NewScheme()
	v1.AddToScheme(s)
	scheduling.AddToScheme(s)
	storage.AddToScheme(s)
	client := dynamicFake.NewSimpleDynamicClient(s)
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
					{Name: "nodes", Namespaced: false, Kind: "Node"},
				},
			},
		},
	}

	mapper := restmapper.NewDiscoveryRESTMapper(resources)
	return client, mapper
}

func getResource(gvk schema.GroupVersionKind, name, namespace string, mapper meta.RESTMapper, client *dynamicFake.FakeDynamicClient) (*unstructured.Unstructured, error) {
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	resource := client.Resource(mapping.Resource).Namespace(namespace)
	return resource.Get(context.Background(), name, metav1.GetOptions{})
}

func setFilter(gvk schema.GroupVersionKind, filter FilteringFunction, mapper meta.RESTMapper) Options {
	m, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		panic(err)
	}

	return Options{
		FilterBeforeCreating: map[schema.GroupVersionResource][]FilteringFunction{
			m.Resource: {filter},
		},
	}
}
