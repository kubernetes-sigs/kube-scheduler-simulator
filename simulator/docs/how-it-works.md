# How the simulator works

This page describes how this kube-scheduler-simulator works.

## 0. starts the simulator.

The simulator server works with the following:
- [kube-apiserver (+ etcd)](kube-apiserver.md)
- scheduler
- pv controller
- [HTTP server](api.md) 

When the simulator server starts, it will start these components with server.

## 1. users request creating resource.

Users can create resources below by communicating with kube-apiserver in simulator via any clients (e.g. kubectl, k8s client library or Web UI)

- Nodes
- Pods
- Persistent Volumes
- Persistent Volume Claims
- Storage Classes
- Priority Classes

## 2. the scheduler schedules a new pod.

When a new pod is created through kube-apiserver, the scheduler starts scheduling.

## 3. the results of score/filter plugins are recorded.

Normally, when score/filter plugins are called from scheduler, they will calculate the results and return results to the scheduler.
But, in the simulator, custom plugins, that behave as score/filter plugin but records result after calculation, are used in scheduler.

## 4. the scheduler binds the pod to a node.

The scheduler finally binds the pod to a node if succeeded, or move the pod back to queue if failed.

The result store will notice that the pod has been scheduled/marked as unscheduled by the scheduler and add the scheduling results to the pod's annotation.