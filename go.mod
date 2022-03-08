module github.com/kubernetes-sigs/kube-scheduler-simulator

go 1.16

replace (
	google.golang.org/grpc => google.golang.org/grpc v1.38.0
	k8s.io/api => k8s.io/api v0.23.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.4
	k8s.io/apiserver => k8s.io/apiserver v0.23.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.4
	k8s.io/client-go => k8s.io/client-go v0.23.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.4
	k8s.io/code-generator => k8s.io/code-generator v0.23.4
	k8s.io/component-base => k8s.io/component-base v0.23.4
	k8s.io/component-helpers => k8s.io/component-helpers v0.23.4
	k8s.io/controller-manager => k8s.io/controller-manager v0.23.4
	k8s.io/cri-api => k8s.io/cri-api v0.23.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.4
	k8s.io/kubectl => k8s.io/kubectl v0.23.4
	k8s.io/kubelet => k8s.io/kubelet v0.23.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.4
	k8s.io/metrics => k8s.io/metrics v0.23.4
	k8s.io/mount-utils => k8s.io/mount-utils v0.23.4
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.23.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.4
)

require (
	github.com/golang/mock v1.5.0
	github.com/golangci/golangci-lint v1.41.1
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.1.2
	github.com/labstack/echo/v4 v4.5.0
	github.com/labstack/gommon v0.3.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	k8s.io/api v1.23.4
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v1.23.4
	k8s.io/apiserver v1.23.4
	k8s.io/client-go v1.23.4
	k8s.io/component-base v1.23.4
	k8s.io/klog/v2 v2.30.0
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf
	k8s.io/kube-scheduler v1.23.4
	k8s.io/kubernetes v1.23.4
)
