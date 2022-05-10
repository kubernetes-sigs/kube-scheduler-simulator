package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/defaultconfig"
)

// ErrEmptyEnv represents the required environment variable don't exist.
var ErrEmptyEnv = errors.New("env is needed, but empty")

// Config is configuration for simulator.
type Config struct {
	Port                  int
	KubeAPIServerURL      string
	EtcdURL               string
	CorsAllowedOriginList []string
	// ExternalImportEnabled indicates whether the simulator will import resources from an existing cluster or not.
	ExternalImportEnabled bool
	// ExternalKubeClientCfg is KubeConfig to get resources from external cluster.
	// This field is non-empty only when ExternalImportEnabled == true.
	ExternalKubeClientCfg *rest.Config
	InitialSchedulerCfg   *v1beta2config.KubeSchedulerConfiguration
}

// NewConfig gets some settings from environment variables.
func NewConfig() (*Config, error) {
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

	return &Config{
		Port:                  port,
		KubeAPIServerURL:      apiurl,
		EtcdURL:               etcdurl,
		CorsAllowedOriginList: corsAllowedOriginList,
		InitialSchedulerCfg:   initialschedulerCfg,
		ExternalImportEnabled: externalimportenabled,
		ExternalKubeClientCfg: externalKubeClientCfg,
	}, nil
}

// getPort gets Port from the environment variable named PORT.
func getPort() (int, error) {
	p := os.Getenv("PORT")
	if p == "" {
		return 0, xerrors.Errorf("get PORT from env: %w", ErrEmptyEnv)
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, xerrors.Errorf("convert PORT of string to int: %w", err)
	}
	return port, nil
}

func getKubeAPIServerURL() string {
	p := os.Getenv("KUBE_API_PORT")
	if p == "" {
		p = "3131"
	}

	h := os.Getenv("KUBE_API_HOST")
	if h == "" {
		h = "127.0.0.1"
	}
	return h + ":" + p
}

func getEtcdURL() (string, error) {
	e := os.Getenv("KUBE_SCHEDULER_SIMULATOR_ETCD_URL")
	if e == "" {
		return "", xerrors.Errorf("get KUBE_SCHEDULER_SIMULATOR_ETCD_URL from env: %w", ErrEmptyEnv)
	}

	return e, nil
}

// getCorsAllowedOriginList fetches CorsAllowedOriginList from the env named CORS_ALLOWED_ORIGIN_LIST.
// This allowed list is applied to kube-apiserver and the simulator server.
//
// e.g. CORS_ALLOWED_ORIGIN_LIST="http://localhost:3000, http://localhost:3001, http://localhost:3002"
//       → return []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:3002"}
func getCorsAllowedOriginList() ([]string, error) {
	e := os.Getenv("CORS_ALLOWED_ORIGIN_LIST")
	if e == "" {
		return nil, xerrors.Errorf("get CORS_ALLOWED_ORIGIN_LIST from env: %w", ErrEmptyEnv)
	}

	urls := parseStringListEnv(e)
	if err := validateURLs(urls); err != nil {
		return nil, xerrors.Errorf("validate origins in CORS_ALLOWED_ORIGIN_LIST: %w", err)
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
// and converts it into *v1beta2config.KubeSchedulerConfiguration.
// KUBE_SCHEDULER_CONFIG_PATH is not required.
// If KUBE_SCHEDULER_CONFIG_PATH is not set, the default configuration of kube-scheduler will be used.
func getSchedulerCfg() (*v1beta2config.KubeSchedulerConfiguration, error) {
	e := os.Getenv("KUBE_SCHEDULER_CONFIG_PATH")
	if e == "" {
		dsc, err := defaultconfig.DefaultSchedulerConfig()
		if err != nil {
			return nil, xerrors.Errorf("create default scheduler config: %w", err)
		}
		return dsc, nil
	}

	data, err := os.ReadFile(e)
	if err != nil {
		return nil, xerrors.Errorf("read scheduler config file: %w", err)
	}

	sc, err := decodeSchedulerCfg(data)
	if err != nil {
		return nil, xerrors.Errorf("decode scheduler config file: %w", err)
	}

	return sc, nil
}

// getExternalImportEnabled reads EXTERNAL_IMPORT_ENABLED and convert it to bool.
// This function will return `true` if `EXTERNAL_IMPORT_ENABLED` is "1".
func getExternalImportEnabled() bool {
	i := os.Getenv("EXTERNAL_IMPORT_ENABLED")
	return i == "1"
}

func decodeSchedulerCfg(buf []byte) (*v1beta2config.KubeSchedulerConfiguration, error) {
	decoder := scheme.Codecs.UniversalDeserializer()
	obj, _, err := decoder.Decode(buf, nil, nil)
	if err != nil {
		return nil, xerrors.Errorf("load an k8s object from buffer: %w", err)
	}

	sc, ok := obj.(*v1beta2config.KubeSchedulerConfiguration)
	if !ok {
		return nil, xerrors.Errorf("convert to *v1beta2config.KubeSchedulerConfiguration, but got unexpected type: %T", obj)
	}

	if err = sc.DecodeNestedObjects(decoder); err != nil {
		return nil, xerrors.Errorf("decode nested plugin args: %w", err)
	}
	return sc, nil
}

func GetKubeClientConfig() (*rest.Config, error) {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, xerrors.Errorf("get client config: %w", err)
	}
	return config, nil
}
