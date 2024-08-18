package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
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

type LeaderElectionConfiguration struct {
	LeaderElect       *bool           `yaml:"leaderElect,omitempty"`
	LeaseDuration     metav1.Duration `yaml:"leaseDuration,omitempty"`
	RenewDeadline     metav1.Duration `yaml:"renewDeadline,omitempty"`
	RetryPeriod       metav1.Duration `yaml:"retryPeriod,omitempty"`
	ResourceLock      string          `yaml:"resourceLock,omitempty"`
	ResourceName      string          `yaml:"resourceName,omitempty"`
	ResourceNamespace string          `yaml:"resourceNamespace,omitempty"`
}

type ClientConnectionConfiguration struct {
	Kubeconfig         string  `yaml:"kubeconfig"`
	AcceptContentTypes string  `yaml:"acceptContentTypes,omitempty"`
	ContentType        string  `yaml:"contentType,omitempty"`
	QPS                float32 `yaml:"qps,omitempty"`
	Burst              int32   `yaml:"burst,omitempty"`
}

type DebuggingConfiguration struct {
	EnableProfiling           *bool `yaml:"enableProfiling,omitempty"`
	EnableContentionProfiling *bool `yaml:"enableContentionProfiling,omitempty"`
}

type KubeSchedulerProfile struct {
	SchedulerName            *string        `yaml:"schedulerName,omitempty"`
	PercentageOfNodesToScore *int32         `yaml:"percentageOfNodesToScore,omitempty"`
	Plugins                  *Plugins       `yaml:"plugins,omitempty"`
	PluginConfig             []PluginConfig `yaml:"pluginConfig,omitempty"`
}

type Plugins struct {
	PreEnqueue PluginSet `yaml:"preEnqueue,omitempty"`
	QueueSort  PluginSet `yaml:"queueSort,omitempty"`
	PreFilter  PluginSet `yaml:"preFilter,omitempty"`
	Filter     PluginSet `yaml:"filter,omitempty"`
	PostFilter PluginSet `yaml:"postFilter,omitempty"`
	PreScore   PluginSet `yaml:"preScore,omitempty"`
	Score      PluginSet `yaml:"score,omitempty"`
	Reserve    PluginSet `yaml:"reserve,omitempty"`
	Permit     PluginSet `yaml:"permit,omitempty"`
	PreBind    PluginSet `yaml:"preBind,omitempty"`
	Bind       PluginSet `yaml:"bind,omitempty"`
	PostBind   PluginSet `yaml:"postBind,omitempty"`
	MultiPoint PluginSet `yaml:"multiPoint,omitempty"`
}

type PluginSet struct {
	Enabled  []Plugin `yaml:"enabled,omitempty"`
	Disabled []Plugin `yaml:"disabled,omitempty"`
}

type Plugin struct {
	Name   string `yaml:"name"`
	Weight *int32 `yaml:"weight,omitempty"`
}

type PluginConfig struct {
	Name string               `yaml:"name"`
	Args runtime.RawExtension `yaml:"args,omitempty"`
}

type Extender struct {
	URLPrefix        string                    `yaml:"urlPrefix"`
	FilterVerb       string                    `yaml:"filterVerb,omitempty"`
	PreemptVerb      string                    `yaml:"preemptVerb,omitempty"`
	PrioritizeVerb   string                    `yaml:"prioritizeVerb,omitempty"`
	Weight           int64                     `yaml:"weight,omitempty"`
	BindVerb         string                    `yaml:"bindVerb,omitempty"`
	EnableHTTPS      bool                      `yaml:"enableHTTPS,omitempty"`
	TLSConfig        *ExtenderTLSConfig        `yaml:"tlsConfig,omitempty"`
	HTTPTimeout      metav1.Duration           `yaml:"httpTimeout,omitempty"`
	NodeCacheCapable bool                      `yaml:"nodeCacheCapable,omitempty"`
	ManagedResources []ExtenderManagedResource `yaml:"managedResources,omitempty"`
	Ignorable        bool                      `yaml:"ignorable,omitempty"`
}

