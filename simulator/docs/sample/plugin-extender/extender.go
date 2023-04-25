package extender

import (
	"context"
	"encoding/json"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/noderesources"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/util"
)

type noderesourcefitPreFilterPluginExtender struct {
	handle plugin.SimulatorHandle
}

// New initializes noderesourcefitPreFilterPluginExtender in plugin.PluginExtenders.
func New(handle plugin.SimulatorHandle) plugin.PluginExtenders {
	e := &noderesourcefitPreFilterPluginExtender{
		handle: handle,
	}

	return plugin.PluginExtenders{PreFilterPluginExtender: e} // only PreFilterPluginExtender
}

func (e *noderesourcefitPreFilterPluginExtender) BeforePreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	klog.Info("execute BeforePreFilter on noderesourcefitPreFilterPluginExtender", "pod", klog.KObj(pod))
	// do nothing.
	return nil, nil
}

// AfterPreFilter checks what noderesource plugin stores into the cyclestate for this scheduling,
// and store it through the SimulatorHandle.
//
// By this func, each Pod will get noderesourcefit-prefilter-data after scheduling like:
// ---
// kind: Pod
// apiVersion: v1
// metadata:
//
//	name: pod-8ldq5
//	namespace: default
//	annotations:
//	  noderesourcefit-prefilter-data: >-
//	    {"MilliCPU":100,"Memory":17179869184,"EphemeralStorage":0,"AllowedPodNumber":0,"ScalarResources":null}
func (e *noderesourcefitPreFilterPluginExtender) AfterPreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, preFilterResult *framework.PreFilterResult, preFilterStatus *framework.Status) (*framework.PreFilterResult, *framework.Status) {
	klog.Info("execute AfterPreFilter on noderesourcefitPreFilterPluginExtender", "pod", klog.KObj(pod))

	c, err := state.Read("PreFilter" + noderesources.Name)
	if err != nil {
		klog.Info("no state data", "pod", klog.KObj(pod))
		return preFilterResult, preFilterStatus
	}

	// use util.PrivateFieldsDecoder to access private fields of c.
	value := util.PrivateFieldsDecoder(c, "Resource")
	data := value.Interface().(framework.Resource)

	j, err := json.Marshal(data)
	if err != nil {
		klog.Info("json marshal failed in extender", "pod", klog.KObj(pod))
		return preFilterResult, preFilterStatus
	}

	prefilterData := string(j)

	// store data via plugin.SimulatorHandle
	e.handle.AddCustomResult(pod.Namespace, pod.Name, "noderesourcefit-prefilter-data", prefilterData)

	return preFilterResult, preFilterStatus
}
