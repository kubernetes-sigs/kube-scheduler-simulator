package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			"EBSLimits",
			"GCEPDLimits",
			"NodeVolumeLimits",
			"AzureDiskLimits",
			"VolumeBinding",
			"VolumeZone",
			"PodTopologySpread",
			"InterPodAffinity",
			"DefaultPreemption",
			"NodeResourcesBalancedAllocation",
			"ImageLocality",
			"DefaultBinder",
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
