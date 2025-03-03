package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// defaultSchedulerCfgPath is where we have the scheduler config in the container by default.
const defaultSchedulerCfgPath = "/config/scheduler.yaml"

// Config is configuration for simulator.
type Config struct {
	Port                  int
	KubeAPIServerURL      string
	EtcdURL               string
	CorsAllowedOriginList []string
	// ExternalImportEnabled indicates whether the simulator will import resources from a target cluster once
	// when it's started.
	ExternalImportEnabled bool
	// ResourceImportLabelSelector is the label selector used to determine which resources from the target cluster should be imported.
	ResourceImportLabelSelector metav1.LabelSelector
	// ResourceSyncEnabled indicates whether the simulator will keep syncing resources from a target cluster.
	ResourceSyncEnabled bool
	// ReplayerEnabled indicates whether the simulator will replay events recorded in a file.
	ReplayerEnabled bool
	// RecordFilePath is the path to the file where the simulator records events.
	RecordFilePath string
	// ExternalKubeClientCfg is KubeConfig to get resources from external cluster.
	// This field should be set when ExternalImportEnabled == true or ResourceSyncEnabled == true.
	ExternalKubeClientCfg *rest.Config
	InitialSchedulerCfg   *configv1.KubeSchedulerConfiguration
}

const (
	// defaultFilePath is the config file path.
	// TODO: move it to somewhere configurable from outside if users want.
	defaultFilePath = "./config.yaml"
)

// NewConfig gets some settings from config file or environment variables.
//
//nolint:cyclop
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

	apiurl, err := getKubeAPIServerURL()
	if err != nil {
		return nil, xerrors.Errorf("get kube API server URL: %w", err)
	}

	externalimportenabled := getExternalImportEnabled()
	resourceSyncEnabled := getResourceSyncEnabled()
	replayerEnabled := getReplayerEnabled()
	recordFilePath := getRecordFilePath()
	externalKubeClientCfg := &rest.Config{}
	if hasTwoOrMoreTrue(externalimportenabled, resourceSyncEnabled, replayerEnabled) {
		return nil, xerrors.Errorf("externalImportEnabled, resourceSyncEnabled and replayerEnabled cannot be used simultaneously.")
	}
	if externalimportenabled || resourceSyncEnabled {
		externalKubeClientCfg, err = clientcmd.BuildConfigFromFlags("", configYaml.KubeConfig)
		if err != nil {
			return nil, xerrors.Errorf("get kube clientconfig: %w", err)
		}
	}

	initialschedulerCfg, err := GetSchedulerCfg()
	if err != nil {
		return nil, xerrors.Errorf("get SchedulerCfg: %w", err)
	}

	return &Config{
		Port:                        port,
		KubeAPIServerURL:            apiurl,
		EtcdURL:                     etcdurl,
		CorsAllowedOriginList:       corsAllowedOriginList,
		InitialSchedulerCfg:         initialschedulerCfg,
		ExternalImportEnabled:       externalimportenabled,
		ResourceImportLabelSelector: configYaml.ResourceImportLabelSelector,
		ExternalKubeClientCfg:       externalKubeClientCfg,
		ResourceSyncEnabled:         resourceSyncEnabled,
		ReplayerEnabled:             replayerEnabled,
		RecordFilePath:              recordFilePath,
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
func getKubeAPIServerURL() (string, error) {
	url := os.Getenv("KUBE_APISERVER_URL")
	if url == "" {
		url = configYaml.KubeAPIServerURL
		if url == "" {
			return "", xerrors.Errorf("get KUBE_APISERVER_URL from config: %w", ErrEmptyConfig)
		}
	}
	return url, nil
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
// Let's say CORS_ALLOWED_ORIGIN_LIST=http://localhost:3000,http://localhost:3001,http://localhost:3002 is given.
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

// GetSchedulerCfg reads KUBE_SCHEDULER_CONFIG_PATH which means initial kube-scheduler configuration
// if empty from the config file.
// and converts it into *configv1.KubeSchedulerConfiguration.
// KUBE_SCHEDULER_CONFIG_PATH is not required.
// If KUBE_SCHEDULER_CONFIG_PATH is not set, the default configuration of kube-scheduler will be used.
func GetSchedulerCfg() (*configv1.KubeSchedulerConfiguration, error) {
	kubeSchedulerConfigPath := os.Getenv("KUBE_SCHEDULER_CONFIG_PATH")
	if kubeSchedulerConfigPath == "" {
		kubeSchedulerConfigPath = configYaml.KubeSchedulerConfigPath
		if kubeSchedulerConfigPath == "" {
			config.SetKubeSchedulerCfgPath(defaultSchedulerCfgPath)
			dsc, err := config.DefaultSchedulerConfig()
			if err != nil {
				return nil, xerrors.Errorf("create default scheduler config: %w", err)
			}
			return dsc, nil
		}
	}
	config.SetKubeSchedulerCfgPath(kubeSchedulerConfigPath)
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

// getResourceSyncEnabled reads RESOURCE_SYNC_ENABLED and converts it to bool
// if empty from the config file.
// This function will return `true` if `RESOURCE_SYNC_ENABLED` is "1".
func getResourceSyncEnabled() bool {
	resourceSyncEnabledString := os.Getenv("RESOURCE_SYNC_ENABLED")
	if resourceSyncEnabledString == "" {
		resourceSyncEnabledString = strconv.FormatBool(configYaml.ResourceSyncEnabled)
	}
	resourceSyncEnabled, _ := strconv.ParseBool(resourceSyncEnabledString)
	return resourceSyncEnabled
}

// getReplayerEnabled reads REPLAYER_ENABLED and converts it to bool
// if empty from the config file.
// This function will return `true` if `REPLAYER_ENABLED` is "1".
func getReplayerEnabled() bool {
	replayerEnabledString := os.Getenv("REPLAYER_ENABLED")
	if replayerEnabledString == "" {
		replayerEnabledString = strconv.FormatBool(configYaml.ReplayerEnabled)
	}
	replayerEnabled, _ := strconv.ParseBool(replayerEnabledString)
	return replayerEnabled
}

// getRecordFilePath reads RECORD_FILE_PATH
// if empty from the config file.
func getRecordFilePath() string {
	recordFilePath := os.Getenv("RECORD_FILE_PATH")
	if recordFilePath == "" {
		recordFilePath = configYaml.RecordFilePath
	}
	return recordFilePath
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

func hasTwoOrMoreTrue(values ...bool) bool {
	count := 0
	for _, v := range values {
		if v {
			count++
		}
	}
	return count >= 2
}
