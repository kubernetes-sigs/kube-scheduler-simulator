package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/fake"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/pod"
)

const (
	namespace1      = "namespace1"
	namespace2      = "namespace2"
	node1           = "node1"
	pod1            = "pod1"
	pod2            = "pod2"
	bogusPod        = "boguspod"
	bogusNS         = "bogusns"
	bogusAnnotation = "bogusannotation"
)

type test struct {
	name                   string
	params                 map[string]string
	prepareFakeClientSetFn func() *fake.Clientset
	wantCode               int
	wantBody               string
	wantErr                bool
}

func TestPodHandler_GetPods(t *testing.T) {

	tests := []test{
		{
			name:   "no pods",
			params: map[string]string{},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			wantCode: http.StatusOK,
			wantErr:  false,
		},
		{
			name:   "all pods",
			params: map[string]string{},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return prepareFakeClientset()
			},
			wantCode: http.StatusOK,
			wantErr:  false,
			wantBody: `{ 
                "metadata": {},
                "items": [
                {
          		    "metadata": {
				 		"annotations": {
				            "not-our-usual-annotation":          "some annotation shows up here",
				            "scheduler-simulator/filter-result": "{\"node-45pvw\": {\"AzureDiskLimits\": \"passed\", \"EBSLimits\": \"passed\"}}",
				            "scheduler-simulator/future-thing": "{\"node-45pvw\": {\"SomethingElse\": \"passed\", \"AnotherThing\": \"passed\"}}",
				            "scheduler-simulator/score-result": "{}"
		                },
                        "creationTimestamp": null,
                        "name":              "pod1",
                        "namespace":         "namespace1"
                    },
                    "spec":   {"containers": null, "nodeName": "node1"},
                    "status": {}
         		},
          		{
          			"metadata": {
          				"creationTimestamp": null,
          				"name":              "pod2",
          				"namespace":         "namespace2"
          			},
                    "spec":   { "containers": null },
          			"status": {}
         		}]
            }`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := tt.prepareFakeClientSetFn()
			rec, recordingContext := prepareRecordingContext(tt)

			err := NewPodHandler(pod.NewPodService(client)).GetPods(recordingContext)
			checkResults(t, tt, err, rec)
		})
	}
}

func TestPodHandler_GetPod(t *testing.T) {

	tests := []test{
		{
			name:   "namespace not found",
			params: map[string]string{"namespace": bogusNS, "name": pod1},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			wantCode: http.StatusNotFound,
			wantErr:  true,
		},
		{
			name:   "pod not found",
			params: map[string]string{"namespace": namespace1, "name": bogusPod},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			wantCode: http.StatusNotFound,
			wantErr:  true,
		},
		{
			name:   "pod found",
			params: map[string]string{"namespace": namespace2, "name": pod2},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return prepareFakeClientset()
			},
			wantCode: http.StatusOK,
			wantErr:  false,
			wantBody: `{ 
			 	"metadata": {
				 		"creationTimestamp": null,
				 		"name":              "pod2",
				 		"namespace":         "namespace2"
		        },
			 	"spec": { "containers": null },
			 	"status": {}
            }`,
		},
		{
			name:   "pod found on node",
			params: map[string]string{"namespace": namespace1, "name": pod1},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return prepareFakeClientset()
			},
			wantCode: http.StatusOK,
			wantErr:  false,
			wantBody: `{ 
			 	"metadata": {
				 		"annotations": {
				            "not-our-usual-annotation":          "some annotation shows up here",
				            "scheduler-simulator/filter-result": "{\"node-45pvw\": {\"AzureDiskLimits\": \"passed\", \"EBSLimits\": \"passed\"}}",
				            "scheduler-simulator/future-thing": "{\"node-45pvw\": {\"SomethingElse\": \"passed\", \"AnotherThing\": \"passed\"}}",
				            "scheduler-simulator/score-result": "{}"
		                },
				 		"creationTimestamp": null,
				 		"name":              "pod1",
				 		"namespace":         "namespace1"
		        },
			 	"spec": { "containers": null, "nodeName": "node1" },
			 	"status": {}
            }`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := tt.prepareFakeClientSetFn()
			rec, recordingContext := prepareRecordingContext(tt)

			err := NewPodHandler(pod.NewPodService(client)).GetPod(recordingContext)
			checkResults(t, tt, err, rec)
		})
	}
}

