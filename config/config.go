package config

import (
	"errors"
	"os"
	"strconv"

	"golang.org/x/xerrors"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler/defaultconfig"
)

// ErrEmptyEnv represents the required environment variable don't exist.
var ErrEmptyEnv = errors.New("env is needed, but empty")

// Config is configuration for simulator.
type Config struct {
	Port                int
	APIURL              string
	EtcdURL             string
	FrontendURL         string
	InitialSchedulerCfg *v1beta2config.KubeSchedulerConfiguration
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

	frontendurl, err := getFrontendURL()
	if err != nil {
		return nil, xerrors.Errorf("get frontend URL: %w", err)
	}

	apiurl := getAPIURL()

	initialschedulerCfg, err := getSchedulerCfg()
	if err != nil {
		return nil, xerrors.Errorf("get SchedulerCfg: %w", err)
	}

	return &Config{
		Port:                port,
		APIURL:              apiurl,
		EtcdURL:             etcdurl,
		FrontendURL:         frontendurl,
		InitialSchedulerCfg: initialschedulerCfg,
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

func getAPIURL() string {
	p := os.Getenv("KUBE_API_PORT")
	if p == "" {
		// we still want the simulator to behave as before,
		// use a local test port.
		p = "0"
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

func getFrontendURL() (string, error) {
	e := os.Getenv("FRONTEND_URL")
	if e == "" {
		return "", xerrors.Errorf("get FRONTEND_URL from env: %w", ErrEmptyEnv)
	}

	return e, nil
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
