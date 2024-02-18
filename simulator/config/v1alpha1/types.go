/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SimulatorConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// Port is the port number on which kube-scheduler-simulator
	// server is started.
	Port int `json:"port,omitempty"`

	EtcdURL string `json:"etcdURL,omitempty"`

	// CorsAllowedOriginList is the list that the simulator server and the internal kube-apiserver use as the allowed
	// origin for CorsAllowedOriginList.
	CorsAllowedOriginList []string `json:"corsAllowedOriginList,omitempty"`

	// KubeConfig is a path to Kubeconfig for your real kubernetes cluster.
	// This configuration is used for importing resources to scheduler simulator.
	KubeConfig string `json:"kubeConfig,omitempty"`

	// This is the URL for kube-apiserver.
	KubeAPIServerURL string `json:"kubeApiServerUrl,omitempty"`

	KubeAPIHost string `json:"kubeApiHost,omitempty"`

	// KubeAPIPort is the port of kube-apiserver that the simulator has internally. Its default value is 3131.
	KubeAPIPort int `json:"kubeApiPort,omitempty"`

	// KubeSchedulerConfigPath is a path to a KubeSchedulerConfiguration file.
	// If passed, the simulator will start the scheduler
	// with that configuration. Or, if you use web UI,
	// you can change the configuration from the web UI as well.
	KubeSchedulerConfigPath string `json:"kubeSchedulerConfigPath,omitempty"`

	// ExternalImportEnabled indicates whether the simulator will
	// import resources from an user cluster's or not.
	// When you set it true, you also have to set KubeConfig envrionment variable.
	// Note that this is still a beta feature.
	ExternalImportEnabled bool `json:"externalImportEnabled,omitempty"`

	// ExternalSyncEnabled indicates whether the simulator will
	// sync resources from an user cluster's or not.
	// When you set it true, you also have to set KubeConfig envrionment variable.
	// Note that this is still a beta feature.
	ExternalSyncEnabled bool `json:"externalSyncEnabled,omitempty"`

	// ExternalSchedulerEnabled indicates whether an external scheduler
	// is used.
	ExternalSchedulerEnabled bool `json:"externalSchedulerEnabled,omitempty"`
}
