## Debuggable scheduler

The "debuggable scheduler" is the core concept of the simulator.

### What's the "debuggable scheduler"?

The debuggable scheduler is the scheduler that schedules Pods like other schedulers, 
but it exposes the scheduling results from each scheduling plugin on Pod annotation.

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: hoge-pod
  annotations:
    scheduler-simulator/bind-result: '{"DefaultBinder":"success"}'
    scheduler-simulator/filter-result: >-
      {"node-282x7":{"AzureDiskLimits":"passed","EBSLimits":"passed","GCEPDLimits":"passed","InterPodAffinity":"passed","NodeAffinity":"passed","NodeName":"passed","NodePorts":"passed","NodeResourcesFit":"passed","NodeUnschedulable":"passed","NodeVolumeLimits":"passed","PodTopologySpread":"passed","TaintToleration":"passed","VolumeBinding":"passed","VolumeRestrictions":"passed","VolumeZone":"passed"},"node-gp9t4":{"AzureDiskLimits":"passed","EBSLimits":"passed","GCEPDLimits":"passed","InterPodAffinity":"passed","NodeAffinity":"passed","NodeName":"passed","NodePorts":"passed","NodeResourcesFit":"passed","NodeUnschedulable":"passed","NodeVolumeLimits":"passed","PodTopologySpread":"passed","TaintToleration":"passed","VolumeBinding":"passed","VolumeRestrictions":"passed","VolumeZone":"passed"}}
    scheduler-simulator/finalscore-result: >-
      {"node-282x7":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"73","PodTopologySpread":"200","TaintToleration":"300","VolumeBinding":"0"},"node-gp9t4":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"73","PodTopologySpread":"200","TaintToleration":"300","VolumeBinding":"0"}}
    scheduler-simulator/permit-result: '{}'
    scheduler-simulator/permit-result-timeout: '{}'
    scheduler-simulator/postfilter-result: '{}'
    scheduler-simulator/prebind-result: '{"VolumeBinding":"success"}'
    scheduler-simulator/prefilter-result: '{}'
    scheduler-simulator/prefilter-result-status: >-
      {"InterPodAffinity":"success","NodeAffinity":"success","NodePorts":"success","NodeResourcesFit":"success","PodTopologySpread":"success","VolumeBinding":"success","VolumeRestrictions":"success"}
    scheduler-simulator/prescore-result: >-
      {"InterPodAffinity":"success","NodeAffinity":"success","NodeNumber":"success","PodTopologySpread":"success","TaintToleration":"success"}
    scheduler-simulator/reserve-result: '{"VolumeBinding":"success"}'
    scheduler-simulator/result-history: >-
      [{"noderesourcefit-prefilter-data":"{\"MilliCPU\":100,\"Memory\":17179869184,\"EphemeralStorage\":0,\"AllowedPodNumber\":0,\"ScalarResources\":null}","scheduler-simulator/bind-result":"{\"DefaultBinder\":\"success\"}","scheduler-simulator/filter-result":"{\"node-282x7\":{\"AzureDiskLimits\":\"passed\",\"EBSLimits\":\"passed\",\"GCEPDLimits\":\"passed\",\"InterPodAffinity\":\"passed\",\"NodeAffinity\":\"passed\",\"NodeName\":\"passed\",\"NodePorts\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"NodeVolumeLimits\":\"passed\",\"PodTopologySpread\":\"passed\",\"TaintToleration\":\"passed\",\"VolumeBinding\":\"passed\",\"VolumeRestrictions\":\"passed\",\"VolumeZone\":\"passed\"},\"node-gp9t4\":{\"AzureDiskLimits\":\"passed\",\"EBSLimits\":\"passed\",\"GCEPDLimits\":\"passed\",\"InterPodAffinity\":\"passed\",\"NodeAffinity\":\"passed\",\"NodeName\":\"passed\",\"NodePorts\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"NodeVolumeLimits\":\"passed\",\"PodTopologySpread\":\"passed\",\"TaintToleration\":\"passed\",\"VolumeBinding\":\"passed\",\"VolumeRestrictions\":\"passed\",\"VolumeZone\":\"passed\"}}","scheduler-simulator/finalscore-result":"{\"node-282x7\":{\"ImageLocality\":\"0\",\"InterPodAffinity\":\"0\",\"NodeAffinity\":\"0\",\"NodeNumber\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"73\",\"PodTopologySpread\":\"200\",\"TaintToleration\":\"300\",\"VolumeBinding\":\"0\"},\"node-gp9t4\":{\"ImageLocality\":\"0\",\"InterPodAffinity\":\"0\",\"NodeAffinity\":\"0\",\"NodeNumber\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"73\",\"PodTopologySpread\":\"200\",\"TaintToleration\":\"300\",\"VolumeBinding\":\"0\"}}","scheduler-simulator/permit-result":"{}","scheduler-simulator/permit-result-timeout":"{}","scheduler-simulator/postfilter-result":"{}","scheduler-simulator/prebind-result":"{\"VolumeBinding\":\"success\"}","scheduler-simulator/prefilter-result":"{}","scheduler-simulator/prefilter-result-status":"{\"InterPodAffinity\":\"success\",\"NodeAffinity\":\"success\",\"NodePorts\":\"success\",\"NodeResourcesFit\":\"success\",\"PodTopologySpread\":\"success\",\"VolumeBinding\":\"success\",\"VolumeRestrictions\":\"success\"}","scheduler-simulator/prescore-result":"{\"InterPodAffinity\":\"success\",\"NodeAffinity\":\"success\",\"NodeNumber\":\"success\",\"PodTopologySpread\":\"success\",\"TaintToleration\":\"success\"}","scheduler-simulator/reserve-result":"{\"VolumeBinding\":\"success\"}","scheduler-simulator/score-result":"{\"node-282x7\":{\"ImageLocality\":\"0\",\"InterPodAffinity\":\"0\",\"NodeAffinity\":\"0\",\"NodeNumber\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"73\",\"PodTopologySpread\":\"0\",\"TaintToleration\":\"0\",\"VolumeBinding\":\"0\"},\"node-gp9t4\":{\"ImageLocality\":\"0\",\"InterPodAffinity\":\"0\",\"NodeAffinity\":\"0\",\"NodeNumber\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"73\",\"PodTopologySpread\":\"0\",\"TaintToleration\":\"0\",\"VolumeBinding\":\"0\"}}","scheduler-simulator/selected-node":"node-282x7"}]
    scheduler-simulator/score-result: >-
      {"node-282x7":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"73","PodTopologySpread":"0","TaintToleration":"0","VolumeBinding":"0"},"node-gp9t4":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"73","PodTopologySpread":"0","TaintToleration":"0","VolumeBinding":"0"}}
    scheduler-simulator/selected-node: node-282x7
