package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/config"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/k8sapiserver"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/pvcontroller"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

// entry point.
func main() {
	if err := startSimulator(); err != nil {
		klog.Fatalf("failed with error on running simulator: %+v", err)
	}
}

// startSimulator starts simulator and needed k8s components.
func startSimulator() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return xerrors.Errorf("get config: %w", err)
	}

	restclientCfg, apiShutdown, err := k8sapiserver.StartAPIServer(cfg.KubeAPIServerURL, cfg.EtcdURL)
	if err != nil {
		return xerrors.Errorf("start API server: %w", err)
	}
	defer apiShutdown()

	client := clientset.NewForConfigOrDie(restclientCfg)

	pvshutdown, err := pvcontroller.StartPersistentVolumeController(client)
	if err != nil {
		return xerrors.Errorf("start pv controller: %w", err)
	}
	defer pvshutdown()

	existingClusterClient := &clientset.Clientset{}
	if cfg.ExternalImportEnabled {
		existingClusterClient, err = clientset.NewForConfig(cfg.ExternalKubeClientCfg)
		if err != nil {
			return xerrors.Errorf("creates a new Clientset for the ExternalKubeClientCfg: %w", err)
		}
	}

	dic := di.NewDIContainer(client, restclientCfg, cfg.InitialSchedulerCfg, cfg.ExternalImportEnabled, existingClusterClient, cfg.ExternalKubeClientCfg)

	if err := dic.SchedulerService().StartScheduler(cfg.InitialSchedulerCfg); err != nil {
		return xerrors.Errorf("start scheduler: %w", err)
	}
	defer dic.SchedulerService().ShutdownScheduler()

	// If ExternalImportEnabled is enabled, the simulator import resources
	// from the existing cluster that indicated by the `KUBECONFIG`.
	if cfg.ExternalImportEnabled {
		ctx := context.Background()
		// This must be called after `StartScheduler`
		if err := dic.ReplicateExistingClusterService().ImportFromExistingCluster(ctx); err != nil {
			return xerrors.Errorf("import existing cluster: %w", err)
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
