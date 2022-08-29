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
  // ControllerStartJitter is the Jitter used when starting controller.
  ControllerStartJitter = 1.0
)

type InitFunc func(controllerCtx ControllerContext) error

// ControllerInitializersFunc is used to create a collection of initializers
// given the loopMode.
//nolint:revive // Intentionally used the same name.
type ControllerInitializersFunc func() (initializers map[string]InitFunc)

var _ ControllerInitializersFunc = NewControllerInitializers

func NewController(client clientset.Interface, c *restclient.Config) (func(), error) {
  ctx, cancel := context.WithCancel(context.Background())

  go Run(client, c, ctx.Done())
  shutdownFunc := func() {
    cancel()
  }

  return shutdownFunc, nil
}

// Run runs the KubeControllerManagerOptions.  This should never exit.
func Run(client clientset.Interface, config *restclient.Config, stopCh <-chan struct{}) {
  run := func(ctx context.Context, initializersFunc ControllerInitializersFunc) {
    controllerContext, err := CreateControllerContext(client, config, ctx.Done())
    if err != nil {
      klog.Fatalf("error building controller context: %v", err)
    }
    controllerInitializers := initializersFunc()
    if err := StartControllers(controllerContext, controllerInitializers); err != nil {
      klog.Fatalf("error starting controllers: %v", err)
    }

    controllerContext.InformerFactory.Start(stopCh)
    controllerContext.ObjectOrMetadataInformerFactory.Start(stopCh)

    close(controllerContext.InformersStarted)

    select {}
  }
  run(context.Background(), NewControllerInitializers)
  panic("unreachable")
}

// StartControllers starts a set of controllers with a specified ControllerContext.
func StartControllers(ctx ControllerContext, controllers map[string]InitFunc) error {
  for controllerName, initFn := range controllers {
    time.Sleep(wait.Jitter(ctx.ComponentConfig.Generic.ControllerStartInterval.Duration, ControllerStartJitter))

    klog.Infof("Starting %q", controllerName)
    err := initFn(ctx)
    if err != nil {
      klog.Errorf("Error starting %q", controllerName)
      return xerrors.Errorf("starting %v: %w", controllerName, err)
    }
    klog.Infof("Started %q", controllerName)
  }
  return nil
}

// NewControllerInitializers is a public map of named controller groups paired to their InitFunc.
// This allows for structured downstream composition and subdivision.
func NewControllerInitializers() map[string]InitFunc {
  controllers := map[string]InitFunc{}
  controllers["deployment"] = StartDeploymentController
  controllers["replicaset"] = StartReplicaSetController
  controllers["persistent-volume"] = StartPersistentVolumeController
  return controllers
}

// ControllerContext defines the context object for controller.
//nolint:revive // Intentionally used the same name.
type ControllerContext struct {
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

  // Stop is the stop channel
  Stop <-chan struct{}

  // InformersStarted is closed after all of the controllers have been initialized and are running.  After this point it is safe,
  // for an individual controller to start the shared informers. Before it is closed, they should not.
  InformersStarted chan struct{}

  // ResyncPeriod generates a duration each time it is invoked; this is so that
  // multiple controllers don't get into lock-step and all hammer the apiserver
  // with list requests simultaneously.
  ResyncPeriod func() time.Duration
}

// CreateControllerContext creates a context struct containing references to resources needed by the controllers.
func CreateControllerContext(client clientset.Interface, config *restclient.Config, stop <-chan struct{}) (ControllerContext, error) {
  clientbuilder := clientbuilder.SimpleControllerClientBuilder{
    ClientConfig: config,
  }
  componentConfig, err := controllermanageroptions.NewDefaultComponentConfig()
  if err != nil {
    return ControllerContext{}, xerrors.Errorf("new default component config: %w", err)
  }
  sharedInformers := informers.NewSharedInformerFactory(client, ResyncPeriod(componentConfig)())

  metadataClient := metadata.NewForConfigOrDie(clientbuilder.ConfigOrDie("metadata-informers"))
  metadataInformers := metadatainformer.NewSharedInformerFactory(metadataClient, ResyncPeriod(componentConfig)())

  ctx := ControllerContext{
    ClientBuilder:                   clientbuilder,
    ComponentConfig:                 componentConfig,
    InformerFactory:                 sharedInformers,
    ObjectOrMetadataInformerFactory: informerfactory.NewInformerFactory(sharedInformers, metadataInformers),
    Stop:                            stop,
    InformersStarted:                make(chan struct{}),
    ResyncPeriod:                    ResyncPeriod(componentConfig),
  }
  return ctx, nil
}

// ResyncPeriod returns a function which generates a duration each time it is
// invoked; this is so that multiple controllers don't get into lock-step and all
// hammer the apiserver with list requests simultaneously.
func ResyncPeriod(c kubectrlmgrconfig.KubeControllerManagerConfiguration) func() time.Duration {
  return func() time.Duration {
    //nolint:gosec // Same usage as kubernetes
    factor := rand.Float64() + 1
    return time.Duration(float64(c.Generic.MinResyncPeriod.Nanoseconds()) * factor)
  }
}
