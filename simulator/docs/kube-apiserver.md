## Kube-apiserver

> [!WARNING]
> To reduce the maintenance burden, the simulator no longer run kube-apiserver internally 
> and we made the simulator require the kube-apiserver outside.
> The simulator needs to launch the kube-apiserver outside.
> We highly recommend using the [KWOK](https://github.com/kubernetes-sigs/kwok).

This page describes about kube-apiserver run in simulator.

### How to communicate with this kube-apiserver

You can use any ways like kubectl, k8s client library, or our Web UI.

The endpoint is "http://localhost:3131" by default. (can be configured by env described in the below section.)

#### kubeconfig

You can use this kubeconfig to communicate with kube-apiserver in the simulator. 

[kubeconfig.yaml](./kubeconfig.yaml)

#### kubectl

You can use the `--server` option. 

```sh
kubectl get pods --server=localhost:3131
```

Of course, you can also use the above kubeconfig.

### How it is configured

#### Environment Variable
The kube-apiserver is configured to expose on the port `KUBE_API_PORT` and on the network interface `KUBE_API_HOST`.

If the two variables are not specified, port `3131` will be used with the localhost `127.0.0.1` address.

#### Server Creation

We create a kube-apiserver instance by utilising the code path in `Kubernetes/cmd/kube-apiserver`, meaning we do not have to maintain any apiserver code.

However, we will have to modify a few things to allow our kube-apiserver to be accessible and usable without too much hassles. We have modified the following options in [file](../k8sapiserver/k8sapiserver.go):

1. Etcd URL - access our Etcd instance.
2. Authorization mode - uses RBAC authorization.
3. Authentication method - to allow anonymous authentication. 
4. Secure Serving - creation of a temporary fake key for *secure serving* and saving the key in a temporary directory.
5. Admission - disabling admission plugins allow us to create nodes without not ready taints, and not having to create default service accounts.
