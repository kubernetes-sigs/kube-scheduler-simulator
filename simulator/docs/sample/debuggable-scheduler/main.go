package main

import (
	"fmt"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/component-base/cli"
	_ "k8s.io/component-base/logs/json/register" // for JSON log format registration
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version" // for version metric registration
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/noderesources"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/docs/sample/nodenumber"
	extender "sigs.k8s.io/kube-scheduler-simulator/simulator/docs/sample/plugin-extender"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/pkg/debuggablescheduler"
)

func main() {
	command, cancelFn, err := debuggablescheduler.NewSchedulerCommand(
		debuggablescheduler.WithPlugin(nodenumber.Name, nodenumber.New),           // Register the custom scheduler plugin.
		debuggablescheduler.WithPluginExtenders(noderesources.Name, extender.New), // [Optional] Register the plugin extender. See /simulator/docs/plugin-extender.md
	)
	if err != nil {
		klog.Info(fmt.Sprintf("failed to build the scheduler command: %+v", err))
		os.Exit(1)
	}
	code := cli.Run(command)

	cancelFn()
	os.Exit(code)
}
