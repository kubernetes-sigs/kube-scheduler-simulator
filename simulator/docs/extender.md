## Extender

This document describes how to use your `Extenders` in the scheduler running in the simulator.
The `Extender` is one of the features of the scheduler's webhook on kubernetes.

The simulator stores the results of each Extender in the annotation of a pod.

(Note that it's not related to the [`plugin-extender`](./plugin-extender.md) which is one of the our simulator's feature. 
(Sorry for the confusing name 😅))

Note: This feature is not available in [external scheduler](./external-scheduler.md).

## How to use

+ Run k8s-scheduler-extender-example in local
First, clone and run the k8s-scheduler-extender-example repository locally. This example demonstrates how to extend Kubernetes' scheduler with a custom extender.

[View the `k8s-scheduler-extender-example` on GitHub](https://github.com/everpeace/k8s-scheduler-extender-example?tab=readme-ov-file)

+ Run Kwok with the Debuggable Scheduler and Enable the Extender
Set up Kwok and configure it to run the debuggable scheduler with the custom extender enabled. Kwok is a tool for running simulated Kubernetes clusters, which allows you to test and debug schedulers efficiently.
To build and start the environment, run the following command:
```sh
$ make docker_build docker_up
```

Step 3: Create a Pod and Verify the Extender's Results
After setting up the environment, create a Pod in your Kubernetes cluster. Once the Pod is scheduled, you can verify that the extender's results are correctly reflected in the Pod's annotations.
To check the annotations, use the following command:
```sh
$ kubectl get pod <POD_NAME> -o yaml
```
```yaml
kind: Pod
apiVersion: v1
metadata:
  name: pod-2rsvz
...
  annotations:
    annotations:
    scheduler-simulator/bind-result: '{"DefaultBinder":"success"}'
    scheduler-simulator/filter-result: '{"node-cl8rv":{"NodeName":"passed","NodeResourcesFit":"passed","NodeUnschedulable":"passed","TaintToleration":"passed"},"node-thcvl":{"NodeName":"passed","NodeResourcesFit":"passed","NodeUnschedulable":"passed","TaintToleration":"passed"}}'
    scheduler-simulator/finalscore-result: '{"node-cl8rv":{"ImageLocality":"0","NodeAffinity":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"70","TaintToleration":"300","VolumeBinding":"0"},"node-thcvl":{"ImageLocality":"0","NodeAffinity":"0","NodeResourcesBalancedAllocation":"52","NodeResourcesFit":"47","TaintToleration":"300","VolumeBinding":"0"}}'
    scheduler-simulator/permit-result: '{}'
    scheduler-simulator/permit-result-timeout: '{}'
    scheduler-simulator/postfilter-result: '{}'
    scheduler-simulator/prebind-result: '{"VolumeBinding":"success"}'
    scheduler-simulator/prefilter-result: '{}'
    scheduler-simulator/prefilter-result-status: '{"AzureDiskLimits":"","EBSLimits":"","GCEPDLimits":"","InterPodAffinity":"","NodeAffinity":"","NodePorts":"","NodeResourcesFit":"success","NodeVolumeLimits":"","PodTopologySpread":"","VolumeBinding":"","VolumeRestrictions":"","VolumeZone":""}'
    scheduler-simulator/prescore-result: '{"InterPodAffinity":"","NodeAffinity":"success","NodeResourcesBalancedAllocation":"success","NodeResourcesFit":"success","PodTopologySpread":"","TaintToleration":"success"}'
    scheduler-simulator/reserve-result: '{"VolumeBinding":"success"}'
    scheduler-simulator/result-history: '[{"scheduler-simulator/bind-result":"{\"DefaultBinder\":\"Operation
      cannot be fulfilled on pods/binding \\\"pod-pvzns\\\": pod pod-pvzns is already
      assigned to node \\\"node-cl8rv\\\"\"}","scheduler-simulator/filter-result":"{\"node-cl8rv\":{\"NodeName\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"TaintToleration\":\"passed\"},\"node-thcvl\":{\"NodeName\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"TaintToleration\":\"passed\"}}","scheduler-simulator/finalscore-result":"{\"node-cl8rv\":{\"ImageLocality\":\"0\",\"NodeAffinity\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"70\",\"TaintToleration\":\"300\",\"VolumeBinding\":\"0\"},\"node-thcvl\":{\"ImageLocality\":\"0\",\"NodeAffinity\":\"0\",\"NodeResourcesBalancedAllocation\":\"52\",\"NodeResourcesFit\":\"47\",\"TaintToleration\":\"300\",\"VolumeBinding\":\"0\"}}","scheduler-simulator/permit-result":"{}","scheduler-simulator/permit-result-timeout":"{}","scheduler-simulator/postfilter-result":"{}","scheduler-simulator/prebind-result":"{\"VolumeBinding\":\"success\"}","scheduler-simulator/prefilter-result":"{}","scheduler-simulator/prefilter-result-status":"{\"AzureDiskLimits\":\"\",\"EBSLimits\":\"\",\"GCEPDLimits\":\"\",\"InterPodAffinity\":\"\",\"NodeAffinity\":\"\",\"NodePorts\":\"\",\"NodeResourcesFit\":\"success\",\"NodeVolumeLimits\":\"\",\"PodTopologySpread\":\"\",\"VolumeBinding\":\"\",\"VolumeRestrictions\":\"\",\"VolumeZone\":\"\"}","scheduler-simulator/prescore-result":"{\"InterPodAffinity\":\"\",\"NodeAffinity\":\"success\",\"NodeResourcesBalancedAllocation\":\"success\",\"NodeResourcesFit\":\"success\",\"PodTopologySpread\":\"\",\"TaintToleration\":\"success\"}","scheduler-simulator/reserve-result":"{\"VolumeBinding\":\"success\"}","scheduler-simulator/score-result":"{\"node-cl8rv\":{\"ImageLocality\":\"0\",\"NodeAffinity\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"70\",\"TaintToleration\":\"0\",\"VolumeBinding\":\"0\"},\"node-thcvl\":{\"ImageLocality\":\"0\",\"NodeAffinity\":\"0\",\"NodeResourcesBalancedAllocation\":\"52\",\"NodeResourcesFit\":\"47\",\"TaintToleration\":\"0\",\"VolumeBinding\":\"0\"}}","scheduler-simulator/selected-node":"node-cl8rv"},{"scheduler-simulator/bind-result":"{\"DefaultBinder\":\"success\"}","scheduler-simulator/filter-result":"{\"node-cl8rv\":{\"NodeName\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"TaintToleration\":\"passed\"},\"node-thcvl\":{\"NodeName\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"TaintToleration\":\"passed\"}}","scheduler-simulator/finalscore-result":"{\"node-cl8rv\":{\"ImageLocality\":\"0\",\"NodeAffinity\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"70\",\"TaintToleration\":\"300\",\"VolumeBinding\":\"0\"},\"node-thcvl\":{\"ImageLocality\":\"0\",\"NodeAffinity\":\"0\",\"NodeResourcesBalancedAllocation\":\"52\",\"NodeResourcesFit\":\"47\",\"TaintToleration\":\"300\",\"VolumeBinding\":\"0\"}}","scheduler-simulator/permit-result":"{}","scheduler-simulator/permit-result-timeout":"{}","scheduler-simulator/postfilter-result":"{}","scheduler-simulator/prebind-result":"{\"VolumeBinding\":\"success\"}","scheduler-simulator/prefilter-result":"{}","scheduler-simulator/prefilter-result-status":"{\"AzureDiskLimits\":\"\",\"EBSLimits\":\"\",\"GCEPDLimits\":\"\",\"InterPodAffinity\":\"\",\"NodeAffinity\":\"\",\"NodePorts\":\"\",\"NodeResourcesFit\":\"success\",\"NodeVolumeLimits\":\"\",\"PodTopologySpread\":\"\",\"VolumeBinding\":\"\",\"VolumeRestrictions\":\"\",\"VolumeZone\":\"\"}","scheduler-simulator/prescore-result":"{\"InterPodAffinity\":\"\",\"NodeAffinity\":\"success\",\"NodeResourcesBalancedAllocation\":\"success\",\"NodeResourcesFit\":\"success\",\"PodTopologySpread\":\"\",\"TaintToleration\":\"success\"}","scheduler-simulator/reserve-result":"{\"VolumeBinding\":\"success\"}","scheduler-simulator/score-result":"{\"node-cl8rv\":{\"ImageLocality\":\"0\",\"NodeAffinity\":\"0\",\"NodeResourcesBalancedAllocation\":\"76\",\"NodeResourcesFit\":\"70\",\"TaintToleration\":\"0\",\"VolumeBinding\":\"0\"},\"node-thcvl\":{\"ImageLocality\":\"0\",\"NodeAffinity\":\"0\",\"NodeResourcesBalancedAllocation\":\"52\",\"NodeResourcesFit\":\"47\",\"TaintToleration\":\"0\",\"VolumeBinding\":\"0\"}}","scheduler-simulator/selected-node":"node-cl8rv"}]'
    scheduler-simulator/score-result: '{"node-cl8rv":{"ImageLocality":"0","NodeAffinity":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"70","TaintToleration":"0","VolumeBinding":"0"},"node-thcvl":{"ImageLocality":"0","NodeAffinity":"0","NodeResourcesBalancedAllocation":"52","NodeResourcesFit":"47","TaintToleration":"0","VolumeBinding":"0"}}'
    scheduler-simulator/selected-node: node-cl8rv
...
```

You can also view the annotation results from the web UI. Simply select the Pod you created and scheduled, then check the Resource Definition section to see the annotations.

