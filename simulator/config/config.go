package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	configv1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/config/v1alpha1"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
)

// ErrEmptyConfig represents the required config variable don't exist.
var ErrEmptyConfig = errors.New("config is required, but empty")

// configYaml represents the value from the config file.
var configYaml = &v1alpha1.SimulatorConfiguration{}

// Config is configuration for simulator.
type Config struct {
	Port                  int
	KubeAPIServerURL      string
	EtcdURL               string
	CorsAllowedOriginList []string
	// ExternalImportEnabled indicates whether the simulator will import resources from an target cluster or not.
	ExternalImportEnabled bool
	// ExternalKubeClientCfg is KubeConfig to get resources from external cluster.
	// This field is non-empty only when ExternalImportEnabled == true.
	ExternalKubeClientCfg *rest.Config
	InitialSchedulerCfg   *configv1.KubeSchedulerConfiguration
	// ExternalSchedulerEnabled indicates whether an external scheduler is enabled.
	ExternalSchedulerEnabled bool
}

const (
	// defaultFilePath is the config file path.
	// TODO: move it to somewhere configurable from outside if users want.
	defaultFilePath = "./config.yaml"
)

// NewConfig gets some settings from config file or environment variables.
func NewConfig() (*Config, error) {
	if err := LoadYamlConfig(defaultFilePath); err != nil {
		return nil, err
	}

	port, err := getPort()
	if err != nil {
		return nil, xerrors.Errorf("get port: %w", err)
	}

	etcdurl, err := getEtcdURL()
	if err != nil {
		return nil, xerrors.Errorf("get etcd URL: %w", err)
	}

	corsAllowedOriginList, err := getCorsAllowedOriginList()
	if err != nil {
		return nil, xerrors.Errorf("get frontend URL: %w", err)
	}

	apiurl := getKubeAPIServerURL()

	externalimportenabled := getExternalImportEnabled()
	externalKubeClientCfg := &rest.Config{}
	if externalimportenabled {
		externalKubeClientCfg, err = GetKubeClientConfig()
		if err != nil {
			return nil, xerrors.Errorf("get kube clientconfig: %w", err)
		}
	}

	initialschedulerCfg, err := getSchedulerCfg()
	if err != nil {
		return nil, xerrors.Errorf("get SchedulerCfg: %w", err)
	}

	externalSchedEnabled := getExternalSchedulerEnabled()

	return &Config{
		Port:                     port,
		KubeAPIServerURL:         apiurl,
		EtcdURL:                  etcdurl,
		CorsAllowedOriginList:    corsAllowedOriginList,
		InitialSchedulerCfg:      initialschedulerCfg,
		ExternalImportEnabled:    externalimportenabled,
		ExternalKubeClientCfg:    externalKubeClientCfg,
		ExternalSchedulerEnabled: externalSchedEnabled,
	}, nil
}

// LoadYamlConfig read the yaml file and set configYaml.
func LoadYamlConfig(configFile string) error {
	if configFile == "" {
		klog.V(1).InfoS("Config file not specified")
		return nil
	}

	conf, err := os.ReadFile(configFile)
	if err != nil {
		return xerrors.Errorf("failed to read config file: %w", err)
	}

	versionedConfig := &v1alpha1.SimulatorConfiguration{}

	decoder := scheme.Codecs.UniversalDecoder(v1alpha1.SchemeGroupVersion)
	if err := runtime.DecodeInto(decoder, conf, versionedConfig); err != nil {
		return xerrors.Errorf("failed decoding simulator's config %w", err)
	}

	configYaml = versionedConfig

	return nil
}

// getPort gets port from environment variable named PORT first, if empty from the config file.
func getPort() (int, error) {
	p := os.Getenv("PORT")
	port, err := strconv.Atoi(p)
	if p == "" || err != nil {
		port = configYaml.Port
		if port == 0 {
			return 0, xerrors.Errorf("get PORT from config: %w", ErrEmptyConfig)
		}
	}
	return port, nil
}

// getKubeAPIServerURL gets KubeAPIServerURL from environment variable first, if empty from the config file.
func getKubeAPIServerURL() string {
	url := os.Getenv("KUBE_APISERVER_URL")
	if url == "" && configYaml.KubeAPIServerURL != "" {
		return configYaml.KubeAPIServerURL
	}
	p := os.Getenv("KUBE_API_PORT")
	if p == "" {
		p = strconv.Itoa(configYaml.KubeAPIPort)
		if p == "" {
			p = "3131"
		}
	}
	h := os.Getenv("KUBE_API_HOST")
	if h == "" {
		h = configYaml.KubeAPIHost
		if h == "" {
			h = "127.0.0.1"
		}
	}

	return fmt.Sprintf("%s:%s", h, p)
}

// getExternalSchedulerEnabled gets ExternalSchedulerEnabled from environment variable first,
// if empty from the config file.
func getExternalSchedulerEnabled() bool {
	e := os.Getenv("EXTERNAL_SCHEDULER_ENABLED")
	b, err := strconv.ParseBool(e)
	if e == "" || err != nil {
		b = configYaml.ExternalSchedulerEnabled
	}
	return b
}

