// Package di organizes the dependencies.
// All services are only initialized on this package.
// di means dependency injection.
package di

import (
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/node"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/persistentvolume"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/persistentvolumeclaim"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/pod"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/priorityclass"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/resources"
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
	resourcesService     ResourcesService
	priorityClassService PriorityClassService
}

// NewDIContainer initializes Container.
// It initializes all service and puts to Container.
func NewDIContainer(client clientset.Interface, restclientCfg *restclient.Config) *Container {
	c := &Container{}

	// initializes each service
	c.pvService = persistentvolume.NewPersistentVolumeService(client)
	c.pvcService = persistentvolumeclaim.NewPersistentVolumeClaimService(client)
	c.storageClassService = storageclass.NewStorageClassService(client)
	c.schedulerService = scheduler.NewSchedulerService(client, restclientCfg)
	c.podService = pod.NewPodService(client)

	c.nodeService = node.NewNodeService(client, c.podService)
	c.resourcesService = resources.NewResourcesService(client, c.podService, c.nodeService, c.pvService, c.pvcService, c.storageClassService, c.schedulerService)

	c.priorityClassService = priorityclass.NewPriorityClassService(client)
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

func (c *Container) ResourcesService() ResourcesService {
	return c.resourcesService
}
