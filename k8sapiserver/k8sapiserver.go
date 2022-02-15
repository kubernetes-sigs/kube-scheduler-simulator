package k8sapiserver

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	extensionsapiserver "k8s.io/apiextensions-apiserver/pkg/apiserver"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	authauthenticator "k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	authenticatorunion "k8s.io/apiserver/pkg/authentication/request/union"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	authorizerunion "k8s.io/apiserver/pkg/authorization/union"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	utilflowcontrol "k8s.io/apiserver/pkg/util/flowcontrol"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"

	restclient "k8s.io/client-go/rest"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"
	aggregatorapiserver "k8s.io/kube-aggregator/pkg/apiserver"
	aggregatorscheme "k8s.io/kube-aggregator/pkg/apiserver/scheme"
	openapicommon "k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/controlplane"
	"k8s.io/kubernetes/pkg/kubeapiserver"
	kubeletclient "k8s.io/kubernetes/pkg/kubelet/client"

	generated "github.com/kubernetes-sigs/kube-scheduler-simulator/k8sapiserver/openapi"
)

// StartAPIServer starts API server, and it make panic when a error happen.
func StartAPIServer(kubeAPIServerURL string, etcdURL string) (func(), error) {
	h := &APIServerHolder{Initialized: make(chan struct{})}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	})

	l, err := net.Listen("tcp", kubeAPIServerURL)
	if err != nil {
		return nil, xerrors.Errorf("announces on the local network address: %w", err)
	}

	s := &httptest.Server{
		Listener: l,
		Config: &http.Server{
			Handler: handler,
		},
	}
	s.Start()
	klog.Info("kube-apiserver is started on :", s.URL)

	c, etcdOpt := NewControlPlaneConfigWithOptions(s.URL, etcdURL)

	clientset, err := clientset.NewForConfig(c.GenericConfig.LoopbackClientConfig)
	if err != nil {
		return nil, xerrors.Errorf("create clientset: %w", err)
	}

	c.ExtraConfig.VersionedInformers = informers.NewSharedInformerFactory(clientset, c.GenericConfig.LoopbackClientConfig.Timeout)

	c.GenericConfig.FlowControl = utilflowcontrol.New(
		c.ExtraConfig.VersionedInformers,
		clientset.FlowcontrolV1beta1(),
		c.GenericConfig.MaxRequestsInFlight+c.GenericConfig.MaxMutatingRequestsInFlight,
		c.GenericConfig.RequestTimeout/4,
	)
	c.ExtraConfig.ServiceIPRange = net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}

	apiExtensionsCfg, err := createAPIExtensionConfig(*c.GenericConfig, *etcdOpt, c.ExtraConfig.VersionedInformers)
	if err != nil {
		return nil, xerrors.Errorf("Create APIExtension Config error: %w", err)
	}

	apiExtensionServer, err := createAPIExtensionsServer(apiExtensionsCfg, genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, xerrors.Errorf("Create APIExtension Server error: %w", err)
	}

	kubeAPIServer, err := createKubeAPIServer(c, apiExtensionServer.GenericAPIServer)
	if err != nil {
		return nil, xerrors.Errorf("Create KubeAPI Server error: %w", err)
	}

	// aggregator comes last in the chain
	aggregatorConfig, err := createAggregatorConfig(*c.GenericConfig, *etcdOpt, apiExtensionsCfg.GenericConfig.SharedInformerFactory)
	if err != nil {
		return nil, xerrors.Errorf("Create Aggregator Config error: %w", err)
	}

	aggregatorServer, err := createAggregatorServer(aggregatorConfig, kubeAPIServer.GenericAPIServer, apiExtensionServer.Informers)
	if err != nil {
		return nil, xerrors.Errorf("Create Aggregator Server error: %w", err)
	}

	_, _, closeFn, err := startChainedAPIServer(c, aggregatorServer, s, h)
	if err != nil {
		return nil, xerrors.Errorf("start API server: %w", err)
	}

	shutdownFunc := func() {
		klog.Info("destroying API server")
		closeFn()
		s.Close()
		klog.Info("destroyed API server")
	}
	return shutdownFunc, nil
}

func createKubeAPIServer(kubeAPIServerConfig *controlplane.Config, delegateAPIServer genericapiserver.DelegationTarget) (*controlplane.Instance, error) {
	m, err := kubeAPIServerConfig.Complete().New(delegateAPIServer)
	if err != nil {
		// We log the error first so that even if closeFn crashes, the error is shown
		klog.Errorf("error in bringing up the apiserver: %v", err)
		return nil, err
	}
	return m, nil
}

func defaultOpenAPIConfig() *openapicommon.Config {
	openAPIConfig := genericapiserver.DefaultOpenAPIConfig(generated.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(legacyscheme.Scheme, extensionsapiserver.Scheme, aggregatorscheme.Scheme))
	openAPIConfig.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:   "Kubernetes",
			Version: "unversioned",
		},
	}
	openAPIConfig.DefaultResponse = &spec.Response{
		ResponseProps: spec.ResponseProps{
			Description: "Default Response.",
		},
	}
	openAPIConfig.GetDefinitions = generated.GetOpenAPIDefinitions

	return openAPIConfig
}

