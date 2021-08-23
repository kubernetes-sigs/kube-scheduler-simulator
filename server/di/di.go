package di

import (
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/node"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/persistentvolume"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/persistentvolumeclaim"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/pod"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/scheduler"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/storageclass"
)

// Container saves dependencies.
type Container struct {
	nodeService         NodeService
	podService          PodService
	pvService           PersistentVolumeService
	pvcService          PersistentVolumeClaimService
	storageClassService StorageClassService
	schedulerService    SchedulerService
}

// NewDIContainer initializes Container.
func NewDIContainer(client clientset.Interface, restclientCfg *restclient.Config) *Container {
	c := &Container{}

	// initialize each service
	c.pvService = persistentvolume.NewPersistentVolumeService(client)
	c.pvcService = persistentvolumeclaim.NewPersistentVolumeClaimService(client)
	c.storageClassService = storageclass.NewStorageClassService(client)
	c.schedulerService = scheduler.NewSchedulerService(client, restclientCfg)
	c.podService = pod.NewPodService(client)

	c.nodeService = node.NewNodeService(client, c.podService)

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

func (c *Container) SchedulerService() SchedulerService {
	return c.schedulerService
}
