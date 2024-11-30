package debuggablescheduler

import (
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"

	simulatorschedulerconfig "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/extender"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin"
)

func NewSchedulerCommand(opts ...Option) (*cobra.Command, func(), error) {
	opt := &options{pluginExtender: map[string]plugin.PluginExtenderInitializer{}, outOfTreeRegistry: map[string]runtime.PluginFactory{}}
	for _, o := range opts {
		o(opt)
	}

	if opt.outOfTreeRegistry != nil {
		simulatorschedulerconfig.SetOutOfTreeRegistries(opt.outOfTreeRegistry)
	}

	configs, err := NewConfigs()
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to NewConfigs(): %w", err)
	}

	// Extender service must be initialized using `KubeSchedulerConfiguration.Extenders` config which is not override for simulator (before calling OverrideExtendersCfgToSimulator()).
	// The override will be do within CreateOptions().
	extenderService, err := extender.New(configs.clientSet, configs.versioned.Extenders, configs.sharedStore)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to New Extender service: %w", err)
	}

	schedulerOpts, cancelFn, err := CreateOptions(configs, opt.pluginExtender)
	if err != nil {
		return nil, cancelFn, err
	}
	// Launch the proxy HTTP server for Extender, which is used to store the Extender's results.
	s := NewExtenderServer(extenderService)
	shutdownFn, err := s.Start(configs.port)
	if err != nil {
		return nil, nil, xerrors.Errorf("start extender proxy server: %w", err)
	}

	cancel := func() {
		cancelFn()
		shutdownFn()
	}
	command := app.NewSchedulerCommand(schedulerOpts...)

	return command, cancel, nil
}

type options struct {
	outOfTreeRegistry runtime.Registry
	pluginExtender    map[string]plugin.PluginExtenderInitializer
}

type Option func(opt *options)

// WithPlugin creates an Option based on plugin name and factory.
func WithPlugin(pluginName string, factory runtime.PluginFactory) Option {
	return func(opt *options) {
		opt.outOfTreeRegistry[pluginName] = factory
	}
}

// WithPluginExtenders creates an Option based on plugin name and plugin extenders.
func WithPluginExtenders(pluginName string, e plugin.PluginExtenderInitializer) Option {
	return func(opt *options) {
		opt.pluginExtender[pluginName] = e
	}
}
