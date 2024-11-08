## Extender

This document describes how to use your `Extenders` in the scheduler running in the simulator.
The `Extender` is one of the features of the scheduler's webhook on kubernetes.

The simulator stores the results of each Extender in the annotation of a pod.

(Note that it's not related to the [`plugin-extender`](./plugin-extender.md) which is one of the our simulator's feature. 
(Sorry for the confusing name 😅))

Note: This feature is not available in [external scheduler](./external-scheduler.md).

## How to use
In this example, we describe how you can run an extender with the simulator, using [k8s-scheduler-extender-example](https://github.com/everpeace/k8s-scheduler-extender-example).

+ Create k8s-scheduler-extender-example's Image: Clone [k8s-scheduler-extender-example](https://github.com/everpeace/k8s-scheduler-extender-example) repository, and follow the step `1 build a docker image` on README.

+ Set up your extender in KubeSchedulerConfiguration either through [`kubeSchedulerConfigPath`](./simulator-server-config.md) or the Web UI.
For example, if you are running the server on http://kube-scheduler-simulator-extender-1:80/scheduler/, your configuration might look like the following:

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

+ Run Simulator:
We have an example [`docker-compose.yaml`](./example/docker-compose.yaml); you can overwrite the [`docker-compose-local.yaml`](../../docker-compose-local.yml) file with this file, but make sure to update the extender's image name there.

To run the simulator, use the following commands:
```sh
$ make docker_build docker_up_local
```

+ Create a Pod and examine your Extender's Results:
The simulator started with the above steps should have your extender(s) enabled. You can create Pod(s) in the simulator and see the result. 
The result shows up in the Pod's annotations `kube-scheduler-simulator.sigs.k8s.io/extender-xxx` like the following:

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: pod-2rsvz
...
annotations:
      kube-scheduler-simulator.sigs.k8s.io/extender-bind-result: '{}'
      kube-scheduler-simulator.sigs.k8s.io/extender-filter-result: '{"http://kube-scheduler-simulator-extender-1:80/scheduler":{"Nodes":{"metadata":{},"items":[{"metadata":{"name":"node-tzjll","generateName":"node-","uid":"a3e39211-2200-4dee-99a8-a27b2ac528b3","resourceVersion":"223","creationTimestamp":"2024-09-25T12:24:50Z","annotations":{"node.alpha.kubernetes.io/ttl":"0"},"managedFields":[{"manager":"kube-controller-manager","operation":"Update","apiVersion":"v1","time":"2024-09-25T12:24:50Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:node.alpha.kubernetes.io/ttl":{}}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2024-09-25T12:24:50Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:generateName":{}}}}]},"spec":{},"status":{"capacity":{"cpu":"4","memory":"32Gi","pods":"110"},"allocatable":{"cpu":"4","memory":"32Gi","pods":"110"},"phase":"Running","conditions":[{"type":"Ready","status":"True","lastHeartbeatTime":null,"lastTransitionTime":null}],"daemonEndpoints":{"kubeletEndpoint":{"Port":0}},"nodeInfo":{"machineID":"","systemUUID":"","bootID":"","kernelVersion":"","osImage":"","containerRuntimeVersion":"","kubeletVersion":"","kubeProxyVersion":"","operatingSystem":"","architecture":""}}}]},"NodeNames":null,"FailedNodes":{},"FailedAndUnresolvableNodes":null,"Error":""}}'
      kube-scheduler-simulator.sigs.k8s.io/extender-preempt-result: '{}'
      kube-scheduler-simulator.sigs.k8s.io/extender-prioritize-result: '{}'
      ....
```

You can also view the annotation results from the web UI. Simply select the Pod you created and scheduled, then check the Resource Definition section to see the annotations.