//nolint:funlen
func NewControlPlaneConfigWithOptions(serverURL, etcdURL string) (*controlplane.Config, *options.EtcdOptions) {
	etcdOptions := options.NewEtcdOptions(storagebackend.NewDefaultConfig(uuid.New().String(), nil))
	etcdOptions.StorageConfig.Transport.ServerList = []string{etcdURL}

	storageConfig := kubeapiserver.NewStorageFactoryConfig()
	storageConfig.APIResourceConfig = serverstorage.NewResourceConfig()
	completedStorageConfig, err := storageConfig.Complete(etcdOptions)
	if err != nil {
		panic(err)
	}
	storageFactory, err := completedStorageConfig.New()
	if err != nil {
		panic(err)
	}

	genericConfig := genericapiserver.NewConfig(legacyscheme.Codecs)
	kubeVersion := version.Get()
	if len(kubeVersion.Major) == 0 {
		kubeVersion.Major = "1"
	}
	if len(kubeVersion.Minor) == 0 {
		kubeVersion.Minor = "22"
	}
	genericConfig.Version = &kubeVersion

	genericConfig.SecureServing = &genericapiserver.SecureServingInfo{Listener: fakeLocalhost443Listener{}}

	err = etcdOptions.ApplyWithStorageFactoryTo(storageFactory, genericConfig)
	if err != nil {
		panic(err)
	}

	cfg := &controlplane.Config{
		GenericConfig: genericConfig,
		ExtraConfig: controlplane.ExtraConfig{
			APIResourceConfigSource: controlplane.DefaultAPIResourceConfigSource(),
			StorageFactory:          storageFactory,
			KubeletClientConfig:     kubeletclient.KubeletClientConfig{Port: 10250},
			APIServerServicePort:    443,
			MasterCount:             1,
		},
	}

	// set the loopback client config
	cfg.GenericConfig.LoopbackClientConfig = &restclient.Config{QPS: 50, Burst: 100, ContentConfig: restclient.ContentConfig{NegotiatedSerializer: legacyscheme.Codecs}}
	cfg.GenericConfig.LoopbackClientConfig.Host = serverURL

	privilegedLoopbackToken := uuid.New().String()
	// wrap any available authorizer
	tokens := make(map[string]*user.DefaultInfo)
	tokens[privilegedLoopbackToken] = &user.DefaultInfo{
		Name:   user.APIServerUser,
		UID:    uuid.New().String(),
		Groups: []string{user.SystemPrivilegedGroup},
	}

	tokenAuthenticator := authenticatorfactory.NewFromTokens(tokens, cfg.GenericConfig.Authentication.APIAudiences)
	cfg.GenericConfig.Authentication.Authenticator = authenticatorunion.New(tokenAuthenticator, authauthenticator.RequestFunc(alwaysEmpty))
	tokenAuthorizer := authorizerfactory.NewPrivilegedGroups(user.SystemPrivilegedGroup)
	cfg.GenericConfig.Authorization.Authorizer = authorizerunion.New(tokenAuthorizer, authorizerfactory.NewAlwaysAllowAuthorizer())

	cfg.GenericConfig.LoopbackClientConfig.BearerToken = privilegedLoopbackToken

	cfg.GenericConfig.PublicAddress = net.ParseIP("192.168.10.4")
	cfg.GenericConfig.SecureServing = &genericapiserver.SecureServingInfo{Listener: fakeLocalhost443Listener{}}

	cfg.GenericConfig.OpenAPIConfig = defaultOpenAPIConfig()

	return cfg, etcdOptions
}

type fakeLocalhost443Listener struct{}

func (fakeLocalhost443Listener) Accept() (net.Conn, error) {
	return nil, nil
}

func (fakeLocalhost443Listener) Close() error {
	return nil
}

func (fakeLocalhost443Listener) Addr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 443,
	}
}

// startAPIServer starts a kubernetes API server and an httpserver to handle api requests.
//nolint:funlen
func startChainedAPIServer(controlPlaneConfig *controlplane.Config, server *aggregatorapiserver.APIAggregator, s *httptest.Server, apiServerReceiver *APIServerHolder) (*controlplane.Instance, *httptest.Server, func(), error) {
	m := &controlplane.Instance{GenericAPIServer: server.GenericAPIServer}

	stopCh := make(chan struct{})
	closeFn := func() {
		if m != nil {
			if err := m.GenericAPIServer.RunPreShutdownHooks(); err != nil {
				klog.Errorf("failed to run pre-shutdown hooks for api server: %v", err)
			}
		}
		close(stopCh)
		s.Close()
	}
	apiServerReceiver.SetAPIServer(m)

	server.PrepareRun()
	m.GenericAPIServer.RunPostStartHooks(stopCh)

	cfg := *controlPlaneConfig.GenericConfig.LoopbackClientConfig
	cfg.ContentConfig.GroupVersion = &schema.GroupVersion{}
	privilegedClient, err := restclient.RESTClientFor(&cfg)
	if err != nil {
		closeFn()
		return nil, nil, nil, xerrors.Errorf("create restclient: %w", err)
	}

	var lastHealthContent []byte
	err = wait.PollImmediate(100*time.Millisecond, 30*time.Second, func() (bool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := privilegedClient.Get().AbsPath("/healthz").Do(ctx)
		status := 0
		result.StatusCode(&status)
		if status == 200 {
			return true, nil
		}
		lastHealthContent, _ = result.Raw()
		return false, nil
	})
	if err != nil {
		closeFn()
		klog.Errorf("last health content: %q", string(lastHealthContent))
		return nil, nil, nil, xerrors.Errorf("last health content: %w", err)
	}

	return m, s, closeFn, nil
}

// alwaysEmpty simulates "no authentication" for old tests.
func alwaysEmpty(_ *http.Request) (*authauthenticator.Response, bool, error) {
	return &authauthenticator.Response{
		User: &user.DefaultInfo{
			Name: "",
		},
	}, true, nil
}

// APIServerHolder implements.
type APIServerHolder struct {
	Initialized chan struct{}
	M           *controlplane.Instance
}

// SetAPIServer assigns the current API server.
func (h *APIServerHolder) SetAPIServer(m *controlplane.Instance) {
	h.M = m
	close(h.Initialized)
}
