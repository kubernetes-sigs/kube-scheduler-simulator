package config

import (
	v1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
)

// DefaultSchedulerConfig creates KubeSchedulerConfiguration default configuration.
func DefaultSchedulerConfig() (*v1.KubeSchedulerConfiguration, error) {
	var versionedCfg v1.KubeSchedulerConfiguration
	scheme.Scheme.Default(&versionedCfg)
	versionedCfg.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))

	return &versionedCfg, nil
}
