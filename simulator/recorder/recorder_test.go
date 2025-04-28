package recorder

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/restmapper"
	appsv1 "k8s.io/kubernetes/pkg/apis/apps/v1"
	schedulingv1 "k8s.io/kubernetes/pkg/apis/scheduling/v1"
	storagev1 "k8s.io/kubernetes/pkg/apis/storage/v1"
	"k8s.io/utils/ptr"
)

func TestRecorder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		resourceToCreate []unstructured.Unstructured
		resourceToUpdate []unstructured.Unstructured
		resourceToDelete []unstructured.Unstructured
		want             []Record
		wantErr          bool
	}{
		{
			name: "should record creating pods",
			resourceToCreate: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "pod-1",
							"namespace": "default",
						},
					},
				},
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "pod-2",
							"namespace": "default",
						},
					},
				},
			},
			want: []Record{
				{
					Event: Add,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-1",
								"namespace": "default",
							},
						},
					},
				},
				{
					Event: Add,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-2",
								"namespace": "default",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should record updating a pod",
			resourceToCreate: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "pod-1",
							"namespace": "default",
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:latest",
								},
							},
						},
					},
				},
			},
			resourceToUpdate: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "pod-1",
							"namespace": "default",
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:latest",
								},
							},
							"nodeName": "node-1",
						},
					},
				},
			},
			want: []Record{
				{
					Event: Add,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-1",
								"namespace": "default",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:latest",
									},
								},
							},
						},
					},
				},
				{
					Event: Update,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-1",
								"namespace": "default",
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:latest",
									},
								},
								"nodeName": "node-1",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should record deleting a pod",
			resourceToCreate: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "pod-1",
							"namespace": "default",
						},
					},
				},
			},
			resourceToDelete: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "pod-1",
							"namespace": "default",
						},
					},
				},
			},
			want: []Record{
				{
					Event: Add,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-1",
								"namespace": "default",
							},
						},
					},
				},
				{
					Event: Delete,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-1",
								"namespace": "default",
							},
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

			filePath := path.Join(t.TempDir(), strings.ReplaceAll(tt.name, " ", "_"))
			defer os.Remove(filePath)

			s := runtime.NewScheme()
			corev1.AddToScheme(s)
			appsv1.AddToScheme(s)
			schedulingv1.AddToScheme(s)
			storagev1.AddToScheme(s)
			client := dynamicFake.NewSimpleDynamicClient(s)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			service := New(client, Options{RecordFile: filePath, FlushInterval: ptr.To(100 * time.Millisecond)})
			err := service.Run(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Record() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = apply(ctx, client, tt.resourceToCreate, tt.resourceToUpdate, tt.resourceToDelete)
			if err != nil {
				t.Fatal(err)
			}

			err = assert(ctx, filePath, tt.want)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func apply(ctx context.Context, client *dynamicFake.FakeDynamicClient, resourceToCreate []unstructured.Unstructured, resourceToUpdate []unstructured.Unstructured, resourceToDelete []unstructured.Unstructured) error {
	for i := range resourceToCreate {
		resource := &resourceToCreate[i]
		gvr, err := findGVR(resource)
		if err != nil {
			return xerrors.Errorf("failed to find GVR: %w", err)
		}
		ns := resource.GetNamespace()

		_, err = client.Resource(gvr).Namespace(ns).Create(ctx, resource, metav1.CreateOptions{})
		if err != nil {
			return xerrors.Errorf("failed to create a pod: %w", err)
		}
	}

	for i := range resourceToUpdate {
		resource := &resourceToUpdate[i]
		gvr, err := findGVR(resource)
		if err != nil {
			return xerrors.Errorf("failed to find GVR: %w", err)
		}
		ns := resource.GetNamespace()

		_, err = client.Resource(gvr).Namespace(ns).Update(ctx, resource, metav1.UpdateOptions{})
		if err != nil {
			return xerrors.Errorf("failed to update a pod: %w", err)
		}
	}

	for i := range resourceToDelete {
		resource := &resourceToDelete[i]
		gvr, err := findGVR(resource)
		if err != nil {
			return xerrors.Errorf("failed to find GVR: %w", err)
		}
		ns := resource.GetNamespace()

		err = client.Resource(gvr).Namespace(ns).Delete(ctx, resource.GetName(), metav1.DeleteOptions{})
		if err != nil {
			return xerrors.Errorf("failed to delete a pod: %w", err)
		}
	}

	return nil
}

func assert(ctx context.Context, filePath string, want []Record) error {
	var finalErr error
	wait.PollUntilContextTimeout(ctx, 100*time.Millisecond, 5*time.Second, false, func(context.Context) (bool, error) {
		file, err := os.Open(filePath)
		if err != nil {
			finalErr = xerrors.Errorf("failed to open the record file: %w", err)
			return false, nil
		}
		defer file.Close()

		got := []Record{}
		reader := bufio.NewReader(file)
		for {
			record, err := loadRecordFromLine(reader)
			if err != nil {
				finalErr = xerrors.Errorf("failed to load record from line: %w", err)
				return false, nil
			}
			if record == nil {
				break
			}

			got = append(got, *record)
		}

		if len(got) != len(want) {
			finalErr = xerrors.Errorf("Service.Record() got = %v, want %v", got, want)
			return false, nil
		}

		for i := range got {
			if got[i].Event != want[i].Event {
				finalErr = xerrors.Errorf("Service.Record() got = %v, want %v", got[i].Event, want[i].Event)
				return true, finalErr
			}

			if diff := cmp.Diff(want[i].Resource, got[i].Resource); diff != "" {
				finalErr = xerrors.Errorf("Service.Record() got = %v, want %v", got[i].Resource, want[i].Resource)
				return true, finalErr
			}
		}

		return true, nil
	})

	return finalErr
}

func loadRecordFromLine(reader *bufio.Reader) (*Record, error) {
	line, err := reader.ReadBytes('\n')
	if len(line) == 0 || err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to read line: %w", err)
	}

	record := &Record{}
	if err := json.Unmarshal(line, record); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal record: %w", err)
	}

	return record, nil
}

var (
	resources = []*restmapper.APIGroupResources{
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
	}
	mapper = restmapper.NewDiscoveryRESTMapper(resources)
)

func findGVR(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	gvk := obj.GroupVersionKind()
	m, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return m.Resource, nil
}