```

The simulator runs the debuggable scheduler built from the upstream scheduler by default,
that's why we can see the scheduling results like the above in the simulator.

## Build a debuggable scheduler from your scheduler

You can use the [`debuggablescheduler` package](../pkg/debuggablescheduler) to transform your scheduler into a debuggable scheduler.

The debuggable scheduler could work anywhere, in the simulator or even in your cluster.

### Change your scheduler

Here, we assume you're registering your custom plugins in your scheduler like this:

```go
// your scheduler's main package
func main() {
	command := app.NewSchedulerCommand(
		app.WithPlugin(yourcustomplugin.Name, yourcustomplugin.New),
	)

	code := cli.Run(command)
	os.Exit(code)
}
```

Then, you need to replace few lines to use the [`debuggablescheduler` package](../pkg/debuggablescheduler).

```go
func main() {
	command, cancelFn, err := debuggablescheduler.NewSchedulerCommand(
        debuggablescheduler.WithPlugin(yourcustomplugin.Name, yourcustomplugin.New),
        debuggablescheduler.WithPluginExtenders(noderesources.Name, extender.New), // [optional] see plugin-extender.md about PluginExtender.
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

As you see, `debuggablescheduler.NewSchedulerCommand` has much similar interface as the `app.NewSchedulerCommand`.
You can register your plugins by `debuggablescheduler.WithPlugin` option.

### The plugin extender

We have the plugin extender feature to expand the debuggable scheduler more.
See [plugin-extender.md](./plugin-extender.md) to know more about it.

### The example debuggable scheduler

We have the sample implementation in [./sample/debuggable-scheduler](./sample/debuggable-scheduler).

You can try to run it in the simulator via [external scheduler](./external-scheduler.md) feature.
See [external-scheduler.md#the-example-external-scheduler](./external-scheduler.md#the-example-external-scheduler).
