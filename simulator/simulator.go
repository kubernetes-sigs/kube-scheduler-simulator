package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/di"
)

// entry point.
func main() {
	if err := startSimulator(); err != nil {
		klog.Fatalf("failed with error on running simulator: %+v", err)
	}
}

// startSimulator starts simulator and needed k8s components.
//
//nolint:funlen,cyclop
func startSimulator() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return xerrors.Errorf("get config: %w", err)
	}

	restCfg := &rest.Config{
		Host: cfg.KubeAPIServerURL,
	}
	client := clientset.NewForConfigOrDie(restCfg)

	importClusterResourceClient := &clientset.Clientset{}
	if cfg.ExternalImportEnabled {
		importClusterResourceClient, err = clientset.NewForConfig(cfg.ExternalKubeClientCfg)
		if err != nil {
			return xerrors.Errorf("creates a new Clientset for the ExternalKubeClientCfg: %w", err)
		}
	}

	etcdclient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{cfg.EtcdURL},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return xerrors.Errorf("create an etcd client: %w", err)
	}

	if _, err := client.CoreV1().Namespaces().Get(context.Background(), "kube-system", metav1.GetOptions{}); err != nil {
		return xerrors.Errorf("get kube-system namespace: %w", err)
	}

	dic, err := di.NewDIContainer(client, etcdclient, restCfg, cfg.InitialSchedulerCfg, cfg.ExternalImportEnabled, importClusterResourceClient, cfg.ExternalSchedulerEnabled, cfg.Port)
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
		ctx := context.Background()
		// This must be called after `StartScheduler`
		if err := dic.ImportClusterResourceService().ImportClusterResources(ctx); err != nil {
			return xerrors.Errorf("import from the target cluster: %w", err)
		}
	}

	// start simulator server
	s := server.NewSimulatorServer(cfg, dic)
	shutdownFn3, err := s.Start(cfg.Port)
	if err != nil {
		return xerrors.Errorf("start simulator server: %w", err)
	}
	defer shutdownFn3()

	// wait the signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	return nil
}
