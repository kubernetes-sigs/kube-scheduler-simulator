// Package di organizes the dependencies.
// All services are only initialized on this package.
// di means dependency injection.
package di

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	configv1 "k8s.io/kube-scheduler/config/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/export"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/node"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/persistentvolume"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/persistentvolumeclaim"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/pod"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/priorityclass"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/replicateexistingcluster"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/reset"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/storageclass"
)

// Container saves and provides dependencies.
type Container struct {
	nodeService                     NodeService
	podService                      PodService
	pvService                       PersistentVolumeService
	pvcService                      PersistentVolumeClaimService
	storageClassService             StorageClassService
	schedulerService                SchedulerService
	exportService                   ExportService
	priorityClassService            PriorityClassService
	resetService                    ResetService
	replicateExistingClusterService ReplicateExistingClusterService
	resourceWatcherService          ResourceWatcherService
}

// NewDIContainer initializes Container.
// It initializes all service and puts to Container.
// If externalImportEnabled is false, the simulator will not use externalClient and will not create ReplicateExistingClusterService.
func NewDIContainer(
	client clientset.Interface,
	etcdclient *clientv3.Client,
	restclientCfg *restclient.Config,
	initialSchedulerCfg *configv1.KubeSchedulerConfiguration,
	externalImportEnabled bool,
	externalClient clientset.Interface,
	externalSchedulerEnabled bool,
	simulatorPort int,
) (*Container, error) {
	c := &Container{}

	// initializes each service
	c.pvService = persistentvolume.NewPersistentVolumeService(client)
	c.pvcService = persistentvolumeclaim.NewPersistentVolumeClaimService(client)
	c.storageClassService = storageclass.NewStorageClassService(client)
	c.schedulerService = scheduler.NewSchedulerService(client, restclientCfg, initialSchedulerCfg, externalSchedulerEnabled, simulatorPort)
	c.podService = pod.NewPodService(client)
	c.nodeService = node.NewNodeService(client, c.podService)
	c.priorityClassService = priorityclass.NewPriorityClassService(client)
	var err error
	c.resetService, err = reset.NewResetService(etcdclient, client, c.schedulerService)
	if err != nil {
		return nil, xerrors.Errorf("initialize reset service: %w", err)
	}
	exportService := export.NewExportService(client, c.podService, c.nodeService, c.pvService, c.pvcService, c.storageClassService, c.priorityClassService, c.schedulerService)
	c.exportService = exportService
	if externalImportEnabled {
		existingClusterExportService := createExportServiceForReplicateExistingClusterService(externalClient, c.schedulerService)
		c.replicateExistingClusterService = replicateexistingcluster.NewReplicateExistingClusterService(exportService, existingClusterExportService)
	}
	c.resourceWatcherService = resourcewatcher.NewService(client)

	return c, nil
}

// NodeService returns NodeService.
func (c *Container) NodeService() NodeService {
	return c.nodeService
}

// PodService returns PodService.
func (c *Container) PodService() PodService {
	return c.podService
}

// StorageClassService returns StorageClassService.
func (c *Container) StorageClassService() StorageClassService {
	return c.storageClassService
}

// PersistentVolumeService returns PersistentVolumeService.
func (c *Container) PersistentVolumeService() PersistentVolumeService {
	return c.pvService
}

// PersistentVolumeClaimService returns PersistentVolumeClaimService.
func (c *Container) PersistentVolumeClaimService() PersistentVolumeClaimService {
	return c.pvcService
}

// SchedulerService returns SchedulerService.
func (c *Container) SchedulerService() SchedulerService {
	return c.schedulerService
}

// PriorityClassService returns PriorityClassService.
func (c *Container) PriorityClassService() PriorityClassService {
	return c.priorityClassService
}

// ExportService returns ExportService.
func (c *Container) ExportService() ExportService {
	return c.exportService
}

// ResetService returns ResetService.
func (c *Container) ResetService() ResetService {
	return c.resetService
}

// ReplicateExistingClusterService returns ReplicateExistingClusterService.
// Note: this service will return nil when `externalImportEnabled` is false.
func (c *Container) ReplicateExistingClusterService() ReplicateExistingClusterService {
	return c.replicateExistingClusterService
}

// ResourceWatcherService returns ResourceWatcherService.
func (c *Container) ResourceWatcherService() ResourceWatcherService {
	return c.resourceWatcherService
}

// ExtenderService returns ExtenderService.
func (c *Container) ExtenderService() ExtenderService {
	return c.schedulerService.ExtenderService()
}

// createExportServiceForReplicateExistingClusterService creates each services
// that will be used for the ExportService for an existing cluster.
func createExportServiceForReplicateExistingClusterService(externalClient clientset.Interface, schedulerService SchedulerService) *export.Service {
	pvService := persistentvolume.NewPersistentVolumeService(externalClient)
	pvcService := persistentvolumeclaim.NewPersistentVolumeClaimService(externalClient)
	storageClassService := storageclass.NewStorageClassService(externalClient)

	podService := pod.NewPodService(externalClient)
	nodeService := node.NewNodeService(externalClient, podService)
	priorityClassService := priorityclass.NewPriorityClassService(externalClient)
	return export.NewExportService(externalClient, podService, nodeService, pvService, pvcService, storageClassService, priorityClassService, schedulerService)
}
