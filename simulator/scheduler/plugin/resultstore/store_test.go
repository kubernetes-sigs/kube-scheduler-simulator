package resultstore

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/scheduler/framework"

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
					selectedNode:    "",
					preScore:        map[string]string{},
					preFilterStatus: map[string]string{},
					preFilterResult: map[string][]string{},
					permit:          map[string]string{},
					permitTimeout:   map[string]string{},
					reserve:         map[string]string{},
					prebind:         map[string]string{},
					bind:            map[string]string{},
					score:           map[string]map[string]string{},
					finalScore:      map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					postFilter: map[string]map[string]string{},
				},
			},
		},
		{
			name: "success with non-empty filter map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalScore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					postFilter: map[string]map[string]string{},
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
					finalScore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node1": {
							"plugin1": PassedFilterMessage,
							"plugin2": PassedFilterMessage,
						},
					},
					postFilter: map[string]map[string]string{},
				},
			},
		},
		{
			name: "success when no map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalScore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
					},
					postFilter: map[string]map[string]string{},
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
					finalScore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					postFilter: map[string]map[string]string{},
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

func TestStore_AddPostFilterResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace         string
		podName           string
		nominatedNodeName string
		pluginName        string
		nodeNames         []string
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
				namespace:         "default",
				podName:           "pod1",
				nominatedNodeName: "node1",
				pluginName:        "plugin1",
				nodeNames:         []string{"node1", "node2"},
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					selectedNode:    "",
					preScore:        map[string]string{},
					preFilterStatus: map[string]string{},
					preFilterResult: map[string][]string{},
					permit:          map[string]string{},
					permitTimeout:   map[string]string{},
					reserve:         map[string]string{},
					prebind:         map[string]string{},
					bind:            map[string]string{},
					score:           map[string]map[string]string{},
					finalScore:      map[string]map[string]string{},
					filter:          map[string]map[string]string{},
					postFilter: map[string]map[string]string{
						"node1": {
							"plugin1": PostFilterNominatedMessage,
						},
						"node2": {},
					},
				},
			},
		},
		{
			name: "success with non-empty postFilter map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalScore: map[string]map[string]string{},
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{
						"node1": {},
					},
				},
			},
			args: args{
				namespace:         "default",
				podName:           "pod1",
				nominatedNodeName: "node1",
				pluginName:        "plugin2",
				nodeNames:         []string{"node1", "node2"},
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalScore: map[string]map[string]string{},
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{
						"node1": {
							"plugin2": PostFilterNominatedMessage,
						},
						"node2": {},
					},
				},
			},
		},
		{
			name: "success when no map for the node",
			resultbefore: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalScore: map[string]map[string]string{},
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{
						"node0": {},
					},
				},
			},
			args: args{
				namespace:         "default",
				podName:           "pod1",
				nominatedNodeName: "node1",
				pluginName:        "plugin2",
				nodeNames:         []string{"node1", "node2"},
			},
			wantResultMap: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalScore: map[string]map[string]string{},
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{
						"node0": {},
						"node1": {
							"plugin2": PostFilterNominatedMessage,
						},
						"node2": {},
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
			s.AddPostFilterResult(tt.args.namespace, tt.args.podName, tt.args.nominatedNodeName, tt.args.pluginName, tt.args.nodeNames)
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
					selectedNode:    "",
					preScore:        map[string]string{},
					preFilterStatus: map[string]string{},
					preFilterResult: map[string][]string{},
					permit:          map[string]string{},
					permitTimeout:   map[string]string{},
					reserve:         map[string]string{},
					prebind:         map[string]string{},
					bind:            map[string]string{},
					filter:          map[string]map[string]string{},
					postFilter:      map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					selectedNode:    "",
					preScore:        map[string]string{},
					preFilterStatus: map[string]string{},
					preFilterResult: map[string][]string{},
					permit:          map[string]string{},
					permitTimeout:   map[string]string{},
					reserve:         map[string]string{},
					prebind:         map[string]string{},
					bind:            map[string]string{},
					filter:          map[string]map[string]string{},
					postFilter:      map[string]map[string]string{},
					score:           map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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
					filter:     map[string]map[string]string{},
					postFilter: map[string]map[string]string{},
					finalScore: map[string]map[string]string{
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

func TestStore_AddStoredResultToPod(t *testing.T) {
	t.Parallel()
	podName := "pod1"
	namespace := "default"
	tests := []struct {
		name    string
		result  map[key]*result
		newObj  *corev1.Pod
		wantpod *corev1.Pod
	}{
		{
			name: "success",
			result: map[key]*result{
				"default/pod1": {
					selectedNode: "node",
					preScore: map[string]string{
						"plugin1": "preScore",
					},
					preFilterStatus: map[string]string{
						"plugin1": "preFilterStatus",
					},
					preFilterResult: map[string][]string{
						"plugin1": {"node1", "node2"},
					},
					permit: map[string]string{
						"plugin1": "permit",
					},
					permitTimeout: map[string]string{
						"plugin1": "1s",
					},
					reserve: map[string]string{
						"plugin1": "reserve",
					},
					prebind: map[string]string{
						"plugin1": "prebind",
					},
					bind: map[string]string{
						"plugin1": "bind",
					},
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					finalScore: map[string]map[string]string{
						"node0": {
							"plugin1": "20",
						},
						"node1": {
							"plugin1": "20",
						},
					},
					postFilter: map[string]map[string]string{
						"node0": {
							"plugin1": PostFilterNominatedMessage,
						},
						"node1": {},
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
						annotation.SelectedNodeAnnotationKey: "node",
						annotation.PreScoreResultAnnotationKey: func() string {
							d, _ := json.Marshal(map[string]string{
								"plugin1": "preScore",
							})
							return string(d)
						}(),
						annotation.PreFilterResultAnnotationKey: func() string {
							d, _ := json.Marshal(map[string][]string{
								"plugin1": {"node1", "node2"},
							})
							return string(d)
						}(),
						annotation.PreFilterStatusResultAnnotationKey: func() string {
							d, _ := json.Marshal(map[string]string{
								"plugin1": "preFilterStatus",
							})
							return string(d)
						}(),
						annotation.PermitStatusResultAnnotationKey: func() string {
							d, _ := json.Marshal(map[string]string{
								"plugin1": "permit",
							})
							return string(d)
						}(),
						annotation.PermitTimeoutResultAnnotationKey: func() string {
							d, _ := json.Marshal(map[string]string{
								"plugin1": "1s",
							})
							return string(d)
						}(),
						annotation.ReserveResultAnnotationKey: func() string {
							d, _ := json.Marshal(map[string]string{
								"plugin1": "reserve",
							})
							return string(d)
						}(),
						annotation.PreBindResultAnnotationKey: func() string {
							d, _ := json.Marshal(map[string]string{
								"plugin1": "prebind",
							})
							return string(d)
						}(),
						annotation.BindResultAnnotationKey: func() string {
							d, _ := json.Marshal(map[string]string{
								"plugin1": "bind",
							})
							return string(d)
						}(),
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
						annotation.PostFilterResultAnnotationKey: func() string {
							r := map[string]map[string]string{
								"node0": {
									"plugin1": PostFilterNominatedMessage,
								},
								"node1": {},
							}
							d, _ := json.Marshal(r)
							return string(d)
						}(),
					},
				},
			},
		},
		{
			name:   "do nothing if store doesn't have data",
			result: map[key]*result{},
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
		},
		{
			name: "success without some data on store",
			result: map[key]*result{
				"default/pod1": {
					score:      map[string]map[string]string{},
					finalScore: map[string]map[string]string{},
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					postFilter: map[string]map[string]string{},
				},
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
						annotation.ScoreResultAnnotationKey:           "{}",
						annotation.FinalScoreResultAnnotationKey:      "{}",
						annotation.PostFilterResultAnnotationKey:      "{}",
						annotation.SelectedNodeAnnotationKey:          "",
						annotation.PreScoreResultAnnotationKey:        "{}",
						annotation.PreFilterResultAnnotationKey:       "{}",
						annotation.PreFilterStatusResultAnnotationKey: "{}",
						annotation.PermitStatusResultAnnotationKey:    "{}",
						annotation.PermitTimeoutResultAnnotationKey:   "{}",
						annotation.ReserveResultAnnotationKey:         "{}",
						annotation.PreBindResultAnnotationKey:         "{}",
						annotation.BindResultAnnotationKey:            "{}",
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
				results: tt.result,
			}
			p := tt.newObj
			s.AddStoredResultToPod(p)

			assert.Equal(t, tt.wantpod, p)
		})
	}
}

func TestStore_AddPreFilterResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace       string
		podName         string
		pluginName      string
		reason          string
		preFilterResult *framework.PreFilterResult
	}
	tests := []struct {
		name          string
		args          args
		wantResultMap map[key]*result
	}{
		{
			name: "success",
			args: args{
				namespace:  "namespace",
				podName:    "pod",
				pluginName: "plugin",
				reason:     "reason",
				preFilterResult: &framework.PreFilterResult{
					NodeNames: sets.NewString("hoge"),
				},
			},
			wantResultMap: func() map[key]*result {
				d := newData()
				d.preFilterResult = map[string][]string{
					"plugin": {"hoge"},
				}
				d.preFilterStatus = map[string]string{
					"plugin": "reason",
				}
				return map[key]*result{
					"namespace/pod": d,
				}
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Store{mu: &sync.Mutex{}, results: map[key]*result{}}
			s.AddPreFilterResult(tt.args.namespace, tt.args.podName, tt.args.pluginName, tt.args.reason, tt.args.preFilterResult)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_AddPreScoreResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace  string
		podName    string
		pluginName string
		reason     string
	}
	tests := []struct {
		name          string
		args          args
		wantResultMap map[key]*result
	}{
		{
			name: "success",
			args: args{
				namespace:  "namespace",
				podName:    "pod",
				pluginName: "plugin",
				reason:     "reason",
			},
			wantResultMap: func() map[key]*result {
				d := newData()
				d.preScore = map[string]string{
					"plugin": "reason",
				}
				return map[key]*result{
					"namespace/pod": d,
				}
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Store{mu: &sync.Mutex{}, results: map[key]*result{}}
			s.AddPreScoreResult(tt.args.namespace, tt.args.podName, tt.args.pluginName, tt.args.reason)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_AddPermitResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace  string
		podName    string
		pluginName string
		status     string
		timeout    time.Duration
	}
	tests := []struct {
		name          string
		wantResultMap map[key]*result
		args          args
	}{
		{
			name: "success",
			args: args{
				namespace:  "namespace",
				podName:    "pod",
				pluginName: "plugin",
				status:     "success",
				timeout:    time.Duration(1), // meaning 1ns
			},
			wantResultMap: func() map[key]*result {
				d := newData()
				d.permit = map[string]string{
					"plugin": "success",
				}
				d.permitTimeout = map[string]string{
					"plugin": "1ns",
				}
				return map[key]*result{
					"namespace/pod": d,
				}
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Store{mu: &sync.Mutex{}, results: map[key]*result{}}
			s.AddPermitResult(tt.args.namespace, tt.args.podName, tt.args.pluginName, tt.args.status, tt.args.timeout)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_AddSelectedNode(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace string
		podName   string
		nodeName  string
	}
	tests := []struct {
		name          string
		args          args
		wantResultMap map[key]*result
	}{
		{
			name: "success",
			args: args{
				namespace: "namespace",
				podName:   "pod",
				nodeName:  "node",
			},
			wantResultMap: func() map[key]*result {
				d := newData()
				d.selectedNode = "node"
				return map[key]*result{
					"namespace/pod": d,
				}
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Store{mu: &sync.Mutex{}, results: map[key]*result{}}
			s.AddSelectedNode(tt.args.namespace, tt.args.podName, tt.args.nodeName)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_AddReserveResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace  string
		podName    string
		pluginName string
		status     string
	}
	tests := []struct {
		name          string
		args          args
		wantResultMap map[key]*result
	}{
		{
			name: "success",
			args: args{
				namespace:  "namespace",
				podName:    "pod",
				pluginName: "plugin",
				status:     "success",
			},
			wantResultMap: func() map[key]*result {
				d := newData()
				d.reserve = map[string]string{
					"plugin": "success",
				}
				return map[key]*result{
					"namespace/pod": d,
				}
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Store{mu: &sync.Mutex{}, results: map[key]*result{}}
			s.AddReserveResult(tt.args.namespace, tt.args.podName, tt.args.pluginName, tt.args.status)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_AddBindResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace  string
		podName    string
		pluginName string
		status     string
	}
	tests := []struct {
		name          string
		args          args
		wantResultMap map[key]*result
	}{
		{
			name: "success",
			args: args{
				namespace:  "namespace",
				podName:    "pod",
				pluginName: "plugin",
				status:     "success",
			},
			wantResultMap: func() map[key]*result {
				d := newData()
				d.bind = map[string]string{
					"plugin": "success",
				}
				return map[key]*result{
					"namespace/pod": d,
				}
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Store{mu: &sync.Mutex{}, results: map[key]*result{}}
			s.AddBindResult(tt.args.namespace, tt.args.podName, tt.args.pluginName, tt.args.status)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_AddPreBindResult(t *testing.T) {
	t.Parallel()
	type args struct {
		namespace  string
		podName    string
		pluginName string
		status     string
	}
	tests := []struct {
		name          string
		args          args
		wantResultMap map[key]*result
	}{
		{
			name: "success",
			args: args{
				namespace:  "namespace",
				podName:    "pod",
				pluginName: "plugin",
				status:     "success",
			},
			wantResultMap: func() map[key]*result {
				d := newData()
				d.prebind = map[string]string{
					"plugin": "success",
				}
				return map[key]*result{
					"namespace/pod": d,
				}
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &Store{mu: &sync.Mutex{}, results: map[key]*result{}}
			s.AddPreBindResult(tt.args.namespace, tt.args.podName, tt.args.pluginName, tt.args.status)
			assert.Equal(t, tt.wantResultMap, s.results)
		})
	}
}

func TestStore_DeleteData(t *testing.T) {
	t.Parallel()
	podName := "pod1"
	namespace := "default"
	tests := []struct {
		name       string
		target     corev1.Pod
		result     map[key]*result
		wantResult map[key]*result
	}{
		{
			name: "success to delete the stored data which has the specified key.",
			target: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			},
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
					finalScore: map[string]map[string]string{
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
				"default/pod2": {
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					finalScore: map[string]map[string]string{
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
			wantResult: map[key]*result{
				"default/pod2": {
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					finalScore: map[string]map[string]string{
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
		{
			name: "do nothing if store doesn't have the data.",
			target: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			},
			result: map[key]*result{
				"default/pod2": {
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					finalScore: map[string]map[string]string{
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
			wantResult: map[key]*result{
				"default/pod2": {
					filter: map[string]map[string]string{
						"node0": {
							"plugin1": PassedFilterMessage,
						},
						"node1": {
							"plugin1": PassedFilterMessage,
						},
					},
					finalScore: map[string]map[string]string{
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
				mu:      new(sync.Mutex),
				results: tt.result,
			}
			s.DeleteData(tt.target)

			assert.Equal(t, tt.wantResult, s.results)
		})
	}
}
