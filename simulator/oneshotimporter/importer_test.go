package oneshotimporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/restmapper"
	scheduling "k8s.io/kubernetes/pkg/apis/scheduling/v1"
	storage "k8s.io/kubernetes/pkg/apis/storage/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourceapplier"
)

func TestService_ImportClusterResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		labelSelector      metav1.LabelSelector
		srcObjects         []*unstructured.Unstructured
		importedObjects    []*unstructured.Unstructured
		notImportedObjects []*unstructured.Unstructured
		wantErr            bool
	}{
		{
			name:          "successfully import resources without label selector",
			labelSelector: metav1.LabelSelector{},
			srcObjects: []*unstructured.Unstructured{
				podWithNameAndLabel("pod", nil),
				podWithNameAndLabel("pod2", nil),
			},
			importedObjects: []*unstructured.Unstructured{
				podWithNameAndLabel("pod", nil),
				podWithNameAndLabel("pod2", nil),
			},
			wantErr: false,
		},
		{
			name: "successfully import resources filtered with label selector",
			labelSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			srcObjects: []*unstructured.Unstructured{
				podWithNameAndLabel("test-pod-1", map[string]string{"app": "test"}),
				podWithNameAndLabel("test-pod-2", map[string]string{"app": "test2"}),
				podWithNameAndLabel("test-pod-3", nil),
			},
			importedObjects: []*unstructured.Unstructured{
				podWithNameAndLabel("test-pod-1", map[string]string{"app": "test"}),
			},
			notImportedObjects: []*unstructured.Unstructured{
				podWithNameAndLabel("test-pod-2", map[string]string{"app": "test2"}),
				podWithNameAndLabel("test-pod-3", nil),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := runtime.NewScheme()
			v1.AddToScheme(s)
			storage.AddToScheme(s)
			scheduling.AddToScheme(s)
			srcClient := fake.NewSimpleDynamicClient(s)
			destClient := fake.NewSimpleDynamicClient(s)
			applier := resourceapplier.New(destClient, mapper, resourceapplier.Options{})
			oneshotImporter := NewService(srcClient, applier)
			for _, obj := range tt.srcObjects {
				gvr, err := findGVR(obj)
				assert.NoError(t, err)
				_, err = srcClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(context.Background(), obj, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			err := oneshotImporter.ImportClusterResources(context.Background(), tt.labelSelector)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			for _, want := range tt.importedObjects {
				gvr, err := findGVR(want)
				assert.NoError(t, err)
				got, err := destClient.Resource(gvr).Namespace(want.GetNamespace()).Get(context.Background(), want.GetName(), metav1.GetOptions{})
				assert.NoError(t, err)
				assert.Equal(t, want.GetName(), got.GetName())
			}
			for _, notWant := range tt.notImportedObjects {
				gvr, err := findGVR(notWant)
				assert.NoError(t, err)
				got, err := destClient.Resource(gvr).Namespace(notWant.GetNamespace()).Get(context.Background(), notWant.GetName(), metav1.GetOptions{})
				assert.Error(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

var mapper = restmapper.NewDiscoveryRESTMapper([]*restmapper.APIGroupResources{
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
})

func findGVR(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	gvk := obj.GroupVersionKind()
	m, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return m.Resource, nil
}

func podWithNameAndLabel(name string, labels map[string]string) *unstructured.Unstructured {
	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "test-container",
						"image": "test-image",
					},
				},
			},
		},
	}

	if labels != nil {
		pod.SetLabels(labels)
	}

	return pod
}
