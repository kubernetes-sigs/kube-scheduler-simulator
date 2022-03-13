package scheduler

import (
	"context"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/profile"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/defaultconfig"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/plugin"
)

// Service manages scheduler.
type Service struct {
	// function to shutdown scheduler.
	shutdownfn func()

	clientset           clientset.Interface
	restclientCfg       *restclient.Config
	initialSchedulerCfg *v1beta2config.KubeSchedulerConfiguration
	currentSchedulerCfg *v1beta2config.KubeSchedulerConfiguration
}

// NewSchedulerService starts scheduler and return *Service.
func NewSchedulerService(client clientset.Interface, restclientCfg *restclient.Config, initialSchedulerCfg *v1beta2config.KubeSchedulerConfiguration) *Service {
	initCfg := initialSchedulerCfg.DeepCopy()
	return &Service{clientset: client, restclientCfg: restclientCfg, initialSchedulerCfg: initCfg}
}

func (s *Service) RestartScheduler(cfg *v1beta2config.KubeSchedulerConfiguration) error {
	s.ShutdownScheduler()

	oldSchedulerCfg := s.currentSchedulerCfg
	if err := s.StartScheduler(cfg); err != nil {
		klog.Infof("failed to start scheduler: %v. restarting with old configuration", err)
		if err2 := s.StartScheduler(oldSchedulerCfg); err2 != nil {
			klog.Warningf("failed to start scheduler with old configuration: %v", err2)
			return xerrors.Errorf("start scheduler: %w, restart scheduler with old configuration: %w", err, err2)
		}
		return xerrors.Errorf("start scheduler: %w", err)
	}
	return nil
}

func (s *Service) ResetScheduler() error {
	return s.RestartScheduler(s.initialSchedulerCfg)
}

// StartScheduler starts scheduler.
func (s *Service) StartScheduler(versionedcfg *v1beta2config.KubeSchedulerConfiguration) error {
	clientSet := s.clientset
	restConfig := s.restclientCfg
	ctx, cancel := context.WithCancel(context.Background())

	informerFactory := scheduler.NewInformerFactory(clientSet, 0)
	evtBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{
		Interface: clientSet.EventsV1(),
	})

	evtBroadcaster.StartRecordingToSink(ctx.Done())

	s.currentSchedulerCfg = versionedcfg.DeepCopy()

	cfg, err := convertConfigurationForSimulator(versionedcfg)
	if err != nil {
		cancel()
		return xerrors.Errorf("convert scheduler config to apply: %w", err)
	}

	registry, err := plugin.NewRegistry(informerFactory, clientSet)
	if err != nil {
		cancel()
		return xerrors.Errorf("plugin registry: %w", err)
	}

	sched, err := scheduler.New(
		clientSet,
		informerFactory,
		profile.NewRecorderFactory(evtBroadcaster),
		ctx.Done(),
		scheduler.WithKubeConfig(restConfig),
		scheduler.WithProfiles(cfg.Profiles...),
		scheduler.WithPercentageOfNodesToScore(cfg.PercentageOfNodesToScore),
		scheduler.WithPodMaxBackoffSeconds(cfg.PodMaxBackoffSeconds),
		scheduler.WithPodInitialBackoffSeconds(cfg.PodInitialBackoffSeconds),
		scheduler.WithExtenders(cfg.Extenders...),
		scheduler.WithParallelism(cfg.Parallelism),
		scheduler.WithFrameworkOutOfTreeRegistry(registry),
	)
	if err != nil {
		cancel()
		return xerrors.Errorf("create scheduler: %w", err)
	}

	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	go sched.Run(ctx)

	s.shutdownfn = cancel

	return nil
}

func (s *Service) ShutdownScheduler() {
	if s.shutdownfn != nil {
		klog.Info("shutdown scheduler...")
		s.shutdownfn()
	}
}

func (s *Service) GetSchedulerConfig() *v1beta2config.KubeSchedulerConfiguration {
	return s.currentSchedulerCfg
}

// convertConfigurationForSimulator convert KubeSchedulerConfiguration to apply scheduler on simulator
// (1) It excludes non-allowed changes. Now, we accept only changes to Profiles.Plugins field.
// (2) It replaces filter/score default-plugins with plugins for simulator.
// (3) It convert KubeSchedulerConfiguration from v1beta2config.KubeSchedulerConfiguration to config.KubeSchedulerConfiguration.
func convertConfigurationForSimulator(versioned *v1beta2config.KubeSchedulerConfiguration) (*config.KubeSchedulerConfiguration, error) {
	if len(versioned.Profiles) == 0 {
		defaultSchedulerName := v1.DefaultSchedulerName
		versioned.Profiles = []v1beta2config.KubeSchedulerProfile{
			{
				SchedulerName: &defaultSchedulerName,
				Plugins:       &v1beta2config.Plugins{},
			},
		}
	}

	for i := range versioned.Profiles {
		if versioned.Profiles[i].Plugins == nil {
			versioned.Profiles[i].Plugins = &v1beta2config.Plugins{}
		}

		plugins, err := plugin.ConvertForSimulator(versioned.Profiles[i].Plugins)
		if err != nil {
			return nil, xerrors.Errorf("convert plugins for simulator: %w", err)
		}
		versioned.Profiles[i].Plugins = plugins

		pluginConfigForSimulatorPlugins, err := plugin.NewPluginConfig(versioned.Profiles[i].PluginConfig)
		if err != nil {
			return nil, xerrors.Errorf("get plugin configs: %w", err)
		}
		versioned.Profiles[i].PluginConfig = pluginConfigForSimulatorPlugins
	}

	defaultCfg, err := defaultconfig.DefaultSchedulerConfig()
	if err != nil {
		return nil, xerrors.Errorf("get default scheduler config: %w", err)
	}

	// set default value to all field other than Profiles.
	defaultCfg.Profiles = versioned.Profiles
	versioned = defaultCfg

	v1beta2.SetDefaults_KubeSchedulerConfiguration(versioned)
	cfg := config.KubeSchedulerConfiguration{}
	if err := scheme.Scheme.Convert(versioned, &cfg, nil); err != nil {
		return nil, xerrors.Errorf("convert configuration: %w", err)
	}
	cfg.SetGroupVersionKind(v1beta2config.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))

	return &cfg, nil
}
