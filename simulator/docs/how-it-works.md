## How the simulator works

This page describes how the simulator works.

### 0. starts the simulator.

The simulator server works with the following:
- [kube-apiserver](kube-apiserver.md)
- etcd
- scheduler
- controllers for core resources
- [HTTP server](api.md) 

In advance, the simulator needs to launch etcd, controllers and kube-apiserver outside.
[KWOK](https://github.com/kubernetes-sigs/kwok) can launch these components all at once, thus we recommend using it.
When the simulator server starts, it will start scheduler and HTTP server.

### 1. users request creating resource.

Users can create resources below by communicating with kube-apiserver in simulator via any clients (e.g. kubectl, k8s client library or Web UI)

- Nodes
- Pods
- Persistent Volumes
- Persistent Volume Claims
- Storage Classes
- Priority Classes
- Namespaces

### 2. the scheduler schedules a new pod.

When a new pod is created through kube-apiserver, the scheduler starts scheduling.

### 3. the results of score/filter plugins are recorded.

Normally, when score/filter plugins are called from scheduler, they will calculate the results and return results to the scheduler.
But, in the simulator, custom plugins, that behave as score/filter plugin but records result after calculation, are used in scheduler.

### 4. the scheduler binds the pod to a node.

The scheduler finally binds the pod to a node if succeeded, or move the pod back to queue if failed.

The result store will notice that the pod has been scheduled/marked as unscheduled by the scheduler and add the scheduling results to the pod's annotation.
