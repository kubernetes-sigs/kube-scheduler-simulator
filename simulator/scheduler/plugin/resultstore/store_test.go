package resultstore

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin/annotation"
)

func TestStore_AddFilterResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace  string
		podName    string
		nodeName   string
		pluginName string
		reason     string
	}
	tests := []struct {
		name          string
		resultbefore  map[key]*result
		args          args
		wantResultMap map[key]*result
	}{
		{
			name:         "success with empty result",
			resultbefore: map[key]*result{},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin1",
				reason:     PassedFilterMessage,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalscore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
				},
			},
		},
		{
			name: "success with non-empty filter map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalscore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
				},
			},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin2",
				reason:     PassedFilterMessage,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalscore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node1": {
							"plugin1": PassedFilterMessage,
							"plugin2": PassedFilterMessage,
						},
					},
				},
			},
		},
		{
			name: "success when no map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalscore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
					},
				},
			},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin1",
				reason:     PassedFilterMessage,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalscore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
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
			s := &Store{
				mu:      new(sync.Mutex),
				results: tt.resultbefore,
			}
			s.AddFilterResult(tt.args.namespace, tt.args.podName, tt.args.nodeName, tt.args.pluginName, tt.args.reason)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_AddScoreResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace  string
		podName    string
		nodeName   string
		pluginName string
		score      int64
	}
	tests := []struct {
		name              string
		resultbefore      map[key]*result
		scorePluginWeight map[string]int32
		args              args
		wantResultMap     map[key]*result
	}{
		{
			name:              "success with empty result",
			resultbefore:      map[key]*result{},
			scorePluginWeight: map[string]int32{"plugin1": 2},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin1",
				score:      10,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node1": {
							"plugin1": "20",
						},
					},
					score: map[string]map[string]string{
						"node1": {
							"plugin1": "10",
						},
					},
				},
			},
		},
		{
			name: "success with non-empty filter map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node1": {
							"plugin1": "30",
						},
					},
					score: map[string]map[string]string{
						"node1": {
							"plugin1": "10",
						},
					},
				},
			},
			scorePluginWeight: map[string]int32{"plugin2": 2},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin2",
				score:      10,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node1": {
							"plugin1": "30",
							"plugin2": "20",
						},
					},
					score: map[string]map[string]string{
						"node1": {
							"plugin1": "10",
							"plugin2": "10",
						},
					},
				},
			},
		},
		{
			name: "success when no map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node0": {
							"plugin1": "20",
						},
					},
					score: map[string]map[string]string{
						"node0": {
							"plugin1": "10",
						},
					},
				},
			},
			scorePluginWeight: map[string]int32{"plugin1": 2},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin1",
				score:      10,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node0": {
							"plugin1": "20",
						},
						"node1": {
							"plugin1": "20",
						},
					},
					score: map[string]map[string]string{
						"node0": {
							"plugin1": "10",
						},
						"node1": {
							"plugin1": "10",
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
			s := &Store{
				mu:                new(sync.Mutex),
				results:           tt.resultbefore,
				scorePluginWeight: tt.scorePluginWeight,
			}
			s.AddScoreResult(tt.args.namespace, tt.args.podName, tt.args.nodeName, tt.args.pluginName, tt.args.score)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_AddNormalizedScoreResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace  string
		podName    string
		nodeName   string
		pluginName string
		score      int64
	}
	tests := []struct {
		name              string
		resultbefore      map[key]*result
		scorePluginWeight map[string]int32
		args              args
		wantResultMap     map[key]*result
	}{
		{
			name:              "success with empty result",
			resultbefore:      map[key]*result{},
			scorePluginWeight: map[string]int32{"plugin1": 2},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin1",
				score:      10,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					score:  map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node1": {
							"plugin1": "20",
						},
					},
				},
			},
		},
		{
			name: "success with non-empty filter map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node1": {
							"plugin1": "30",
						},
					},
				},
			},
			scorePluginWeight: map[string]int32{"plugin2": 2},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin2",
				score:      10,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node1": {
							"plugin1": "30",
							"plugin2": "20",
						},
					},
				},
			},
		},
		{
			name: "success when no map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node0": {
							"plugin1": "20",
						},
					},
				},
			},
			scorePluginWeight: map[string]int32{"plugin1": 2},
			args: args{
				namespace:  "default",
				podName:    "pod1",
				nodeName:   "node1",
				pluginName: "plugin1",
				score:      10,
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{},
					finalscore: map[string]map[string]string{
						"node0": {
							"plugin1": "20",
						},
						"node1": {
							"plugin1": "20",
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
			s := &Store{
				mu:                new(sync.Mutex),
				results:           tt.resultbefore,
				scorePluginWeight: tt.scorePluginWeight,
			}
			s.AddNormalizedScoreResult(tt.args.namespace, tt.args.podName, tt.args.nodeName, tt.args.pluginName, tt.args.score)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_addSchedulingResultToPod(t *testing.T) {
	t.Parallel()
	podName := "pod1"
	namespace := "default"
	tests := []struct {
		name                       string
		result                     map[key]*result
		prepareFakeClientSetFn     func() *fake.Clientset
		newObj                     interface{}
		wantpod                    *corev1.Pod
		resultRemainsAfterExecFunc bool
		wanterr                    bool
	}{
		{
			name: "success",
			result: map[key]*result{
				"default/pod1": {
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					finalscore: map[string]map[string]string{
						"node0": {
							"plugin1": "20",
						},
						"node1": {
							"plugin1": "20",
						},
					},
					score: map[string]map[string]string{
						"node0": {
							"plugin1": "10",
						},
						"node1": {
							"plugin1": "10",
						},
					},
				},
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(namespace).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				}, metav1.CreateOptions{})

				return c
			},
			newObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			},
			wantpod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					Annotations: map[string]string{
						annotation.FilterResultAnnotationKey: func() string {
							r := map[string]map[string]string{
								"node0": {
									"plugin1": PassedFilterMessage,
								},
								"node1": {
									"plugin1": PassedFilterMessage,
								},
							}
							d, _ := json.Marshal(r)
							return string(d)
						}(),
						annotation.ScoreResultAnnotationKey: func() string {
							r := map[string]map[string]string{
								"node0": {
									"plugin1": "10",
								},
								"node1": {
									"plugin1": "10",
								},
							}
							d, _ := json.Marshal(r)
							return string(d)
						}(),
						annotation.FinalScoreResultAnnotationKey: func() string {
							r := map[string]map[string]string{
								"node0": {
									"plugin1": "20",
								},
								"node1": {
									"plugin1": "20",
								},
							}
							d, _ := json.Marshal(r)
							return string(d)
						}(),
					},
				},
			},
			resultRemainsAfterExecFunc: false,
		},
		{
			name:   "do nothing if store doesn't have data",
			result: map[key]*result{},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(namespace).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				}, metav1.CreateOptions{})

				return c
			},
			newObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			},
			wantpod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			},
			resultRemainsAfterExecFunc: false,
		},
		{
			name: "success without some data on store",
			result: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalscore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
				},
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(namespace).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				}, metav1.CreateOptions{})

				return c
			},
			newObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			},
			wantpod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					Annotations: map[string]string{
						annotation.FilterResultAnnotationKey: func() string {
							r := map[string]map[string]string{
								"node0": {
									"plugin1": PassedFilterMessage,
								},
								"node1": {
									"plugin1": PassedFilterMessage,
								},
							}
							d, _ := json.Marshal(r)
							return string(d)
						}(),
						annotation.ScoreResultAnnotationKey:      "{}",
						annotation.FinalScoreResultAnnotationKey: "{}",
					},
				},
			},
			resultRemainsAfterExecFunc: false,
		},
		{
			name: "fail if client failed to update the pod",
			result: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalscore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
				},
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.CoreV1().Pods(namespace).Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName + "1", // To cause the update to fail.
						Namespace: namespace,
					},
				}, metav1.CreateOptions{})

				return c
			},
			newObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			},
			wantpod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName + "1",
					Namespace: namespace,
				},
			},
			resultRemainsAfterExecFunc: true,
			wanterr:                    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.prepareFakeClientSetFn()
			s := &Store{
				mu:      new(sync.Mutex),
				results: tt.result,
				client:  c,
			}
			s.addSchedulingResultToPod(nil, tt.newObj)

			if !tt.wanterr {
				p, _ := c.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
				assert.Equal(t, tt.wantpod, p)
			}

			if _, ok := s.results["default/pod1"]; ok != tt.resultRemainsAfterExecFunc {
				if ok {
					t.Fatal("result should be deleted")
				}
				t.Fatalf("result should be left")
			}
		})
	}
}
