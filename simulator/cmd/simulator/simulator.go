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
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/recorder"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/replayer"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourceapplier"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/di"
)

const (
	kubeAPIServerPollInterval = 5 * time.Second
	kubeAPIServerReadyTimeout = 2 * time.Minute
	importTimeout             = 2 * time.Minute
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
	dynamicClient := dynamic.NewForConfigOrDie(restCfg)
	discoverClient := discovery.NewDiscoveryClient(client.RESTClient())
	cachedDiscoveryClient := memory.NewMemCacheClient(discoverClient)
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscoveryClient)

	importClusterResourceClient := &clientset.Clientset{}
	var importClusterDynamicClient dynamic.Interface
	if cfg.ExternalImportEnabled || cfg.ResourceSyncEnabled || cfg.RecorderEnabled {
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = wait.PollUntilContextTimeout(ctx, kubeAPIServerPollInterval, kubeAPIServerReadyTimeout, true, func(ctx context.Context) (bool, error) {
		_, err := client.CoreV1().Namespaces().Get(context.Background(), "kube-system", metav1.GetOptions{})
		if err != nil {
			klog.Infof("waiting for kube-system namespace to be ready: %v", err)
			return false, nil
		}
		klog.Info("kubeapi-server is ready")
		return true, nil
	})
	if err != nil {
		return xerrors.Errorf("kubeapi-server is not ready: %w", err)
	}

	recorderOptions := recorder.Options{RecordDir: cfg.RecordFilePath}
	replayerOptions := replayer.Options{Path: cfg.RecordFilePath}
	resourceApplierOptions := resourceapplier.Options{}

	dic, err := di.NewDIContainer(client, dynamicClient, restMapper, etcdclient, restCfg, cfg.InitialSchedulerCfg, cfg.ExternalImportEnabled, cfg.ResourceSyncEnabled, cfg.RecorderEnabled, cfg.ReplayerEnabled, importClusterResourceClient, importClusterDynamicClient, cfg.Port, resourceApplierOptions, recorderOptions, replayerOptions)
	if err != nil {
		return xerrors.Errorf("create di container: %w", err)
	}

	// If ExternalImportEnabled is enabled, the simulator import resources
	// from the target cluster that indicated by the `KUBECONFIG`.
	if cfg.ExternalImportEnabled {
		// This must be called after `StartScheduler`
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, importTimeout)
		defer timeoutCancel()
		if err := dic.OneshotClusterResourceImporter().ImportClusterResources(timeoutCtx, cfg.ResourceImportLabelSelector); err != nil {
			return xerrors.Errorf("import from the target cluster: %w", err)
		}
	}

	// If ReplayEnabled is enabled, the simulator replays the recorded resources.
	if cfg.ReplayerEnabled {
		if err := dic.ReplayService().Replay(ctx); err != nil {
			return xerrors.Errorf("replay resources: %w", err)
		}
	}

	dic.SchedulerService().SetSchedulerConfig(cfg.InitialSchedulerCfg)

	if cfg.ResourceSyncEnabled {
		// Start the resource syncer to sync resources from the target cluster.
		if err = dic.ResourceSyncer().Run(ctx); err != nil {
			return xerrors.Errorf("start syncing: %w", err)
		}
	}

	if cfg.RecorderEnabled {
		// Start the recorder to record events in the target cluster.
		if err = dic.RecorderService().Run(ctx); err != nil {
			return xerrors.Errorf("start recording: %w", err)
		}
	}

	// start simulator server
	s := server.NewSimulatorServer(cfg, dic)
	shutdownFn, err := s.Start(cfg.Port)
	if err != nil {
		return xerrors.Errorf("start simulator server: %w", err)
	}
	defer shutdownFn()

	// wait the signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	return nil
}
