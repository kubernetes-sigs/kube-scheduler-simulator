package resultstore

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/extender/annotation"
)

func TestStore_AddStoredResultToPod(t *testing.T) {
	t.Parallel()
	podName := "pod1"
	namespace := "default"
	tests := []struct {
		name           string
		result         map[key]*result
		newObj         *corev1.Pod
		wantAnnotation map[string]string
	}{
		{
			name: "success",
			result: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{
						"node0": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"node0": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"node0": {
							Error: "myerror",
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
			wantAnnotation: map[string]string{
				annotation.ExtenderFilterResultAnnotationKey: func() string {
					r := map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					}
					d, _ := json.Marshal(r)
					return string(d)
				}(),
				annotation.ExtenderPrioritizeResultAnnotationKey: func() string {
					r := map[string]extenderv1.HostPriorityList{
						"node0": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					}
					d, _ := json.Marshal(r)
					return string(d)
				}(),
				annotation.ExtenderPreemptResultAnnotationKey: func() string {
					r := map[string]extenderv1.ExtenderPreemptionResult{
						"node0": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					}
					d, _ := json.Marshal(r)
					return string(d)
				}(),
				annotation.ExtenderBindResultAnnotationKey: func() string {
					r := map[string]extenderv1.ExtenderBindingResult{
						"node0": {
							Error: "myerror",
						},
					}
					d, _ := json.Marshal(r)
					return string(d)
				}(),
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
		},
		{
			name: "success without some data on store",
			result: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			newObj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
			},
			wantAnnotation: map[string]string{
				annotation.ExtenderFilterResultAnnotationKey: func() string {
					r := map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					}
					d, _ := json.Marshal(r)
					return string(d)
				}(),
				annotation.ExtenderPrioritizeResultAnnotationKey: "{}",
				annotation.ExtenderPreemptResultAnnotationKey:    "{}",
				annotation.ExtenderBindResultAnnotationKey:       "{}",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &store{
				mu:      new(sync.Mutex),
				results: tt.result,
			}
			p := tt.newObj
			result := s.AddStoredResultToPod(p)

			assert.Equal(t, tt.wantAnnotation, result)
		})
	}
}

func TestStore_AddFilterResult(t *testing.T) {
	t.Parallel()
	podName := "pod1"
	namespace := "default"
	tests := []struct {
		name          string
		hostname      string
		args          extenderv1.ExtenderArgs
		filterResult  extenderv1.ExtenderFilterResult
		prepareResult map[key]*result
		wantResult    map[key]*result
	}{
		{
			name:     "success to add the result",
			hostname: "extenderserver",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			filterResult: extenderv1.ExtenderFilterResult{
				Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
				NodeNames:                  &[]string{"node1"},
				FailedNodes:                map[string]string{"foo": "bar"},
				FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
				Error:                      "myerror",
			},
			prepareResult: map[key]*result{},
			wantResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "overwrite to the already stored data which has the same key and hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			filterResult: extenderv1.ExtenderFilterResult{
				Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
				NodeNames:                  &[]string{"node1"},
				FailedNodes:                map[string]string{"foo": "bar"},
				FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
				Error:                      "myerror",
			},
			prepareResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename0"}}}},
							NodeNames:                  &[]string{"node0"},
							FailedNodes:                map[string]string{"foo": "foo"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "foo"},
							Error:                      "",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "shouldn't overwrite to the already stored data which has the same key and different hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			filterResult: extenderv1.ExtenderFilterResult{
				Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
				NodeNames:                  &[]string{"node1"},
				FailedNodes:                map[string]string{"foo": "bar"},
				FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
				Error:                      "myerror",
			},
			prepareResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"different-extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename0"}}}},
							NodeNames:                  &[]string{"node0"},
							FailedNodes:                map[string]string{"foo": "foo"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "foo"},
							Error:                      "",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
						"different-extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename0"}}}},
							NodeNames:                  &[]string{"node0"},
							FailedNodes:                map[string]string{"foo": "foo"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "foo"},
							Error:                      "",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "overwrite to the already stored data which has the different key and same hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			filterResult: extenderv1.ExtenderFilterResult{
				Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
				NodeNames:                  &[]string{"node1"},
				FailedNodes:                map[string]string{"foo": "bar"},
				FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
				Error:                      "myerror",
			},
			prepareResult: map[key]*result{
				"default/pod2": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename0"}}}},
							NodeNames:                  &[]string{"node0"},
							FailedNodes:                map[string]string{"foo": "foo"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "foo"},
							Error:                      "",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
				"default/pod2": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"extenderserver": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename0"}}}},
							NodeNames:                  &[]string{"node0"},
							FailedNodes:                map[string]string{"foo": "foo"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "foo"},
							Error:                      "",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind:       map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &store{
				mu:      new(sync.Mutex),
				results: tt.prepareResult,
			}
			s.AddFilterResult(tt.args, tt.filterResult, tt.hostname)

			assert.Equal(t, tt.wantResult, s.results)
		})
	}
}

