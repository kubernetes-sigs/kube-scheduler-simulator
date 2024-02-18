package app

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/di"
)

// startSimulator starts simulator and needed k8s components.
// It should be called from the entry point basically, or from the integration test.
//
//nolint:funlen,cyclop
func StartSimulator(ctx context.Context, cfg *config.Config) error {
	restCfg := &rest.Config{
		Host: cfg.KubeAPIServerURL,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	client := clientset.NewForConfigOrDie(restCfg)
	dynamicClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return xerrors.Errorf("creates dynamic clientset: %w", err)
	}
	discoveryClient := discovery.NewDiscoveryClient(client.RESTClient())

	importClusterResourceClient := &clientset.Clientset{}
	var importClusterDynamicClient dynamic.Interface
	if cfg.ExternalImportEnabled || cfg.ResourceSyncEnabled {
		importClusterResourceClient, err = clientset.NewForConfig(cfg.ExternalKubeClientCfg)
		if err != nil {
			return xerrors.Errorf("creates a new Clientset for the ExternalKubeClientCfg: %w", err)
		}

		importClusterDynamicClient, err = dynamic.NewForConfig(cfg.ExternalKubeClientCfg)
		if err != nil {
			return xerrors.Errorf("creates a new dynamic Clientset for the ExternalKubeClientCfg: %w", err)
		}
	}

	etcdclient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{cfg.EtcdURL},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return xerrors.Errorf("create an etcd client: %w", err)
	}

	// need to sleep here to make all controllers create initial resources. (like "system-" priorityclass.)
	if _, err := client.CoreV1().Namespaces().Get(context.Background(), "default", metav1.GetOptions{}); err != nil {
		return xerrors.Errorf("get kube-system namespace: %w", err)
	}

	dic, err := di.NewDIContainer(client, dynamicClient, discoveryClient, etcdclient, restCfg, cfg.InitialSchedulerCfg, cfg.ExternalImportEnabled, cfg.ResourceSyncEnabled, importClusterResourceClient, importClusterDynamicClient, cfg.ExternalSchedulerEnabled, cfg.Port)
	if err != nil {
		return xerrors.Errorf("create di container: %w", err)
	}
	if !cfg.ExternalSchedulerEnabled {
		if err := dic.SchedulerService().StartScheduler(cfg.InitialSchedulerCfg); err != nil {
			return xerrors.Errorf("start scheduler: %w", err)
		}
		defer dic.SchedulerService().ShutdownScheduler()
	}

	// If ExternalImportEnabled is enabled, the simulator import resources
	// from the target cluster that indicated by the `KUBECONFIG`.
	if cfg.ExternalImportEnabled {
		// This must be called after `StartScheduler`
		if err := dic.OneshotClusterResourceImporter().ImportClusterResources(ctx); err != nil {
			return xerrors.Errorf("import from the target cluster: %w", err)
		}
	}

	if cfg.ResourceSyncEnabled {
		// Start the resource syncer to sync resources from the target cluster.
		go dic.ResourceSyncer().Run(ctx)
	}

	// start simulator server
	s := server.NewSimulatorServer(cfg, dic)
	shutdownFn3, err := s.Start(cfg.Port)
	if err != nil {
		return xerrors.Errorf("start simulator server: %w", err)
	}
	defer shutdownFn3()

	// Block until ctx is canceled
	<-ctx.Done()

	return nil
}
