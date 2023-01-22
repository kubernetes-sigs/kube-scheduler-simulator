package controller

import (
	"context"
	"math/rand"
	"time"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	restclient "k8s.io/client-go/rest"
	"k8s.io/controller-manager/pkg/clientbuilder"
	"k8s.io/controller-manager/pkg/informerfactory"
	"k8s.io/klog/v2"
	controllermanageroptions "k8s.io/kubernetes/cmd/kube-controller-manager/app/options"
	kubectrlmgrconfig "k8s.io/kubernetes/pkg/controller/apis/config"
)

const (
	// controllerStartJitter is the Jitter used when starting controller.
	controllerStartJitter = 1.0
)

// initFunc is the func to start a controller.
type initFunc func(ctx context.Context, controllerCtx controllerContext) error

// RunControllers runs all controllers.
func RunController(client clientset.Interface, cfg *restclient.Config) (func(), error) {
	controllerCtx, err := createControllerContext(client, cfg)
	if err != nil {
		return nil, xerrors.Errorf("building controller context: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go run(controllerCtx, ctx.Done())
	shutdownFunc := func() {
		klog.Info("shutdown controllers...")
		cancel()
	}

	return shutdownFunc, nil
}

// run runs the KubeControllerManagerOptions.
func run(controllerCtx controllerContext, stopCh <-chan struct{}) {
	controllerInitializers := newControllerInitializers()
	if err := startControllers(context.Background(), controllerCtx, controllerInitializers); err != nil {
		klog.Fatalf("error starting controllers: %v", err)
	}
	controllerCtx.InformerFactory.Start(stopCh)
	controllerCtx.ObjectOrMetadataInformerFactory.Start(stopCh)

	close(controllerCtx.InformersStarted)
}

// startControllers starts a set of controllers with a specified controllerContext.
func startControllers(ctx context.Context, contrllerCtx controllerContext, controllers map[string]initFunc) error {
	for controllerName, initFn := range controllers {
		time.Sleep(wait.Jitter(contrllerCtx.ComponentConfig.Generic.ControllerStartInterval.Duration, controllerStartJitter))

		klog.Infof("Starting %q", controllerName)
		err := initFn(ctx, contrllerCtx)
		if err != nil {
			klog.Errorf("Error starting %q", controllerName)
			return xerrors.Errorf("starting %v: %w", controllerName, err)
		}
		klog.Infof("Started %q", controllerName)
	}
	return nil
}

// newControllerInitializers is a public map of named controller groups paired to their initFunc.
// This allows for structured downstream composition and subdivision.
func newControllerInitializers() map[string]initFunc {
	controllers := map[string]initFunc{}
	controllers["deployment"] = startDeploymentController
	controllers["replicaset"] = startReplicaSetController
	controllers["persistent-volume"] = startPersistentVolumeController
	return controllers
}

// controllerContext defines the context object for controller.
type controllerContext struct {
	// ClientBuilder will provide a client for this controller to use
	ClientBuilder clientbuilder.ControllerClientBuilder

	ComponentConfig kubectrlmgrconfig.KubeControllerManagerConfiguration

	// InformerFactory gives access to informers for the controller.
	InformerFactory informers.SharedInformerFactory

	// ObjectOrMetadataInformerFactory gives access to informers for typed resources
	// and dynamic resources by their metadata. All generic controllers currently use
	// object metadata - if a future controller needs access to the full object this
	// would become GenericInformerFactory and take a dynamic client.
	ObjectOrMetadataInformerFactory informerfactory.InformerFactory

	// InformersStarted is closed after all of the controllers have been initialized and are running.  After this point it is safe,
	// for an individual controller to start the shared informers. Before it is closed, they should not.
	InformersStarted chan struct{}

	// ResyncPeriod generates a duration each time it is invoked; this is so that
	// multiple controllers don't get into lock-step and all hammer the apiserver
	// with list requests simultaneously.
	ResyncPeriod func() time.Duration
}

// createControllerContext creates a context struct containing references to resources needed by the controllers.
func createControllerContext(client clientset.Interface, config *restclient.Config) (controllerContext, error) {
	clientbuilder := clientbuilder.SimpleControllerClientBuilder{
		ClientConfig: config,
	}
	componentConfig, err := controllermanageroptions.NewDefaultComponentConfig()
	if err != nil {
		return controllerContext{}, xerrors.Errorf("new default component config: %w", err)
	}
	sharedInformers := informers.NewSharedInformerFactory(client, resyncPeriod(componentConfig)())

	metadataClient := metadata.NewForConfigOrDie(clientbuilder.ConfigOrDie("metadata-informers"))
	metadataInformers := metadatainformer.NewSharedInformerFactory(metadataClient, resyncPeriod(componentConfig)())

	ctx := controllerContext{
		ClientBuilder:                   clientbuilder,
		ComponentConfig:                 componentConfig,
		InformerFactory:                 sharedInformers,
		ObjectOrMetadataInformerFactory: informerfactory.NewInformerFactory(sharedInformers, metadataInformers),
		InformersStarted:                make(chan struct{}),
		ResyncPeriod:                    resyncPeriod(componentConfig),
	}
	return ctx, nil
}

// resyncPeriod returns a function which generates a duration each time it is
// invoked; this is so that multiple controllers don't get into lock-step and all
// hammer the apiserver with list requests simultaneously.
func resyncPeriod(c kubectrlmgrconfig.KubeControllerManagerConfiguration) func() time.Duration {
	return func() time.Duration {
		//nolint:gosec // Same usage as kubernetes
		factor := rand.Float64() + 1
		return time.Duration(float64(c.Generic.MinResyncPeriod.Nanoseconds()) * factor)
	}
}
