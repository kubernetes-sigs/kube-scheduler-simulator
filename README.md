# Kubernetes scheduler simulator

Hello world. Here is Kubernetes scheduler simulator.

Nowadays, the scheduler is configurable/extendable in the multiple ways:
- configure with [KubeSchedulerConfiguration](https://kubernetes.io/docs/reference/scheduling/config/)
- add Plugins of [Scheduling Framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/)
- add [Extenders](https://github.com/kubernetes/enhancements/tree/5320deb4834c05ad9fb491dcd361f952727ece3e/keps/sig-scheduling/1819-scheduler-extender)
- etc...

But, unfortunately, not all configurations/expansions yield good results.
Those who customize the scheduler need to make sure their scheduler is working as expected,
and doesn't have an unacceptably negative impact on the scheduling.

In real Kubernetes, we cannot know the results of scheduling in detail without reading the logs,
which usually requires privileged access to the control plane.

That's why we are developing a simulator for kube-scheduler
-- you can try out the behavior of the scheduler while checking which plugin made what decision for which Node.

## Getting started 

```shell
git clone git@github.com:kubernetes-sigs/kube-scheduler-simulator.git
cd kube-scheduler-simulator
# pull images from the registry and run up all components
make docker_up 
# All things up! 
```

You can access the simulator in http://localhost:3000.

Please see [this doc](./simulator/docs/running-simulator.md) to know other ways to run up the simulator.

## Simulator's architecture

We have several components:
- Simulator (in `/simulator`): the core implementation of the simulator
- Web UI (in `/web`): the Web client of the simulator
- Coming soon... :)  (see [./keps](./keps) to see some nice ideas we're working on)

### Simulator

Simulator is kube-apiserver + scheduler + the HTTP server which mainly for the web UI.

There are several ways to integrate your scheduler into the simulator.
See [integrate-your-scheduler.md](simulator/docs/integrate-your-scheduler.md).

You can create any resources by communicating with kube-apiserver in any ways (kubectl, k8s client library, or web UI described next)
and see how your Pods are scheduled.

When you create Pods, Pods will get annotations from the simulator which contains the scheduling results per plugins or extenders.

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

You can utilize these results to understand your scheduler, check/test your configurations or customized scheduler's behavior.

Further expansion, you can export internal state more, change specific behaviours on plugins etc by implementing [PluginExtender](./simulator/docs/plugin-extender.md).

The simulator has its own configuration,
you can refer to the [documentation](./simulator/docs/simulator-server-config.md).

See the following docs to know more about simulator:
- [import-cluster-resources.md](./simulator/docs/import-cluster-resources.md): describes how you can import resources in your cluster to the simulator so that you can simulate scheduling based on your cluster's situation.
- [how-it-works.md](simulator/docs/how-it-works.md): describes about how the simulator works.
- [kube-apiserver.md](simulator/docs/kube-apiserver.md): describe about kube-apiserver in simulator. (how you can configure and access)
- [api.md](simulator/docs/api.md): describes about HTTP server the simulator has. (mainly for the webUI)

### Web UI

Web UI is the easiest way to check the scheduler's behavior.

Nice table view for the scheduling result, the scheduler configuration reload feature; 
you can access every core features of the simulator with human-friendly ways!

![list resources](simulator/docs/images/resources.png)

It has a cool yaml editor to create/edit resources.

![create node](simulator/docs/images/create-node.png)

And, after pods are scheduled, you can see the results of

- Each Filter plugins
- Each Score plugins
- Final score (normalized and applied Plugin Weight)

![result](simulator/docs/images/result.jpg)

Also, You can change the configuration of the scheduler through [KubeSchedulerConfiguration](https://kubernetes.io/docs/reference/scheduling/config/) in Web UI.

(Note: changes to any fields other than `.profiles` are disabled on simulator, 
since they do not affect the results of the scheduling.)

![configure scheduler](simulator/docs/images/schedulerconfiguration.png)

Also, you can take a snapshot of the resources so that you can share the situation or load them later.

![snapshot button](simulator/docs/images/webUI-snapshot.png)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](http://slack.k8s.io/)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-dev)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[creative commons 4.0]: https://git.k8s.io/website/LICENSE
