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

	// This is the port number on which kube-scheduler-simulator
	// server is started.
	Port int `json:"port,omitempty"`

	// This is the URL for etcd. The simulator runs kube-apiserver
	// internally, and the kube-apiserver uses this etcd.
	EtcdURL string `json:"etcdURL,omitempty"`

	// This URL represents the URL once web UI is started.
	// The simulator and internal kube-apiserver set the allowed
	// origin for CorsAllowedOriginList
	CorsAllowedOriginList []string `json:"corsAllowedOriginList,omitempty"`

	// This is for the beta feature "Importing cluster's resources".
	// This variable is used to find Kubeconfig required to access your
	// cluster for importing resources to scheduler simulator.
	KubeConfig string `json:"kubeConfig,omitempty"`

	// This is the host of kube-apiserver which the simulator
	// starts internally. Its default value is 127.0.0.1.
	KubeAPIHost string `json:"kubeApiHost,omitempty"`

	// This is the port of kube-apiserver. Its default value is 3131.
	KubeAPIPort int `json:"kubeApiPort,omitempty"`

	// The path to a KubeSchedulerConfiguration file.
	// If passed, the simulator will start the scheduler
	// with that configuration. Or, if you use web UI,
	// you can change the configuration from the web UI as well.
	KubeSchedulerConfigPath string `json:"kubeSchedulerConfigPath,omitempty"`

	// This variable indicates whether the simulator will
	// import resources from an user cluster's or not.
	// Note, this is still a beta feature.
	ExternalImportEnabled bool `json:"externalImportEnabled,omitempty"`

	// This variable indicates whether an external scheduler
	// is used.
	ExternalSchedulerEnabled bool `json:"externalSchedulerEnabled,omitempty"`
}