type ExtenderManagedResource struct {
	Name               string `yaml:"name"`
	IgnoredByScheduler bool   `yaml:"ignoredByScheduler,omitempty"`
}

type ExtenderTLSConfig struct {
	Insecure   bool   `yaml:"insecure,omitempty"`
	ServerName string `yaml:"serverName,omitempty"`
	CertFile   string `yaml:"certFile,omitempty"`
	KeyFile    string `yaml:"keyFile,omitempty"`
	CAFile     string `yaml:"caFile,omitempty"`
	CertData   []byte `yaml:"certData,omitempty"`
	KeyData    []byte `yaml:"keyData,omitempty"`
	CAData     []byte `yaml:"caData,omitempty"`
}

func convertLeaderElection(cfg componentbaseconfigv1alpha1.LeaderElectionConfiguration) *LeaderElectionConfiguration {
	return &LeaderElectionConfiguration{
		LeaderElect:       cfg.LeaderElect,
		LeaseDuration:     cfg.LeaseDuration,
		RenewDeadline:     cfg.RenewDeadline,
		RetryPeriod:       cfg.RetryPeriod,
		ResourceLock:      cfg.ResourceLock,
		ResourceName:      cfg.ResourceName,
		ResourceNamespace: cfg.ResourceNamespace,
	}
}

func convertClientConnection(cfg componentbaseconfigv1alpha1.ClientConnectionConfiguration) *ClientConnectionConfiguration {
	return &ClientConnectionConfiguration{
		Kubeconfig:         cfg.Kubeconfig,
		AcceptContentTypes: cfg.AcceptContentTypes,
		ContentType:        cfg.ContentType,
		QPS:                cfg.QPS,
		Burst:              cfg.Burst,
	}
}

func convertDebuggingConfiguration(cfg componentbaseconfigv1alpha1.DebuggingConfiguration) DebuggingConfiguration {
	return DebuggingConfiguration{
		EnableProfiling:           cfg.EnableProfiling,
		EnableContentionProfiling: cfg.EnableContentionProfiling,
	}
}

func convertProfiles(profiles []v1.KubeSchedulerProfile) []KubeSchedulerProfile {
	var result []KubeSchedulerProfile
	for _, p := range profiles {
		result = append(result, KubeSchedulerProfile{
			SchedulerName:            p.SchedulerName,
			PercentageOfNodesToScore: p.PercentageOfNodesToScore,
			Plugins:                  convertPlugins(p.Plugins),
			PluginConfig:             convertPluginConfig(p.PluginConfig),
		})
	}
	return result
}

func convertPlugins(plugins *v1.Plugins) *Plugins {
	if plugins == nil {
		return nil
	}
	return &Plugins{
		PreEnqueue: convertPluginSet(plugins.PreEnqueue),
		QueueSort:  convertPluginSet(plugins.QueueSort),
		PreFilter:  convertPluginSet(plugins.PreFilter),
		Filter:     convertPluginSet(plugins.Filter),
		PostFilter: convertPluginSet(plugins.PostFilter),
		PreScore:   convertPluginSet(plugins.PreScore),
		Score:      convertPluginSet(plugins.Score),
		Reserve:    convertPluginSet(plugins.Reserve),
		Permit:     convertPluginSet(plugins.Permit),
		PreBind:    convertPluginSet(plugins.PreBind),
		Bind:       convertPluginSet(plugins.Bind),
		PostBind:   convertPluginSet(plugins.PostBind),
		MultiPoint: convertPluginSet(plugins.MultiPoint),
	}
}

func convertPluginSet(set v1.PluginSet) PluginSet {
	return PluginSet{
		Enabled:  convertPluginsArray(set.Enabled),
		Disabled: convertPluginsArray(set.Disabled),
	}
}

func convertPluginsArray(plugins []v1.Plugin) []Plugin {
	var result []Plugin
	for _, p := range plugins {
		result = append(result, Plugin{
			Name:   p.Name,
			Weight: p.Weight,
		})
	}
	return result
}

func convertPluginConfig(config []v1.PluginConfig) []PluginConfig {
	var result []PluginConfig
	for _, c := range config {
		result = append(result, PluginConfig{
			Name: c.Name,
			Args: c.Args,
		})
	}
	return result
}

