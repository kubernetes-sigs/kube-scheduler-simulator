package extender

import (
	"io"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	// just in case you want default correct return value
	return &http.Response{}, nil
}

func TestHttpExtender_send(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		verb                    string
		args                    interface{}
		prepareMockHTTPClientFn func() HTTPClient
		extenderFn              func(m HTTPClient) *extender
		wantErr                 bool
		wantResult              interface{}
	}{
		{
			name: "reflect response in the result",
			verb: "/PreemptVerb",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
					},
				},
			},
			prepareMockHTTPClientFn: func() HTTPClient {
				return &MockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(`{"Error":"myerror"}`)),
						}, nil
					},
				}
			},
			extenderFn: func(m HTTPClient) *extender {
				return &extender{
					extenderURL:      "http://example.com/extender",
					preemptVerb:      "/PreemptVerb",
					filterVerb:       "/FilterVerb",
					prioritizeVerb:   "/PrioritizeVerb",
					bindVerb:         "/BindVerb",
					weight:           1,
					client:           m,
					nodeCacheCapable: false,
					managedResources: sets.NewString(),
				}
			},
			wantErr: false,
			wantResult: extenderv1.ExtenderFilterResult{
				Nodes:                      nil,
				NodeNames:                  nil,
				FailedNodes:                nil,
				FailedAndUnresolvableNodes: nil,
				Error:                      "myerror",
			},
		},
		{
			name: "return an error if the args is nil",
			verb: "/PreemptVerb",
			args: math.Inf(1),
			prepareMockHTTPClientFn: func() HTTPClient {
				return &MockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return nil, nil
					},
				}
			},
			extenderFn: func(m HTTPClient) *extender {
				return &extender{
					extenderURL:      "http://example.com/extender",
					preemptVerb:      "/PreemptVerb",
					filterVerb:       "/FilterVerb",
					prioritizeVerb:   "/PrioritizeVerb",
					bindVerb:         "/BindVerb",
					weight:           1,
					client:           m,
					nodeCacheCapable: false,
					managedResources: sets.NewString(),
				}
			},
			wantErr: true,
		},
		{
			name: "return an error if NewRequest method returns an error",
			verb: "/PreemptVerb",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
					},
				},
			},
			prepareMockHTTPClientFn: func() HTTPClient {
				return &MockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							Body: io.NopCloser(strings.NewReader(`{"Error":"myerror"}`)),
						}, nil
					},
				}
			},
			extenderFn: func(m HTTPClient) *extender {
				return &extender{
					extenderURL:      "Unexpected URLs that cause errors",
					preemptVerb:      "/PreemptVerb",
					filterVerb:       "/FilterVerb",
					prioritizeVerb:   "/PrioritizeVerb",
					bindVerb:         "/BindVerb",
					weight:           1,
					client:           m,
					nodeCacheCapable: false,
					managedResources: sets.NewString(),
				}
			},
			wantErr: true,
		},
		{
			name: "return an error if Do method returns an error",
			verb: "/PreemptVerb",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
					},
				},
			},
			prepareMockHTTPClientFn: func() HTTPClient {
				return &MockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return nil, xerrors.New("Do returns an error")
					},
				}
			},
			extenderFn: func(m HTTPClient) *extender {
				return &extender{
					extenderURL:      "http://example.com/extender",
					preemptVerb:      "/PreemptVerb",
					filterVerb:       "/FilterVerb",
					prioritizeVerb:   "/PrioritizeVerb",
					bindVerb:         "/BindVerb",
					weight:           1,
					client:           m,
					nodeCacheCapable: false,
					managedResources: sets.NewString(),
				}
			},
			wantErr: true,
		},
		{
			name: "return an error if the status code of response is StatusInternalServerError",
			verb: "/PreemptVerb",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
					},
				},
			},
			prepareMockHTTPClientFn: func() HTTPClient {
				return &MockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(strings.NewReader(`{"Error":"myerror"}`)),
						}, nil
					},
				}
			},
			extenderFn: func(m HTTPClient) *extender {
				return &extender{
					extenderURL:      "http://example.com/extender",
					preemptVerb:      "/PreemptVerb",
					filterVerb:       "/FilterVerb",
					prioritizeVerb:   "/PrioritizeVerb",
					bindVerb:         "/BindVerb",
					weight:           1,
					client:           m,
					nodeCacheCapable: false,
					managedResources: sets.NewString(),
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockClient := tt.prepareMockHTTPClientFn()
			e := tt.extenderFn(mockClient)
			var result extenderv1.ExtenderFilterResult
			err := e.send(tt.verb, tt.args, &result)
			if (err != nil) != tt.wantErr {
				t.Fatalf("send() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
