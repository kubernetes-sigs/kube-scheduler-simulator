package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
)

// kubeSchedulerConfigPath represents the file path to the scheduler configuration.
var kubeSchedulerConfigPath string

// DefaultSchedulerConfig creates KubeSchedulerConfiguration default configuration.
func DefaultSchedulerConfig() (*v1.KubeSchedulerConfiguration, error) {
	var versionedCfg v1.KubeSchedulerConfiguration
	scheme.Scheme.Default(&versionedCfg)
	versionedCfg.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))

	return &versionedCfg, nil
}

// SetKubeSchedulerCfg set Scheduler config path.
func SetKubeSchedulerCfg(path string) {
	kubeSchedulerConfigPath = path
}

func UpdateSchedulerConfig(cfg *v1.KubeSchedulerConfiguration) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal jsonData: %w", err)
	}

	var yamlData map[string]interface{}
	decoder := yaml.NewDecoder((bytes.NewReader(jsonData)))
	err = decoder.Decode(&yamlData)
	if err != nil {
		return fmt.Errorf("failed to decode jsonData: %w", err)
	}

	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	if err := os.WriteFile(kubeSchedulerConfigPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
