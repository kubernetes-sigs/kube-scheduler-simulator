## External scheduler

The external scheduler is the feature of the simulator 
to allow you to run any of your scheduler in the simulator instead of the default one.

Before diving into the description, you need to know [debuggable scheduler](./debuggable-scheduler.md).

### connect your scheduler to the simulator

Let's connect your scheduler to the simulator.

First, you need to set `externalSchedulerEnabled: true` on [the simulator config](../config.yaml)
so that the scheduler, running in the simulator by default, won't get started.

Next, you need to connect your scheduler into the simulator's kube-apiserver via KubeSchedulerConfig:

```yaml
kind: KubeSchedulerConfiguration
apiVersion: kubescheduler.config.k8s.io/v1
clientConnection:
  kubeconfig: ./path/to/kubeconfig.yaml
```

You can use this [kubeconfig.yaml](./kubeconfig.yaml) to communicate with the simulator's kube-apiserver.

### The example external scheduler

You can see how the external scheduler can be set up 
with the sample debuggable scheduler implementation in [./sample/debuggable-scheduler](./sample/debuggable-scheduler).

prerequisite:
1. set `externalSchedulerEnabled: true` on [the simulator config](../config.yaml)
2. run the simulator 

```shell
cd sample/debuggable-scheduler
go run main.go --config scheduler.yaml
```

You'll see the simulator is working with the external scheduler.