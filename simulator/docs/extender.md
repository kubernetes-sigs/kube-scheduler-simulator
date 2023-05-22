## Extender

This document describes how to use your `Extenders` in the scheduler running in the simulator.
The `Extender` is one of the features of the scheduler's webhook on kubernetes.

The simulator stores the results of each Extender in the annotation of a pod.

(Note that it's not related to the [`plugin-extender`](./plugin-extender.md) which is one of the our simulator's feature. 
(Sorry for the confusing name 😅))

Note: This function uses api-server on the simulator to store the results.
Therefore, **if the scheduler is not connected to the simulator's api-server,
this feature is not available.**

## How to use

You need to configure your extender in KubeSchedulerConfig.
(via [the simulator config](./simulator-server-config.md) or WebUI)

(No required special configuration is for the simulator to use this feature.)

For example, if you run the server on `http://localhost:8080/scheduler/`,
the configuration will look like this.

```yaml
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
  - schedulerName: default-scheduler
extenders:
  - urlPrefix: "http://localhost:8080/scheduler/"
    filterVerb: "predicates/always_true"
    prioritizeVerb: "priorities/zero_score"
    preemptVerb: "preemption"
    bindVerb: ""
    weight: 1
    enableHTTPS: false
    nodeCacheCapable: false
```

After the above settings are made, when the simulator is started and the pod is scheduled, 
you will see each Pod gets many results on the annotation like this:

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: pod-2rsvz
...
  annotations:
    scheduler-simulator/bind-result: '{"DefaultBinder":"success"}'
    scheduler-simulator/extender-bind-result: '{}'
    scheduler-simulator/extender-filter-result: >-
      {"http://localhost:8080/scheduler/":{"Nodes":{"metadata":{},"items":[{"metadata":{"name":"node-sc9ns","generateName":"node-","uid":"4b008c90-971e-4816-a0f4-dc1a3b6e856e","resourceVersion":"208","creationTimestamp":"2023-03-03T16:03:50Z","managedFields":[{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2023-03-03T16:03:50Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:generateName":{}}}}]},"spec":{},"status":{"capacity":{"cpu":"4","memory":"32Gi","pods":"110"},"allocatable":{"cpu":"4","memory":"32Gi","pods":"110"},"phase":"Running","conditions":[{"type":"Ready","status":"True","lastHeartbeatTime":null,"lastTransitionTime":null}],"daemonEndpoints":{"kubeletEndpoint":{"Port":0}},"nodeInfo":{"machineID":"","systemUUID":"","bootID":"","kernelVersion":"","osImage":"","containerRuntimeVersion":"","kubeletVersion":"","kubeProxyVersion":"","operatingSystem":"","architecture":""}}},{"metadata":{"name":"node-pwzdq","generateName":"node-","uid":"b24f918d-94ae-4c35-9e2c-2376998dbede","resourceVersion":"209","creationTimestamp":"2023-03-03T16:03:53Z","managedFields":[{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2023-03-03T16:03:53Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:generateName":{}}}}]},"spec":{},"status":{"capacity":{"cpu":"4","memory":"32Gi","pods":"110"},"allocatable":{"cpu":"4","memory":"32Gi","pods":"110"},"phase":"Running","conditions":[{"type":"Ready","status":"True","lastHeartbeatTime":null,"lastTransitionTime":null}],"daemonEndpoints":{"kubeletEndpoint":{"Port":0}},"nodeInfo":{"machineID":"","systemUUID":"","bootID":"","kernelVersion":"","osImage":"","containerRuntimeVersion":"","kubeletVersion":"","kubeProxyVersion":"","operatingSystem":"","architecture":""}}}]},"NodeNames":null,"FailedNodes":{},"FailedAndUnresolvableNodes":null,"Error":""}}
    scheduler-simulator/extender-preempt-result: '{}'
    scheduler-simulator/extender-prioritize-result: >-
      {"http://localhost:8080/scheduler/":[{"Host":"node-sc9ns","Score":0},{"Host":"node-pwzdq","Score":0}]}
    scheduler-simulator/score-result: >-
      {"node-282x7":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"52","NodeResourcesFit":"47","PodTopologySpread":"0","TaintToleration":"0","VolumeBinding":"0"},"node-gp9t4":{"ImageLocality":"0","InterPodAffinity":"0","NodeAffinity":"0","NodeNumber":"0","NodeResourcesBalancedAllocation":"76","NodeResourcesFit":"73","PodTopologySpread":"0","TaintToleration":"0","VolumeBinding":"0"}}
...
```


