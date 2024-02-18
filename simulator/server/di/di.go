// Package di organizes the dependencies.
// All services are only initialized on this package.
// di means dependency injection.
package di

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	configv1 "k8s.io/kube-scheduler/config/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/clusterresourceimporter/oneshotimporter"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/clusterresourceimporter/syncer"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/reset"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/snapshot"
)

// Container saves and provides dependencies.
type Container struct {
	schedulerService               SchedulerService
	snapshotService                SnapshotService
	resetService                   ResetService
	oneshotClusterResourceImporter OneShotClusterResourceImporter
	resourceSyncer                 ResourceSyncer
	resourceWatcherService         ResourceWatcherService
}

// NewDIContainer initializes Container.
// It initializes all service and puts to Container.
// Only when externalImportEnabled is true, the simulator uses externalClient and creates ImportClusterResourceService.
func NewDIContainer(
	client clientset.Interface,
	dynamicClient dynamic.Interface,
	discoveryCli discovery.DiscoveryInterface,
	etcdclient *clientv3.Client,
	restclientCfg *restclient.Config,
	initialSchedulerCfg *configv1.KubeSchedulerConfiguration,
	externalImportEnabled bool,
	resourceSyncEnabled bool,
	externalClient clientset.Interface,
	externalDynamicClient dynamic.Interface,
	externalSchedulerEnabled bool,
	simulatorPort int,
) (*Container, error) {
	c := &Container{}

	// initializes each service
	c.schedulerService = scheduler.NewSchedulerService(client, restclientCfg, initialSchedulerCfg, externalSchedulerEnabled, simulatorPort)
	var err error
	c.resetService, err = reset.NewResetService(etcdclient, client, c.schedulerService)
	if err != nil {
		return nil, xerrors.Errorf("initialize reset service: %w", err)
	}
	snapshotSvc := snapshot.NewService(client, c.schedulerService)
	c.snapshotService = snapshotSvc
	if externalImportEnabled {
		extSnapshotSvc := snapshot.NewService(externalClient, c.schedulerService)
		c.oneshotClusterResourceImporter = oneshotimporter.NewService(snapshotSvc, extSnapshotSvc)
	}
	if resourceSyncEnabled {
		c.resourceSyncer = syncer.New(externalDynamicClient, dynamicClient, discoveryCli)
	}
	c.resourceWatcherService = resourcewatcher.NewService(client)

	return c, nil
}

// SchedulerService returns SchedulerService.
func (c *Container) SchedulerService() SchedulerService {
	return c.schedulerService
}

// ExportService returns ExportService.
func (c *Container) ExportService() SnapshotService {
	return c.snapshotService
}

// ResetService returns ResetService.
func (c *Container) ResetService() ResetService {
	return c.resetService
}

// OneshotClusterResourceImporter returns OneshotClusterResourceImporter.
// Note: this service will return nil when `externalImportEnabled` is false.
func (c *Container) OneshotClusterResourceImporter() OneShotClusterResourceImporter {
	return c.oneshotClusterResourceImporter
}

// ResourceSyncer returns ResourceSyncer.
func (c *Container) ResourceSyncer() ResourceSyncer {
	return c.resourceSyncer
}

// ResourceWatcherService returns ResourceWatcherService.
func (c *Container) ResourceWatcherService() ResourceWatcherService {
	return c.resourceWatcherService
}

// ExtenderService returns ExtenderService.
func (c *Container) ExtenderService() ExtenderService {
	return c.schedulerService.ExtenderService()
}
