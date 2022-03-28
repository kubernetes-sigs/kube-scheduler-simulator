# How the simulator works

This page describes how this kube-scheduler-simulator works.

## 0. starts the simulator.

The simulator server works with k8s components below.
- kube-apiserver (+ etcd)
- scheduler
- pv controller

When the simulator server starts, it will start these components with server.

## 0. configure the scheduler.

Users can configure the scheduler on simulator from web-ui.

## 1. users request creating resource.

Users can create resources below.
- Nodes
- Pods
- Persistent Volumes
- Persistent Volume Claims
- Storage Classes
- Priority Classes

When users request to create resources, the simulator's frontend will create it with requesting kube-apiserver.

## 2. the scheduler schedules a new pod.

When a new pod is created through kube-apiserver, the scheduler will notice that the pod has been created and start scheduling.

## 3. the results of score/filter plugins are recorded.

When score/filter plugins called from scheduler, they will do their work, record the results and return results to the scheduler.

We create custom-plugins that behave like default score/filter plugins but records result after score/filter and use these plugins instead of default score/filter plugins, i.e. we enable these custom-plugins and disable all score/filter plugin.

## 4. the scheduler binds the pod to a node.

The scheduler finally binds the pod to a node if scheduling is succeeded, or adds some information on the pod if failed.

At that time, the result store will notice that the pod has been scheduled/marked-as-unscheduled by the scheduler and add the scheduling results to the pod's annotation.

The simulator's frontend can see the scheduling result with the annotation.
