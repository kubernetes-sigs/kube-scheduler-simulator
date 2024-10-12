package scheduler

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	configv1 "k8s.io/kube-scheduler/config/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	apiconfigv1 "k8s.io/kubernetes/pkg/scheduler/apis/config/v1"

	simulatorconfig "sigs.k8s.io/kube-scheduler-simulator/simulator/config"
	simulatorschedconfig "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
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

func restartContainer(ctx context.Context, cli *client.Client, cfg *configv1.KubeSchedulerConfiguration) error {
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return xerrors.Errorf("failed to get container list: %w", err)
	}
	for _, c := range containers {
		if c.Names[0] != "/simulator-scheduler" {
			continue
		}
		if err := simulatorschedconfig.UpdateSchedulerConfig(cfg); err != nil {
			return xerrors.Errorf("read old scheduler.yaml: %w", err)
		}

		if err := cli.ContainerRestart(ctx, c.ID, container.StopOptions{}); err != nil {
			return xerrors.Errorf("failed restart container: %w", err)
		}
		inspect, err := cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			return xerrors.Errorf("failed get container inspect: %w", err)
		}
		if inspect.State.Status != "running" {
			return xerrors.Errorf("restart container status is not running")
		}
		break
	}

	return nil
}

// RestartScheduler restarts the debuggable scheduler with a new config.
// Specifically, it updates the config file, which is also mounted on the debuggable scheduler,
// and then restart the debuggable scheduler.
func (s *Service) RestartScheduler(cfg *configv1.KubeSchedulerConfiguration) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return xerrors.Errorf("failed to create docker client: %w", err)
	}

	oldCfg, err := simulatorconfig.GetSchedulerCfg()
	if err != nil {
		return xerrors.Errorf("read old scheduler.yaml: %w", err)
	}

	if err := restartContainer(ctx, cli, cfg); err != nil {
		klog.Errorf("failed to apply new scheduler config: %w", err)
		// If failing restarting the container, we roll back to the old config.
		if err := restartContainer(ctx, cli, oldCfg); err != nil {
			return xerrors.Errorf("oldConfig restart failed: %w", err)
		}
	}
	s.currentSchedulerCfg = cfg.DeepCopy()
	return nil
}

func (s *Service) ResetScheduler() error {
	return s.RestartScheduler(s.initialSchedulerCfg.DeepCopy())
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
