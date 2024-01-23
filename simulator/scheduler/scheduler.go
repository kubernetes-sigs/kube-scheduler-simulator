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
	configv1 "k8s.io/kube-scheduler/config/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"k8s.io/kubernetes/pkg/scheduler"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	apiconfigv1 "k8s.io/kubernetes/pkg/scheduler/apis/config/v1"
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
	initialSchedulerCfg *configv1.KubeSchedulerConfiguration
	currentSchedulerCfg *configv1.KubeSchedulerConfiguration
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
func NewSchedulerService(client clientset.Interface, restclientCfg *restclient.Config, initialSchedulerCfg *configv1.KubeSchedulerConfiguration, externalSchedulerEnabled bool, simulatorPort int) *Service {
	if externalSchedulerEnabled {
		return &Service{disabled: true}
	}

	// sharedStore has some resultstores which are referenced by Registry of Plugins and Extenders.
	sharedStore := storereflector.New()

	initCfg := initialSchedulerCfg.DeepCopy()
	return &Service{clientset: client, restclientCfg: restclientCfg, initialSchedulerCfg: initCfg, sharedStore: sharedStore, simulatorPort: simulatorPort}
}

func (s *Service) RestartScheduler(cfg *configv1.KubeSchedulerConfiguration) error {
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
//nolint:funlen,cyclop
func (s *Service) StartScheduler(versionedcfg *configv1.KubeSchedulerConfiguration) (retErr error) {
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

	// Override the Extenders config so that the connection is directed to the simulator server.
	extender.OverrideExtendersCfgToSimulator(versionedcfg, s.simulatorPort)

	versioned, err := ConvertConfigurationForSimulator(versionedcfg)
	if err != nil {
		return xerrors.Errorf("convert scheduler config to apply: %w", err)
	}

	cfg, err := ConvertSchedulerConfigToInternalConfig(versioned)
	if err != nil {
		return xerrors.Errorf("convert scheduler config to internal one: %w", err)
	}

	cfg, err = filterOutNonAllowedChangesOnCfg(cfg)
	if err != nil {
		return xerrors.Errorf("filter out non allowed changes: %w", err)
	}

	registry, err := plugin.NewRegistry(s.sharedStore, cfg, nil)
	if err != nil {
		return xerrors.Errorf("plugin registry: %w", err)
	}

	if s.sharedStore != nil {
		// Resister the event handler function to store the result stored in the sharedStore in pod.
		if err := s.sharedStore.ResisterResultSavingToInformer(clientSet, ctx.Done()); err != nil {
			return xerrors.Errorf("ResisterResultSavingToInformer of sharedStore: %w", err)
		}
	}

	sched, err := scheduler.New(
		ctx,
		clientSet,
		informerFactory,
		dynInformerFactory,
		profile.NewRecorderFactory(evtBroadcaster),
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

func (s *Service) GetSchedulerConfig() (*configv1.KubeSchedulerConfiguration, error) {
	if s.disabled {
		return nil, xerrors.Errorf("an external scheduler is enabled: %w", ErrServiceDisabled)
	}

	return s.currentSchedulerCfg, nil
}

// ExtenderService returns ExtenderService interface.
func (s *Service) ExtenderService() ExtenderService {
	return s.extenderService
}

// ConvertConfigurationForSimulator convert KubeSchedulerConfiguration to apply scheduler on simulator
// (1) It replaces all default-plugins with plugins for simulator.
// (2) It replaces Extenders config so that the connection is directed to the simulator server.
// (3) It converts KubeSchedulerConfiguration from configv1.KubeSchedulerConfiguration to config.KubeSchedulerConfiguration.
func ConvertConfigurationForSimulator(versioned *configv1.KubeSchedulerConfiguration) (*configv1.KubeSchedulerConfiguration, error) {
	if len(versioned.Profiles) == 0 {
		defaultSchedulerName := v1.DefaultSchedulerName
		versioned.Profiles = []configv1.KubeSchedulerProfile{
			{
				SchedulerName: &defaultSchedulerName,
				Plugins:       &configv1.Plugins{},
			},
		}
	}

	for i := range versioned.Profiles {
		if versioned.Profiles[i].Plugins == nil {
			versioned.Profiles[i].Plugins = &configv1.Plugins{}
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

	apiconfigv1.SetDefaults_KubeSchedulerConfiguration(versioned)

	return versioned, nil
}

func ConvertSchedulerConfigToInternalConfig(versioned *configv1.KubeSchedulerConfiguration) (*config.KubeSchedulerConfiguration, error) {
	cfg := config.KubeSchedulerConfiguration{}
	if err := scheme.Scheme.Convert(versioned, &cfg, nil); err != nil {
		return nil, xerrors.Errorf("convert configuration: %w", err)
	}
	cfg.SetGroupVersionKind(configv1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))

	return &cfg, nil
}

// filterOutNonAllowedChangesOnCfg excludes non-allowed changes.
// Now, we accept only changes to Profiles.Plugins and Extenders fields.
func filterOutNonAllowedChangesOnCfg(originalCfg *config.KubeSchedulerConfiguration) (*config.KubeSchedulerConfiguration, error) {
	defaultCfg, err := simulatorschedconfig.DefaultSchedulerConfig()
	if err != nil {
		return nil, xerrors.Errorf("get default scheduler config: %w", err)
	}
	apiconfigv1.SetDefaults_KubeSchedulerConfiguration(defaultCfg)
	defaultconvertedcfg := config.KubeSchedulerConfiguration{}
	if err := scheme.Scheme.Convert(defaultCfg, &defaultconvertedcfg, nil); err != nil {
		return nil, xerrors.Errorf("convert configuration: %w", err)
	}
	originalCfg.SetGroupVersionKind(configv1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))

	// set default value to all field other than Profiles and Extenders.
	defaultconvertedcfg.Profiles = originalCfg.Profiles
	defaultconvertedcfg.Extenders = originalCfg.Extenders

	return &defaultconvertedcfg, nil
}
