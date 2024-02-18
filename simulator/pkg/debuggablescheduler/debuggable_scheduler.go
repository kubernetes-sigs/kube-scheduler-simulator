package debuggablescheduler

import (
	"context"
	"flag"
	"os"

	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	componentbaseconfig "k8s.io/component-base/config"
	_ "k8s.io/component-base/logs/json/register" // for JSON log format registration
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version" // for version metric registration
	"k8s.io/klog/v2"
	v1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
	configv1 "k8s.io/kubernetes/pkg/scheduler/apis/config/v1"
	configv1beta3 "k8s.io/kubernetes/pkg/scheduler/apis/config/v1beta3"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"

	simulatorconfig "sigs.k8s.io/kube-scheduler-simulator/simulator/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	simulatorschedulerconfig "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/storereflector"
)

// CreateOptionForOutOfTreePlugin creates the option which can be help with running the external scheduler.
// It does:
// - create the wrapped plugin registries and return the registries as app.Option
// - initialize and start the store reflector.
// - the scheduler config conversion
//   - reads the scheduling config passed from users (or use the default config)
//   - converts it for enabling wrapped plugins
//   - makes the defaulting func of the KubeSchedulerConfig always returning the converted one. We can let the scheduler use the converted configuration under any circumstances because the scheduler will always use this defaulting func to load the configuration.
//
//nolint:funlen,cyclop
func CreateOptionForOutOfTreePlugin(outOfTreePluginRegistry runtime.Registry, pluginExtender map[string]plugin.PluginExtenderInitializer, schedulerConfigPath *string) ([]app.Option, func(), error) {
	if outOfTreePluginRegistry != nil {
		simulatorschedulerconfig.SetOutOfTreeRegistries(outOfTreePluginRegistry)
	}

	// flags defined in the upstream scheduler
	configFile := flag.String("config", "", "")
	master := flag.String("master", "", "")
	flag.Parse()

	var versionedcfg *v1.KubeSchedulerConfiguration
	var err error
	if *configFile == "" {
		if schedulerConfigPath != nil {
			versionedcfg, err = loadConfigFromFile(*schedulerConfigPath)
			if err != nil {
				return nil, nil, xerrors.Errorf("load scheduler config: %w", err)
			}
		} else {
			versionedcfg, err = simulatorschedulerconfig.DefaultSchedulerConfig()
			if err != nil {
				return nil, nil, xerrors.Errorf("get default scheduler config: %w", err)
			}
		}
	} else {
		versionedcfg, err = loadConfigFromFile(*configFile)
		if err != nil {
			return nil, nil, xerrors.Errorf("load scheduler config: %w", err)
		}
	}

	versioned, err := scheduler.ConvertConfigurationForSimulator(versionedcfg)
	if err != nil {
		return nil, nil, xerrors.Errorf("convert scheduler config to apply: %w", err)
	}

	internalCfg, err := scheduler.ConvertSchedulerConfigToInternalConfig(versioned)
	if err != nil {
		return nil, nil, xerrors.Errorf("convert scheduler config to internal one: %w", err)
	}

	sharedStore := storereflector.New()

	registry, err := plugin.NewRegistry(sharedStore, internalCfg, pluginExtender)
	if err != nil {
		return nil, nil, xerrors.Errorf("convert scheduler config to apply: %w", err)
	}
	kubeconfig, err := simulatorconfig.GetKubeClientConfig()
	if err != nil {
		return nil, nil, xerrors.Errorf("get kubeconfig: %w", err)
	}
	if internalCfg.ClientConnection.Kubeconfig != "" {
		kubeconfig, err = createKubeConfig(internalCfg.ClientConnection, *master)
		if err != nil {
			return nil, nil, xerrors.Errorf("get kubeconfig specified in config: %w", err)
		}
	}
	clientSet, err := clientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, nil, xerrors.Errorf("creates a new Clientset for kubeconfig: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := sharedStore.ResisterResultSavingToInformer(clientSet, ctx.Done()); err != nil {
		return nil, cancel, xerrors.Errorf("ResisterResultSavingToInformer of sharedStore: %w", err)
	}

	// black magic: We need to use the scheduler config converted for the simulator in the external scheduler.
	// Here, we overwrite the defaulting func for KubeSchedulerConfiguration,
	// so that user's config will be replaced with the one we created here
	// when the scheduler loads the scheduler config
	// or when loading the default scheduler config.
	scheme.Scheme.AddTypeDefaultingFunc(&v1.KubeSchedulerConfiguration{}, func(obj interface{}) {
		c, ok := obj.(*v1.KubeSchedulerConfiguration)
		if !ok {
			panic("unexpected type")
		}
		configv1.SetObjectDefaults_KubeSchedulerConfiguration(c)
		c.Profiles = versioned.Profiles
	})

	return generateWithPluginOptions(registry), cancel, nil
}

// createKubeConfig creates a kubeConfig from the given config and masterOverride.
func createKubeConfig(config componentbaseconfig.ClientConnectionConfiguration, masterOverride string) (*restclient.Config, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags(masterOverride, config.Kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeConfig.DisableCompression = true
	kubeConfig.AcceptContentTypes = config.AcceptContentTypes
	kubeConfig.ContentType = config.ContentType
	kubeConfig.QPS = config.QPS
	kubeConfig.Burst = int(config.Burst)

	return kubeConfig, nil
}

func generateWithPluginOptions(registry map[string]runtime.PluginFactory) []app.Option {
	opt := make([]app.Option, 0, len(registry))
	for k, r := range registry {
		opt = append(opt, app.WithPlugin(k, r))
	}
	return opt
}

func loadConfigFromFile(file string) (*v1.KubeSchedulerConfiguration, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return loadConfig(data)
}

func loadConfig(data []byte) (*v1.KubeSchedulerConfiguration, error) {
	// The UniversalDecoder runs defaulting and returns the internal type by default.
	obj, gvk, err := scheme.Codecs.UniversalDecoder().Decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	if cfgObj, ok := obj.(*config.KubeSchedulerConfiguration); ok {
		// We don't set this field in pkg/scheduler/apis/config/{version}/conversion.go
		// because the field will be cleared later by API machinery during
		// conversion. See KubeSchedulerConfiguration internal type definition for
		// more details.
		cfgObj.TypeMeta.APIVersion = gvk.GroupVersion().String()
		if cfgObj.TypeMeta.APIVersion == configv1beta3.SchemeGroupVersion.String() {
			klog.InfoS("KubeSchedulerConfiguration v1beta3 is deprecated in v1.26, will be removed in v1.29")
		}

		return convertSchedulerConfigToV1Config(cfgObj)
	}
	return nil, xerrors.Errorf("couldn't decode as KubeSchedulerConfiguration, got %s", gvk)
}

func convertSchedulerConfigToV1Config(versioned *config.KubeSchedulerConfiguration) (*v1.KubeSchedulerConfiguration, error) {
	cfg := v1.KubeSchedulerConfiguration{}
	if err := scheme.Scheme.Convert(versioned, &cfg, nil); err != nil {
		return nil, xerrors.Errorf("convert configuration: %w", err)
	}

	return &cfg, nil
}