// getEtcdURL gets EtcdURL from environment variable first,
// if empty from the config file.
func getEtcdURL() (string, error) {
	e := os.Getenv("KUBE_SCHEDULER_SIMULATOR_ETCD_URL")
	if e == "" {
		e = configYaml.EtcdURL
		if e == "" {
			return "", xerrors.Errorf("get KUBE_SCHEDULER_SIMULATOR_ETCD_URL from config: %w", ErrEmptyConfig)
		}
	}
	return e, nil
}

// getCorsAllowedOriginList fetches CorsAllowedOriginList from the env named CORS_ALLOWED_ORIGIN_LIST
// if empty from the config file.
// This allowed list is applied to kube-apiserver and the simulator server.
//
// Let's say CORS_ALLOWED_ORIGIN_LIST="http://localhost:3000, http://localhost:3001, http://localhost:3002" are given.
// Then, getCorsAllowedOriginList returns []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:3002"}
func getCorsAllowedOriginList() ([]string, error) {
	e := os.Getenv("CORS_ALLOWED_ORIGIN_LIST")
	urls := parseStringListEnv(e)

	if err := validateURLs(urls); e == "" || err != nil {
		urls = configYaml.CorsAllowedOriginList
	}
	if err := validateURLs(urls); err != nil {
		return nil, xerrors.Errorf("validate origins in cors-allowed-origin-list: %w", err)
	}

	return urls, nil
}

// validateURLs checks if all URLs in slice is valid or not.
func validateURLs(urls []string) error {
	for _, u := range urls {
		_, err := url.ParseRequestURI(u)
		if err != nil {
			return xerrors.Errorf("parse request uri: %w", err)
		}
	}
	return nil
}

func parseStringListEnv(e string) []string {
	list := strings.Split(e, ",")
	for i := range list {
		// remove space
		list[i] = strings.TrimSpace(list[i])
	}

	return list
}

// getSchedulerCfg reads KUBE_SCHEDULER_CONFIG_PATH which means initial kube-scheduler configuration
// if empty from the config file.
// and converts it into *configv1.KubeSchedulerConfiguration.
// KUBE_SCHEDULER_CONFIG_PATH is not required.
// If KUBE_SCHEDULER_CONFIG_PATH is not set, the default configuration of kube-scheduler will be used.
func getSchedulerCfg() (*configv1.KubeSchedulerConfiguration, error) {
	kubeSchedulerConfigPath := os.Getenv("KUBE_SCHEDULER_CONFIG_PATH")
	if kubeSchedulerConfigPath == "" {
		kubeSchedulerConfigPath = configYaml.KubeSchedulerConfigPath
		if kubeSchedulerConfigPath == "" {
			dsc, err := config.DefaultSchedulerConfig()
			if err != nil {
				return nil, xerrors.Errorf("create default scheduler config: %w", err)
			}
			return dsc, nil
		}
	}
	data, err := os.ReadFile(kubeSchedulerConfigPath)
	if err != nil {
		return nil, xerrors.Errorf("read scheduler config file: %w", err)
	}

	sc, err := decodeSchedulerCfg(data)
	if err != nil {
		return nil, xerrors.Errorf("decode scheduler config file: %w", err)
	}

	return sc, nil
}

// getExternalImportEnabled reads EXTERNAL_IMPORT_ENABLED and convert it to bool
// if empty from the config file.
// This function will return `true` if `EXTERNAL_IMPORT_ENABLED` is "1".
func getExternalImportEnabled() bool {
	isExternalImportEnabledString := os.Getenv("EXTERNAL_IMPORT_ENABLED")
	if isExternalImportEnabledString == "" {
		isExternalImportEnabledString = strconv.FormatBool(configYaml.ExternalImportEnabled)
	}
	isExternalImportEnabled, _ := strconv.ParseBool(isExternalImportEnabledString)
	return isExternalImportEnabled
}

func decodeSchedulerCfg(buf []byte) (*configv1.KubeSchedulerConfiguration, error) {
	decoder := scheme.Codecs.UniversalDeserializer()
	obj, _, err := decoder.Decode(buf, nil, nil)
	if err != nil {
		return nil, xerrors.Errorf("load an k8s object from buffer: %w", err)
	}

	sc, ok := obj.(*configv1.KubeSchedulerConfiguration)
	if !ok {
		return nil, xerrors.Errorf("convert to *configv1.KubeSchedulerConfiguration, but got unexpected type: %T", obj)
	}

	if err = sc.DecodeNestedObjects(decoder); err != nil {
		return nil, xerrors.Errorf("decode nested plugin args: %w", err)
	}
	return sc, nil
}

func GetKubeClientConfig() (*rest.Config, error) {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, xerrors.Errorf("get client config: %w", err)
	}
	return clientConfig, nil
}