func TestPodHandler_GetPodMetadataAnnotations(t *testing.T) {

	tests := []test{
		{
			name:   "no pods",
			params: map[string]string{"namespace": bogusNS, "name": bogusPod, "annotation": bogusAnnotation},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			wantCode: http.StatusNotFound,
			wantErr:  true,
		},
		{
			name:   "pod found but not annotation",
			params: map[string]string{"namespace": namespace1, "name": pod1, "annotation": bogusAnnotation},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return prepareFakeClientset()
			},
			wantCode: http.StatusNotFound,
			wantErr:  true,
		},
		{
			name:   "pod found on node with specific annotation",
			params: map[string]string{"namespace": namespace1, "name": pod1, "annotation": "scheduler-simulator/filter-result"},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return prepareFakeClientset()
			},
			wantCode: http.StatusOK,
			wantErr:  false,
			wantBody: `{ 
                "node-45pvw": {
                    "AzureDiskLimits": "passed",
                    "EBSLimits": "passed"
		        }
            }`,
		},
		{
			name:   "pod found on node with scheduler-simulator annotations",
			params: map[string]string{"namespace": namespace1, "name": pod1, "annotation": "scheduler-simulator"},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return prepareFakeClientset()
			},
			wantCode: http.StatusOK,
			wantErr:  false,
			wantBody: `{ 
        	    "scheduler-simulator/filter-result": {
        		    "node-45pvw": {
                        "AzureDiskLimits": "passed",
                        "EBSLimits": "passed"
                    }
        	    },
        	    "scheduler-simulator/future-thing": {
        		    "node-45pvw": {
                        "AnotherThing": "passed",
                        "SomethingElse": "passed"
                    }
        	    },
        	    "scheduler-simulator/score-result": {}
            }`,
		},
		{
			name:   "pod found on node with some random annotation",
			params: map[string]string{"namespace": namespace1, "name": pod1, "annotation": "not-our-usual-annotation"},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return prepareFakeClientset()
			},
			wantCode: http.StatusOK,
			wantErr:  false,
			wantBody: "some annotation shows up here",
		},
		{
			name:   "pod found on node, dumping all annotations",
			params: map[string]string{"namespace": namespace1, "name": pod1, "annotation": ""},
			prepareFakeClientSetFn: func() *fake.Clientset {
				return prepareFakeClientset()
			},
			wantCode: http.StatusOK,
			wantErr:  false,
			wantBody: `{
				"not-our-usual-annotation":          "some annotation shows up here",
				"scheduler-simulator/filter-result": "{\"node-45pvw\": {\"AzureDiskLimits\": \"passed\", \"EBSLimits\": \"passed\"}}",
				"scheduler-simulator/future-thing": "{\"node-45pvw\": {\"SomethingElse\": \"passed\", \"AnotherThing\": \"passed\"}}",
				"scheduler-simulator/score-result": "{}"
			}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := tt.prepareFakeClientSetFn()
			rec, recordingContext := prepareRecordingContext(tt)

			err := NewPodHandler(pod.NewPodService(client)).GetPodMetaDataAnnotations(recordingContext)
			checkResults(t, tt, err, rec)
		})
	}
}

// checkResults logs and fails if the expected err, code, and body are not returned/recorded
func checkResults(t *testing.T, tt test, err error, rec *httptest.ResponseRecorder) {
	t.Helper()
	if (err != nil) != tt.wantErr {
		t.Fatalf("%v test: error=%v, wantErr=%v", tt.name, err, tt.wantErr)
	}

	// HTTP status code is either recorded or it is in the error
	code := rec.Code
	if err != nil {
		httpError := err.(*echo.HTTPError)
		code = httpError.Code
	}
	if code != tt.wantCode {
		t.Fatalf("%v test: mismatch code=%v, wantCode=%v", tt.name, code, tt.wantCode)
	}

	// Compare recorded body to wantBody (not testing error message)
	if err == nil && tt.wantBody != "" {
		var want, got interface{}
		if err := json.Unmarshal([]byte(tt.wantBody), &want); err != nil {
			// can test for things we cannot unmarshal
			want = tt.wantBody
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			// can test for things we cannot unmarshal
			got = rec.Body
		}
		diffResponse := cmp.Diff(want, got)
		if diffResponse != "" {
			t.Fatalf("%v test: body mismatch (-want +got):\n%v", tt.name, diffResponse)
		}
	}
}

func prepareRecordingContext(tt struct {
	name                   string
	params                 map[string]string
	prepareFakeClientSetFn func() *fake.Clientset
	wantCode               int
	wantBody               string
	wantErr                bool
}) (*httptest.ResponseRecorder, echo.Context) {
	req := httptest.NewRequest(http.MethodGet, "/dummy", nil) // Just enough for context and handler
	rec := httptest.NewRecorder()
	recordingContext := echo.New().NewContext(req, rec)
	n := len(tt.params)
	if n > 0 {
		names := make([]string, 0, n)
		values := make([]string, 0, n)
		for k, v := range tt.params {
			names = append(names, k)
			values = append(values, v)
		}
		recordingContext.SetParamNames(names...)
		recordingContext.SetParamValues(values...)
	}
	return rec, recordingContext
}

func prepareFakeClientset() *fake.Clientset {
	c := fake.NewSimpleClientset()
	c.CoreV1().Pods(namespace1).Create(context.Background(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: pod1,
			Annotations: map[string]string{
				"scheduler-simulator/filter-result": `{"node-45pvw": {"AzureDiskLimits": "passed", "EBSLimits": "passed"}}`,
				"scheduler-simulator/score-result":  `{}`,
				"scheduler-simulator/future-thing":  `{"node-45pvw": {"SomethingElse": "passed", "AnotherThing": "passed"}}`,
				"not-our-usual-annotation":          "some annotation shows up here",
			},
		},
		Spec: corev1.PodSpec{
			NodeName: node1,
		},
	}, metav1.CreateOptions{})
	c.CoreV1().Pods(namespace2).Create(context.Background(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: pod2,
		},
	}, metav1.CreateOptions{})
	return c
}
