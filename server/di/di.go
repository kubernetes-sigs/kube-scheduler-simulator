// Package di organizes the dependencies.
// All services are only initialized on this package.
// di means dependency injection.
package di

import (
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	v1beta3config "k8s.io/kube-scheduler/config/v1beta3"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/node"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/persistentvolume"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/persistentvolumeclaim"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/pod"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/priorityclass"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/replicateexistingcluster"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/reset"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/storageclass"
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
}

// NewDIContainer initializes Container.
// It initializes all service and puts to Container.
// If externalImportEnabled is false, the simulator will not use externalClient and will not create ReplicateExistingClusterService.
func NewDIContainer(client clientset.Interface, restclientCfg *restclient.Config, initialSchedulerCfg *v1beta3config.KubeSchedulerConfiguration, externalImportEnabled bool, externalClient clientset.Interface, externalRestClientCfg *restclient.Config) *Container {
	c := &Container{}

	// initializes each service
	c.pvService = persistentvolume.NewPersistentVolumeService(client)
	c.pvcService = persistentvolumeclaim.NewPersistentVolumeClaimService(client)
	c.storageClassService = storageclass.NewStorageClassService(client)
	c.schedulerService = scheduler.NewSchedulerService(client, restclientCfg, initialSchedulerCfg)
	c.podService = pod.NewPodService(client)
	c.nodeService = node.NewNodeService(client, c.podService)
	c.priorityClassService = priorityclass.NewPriorityClassService(client)

	deleteServices := map[string]reset.DeleteService{
		"node":                    c.nodeService,
		"pod":                     c.podService,
		"persistent volume":       c.pvService,
		"persistent volume claim": c.pvcService,
		"storage class":           c.storageClassService,
		"priority class":          c.priorityClassService,
	}
	c.resetService = reset.NewResetService(client, deleteServices, c.schedulerService)
	exportService := export.NewExportService(client, c.podService, c.nodeService, c.pvService, c.pvcService, c.storageClassService, c.priorityClassService, c.schedulerService)
	c.exportService = exportService
	if externalImportEnabled {
		existingClusterExportService := createExportServiceForReplicateExistingClusterService(externalClient, externalRestClientCfg)
		c.replicateExistingClusterService = replicateexistingcluster.NewReplicateExistingClusterService(exportService, existingClusterExportService)
	}
	return c
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

// createExportServiceForReplicateExistingClusterService creates each services
// that will be used for the ExportService for an existing cluster.
func createExportServiceForReplicateExistingClusterService(externalClient clientset.Interface, externalRestClientCfg *restclient.Config) *export.Service {
	pvService := persistentvolume.NewPersistentVolumeService(externalClient)
	pvcService := persistentvolumeclaim.NewPersistentVolumeClaimService(externalClient)
	storageClassService := storageclass.NewStorageClassService(externalClient)

	// ReplicateExistingClusterService will not use the SchedulerService of the existing cluster.
	// Therefore, this is ok to pass an empty struct.
	schedulerService := scheduler.NewSchedulerService(externalClient, externalRestClientCfg, &v1beta3config.KubeSchedulerConfiguration{})
	podService := pod.NewPodService(externalClient)
	nodeService := node.NewNodeService(externalClient, podService)
	priorityClassService := priorityclass.NewPriorityClassService(externalClient)
	return export.NewExportService(externalClient, podService, nodeService, pvService, pvcService, storageClassService, priorityClassService, schedulerService)
}
