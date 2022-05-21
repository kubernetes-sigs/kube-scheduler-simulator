package config

import (
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
)

// DefaultSchedulerConfig creates KubeSchedulerConfiguration default configuration.
func DefaultSchedulerConfig() (*v1beta2.KubeSchedulerConfiguration, error) {
	var versionedCfg v1beta2.KubeSchedulerConfiguration
	scheme.Scheme.Default(&versionedCfg)
	versionedCfg.SetGroupVersionKind(v1beta2.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))

	return &versionedCfg, nil
}
