package scheduler

import (
	"context"
	"errors"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/profile"

	simulatorschedconfig "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/extender"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/storereflector"
)

// Service manages scheduler.
type Service struct {
	// function to shutdown scheduler.
	shutdownfn func()

	// disabled represents if this Service is disabled.
	// If externalSchedulerEnabled, it'll be true
	// because we don't need to start scheduler, and we cannot change an external scheduler's config in that case.
	disabled bool

	clientset           clientset.Interface
	restclientCfg       *restclient.Config
	initialSchedulerCfg *v1beta2config.KubeSchedulerConfiguration
	currentSchedulerCfg *v1beta2config.KubeSchedulerConfiguration
	extenderService     ExtenderService
	sharedStore         storereflector.Reflector
	simulatorPort       int
}

type ExtenderService interface {
	Filter(id int, args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
	Prioritize(id int, args extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error)
	Preempt(id int, args extenderv1.ExtenderPreemptionArgs) (*extenderv1.ExtenderPreemptionResult, error)
	Bind(id int, args extenderv1.ExtenderBindingArgs) (*extenderv1.ExtenderBindingResult, error)
}

var ErrServiceDisabled = errors.New("scheduler service is disabled")

// NewSchedulerService starts scheduler and return *Service.
func NewSchedulerService(client clientset.Interface, restclientCfg *restclient.Config, initialSchedulerCfg *v1beta2config.KubeSchedulerConfiguration, externalSchedulerEnabled bool, simulatorPort int) *Service {
	if externalSchedulerEnabled {
		return &Service{disabled: true}
	}

	// sharedStore has some resultstores which are referenced by Registry of Plugins and Extenders.
	sharedStore := storereflector.New()

	initCfg := initialSchedulerCfg.DeepCopy()
	return &Service{clientset: client, restclientCfg: restclientCfg, initialSchedulerCfg: initCfg, sharedStore: sharedStore, simulatorPort: simulatorPort}
}

func (s *Service) RestartScheduler(cfg *v1beta2config.KubeSchedulerConfiguration) error {
	if s.disabled {
		return xerrors.Errorf("an external scheduler is enabled: %w", ErrServiceDisabled)
	}

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
	return s.RestartScheduler(s.initialSchedulerCfg.DeepCopy())
}

// StartScheduler starts scheduler.
//
//nolint:funlen
func (s *Service) StartScheduler(versionedcfg *v1beta2config.KubeSchedulerConfiguration) (retErr error) {
	clientSet := s.clientset
	restConfig := s.restclientCfg
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if retErr != nil {
			cancel()
		}
	}()
	informerFactory := scheduler.NewInformerFactory(clientSet, 0)
	var dynInformerFactory dynamicinformer.DynamicSharedInformerFactory
	if restConfig != nil {
		dynClient := dynamic.NewForConfigOrDie(restConfig)
		dynInformerFactory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynClient, 0, v1.NamespaceAll, nil)
	}
	evtBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{
		Interface: clientSet.EventsV1(),
	})
	evtBroadcaster.StartRecordingToSink(ctx.Done())

	s.currentSchedulerCfg = versionedcfg.DeepCopy()

	var err error
	// Extender service must be initialized using unconverted config.
	s.extenderService, err = extender.New(clientSet, versionedcfg.Extenders, s.sharedStore)
	if err != nil {
		return xerrors.Errorf("New extender service: %w", err)
	}

	cfg, err := convertConfigurationForSimulator(versionedcfg, s.simulatorPort)
	if err != nil {
		return xerrors.Errorf("convert scheduler config to apply: %w", err)
	}
	registry, err := plugin.NewRegistry(s.sharedStore)
	if err != nil {
		return xerrors.Errorf("plugin registry: %w", err)
	}

	sched, err := scheduler.New(
		clientSet,
		informerFactory,
		dynInformerFactory,
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
		return xerrors.Errorf("create scheduler: %w", err)
	}

	informerFactory.Start(ctx.Done())
	if dynInformerFactory != nil {
		dynInformerFactory.Start(ctx.Done())
	}
	informerFactory.WaitForCacheSync(ctx.Done())
	if dynInformerFactory != nil {
		dynInformerFactory.WaitForCacheSync(ctx.Done())
	}

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

func (s *Service) GetSchedulerConfig() (*v1beta2config.KubeSchedulerConfiguration, error) {
	if s.disabled {
		return nil, xerrors.Errorf("an external scheduler is enabled: %w", ErrServiceDisabled)
	}

	return s.currentSchedulerCfg, nil
}

// ExtenderService returns ExtenderService interface.
func (s *Service) ExtenderService() ExtenderService {
	return s.extenderService
}

// convertConfigurationForSimulator convert KubeSchedulerConfiguration to apply scheduler on simulator
// (1) It excludes non-allowed changes. Now, we accept only changes to Profiles.Plugins field.
// (2) It replaces all default-plugins with plugins for simulator.
// (3) It replaces Extenders config so that the connection is directed to the simulator server.
// (4) It converts KubeSchedulerConfiguration from v1beta2config.KubeSchedulerConfiguration to config.KubeSchedulerConfiguration.
func convertConfigurationForSimulator(versioned *v1beta2config.KubeSchedulerConfiguration, simulatorPort int) (*config.KubeSchedulerConfiguration, error) {
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

	// Override the Extenders config so that the connection is directed to the simulator server.
	extender.OverrideExtendersCfgToSimulator(versioned, simulatorPort)

	defaultCfg, err := simulatorschedconfig.DefaultSchedulerConfig()
	if err != nil {
		return nil, xerrors.Errorf("get default scheduler config: %w", err)
	}

	// set default value to all field other than Profiles and Extenders.
	defaultCfg.Profiles = versioned.Profiles
	defaultCfg.Extenders = versioned.Extenders
	versioned = defaultCfg

	v1beta2.SetDefaults_KubeSchedulerConfiguration(versioned)
	cfg := config.KubeSchedulerConfiguration{}
	if err := scheme.Scheme.Convert(versioned, &cfg, nil); err != nil {
		return nil, xerrors.Errorf("convert configuration: %w", err)
	}
	cfg.SetGroupVersionKind(v1beta2config.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))

	return &cfg, nil
}
