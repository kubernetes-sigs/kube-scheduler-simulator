## the external scheduler

This document describes how to use the external scheduler instead of the scheduler running in the simulator.
We use the [`externalscheduler` package](../pkg/externalscheduler).

Note that the scheduler built with the [`externalscheduler` package](../pkg/externalscheduler) will work much like the normal scheduler.
But, it does some additional effort inside to export the scheduling results. So, we'd recommend using it in the dev env only.

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
        externalscheduler.WithPluginExtenders(noderesources.Name, extender.New), // see plugin-extender.md about PluginExtender.
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

### [Optional] Connect your scheduler to the kube-apiserver in the simulator

If you want to connect your scheduler in your cluster's kube-apiserver, skip this. 
(But, note that it, of course, means your Pods on your cluster will be actually scheduled by the scheduler)

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