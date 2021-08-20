package config

import (
	"errors"
	"os"
	"strconv"

	"golang.org/x/xerrors"
)

// ErrEmptyEnv represents the required environment variable don't exist.
var ErrEmptyEnv = errors.New("env is needed, but empty")

// Config is configuration for simulator.
type Config struct {
	Port        int
	EtcdURL     string
	FrontendURL string
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

	return &Config{
		Port:        port,
		EtcdURL:     etcdurl,
		FrontendURL: frontendurl,
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
