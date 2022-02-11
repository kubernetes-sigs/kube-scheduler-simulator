// Package di organizes the dependencies.
// All services are only initialized on this package.
// di means dependency injection.
package di

import (
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/node"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/persistentvolume"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/persistentvolumeclaim"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/pod"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/priorityclass"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/reset"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/storageclass"
)

// Container saves and provides dependencies.
type Container struct {
	nodeService          NodeService
	podService           PodService
	pvService            PersistentVolumeService
	pvcService           PersistentVolumeClaimService
	storageClassService  StorageClassService
	schedulerService     SchedulerService
	exportService        ExportService
	priorityClassService PriorityClassService
	resetService         ResetService
}

// NewDIContainer initializes Container.
// It initializes all service and puts to Container.
func NewDIContainer(client clientset.Interface, restclientCfg *restclient.Config, initialSchedulerCfg *v1beta2config.KubeSchedulerConfiguration) *Container {
	c := &Container{}

	// initializes each service
	c.pvService = persistentvolume.NewPersistentVolumeService(client)
	c.pvcService = persistentvolumeclaim.NewPersistentVolumeClaimService(client)
	c.storageClassService = storageclass.NewStorageClassService(client)
	c.schedulerService = scheduler.NewSchedulerService(client, restclientCfg, initialSchedulerCfg)
	c.podService = pod.NewPodService(client)

	c.nodeService = node.NewNodeService(client, c.podService)

	c.priorityClassService = priorityclass.NewPriorityClassService(client)
	c.exportService = export.NewExportService(client, c.podService, c.nodeService, c.pvService, c.pvcService, c.storageClassService, c.priorityClassService, c.schedulerService)
	c.resetService = reset.NewResetService(client, c.nodeService, c.pvService, c.pvcService, c.storageClassService, c.priorityClassService, c.schedulerService)
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
