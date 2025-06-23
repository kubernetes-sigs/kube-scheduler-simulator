package config

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	"k8s.io/utils/strings/slices"
)

func TestInTreeMultiPointPluginSet(t *testing.T) {
	t.Parallel()
	t.Run("", func(t *testing.T) {
		wantEnabled := []string{
			"PrioritySort",
			"NodeUnschedulable",
			"NodeName",
			"TaintToleration",
			"NodeAffinity",
			"NodePorts",
			"NodeResourcesFit",
			"VolumeRestrictions",
			"NodeVolumeLimits",
			"VolumeBinding",
			"VolumeZone",
			"PodTopologySpread",
			"InterPodAffinity",
			"DefaultPreemption",
			"NodeResourcesBalancedAllocation",
			"ImageLocality",
			"DefaultBinder",
			"SchedulingGates",
		}
		wantDisabled := []string{}

		mp, err := InTreeMultiPointPluginSet()
		assert.NoError(t, err, "check error")

		var count int
		for _, p := range mp.Enabled {
			if !slices.Contains(wantEnabled, p.Name) {
				t.Errorf("unexpected enabled plugin name is contained: name=%s", p.Name)
			}
			count++
		}
		assert.Equal(t, len(wantEnabled), count, "check sum of default enabled plugin")

		count = 0
		for _, p := range mp.Disabled {
			if !slices.Contains(wantDisabled, p.Name) {
				t.Errorf("unexpected disabled plugin name is contained: name=%s", p.Name)
			}
			count++
		}
		assert.Equal(t, len(wantDisabled), count, "check sum of default disabled plugin")
	})
}

//nolint:paralleltest // cannot use t.Parallel because SetOutOfTreeRegistries affects other test cases.
func TestRegisteredMultiPointPluginNames(t *testing.T) {
	tests := []struct {
		name              string
		outOfTreeRegistry runtime.Registry
		want              []string
		wantErr           bool
	}{
		{
			name: "success",
			want: []string{
				"PrioritySort",
				"SchedulingGates",
				"NodeName",
				"TaintToleration",
				"NodeAffinity",
				"NodeUnschedulable",
				"NodeResourcesBalancedAllocation",
				"ImageLocality",
				"InterPodAffinity",
				"NodeResourcesFit",
				"PodTopologySpread",
				"DefaultBinder",
				"VolumeBinding",
				"NodePorts",
				"VolumeRestrictions",
				"NodeVolumeLimits",
				"VolumeZone",
				"DefaultPreemption",
			},
			wantErr: false,
		},
		{
			name: "success with out of tree",
			want: []string{
				"PrioritySort",
				"SchedulingGates",
				"NodeName",
				"TaintToleration",
				"NodeAffinity",
				"NodeUnschedulable",
				"NodeResourcesBalancedAllocation",
				"ImageLocality",
				"InterPodAffinity",
				"NodeResourcesFit",
				"PodTopologySpread",
				"DefaultBinder",
				"VolumeBinding",
				"NodePorts",
				"VolumeRestrictions",
				"NodeVolumeLimits",
				"VolumeZone",
				"DefaultPreemption",
				"custom", // added.
			},
			outOfTreeRegistry: map[string]runtime.PluginFactory{
				"custom": nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetOutOfTreeRegistries(tt.outOfTreeRegistry)
			got, err := RegisteredMultiPointPluginNames()
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisteredPluginNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.SliceStable(got, func(i, j int) bool {
				return got[i] < got[j]
			})
			sort.SliceStable(tt.want, func(i, j int) bool {
				return tt.want[i] < tt.want[j]
			})
			assert.Equal(t, tt.want, got)
		})
	}
}
