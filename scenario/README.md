# scenario

**This Scenario is currently under development.** 

**All the following descriptions are subject to change as they are undecided.** 
**And, the implementations are subject to breaking changes without notice.**

This module contains the Scenario CRD and the controller for it.
The Scenario allows you to write scenario for scenario-based simulation of scheduler.

## Description

This Scenario and the controller are discussed and designed in [KEP-140: Scenario-based simulation](../keps/140-scenario-based-simulation).

It's designed mainly to run with simulator. 
The simulator won't create any actual resources, so you can run the simulation for your scheduler without any huge real resources.
You can see more details about simulator in [../simulator](../simulator).

The Scenario and the controller can be run in your cluster, but **we do not recommend to do it in your real cluster**.
(probably no problem on running with a fake cluster like minikube or kind)

By using it on your cluster, you can check your scheduler's behavior by directly using scheduler in your cluster.

**Several important notes for when you want to run it on your cluster** (these are why we don't recommend):
- The controller **removes all resources** in your cluster.
- The controller creates the actual resources in your cluster for simulation.
- The size of Scenario resources sometimes becomes big since it contains both defined scenario and the simulation result
  - the big resource may degrade the performance of etcd on cluster.

TODO: add the note about [the simulator operator](../keps/159-scheduler-simulator-operator) and [SchedulerSimulation](../keps/184-scheduler-simulation).

## Getting Started 

### with simulator

TODO: write it how to run it with simulator.

### with your cluster

**You must read the important notes on [Description section](#Description).**

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

#### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/scenario:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/scenario:tag
```

#### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

#### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html).
