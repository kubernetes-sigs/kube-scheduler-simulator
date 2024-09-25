## Extender

This document describes how to use your `Extenders` in the scheduler running in the simulator.
The `Extender` is one of the features of the scheduler's webhook on kubernetes.

The simulator stores the results of each Extender in the annotation of a pod.

(Note that it's not related to the [`plugin-extender`](./plugin-extender.md) which is one of the our simulator's feature. 
(Sorry for the confusing name 😅))

Note: This feature is not available in [external scheduler](./external-scheduler.md).

## How to use

+ Create k8s-scheduler-extender-example's Image in local
First, clone and run the [k8s-scheduler-extender-example](https://github.com/everpeace/k8s-scheduler-extender-example) repository from GitHub. After cloning the repository, follow the step labeled `1 build a docker image`. This will create a Docker image for local use.

+ Set Up DebuggableSechduler Extender settings
You need to configure your extender in the KubeSchedulerConfig (either through the simulator config or the Web UI). (Note: No special configuration is required to enable this feature in the simulator.)

For example, if you are running the server on http://kube-scheduler-simulator-extender-1:80/scheduler/, your configuration might look like this: (If you're using the example [`docker-compose`](./example/docker-compose.yaml), this configuration is already set up for you.)

```yaml
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
  - schedulerName: default-scheduler
extenders:
  - urlPrefix: "http://kube-scheduler-simulator-extender-1:80/scheduler"
    filterVerb: "predicates/always_true"
    prioritizeVerb: "priorities/zero_score"
    preemptVerb: "preemption"
    bindVerb: ""
    weight: 1
    enableHTTPS: false
    nodeCacheCapable: false
```

+ Run Simulator
If you are using [`docker-compose`](./example/docker-compose.yaml), overwrite the [`docker-compose-local.yaml`](../../docker-compose-local.yml) file with your local settings. Make sure to update the extender's image name in the docker-compose-local file.

To run the simulator, use the following commands:
```sh
$ make docker_build docker_up
```

+ Create a Pod and Verify the Extender's Results
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
      scheduler-simulator/bind-result: '{"DefaultBinder":"success"}'
      scheduler-simulator/extender-bind-result: '{}'
      scheduler-simulator/extender-filter-result: '{"http://kube-scheduler-simulator-extender-1:80/scheduler":{"Nodes":{"metadata":{},"items":[{"metadata":{"name":"node-tzjll","generateName":"node-","uid":"a3e39211-2200-4dee-99a8-a27b2ac528b3","resourceVersion":"223","creationTimestamp":"2024-09-25T12:24:50Z","annotations":{"node.alpha.kubernetes.io/ttl":"0"},"managedFields":[{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2024-09-25T12:24:50Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:node.alpha.kubernetes.io/ttl":{}}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2024-09-25T12:24:50Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:generateName":{}}}}]},"spec":{},"status":{"capacity":{"cpu":"4","memory":"32Gi","pods":"110"},"allocatable":{"cpu":"4","memory":"32Gi","pods":"110"},"phase":"Running","conditions":[{"type":"Ready","status":"True","lastHeartbeatTime":null,"lastTransitionTime":null}],"daemonEndpoints":{"kubeletEndpoint":{"Port":0}},"nodeInfo":{"machineID":"","systemUUID":"","bootID":"","kernelVersion":"","osImage":"","containerRuntimeVersion":"","kubeletVersion":"","kubeProxyVersion":"","operatingSystem":"","architecture":""}}}]},"NodeNames":null,"FailedNodes":{},"FailedAndUnresolvableNodes":null,"Error":""}}'
      scheduler-simulator/extender-preempt-result: '{}'
      scheduler-simulator/extender-prioritize-result: '{}'
      scheduler-simulator/filter-result: '{"node-tzjll":{"NodeName":"passed","NodeResourcesFit":"passed","NodeUnschedulable":"passed","TaintToleration":"passed"}}'
      scheduler-simulator/finalscore-result: '{}'
      scheduler-simulator/permit-result: '{}'
      scheduler-simulator/permit-result-timeout: '{}'
      scheduler-simulator/postfilter-result: '{}'
      scheduler-simulator/prebind-result: '{"VolumeBinding":"success"}'
      scheduler-simulator/prefilter-result: '{}'
      scheduler-simulator/prefilter-result-status: '{"AzureDiskLimits":"","EBSLimits":"","GCEPDLimits":"","InterPodAffinity":"","NodeAffinity":"","NodePorts":"","NodeResourcesFit":"success","NodeVolumeLimits":"","PodTopologySpread":"","VolumeBinding":"","VolumeRestrictions":"","VolumeZone":""}'
      scheduler-simulator/prescore-result: '{}'
      scheduler-simulator/reserve-result: '{"VolumeBinding":"success"}'
      scheduler-simulator/result-history: '[{"scheduler-simulator/bind-result":"{\"DefaultBinder\":\"success\"}","scheduler-simulator/extender-bind-result":"{}","scheduler-simulator/extender-filter-result":"{\"http://kube-scheduler-simulator-extender-1:80/scheduler\":{\"Nodes\":{\"metadata\":{},\"items\":[{\"metadata\":{\"name\":\"node-tzjll\",\"generateName\":\"node-\",\"uid\":\"a3e39211-2200-4dee-99a8-a27b2ac528b3\",\"resourceVersion\":\"223\",\"creationTimestamp\":\"2024-09-25T12:24:50Z\",\"annotations\":{\"node.alpha.kubernetes.io/ttl\":\"0\"},\"managedFields\":[{\"manager\":\"kube-controller-manager\",\"operation\":\"Update\",\"apiVersion\":\"v1\",\"time\":\"2024-09-25T12:24:50Z\",\"fieldsType\":\"FieldsV1\",\"fieldsV1\":{\"f:metadata\":{\"f:annotations\":{\".\":{},\"f:node.alpha.kubernetes.io/ttl\":{}}}}},{\"manager\":\"simulator\",\"operation\":\"Update\",\"apiVersion\":\"v1\",\"time\":\"2024-09-25T12:24:50Z\",\"fieldsType\":\"FieldsV1\",\"fieldsV1\":{\"f:metadata\":{\"f:generateName\":{}}}}]},\"spec\":{},\"status\":{\"capacity\":{\"cpu\":\"4\",\"memory\":\"32Gi\",\"pods\":\"110\"},\"allocatable\":{\"cpu\":\"4\",\"memory\":\"32Gi\",\"pods\":\"110\"},\"phase\":\"Running\",\"conditions\":[{\"type\":\"Ready\",\"status\":\"True\",\"lastHeartbeatTime\":null,\"lastTransitionTime\":null}],\"daemonEndpoints\":{\"kubeletEndpoint\":{\"Port\":0}},\"nodeInfo\":{\"machineID\":\"\",\"systemUUID\":\"\",\"bootID\":\"\",\"kernelVersion\":\"\",\"osImage\":\"\",\"containerRuntimeVersion\":\"\",\"kubeletVersion\":\"\",\"kubeProxyVersion\":\"\",\"operatingSystem\":\"\",\"architecture\":\"\"}}}]},\"NodeNames\":null,\"FailedNodes\":{},\"FailedAndUnresolvableNodes\":null,\"Error\":\"\"}}","scheduler-simulator/extender-preempt-result":"{}","scheduler-simulator/extender-prioritize-result":"{}","scheduler-simulator/filter-result":"{\"node-tzjll\":{\"NodeName\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"TaintToleration\":\"passed\"}}","scheduler-simulator/finalscore-result":"{}","scheduler-simulator/permit-result":"{}","scheduler-simulator/permit-result-timeout":"{}","scheduler-simulator/postfilter-result":"{}","scheduler-simulator/prebind-result":"{\"VolumeBinding\":\"success\"}","scheduler-simulator/prefilter-result":"{}","scheduler-simulator/prefilter-result-status":"{\"AzureDiskLimits\":\"\",\"EBSLimits\":\"\",\"GCEPDLimits\":\"\",\"InterPodAffinity\":\"\",\"NodeAffinity\":\"\",\"NodePorts\":\"\",\"NodeResourcesFit\":\"success\",\"NodeVolumeLimits\":\"\",\"PodTopologySpread\":\"\",\"VolumeBinding\":\"\",\"VolumeRestrictions\":\"\",\"VolumeZone\":\"\"}","scheduler-simulator/prescore-result":"{}","scheduler-simulator/reserve-result":"{\"VolumeBinding\":\"success\"}","scheduler-simulator/score-result":"{}","scheduler-simulator/selected-node":"node-tzjll"}]'
      scheduler-simulator/score-result: '{}'
      scheduler-simulator/selected-node: node-tzjll
...
```

You can also view the annotation results from the web UI. Simply select the Pod you created and scheduled, then check the Resource Definition section to see the annotations.

