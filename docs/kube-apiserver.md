# Kube-apiserver

This page describes how the simulator's kube-apiserver is configured.

## Pre-requisite

- Simulator user should have permission to create directory under `/var/run`.

## How it works

### Environment Variable
The kube-apiserver is configured to expose on the port `KUBE_API_PORT` and on the network interface `KUBE_API_HOST`.

If the two variables are not specified, a random port will be chosen under the localhost `127.0.0.1` .

### Server Creation

We create a kube-apiserver instance by utilising the code path in `Kubernetes/cmd/kube-apiserver`, meaning we do not have to maintain any apiserver code.

However, we will have to modify a few things to allow our kube-apiserver to be accessible and usable without too much hassles. We have modified the following options in [file](../k8sapiserver/k8sapiserver.go):

1. Etcd URL - access our Etcd instance.
2. Authorization mode - uses RBAC authorization.
3. Authentication method - to allow anonymous authentication. 
4. Secure Serving - creation of a fake key for *secure serving* - hence the need for permission to create directory under `var/run`.
5. Admission - disabling admission plugins allow us to create nodes without not ready taints, and not having to create default service accounts.