func TestStore_AddPrioritizeResult(t *testing.T) {
	t.Parallel()
	podName := "pod1"
	namespace := "default"
	tests := []struct {
		name          string
		hostname      string
		args          extenderv1.ExtenderArgs
		prioritize    extenderv1.HostPriorityList
		prepareResult map[key]*result
		wantResult    map[key]*result
	}{
		{
			name:     "success to add the result",
			hostname: "extenderserver",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			prioritize: extenderv1.HostPriorityList{
				{
					Host:  "node1",
					Score: 1.0,
				},
			},
			prepareResult: map[key]*result{},
			wantResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{
						"extenderserver": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{},
					bind:    map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "overwrite to the already stored data which has the same key and hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			prioritize: extenderv1.HostPriorityList{
				{
					Host:  "node1",
					Score: 1.0,
				},
			},
			prepareResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{
						"extenderserver": {
							{
								Host:  "node0",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{},
					bind:    map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{
						"extenderserver": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{},
					bind:    map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "shouldn't overwrite to the already stored data which has the same key and different hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			prioritize: extenderv1.HostPriorityList{
				{
					Host:  "node1",
					Score: 1.0,
				},
			},
			prepareResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{
						"different-extenderserver": {
							{
								Host:  "node0",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{},
					bind:    map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{
						"extenderserver": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
						"different-extenderserver": {
							{
								Host:  "node0",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{},
					bind:    map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "overwrite to the already stored data which has the different key and same hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			prioritize: extenderv1.HostPriorityList{
				{
					Host:  "node1",
					Score: 1.0,
				},
			},
			prepareResult: map[key]*result{
				"default/pod2": {
					filter: map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{
						"extenderserver": {
							{
								Host:  "node0",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{},
					bind:    map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter: map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{
						"extenderserver": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{},
					bind:    map[string]extenderv1.ExtenderBindingResult{},
				},
				"default/pod2": {
					filter: map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{
						"extenderserver": {
							{
								Host:  "node0",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{},
					bind:    map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &store{
				mu:      new(sync.Mutex),
				results: tt.prepareResult,
			}
			s.AddPrioritizeResult(tt.args, tt.prioritize, tt.hostname)

			assert.Equal(t, tt.wantResult, s.results)
		})
	}
}

func TestStore_AddPreemptResult(t *testing.T) {
	t.Parallel()
	podName := "pod1"
	namespace := "default"
	tests := []struct {
		name             string
		hostname         string
		args             extenderv1.ExtenderPreemptionArgs
		preemptionResult extenderv1.ExtenderPreemptionResult
		prepareResult    map[key]*result
		wantResult       map[key]*result
	}{
		{
			name:     "success to add the result",
			hostname: "extenderserver",
			args: extenderv1.ExtenderPreemptionArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			preemptionResult: extenderv1.ExtenderPreemptionResult{
				NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
			},
			prepareResult: map[key]*result{},
			wantResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "overwrite to the already stored data which has the same key and hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderPreemptionArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			preemptionResult: extenderv1.ExtenderPreemptionResult{
				NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
			},
			prepareResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"bar": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "shouldn't overwrite to the already stored data which has the same key and different hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderPreemptionArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			preemptionResult: extenderv1.ExtenderPreemptionResult{
				NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
			},
			prepareResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"different-extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"bar": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
						"different-extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"bar": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
		{
			name:     "overwrite to the already stored data which has the different key and same hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderPreemptionArgs{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      podName,
						Namespace: namespace,
					},
				},
			},
			preemptionResult: extenderv1.ExtenderPreemptionResult{
				NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
			},
			prepareResult: map[key]*result{
				"default/pod2": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"bar": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{},
				},
				"default/pod2": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"extenderserver": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"bar": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &store{
				mu:      new(sync.Mutex),
				results: tt.prepareResult,
			}
			s.AddPreemptResult(tt.args, tt.preemptionResult, tt.hostname)

			assert.Equal(t, tt.wantResult, s.results)
		})
	}
}

func TestStore_AddBindResult(t *testing.T) {
	t.Parallel()
	podName := "pod1"
	namespace := "default"
	tests := []struct {
		name          string
		hostname      string
		args          extenderv1.ExtenderBindingArgs
		bindingResult extenderv1.ExtenderBindingResult
		prepareResult map[key]*result
		wantResult    map[key]*result
	}{
		{
			name:     "success to add the result",
			hostname: "extenderserver",
			args: extenderv1.ExtenderBindingArgs{
				PodName:      podName,
				PodNamespace: namespace,
			},
			bindingResult: extenderv1.ExtenderBindingResult{
				Error: "myerror",
			},
			prepareResult: map[key]*result{},
			wantResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"extenderserver": {
							Error: "myerror",
						},
					},
				},
			},
		},
		{
			name:     "overwrite to the already stored data which has the same key and hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderBindingArgs{
				PodName:      podName,
				PodNamespace: namespace,
			},
			bindingResult: extenderv1.ExtenderBindingResult{
				Error: "myerror",
			},
			prepareResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"extenderserver": {
							Error: "myerror1",
						},
					},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"extenderserver": {
							Error: "myerror",
						},
					},
				},
			},
		},
		{
			name:     "shouldn't overwrite to the already stored data which has the same key and different hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderBindingArgs{
				PodName:      podName,
				PodNamespace: namespace,
			},
			bindingResult: extenderv1.ExtenderBindingResult{
				Error: "myerror",
			},
			prepareResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"different-extenderserver": {
							Error: "myerror1",
						},
					},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"extenderserver": {
							Error: "myerror",
						},
						"different-extenderserver": {
							Error: "myerror1",
						},
					},
				},
			},
		},
		{
			name:     "overwrite to the already stored data which has the different key and same hostname",
			hostname: "extenderserver",
			args: extenderv1.ExtenderBindingArgs{
				PodName:      podName,
				PodNamespace: namespace,
			},
			bindingResult: extenderv1.ExtenderBindingResult{
				Error: "myerror",
			},
			prepareResult: map[key]*result{
				"default/pod2": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"extenderserver": {
							Error: "myerror",
						},
					},
				},
			},
			wantResult: map[key]*result{
				"default/pod1": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"extenderserver": {
							Error: "myerror",
						},
					},
				},
				"default/pod2": {
					filter:     map[string]extenderv1.ExtenderFilterResult{},
					prioritize: map[string]extenderv1.HostPriorityList{},
					preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"extenderserver": {
							Error: "myerror",
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
			s := &store{
				mu:      new(sync.Mutex),
				results: tt.prepareResult,
			}
			s.AddBindResult(tt.args, tt.bindingResult, tt.hostname)

			assert.Equal(t, tt.wantResult, s.results)
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
					filter: map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{
						"node0": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"node0": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"node0": {
							Error: "myerror",
						},
					},
				},
				"default/pod2": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{
						"node0": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"node0": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"node0": {
							Error: "myerror",
						},
					},
				},
			},
			wantResult: map[key]*result{
				"default/pod2": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{
						"node0": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"node0": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"node0": {
							Error: "myerror",
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
					filter: map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{
						"node0": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"node0": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"node0": {
							Error: "myerror",
						},
					},
				},
			},
			wantResult: map[key]*result{
				"default/pod2": {
					filter: map[string]extenderv1.ExtenderFilterResult{
						"node0": {
							Nodes:                      &corev1.NodeList{Items: []corev1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "nodename"}}}},
							NodeNames:                  &[]string{"node1"},
							FailedNodes:                map[string]string{"foo": "bar"},
							FailedAndUnresolvableNodes: map[string]string{"baz": "qux"},
							Error:                      "myerror",
						},
					},
					prioritize: map[string]extenderv1.HostPriorityList{
						"node0": {
							{
								Host:  "node1",
								Score: 1.0,
							},
						},
					},
					preempt: map[string]extenderv1.ExtenderPreemptionResult{
						"node0": {
							NodeNameToMetaVictims: map[string]*extenderv1.MetaVictims{"foo": {Pods: []*extenderv1.MetaPod{{UID: "myuid"}}, NumPDBViolations: 1}},
						},
					},
					bind: map[string]extenderv1.ExtenderBindingResult{
						"node0": {
							Error: "myerror",
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
			s := &store{
				mu:      new(sync.Mutex),
				results: tt.result,
			}
			s.DeleteData(tt.target)

			assert.Equal(t, tt.wantResult, s.results)
		})
	}
}
