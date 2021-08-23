package defaultconfig

import (
	"golang.org/x/xerrors"
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"
)

// DefaultSchedulerConfig creates KubeSchedulerConfiguration default configuration.
func DefaultSchedulerConfig() (*v1beta2.KubeSchedulerConfiguration, error) {
	var versionedCfg v1beta2.KubeSchedulerConfiguration
	scheme.Scheme.Default(&versionedCfg)

	return &versionedCfg, nil
}

func DefaultFilterPlugins() ([]v1beta2.Plugin, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return nil, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Filter.Enabled, nil
}

func DefaultScorePlugins() ([]v1beta2.Plugin, error) {
	defaultConfig, err := DefaultSchedulerConfig()
	if err != nil || len(defaultConfig.Profiles) != 1 {
		// default Config should only have default-scheduler configuration.
		return nil, xerrors.Errorf("get default scheduler configuration: %w", err)
	}
	return defaultConfig.Profiles[0].Plugins.Score.Enabled, nil
}
