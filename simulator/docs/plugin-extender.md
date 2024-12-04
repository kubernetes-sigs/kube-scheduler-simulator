## Plugin extenders

**The plugin extender can be used only with [the external scheduler](./external-scheduler.md).**

The simulator has the concept "Plugin Extenders" which allows you to:
- export plugin's internal state more
- change specific behaviours on particular plugin by injecting the result
- etc...

(Note that it's not related to the scheduler's webhook-based extension which is also called ["extender"](./extender.md). 
(Sorry for the confusing name 😅))

The Plugin Extenders has `BeforeXXX` and `AfterXXX` for each extension point. (XXX = any extension points. e.g., Filter, Score..etc)

For example, `BeforeFilter` is literally called before Filter plugin,
and `AfterFilter` func is called after Filter plugin.

There are multiple interfaces named `XXXXPluginExtender`.

```go
// FilterPluginExtender is the extender for Filter plugin.
type FilterPluginExtender interface {
	// BeforeFilter is a function that runs before the Filter method of the original plugin.
	// If BeforeFilter returns non-success status, the simulator plugin doesn't run the Filter method of the original plugin and return that status.
	BeforeFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status
	// AfterFilter is a function that is run after the Filter method of the original plugin.
	// A Filter of the simulator plugin finally returns the status returned from AfterFilter.
	AfterFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo, filterResult *framework.Status) *framework.Status
}
```

### export something in each Pod's annotation via `SimulatorHandle`

Each PluginExtender can have `SimulatorHandle`, and you can export some internal state through `SimulatorHandle`.

Example:

```go
func (e *noderesourcefitPreFilterPluginExtender) AfterPreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, preFilterResult *framework.PreFilterResult, preFilterStatus *framework.Status) (*framework.PreFilterResult, *framework.Status) {
	// see ./sample/extender/extender.go
	//...
    e.handle.AddCustomResult(pod.Namespace, pod.Name, "noderesourcefit-prefilter-data", prefilterData)
}
```

If you use the above extender, 
each Pod will get `"noderesourcefit-prefilter-data": prefilterData` annotation in each scheduling like other scheduling results.

### use plugin extender

**Currently, the plugin extender can be used only in [the external scheduler](./external-scheduler.md).**

You can use `debuggablescheduler.WithPluginExtenders` option in `debuggablescheduler.NewSchedulerCommand`
to enable some PluginExtender in particular plugin.

```go
func main() {
	command, cancelFn, err := debuggablescheduler.NewSchedulerCommand(
        debuggablescheduler.WithPluginExtenders(noderesources.Name, extender.New),
    )
    if err != nil {
        klog.Info(fmt.Sprintf("failed to build the scheduler command: %+v", err))
        os.Exit(1)
    }
    code := cli.Run(command)
    cancelFn()
    os.Exit(code)
}
```

### The example plugin extender 

We have the sample plugin extender implementation in [./sample/extender](./sample/plugin-extender).

