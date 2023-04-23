## External scheduler

This document describes how to use the external scheduler instead of the scheduler running in the simulator.

We use the [`externalscheduler` package](../pkg/externalscheduler); 
the scheduler built with the [`externalscheduler` package](../pkg/externalscheduler) will export the scheduling results on each Pod annotation.

### Use cases

- Running your scheduler instead of the default one in the simulator 
  - You can still see the scheduling results in web UI as well!
- Running your scheduler with the simulator feature in your cluster
  - All Pods, scheduled by this scheduler, will get the scheduler results on its annotation while each scheduling is done as usual.
  - Note that it has performance overhead in each scheduling cycle 
since the scheduler needs to make additional effort to export the scheduling results.

### Change your scheduler

Here, we assume you're registering your custom plugins in your scheduler like this:

```go
// your scheduler's main package
func main() {
	command := app.NewSchedulerCommand(
		app.WithPlugin(yourcustomplugin.Name, yourcustomplugin.New),
	)

	code := cli.Run(command)
	os.Exit(code)
}
```

Then, you need to replace few lines to use the [`externalscheduler` package](../pkg/externalscheduler).

```go
func main() {
	command, cancelFn, err := externalscheduler.NewSchedulerCommand(
        externalscheduler.WithPlugin(yourcustomplugin.Name, yourcustomplugin.New),
        externalscheduler.WithPluginExtenders(noderesources.Name, extender.New), // [optional] see plugin-extender.md about PluginExtender.
    )
    if err != nil {
        klog.Info(fmt.Sprintf("failed to build the scheduler command: %+v", err))
        os.Exit(1)
    }
    code := cli.Run(command)
    cancelFn()
    os.Exit(code)
}
```

As you see, `externalscheduler.NewSchedulerCommand` has much similar interface as the `app.NewSchedulerCommand`.
You can register your plugins by `externalscheduler.WithPlugin` option.

Via this step, all Pods scheduled by this scheduler will get the scheduling results in the annotation like in the simulator!

### Connect your scheduler to the kube-apiserver in the simulator

If you are here to run the scheduler built with [`externalscheduler` package](../pkg/externalscheduler) in your cluster, 
you don't need to follow this step.

Let's connect your scheduler into the simulator.

First, you need to set `externalSchedulerEnabled: true` on [the simulator config](../config.yaml)
so that the scheduler in the simulator won't get started.

Next, you need to connect your scheduler into the simulator's kube-apiserver via KubeSchedulerConfig:

```yaml
kind: KubeSchedulerConfiguration
apiVersion: kubescheduler.config.k8s.io/v1
clientConnection:
  kubeconfig: ./path/to/kubeconfig.yaml
```

You can use this [kubeconfig.yaml](./kubeconfig.yaml) to communicate with the simulator's kube-apiserver.

### The example external scheduler

We have the sample external scheduler implementation in [./sample/external-scheduler](./sample/external-scheduler).

prerequisite:
1. set `externalSchedulerEnabled: true` on [the simulator config](../config.yaml)
2. run the simulator 

```shell
cd sample/external-scheduler
go run main.go --config scheduler.yaml
```

You'll see the simulator is working with the external scheduler.