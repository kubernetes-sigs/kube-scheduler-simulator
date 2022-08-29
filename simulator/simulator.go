package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/controller"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/k8sapiserver"
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
func startSimulator() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return xerrors.Errorf("get config: %w", err)
	}

	restclientCfg, apiShutdown, err := k8sapiserver.StartAPIServer(cfg.KubeAPIServerURL, cfg.EtcdURL, cfg.CorsAllowedOriginList)
	if err != nil {
		return xerrors.Errorf("start API server: %w", err)
	}
	defer apiShutdown()

	client := clientset.NewForConfigOrDie(restclientCfg)

	ctrlerShutdown, err := controller.RunController(client, restclientCfg)
	if err != nil {
		return xerrors.Errorf("start controllers: %w", err)
	}
	defer ctrlerShutdown()

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