Please follow [this](./external-scheduler.md#the-example-external-scheduler) 
to see how this sample plugin extender works with the external scheduler.

You will see each Pod gets `noderesourcefit-prefilter-data` annotation along with other scheduling results like this:

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: pod-8ldq5
  namespace: default
  annotations:
    noderesourcefit-prefilter-data: >-
      {"MilliCPU":100,"Memory":17179869184,"EphemeralStorage":0,"AllowedPodNumber":0,"ScalarResources":null}
    kube-scheduler-simulator.sigs.k8s.io/bind-result: '{"DefaultBinder":"success"}'
    kube-scheduler-simulator.sigs.k8s.io/filter-result: >-
      {"node-282x7":{"AzureDiskLimits":"passed","EBSLimits":"passed","GCEPDLimits":"passed","InterPodAffinity":"passed","NodeAffinity":"passed","NodeName":"passed","NodePorts":"passed","NodeResourcesFit":"passed","NodeUnschedulable":"passed","NodeVolumeLimits":"passed","PodTopologySpread":"passed","TaintToleration":"passed","VolumeBinding":"passed","VolumeRestrictions":"passed","VolumeZone":"passed"},"node-gp9t4":{"AzureDiskLimits":"passed","EBSLimits":"passed","GCEPDLimits":"passed","InterPodAffinity":"passed","NodeAffinity":"passed","NodeName":"passed","NodePorts":"passed","NodeResourcesFit":"passed","NodeUnschedulable":"passed","NodeVolumeLimits":"passed","PodTopologySpread":"passed","TaintToleration":"passed","VolumeBinding":"passed","VolumeRestrictions":"passed","VolumeZone":"passed"}}
    kube-scheduler-simulator.sigs.k8s.io/finalscore-result: >-
      {"node-282x7":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"52","NodeResourcesFit":"47","PodTopologySpread":"200","TaintToleration":"300","VolumeBinding":"0"},"node-gp9t4":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"73","PodTopologySpread":"200","TaintToleration":"300","VolumeBinding":"0"}}
    kube-scheduler-simulator.sigs.k8s.io/permit-result: '{}'
    kube-scheduler-simulator.sigs.k8s.io/permit-result-timeout: '{}'
    kube-scheduler-simulator.sigs.k8s.io/postfilter-result: '{}'
    kube-scheduler-simulator.sigs.k8s.io/prebind-result: '{"VolumeBinding":"success"}'
    kube-scheduler-simulator.sigs.k8s.io/prefilter-result: '{}'
    kube-scheduler-simulator.sigs.k8s.io/prefilter-result-status: >-
      {"InterPodAffinity":"success","NodeAffinity":"success","NodePorts":"success","NodeResourcesFit":"success","PodTopologySpread":"success","VolumeBinding":"success","VolumeRestrictions":"success"}
    kube-scheduler-simulator.sigs.k8s.io/prescore-result: >-
      {"InterPodAffinity":"success","NodeAffinity":"success","NodeNumber":"success","PodTopologySpread":"success","TaintToleration":"success"}
    kube-scheduler-simulator.sigs.k8s.io/reserve-result: '{"VolumeBinding":"success"}'
    kube-scheduler-simulator.sigs.k8s.io/result-history: >-
      [{"noderesourcefit-prefilter-data":"{\"MilliCPU\":100,\"Memory\":17179869184,\"EphemeralStorage\":0,\"AllowedPodNumber\":0,\"ScalarResources\":null}","kube-scheduler-simulator.sigs.k8s.io/bind-result":"{\"DefaultBinder\":\"success\"}","kube-scheduler-simulator.sigs.k8s.io/filter-result":"{\"node-282x7\":{\"AzureDiskLimits\":\"passed\",\"EBSLimits\":\"passed\",\"GCEPDLimits\":\"passed\",\"InterPodAffinity\":\"passed\",\"NodeAffinity\":\"passed\",\"NodeName\":\"passed\",\"NodePorts\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"NodeVolumeLimits\":\"passed\",\"PodTopologySpread\":\"passed\",\"TaintToleration\":\"passed\",\"VolumeBinding\":\"passed\",\"VolumeRestrictions\":\"passed\",\"VolumeZone\":\"passed\"},\"node-gp9t4\":{\"AzureDiskLimits\":\"passed\",\"EBSLimits\":\"passed\",\"GCEPDLimits\":\"passed\",\"InterPodAffinity\":\"passed\",\"NodeAffinity\":\"passed\",\"NodeName\":\"passed\",\"NodePorts\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"NodeVolumeLimits\":\"passed\",\"PodTopologySpread\":\"passed\",\"TaintToleration\":\"passed\",\"VolumeBinding\":\"passed\",\"VolumeRestrictions\":\"passed\",\"VolumeZone\":\"passed\"}}","kube-scheduler-simulator.sigs.k8s.io/finalscore-result":"{\"node-282x7\":{\"ImageLocality\":\"0\",\"InterPodAffinity\":\"0\",\"NodeAffinity\":\"0\",\"NodeNumber\":\"0\",\"NodeResourcesBalancedAllocation\":\"52\",\"NodeResourcesFit\":\"47\",\"PodTopologySpread\":\"200\",\"TaintToleration\":\"300\",\"VolumeBinding\":\"0\"},\"node-gp9t4\":{\"ImageLocality\":\"0\",\"InterPodAffinity\":\"0\",\"NodeAffinity\":\"0\",\"NodeNumber\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"73\",\"PodTopologySpread\":\"200\",\"TaintToleration\":\"300\",\"VolumeBinding\":\"0\"}}","kube-scheduler-simulator.sigs.k8s.io/permit-result":"{}","kube-scheduler-simulator.sigs.k8s.io/permit-result-timeout":"{}","kube-scheduler-simulator.sigs.k8s.io/postfilter-result":"{}","kube-scheduler-simulator.sigs.k8s.io/prebind-result":"{\"VolumeBinding\":\"success\"}","kube-scheduler-simulator.sigs.k8s.io/prefilter-result":"{}","kube-scheduler-simulator.sigs.k8s.io/prefilter-result-status":"{\"InterPodAffinity\":\"success\",\"NodeAffinity\":\"success\",\"NodePorts\":\"success\",\"NodeResourcesFit\":\"success\",\"PodTopologySpread\":\"success\",\"VolumeBinding\":\"success\",\"VolumeRestrictions\":\"success\"}","kube-scheduler-simulator.sigs.k8s.io/prescore-result":"{\"InterPodAffinity\":\"success\",\"NodeAffinity\":\"success\",\"NodeNumber\":\"success\",\"PodTopologySpread\":\"success\",\"TaintToleration\":\"success\"}","kube-scheduler-simulator.sigs.k8s.io/reserve-result":"{\"VolumeBinding\":\"success\"}","kube-scheduler-simulator.sigs.k8s.io/score-result":"{\"node-282x7\":{\"ImageLocality\":\"0\",\"InterPodAffinity\":\"0\",\"NodeAffinity\":\"0\",\"NodeNumber\":\"0\",\"NodeResourcesBalancedAllocation\":\"52\",\"NodeResourcesFit\":\"47\",\"PodTopologySpread\":\"0\",\"TaintToleration\":\"0\",\"VolumeBinding\":\"0\"},\"node-gp9t4\":{\"ImageLocality\":\"0\",\"InterPodAffinity\":\"0\",\"NodeAffinity\":\"0\",\"NodeNumber\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"73\",\"PodTopologySpread\":\"0\",\"TaintToleration\":\"0\",\"VolumeBinding\":\"0\"}}","kube-scheduler-simulator.sigs.k8s.io/selected-node":"node-gp9t4"}]
    kube-scheduler-simulator.sigs.k8s.io/score-result: >-
      {"node-282x7":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"52","NodeResourcesFit":"47","PodTopologySpread":"0","TaintToleration":"0","VolumeBinding":"0"},"node-gp9t4":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"73","PodTopologySpread":"0","TaintToleration":"0","VolumeBinding":"0"}}
    kube-scheduler-simulator.sigs.k8s.io/selected-node: node-gp9t4
```
