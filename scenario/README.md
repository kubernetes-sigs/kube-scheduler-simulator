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

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/scenario:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/scenario:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/scenario:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/scenario/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)