func convertExtenders(extenders []v1.Extender) []Extender {
	var result []Extender
	for _, e := range extenders {
		result = append(result, Extender{
			URLPrefix:        e.URLPrefix,
			FilterVerb:       e.FilterVerb,
			PreemptVerb:      e.PreemptVerb,
			PrioritizeVerb:   e.PrioritizeVerb,
			Weight:           e.Weight,
			BindVerb:         e.BindVerb,
			EnableHTTPS:      e.EnableHTTPS,
			TLSConfig:        convertExtenderTLSConfig(e.TLSConfig),
			HTTPTimeout:      e.HTTPTimeout,
			NodeCacheCapable: e.NodeCacheCapable,
			ManagedResources: convertExtenderManagedResources(e.ManagedResources),
			Ignorable:        e.Ignorable,
		})
	}
	return result
}

func convertExtenderManagedResources(resources []v1.ExtenderManagedResource) []ExtenderManagedResource {
	var result []ExtenderManagedResource
	for _, r := range resources {
		result = append(result, ExtenderManagedResource{
			Name:               r.Name,
			IgnoredByScheduler: r.IgnoredByScheduler,
		})
	}
	return result
}

func convertExtenderTLSConfig(cfg *v1.ExtenderTLSConfig) *ExtenderTLSConfig {
	if cfg == nil {
		return nil
	}
	return &ExtenderTLSConfig{
		Insecure:   cfg.Insecure,
		ServerName: cfg.ServerName,
		CertFile:   cfg.CertFile,
		KeyFile:    cfg.KeyFile,
		CAFile:     cfg.CAFile,
		CertData:   cfg.CertData,
		KeyData:    cfg.KeyData,
		CAData:     cfg.CAData,
	}
}

func WriteConfig(cfg *v1.KubeSchedulerConfiguration) error {
	customCfg := struct {
		APIVersion               string                         `yaml:"apiVersion"`
		Kind                     string                         `yaml:"kind"`
		Parallelism              *int32                         `yaml:"parallelism,omitempty"`
		LeaderElection           *LeaderElectionConfiguration   `yaml:"leaderElection,omitempty"`
		ClientConnection         *ClientConnectionConfiguration `yaml:"clientConnection,omitempty"`
		DebuggingConfiguration   DebuggingConfiguration         `yaml:",inline"`
		PercentageOfNodesToScore *int32                         `yaml:"percentageOfNodesToScore,omitempty"`
		PodInitialBackoffSeconds *int64                         `yaml:"podInitialBackoffSeconds,omitempty"`
		PodMaxBackoffSeconds     *int64                         `yaml:"podMaxBackoffSeconds,omitempty"`
		Profiles                 []KubeSchedulerProfile         `yaml:"profiles,omitempty"`
		Extenders                []Extender                     `yaml:"extenders,omitempty"`
		DelayCacheUntilActive    *bool                          `yaml:"delayCacheUntilActive,omitempty"`
	}{
		APIVersion:               cfg.APIVersion,
		Kind:                     cfg.Kind,
		Parallelism:              cfg.Parallelism,
		LeaderElection:           convertLeaderElection(cfg.LeaderElection),
		ClientConnection:         convertClientConnection(cfg.ClientConnection),
		DebuggingConfiguration:   convertDebuggingConfiguration(cfg.DebuggingConfiguration),
		PercentageOfNodesToScore: cfg.PercentageOfNodesToScore,
		PodInitialBackoffSeconds: cfg.PodInitialBackoffSeconds,
		PodMaxBackoffSeconds:     cfg.PodMaxBackoffSeconds,
		Profiles:                 convertProfiles(cfg.Profiles),
		Extenders:                convertExtenders(cfg.Extenders),
		DelayCacheUntilActive:    &cfg.DelayCacheUntilActive,
	}

	kubeSchedulerConfigPath := os.Getenv("KUBE_SCHEDULER_CONFIG_PATH")
	data, err := yaml.Marshal(customCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}

	if err := os.WriteFile(kubeSchedulerConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
