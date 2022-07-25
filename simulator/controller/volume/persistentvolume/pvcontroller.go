package controller

import (
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/controller/volume/persistentvolume"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/hostpath"
	"k8s.io/kubernetes/pkg/volume/local"
)

func StartPersistentVolumeController(client clientset.Interface, stopCh <-chan struct{}) error {
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	params := persistentvolume.ControllerParameters{
		KubeClient:                client,
		SyncPeriod:                1 * time.Second,
		VolumePlugins:             append(local.ProbeVolumePlugins(), hostpath.ProbeVolumePlugins(volume.VolumeConfig{})...),
		VolumeInformer:            informerFactory.Core().V1().PersistentVolumes(),
		ClaimInformer:             informerFactory.Core().V1().PersistentVolumeClaims(),
		ClassInformer:             informerFactory.Storage().V1().StorageClasses(),
		PodInformer:               informerFactory.Core().V1().Pods(),
		NodeInformer:              informerFactory.Core().V1().Nodes(),
		EnableDynamicProvisioning: true,
	}
	volumeController, err := persistentvolume.NewController(params)
	if err != nil {
		return fmt.Errorf("construct persistentvolume controller: %w", err)
	}

	go volumeController.Run(stopCh)
	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	return nil
}
