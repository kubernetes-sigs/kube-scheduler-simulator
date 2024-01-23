package k8sapiserver

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"

	"golang.org/x/xerrors"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/kube-aggregator/pkg/apiserver"
	apiserverapp "k8s.io/kubernetes/cmd/kube-apiserver/app"
	apiserverappopts "k8s.io/kubernetes/cmd/kube-apiserver/app/options"
	"k8s.io/kubernetes/pkg/controlplane"
	authzmodes "k8s.io/kubernetes/pkg/kubeapiserver/authorizer/modes"
	apiserveropts "k8s.io/kubernetes/pkg/kubeapiserver/options"
)

// This key is for testing purposes only and is not considered secure.
//
//nolint:gosec
const ecdsaPrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIEZmTmUhuanLjPA2CLquXivuwBDHTt5XYwgIr/kA1LtRoAoGCCqGSM49
AwEHoUQDQgAEH6cuzP8XuD5wal6wf9M6xDljTOPLX2i8uIp/C/ASqiIGUeeKQtX0
/IR3qCXyThP/dbCiHrF3v1cuhBOHY8CLVg==
-----END EC PRIVATE KEY-----`

// StartAPIServer starts both the secure k8sAPIServer and proxy server to handle insecure serving, and it make panic when a error happen.
func StartAPIServer(kubeAPIServerURL, etcdURL string, corsAllowedOriginList []string) (*restclient.Config, func(), error) {
	h := &APIServerHolder{Initialized: make(chan struct{})}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.Header.Set("User-Agent", restclient.DefaultKubernetesUserAgent())
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	})

	l, err := net.Listen("tcp", kubeAPIServerURL)
	if err != nil {
		return nil, nil, xerrors.Errorf("announces on the local network address: %w", err)
	}

	s := &httptest.Server{
		Listener: l,
		//nolint:gosec // we don't need to care about the slowloris attack here.
		Config: &http.Server{
			Handler: handler,
		},
	}
	s.Start()
	klog.InfoS("starting proxy server", "URL", s.URL)

	aggregatorServer, cleanUpFunc, err := createK8SAPIChainedServer(etcdURL, corsAllowedOriginList)
	if err != nil {
		return nil, nil, xerrors.Errorf("start k8s api chained server: %w", err)
	}

	closeFn, err := setUpHandlerAndRun(aggregatorServer, s, h)
	if err != nil {
		return nil, nil, xerrors.Errorf("run aggregator server: %w", err)
	}

	err = createClusterRoleAndRoleBindings(aggregatorServer.GenericAPIServer.LoopbackClientConfig)
	if err != nil {
		return nil, nil, err
	}

	shutdownFunc := func() {
		klog.Info("destroying API server")
		cleanUpFunc()
		closeFn()
		s.Close()
		klog.Info("destroyed API server")
	}

	cfg := &restclient.Config{
		Host:          s.URL,
		ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}},
		QPS:           5000.0,
		Burst:         5000,
	}

	return cfg, shutdownFunc, nil
}

func setUpHandlerAndRun(aggregatorServer *apiserver.APIAggregator, s *httptest.Server, h *APIServerHolder) (func(), error) {
	var m *controlplane.Instance
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
	m = &controlplane.Instance{
		GenericAPIServer: aggregatorServer.GenericAPIServer,
	}
	h.SetAPIServer(m)

	prepared, err := aggregatorServer.PrepareRun()
	if err != nil {
		return nil, err
	}

	go func() {
		if err := prepared.Run(stopCh); err != nil {
			klog.Errorf("run aggregator server: %v", err)
		}
	}()

	return closeFn, nil
}

const EtcdPrefix = "kube-scheduler-simulator/"

func createK8SAPIServerOpts(etcdURL string, corsAllowedOriginList []string) (*apiserverappopts.ServerRunOptions, func(), error) {
	serverOpts := apiserverappopts.NewServerRunOptions()

	// Allow privileged containers.
	serverOpts.AllowPrivileged = true

	// set up etcd
	serverOpts.Etcd.StorageConfig.Transport.ServerList = []string{etcdURL}
	serverOpts.Etcd.StorageConfig.Prefix = EtcdPrefix

	// set up CORS
	serverOpts.GenericServerRunOptions.CorsAllowedOriginList = corsAllowedOriginList

	// set up RBAC authorization and annoymous auth.
	serverOpts.Authorization.Modes = []string{authzmodes.ModeRBAC}
	serverOpts.Authentication.Anonymous.Allow = true
	err := serverOpts.APIEnablement.RuntimeConfig.Set("api/all=true")
	if err != nil {
		return nil, nil, xerrors.Errorf("k8s api server set runtime config: %w", err)
	}

	// setup fake key for secure serving
	saSigningKeyFile, err := os.CreateTemp("/tmp", "insecure_test_key")
	if err != nil {
		return nil, nil, xerrors.Errorf("create temp file failed: %v", err)
	}

	if err = os.WriteFile(saSigningKeyFile.Name(), []byte(ecdsaPrivateKey), 0o600); err != nil {
		return nil, nil, xerrors.Errorf("write file %s failed: %v", saSigningKeyFile.Name(), err)
	}
	serverOpts.ServiceAccountSigningKeyFile = saSigningKeyFile.Name()
	serverOpts.Authentication.ServiceAccounts.Issuers = []string{"https://foo.bar.example.com"}
	serverOpts.Authentication.ServiceAccounts.KeyFiles = []string{saSigningKeyFile.Name()}

	// disable admission plugins to avoid node taints, service account, cert approval etc.
	serverOpts.Admission.GenericAdmission.DisablePlugins = apiserveropts.AllOrderedPlugins
	serverOpts.Admission.GenericAdmission.EnablePlugins = []string{
		// This plugin is needed to use PriorityClass.
		// https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#priority
		"Priority",
	}

	// add cert directory to avoid permission error
	certDir, err := os.MkdirTemp("", "apiserver-certs")
	if err != nil {
		return nil, nil, xerrors.Errorf("create temp certificate dir: %w", err)
	}
	serverOpts.SecureServing.ServerCert.CertDirectory = certDir

	cleanupFunc := func() {
		os.RemoveAll(saSigningKeyFile.Name())
		os.RemoveAll(certDir)
	}

	return serverOpts, cleanupFunc, nil
}

func createK8SAPIChainedServer(etcdURL string, corsAllowedOriginList []string) (*apiserver.APIAggregator, func(), error) {
	serverOpts, cleanupFunc, err := createK8SAPIServerOpts(etcdURL, corsAllowedOriginList)
	if err != nil {
		return nil, nil, err
	}

	completedOpts, err := serverOpts.Complete()
	if err != nil {
		return nil, nil, xerrors.Errorf("complete k8s api server options: %w", err)
	}

	config, err := apiserverapp.NewConfig(completedOpts)
	if err != nil {
		return nil, nil, err
	}
	completed, err := config.Complete()
	if err != nil {
		return nil, nil, err
	}

	aggregatorServer, err := apiserverapp.CreateServerChain(completed)
	if err != nil {
		return nil, nil, err
	}

	return aggregatorServer, cleanupFunc, nil
}

func createClusterRoleAndRoleBindings(loopbackCfg *restclient.Config) error {
	client, err := kubernetes.NewForConfig(loopbackCfg)
	if err != nil {
		return xerrors.Errorf("failed to create a client: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create ClusterRoles and ClusterRoleBindings for annoymous user
	// so we can query the proxy server without any cert.
	cr, err := client.RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "annoymoususerrole"},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs: []string{
					"*",
				},
				APIGroups: []string{
					"*",
				},
				Resources: []string{
					"*",
				},
			}, {
				NonResourceURLs: []string{
					"*",
				},
				Verbs: []string{
					"*",
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return xerrors.Errorf("create RBAC cluster roles: %w", err)
	}

	_, err = client.RbacV1().ClusterRoleBindings().Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "annoymoususerrolebinding"},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.UserKind,
				APIGroup:  rbacv1.GroupName,
				Name:      user.Anonymous,
				Namespace: "",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     cr.Name,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		return xerrors.Errorf("create RBAC cluster role bindings: %w", err)
	}

	return nil
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
