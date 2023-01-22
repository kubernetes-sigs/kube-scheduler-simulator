package controller

import (
	"context"
	"time"

	"golang.org/x/xerrors"
	"k8s.io/kubernetes/pkg/controller/volume/persistentvolume"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/hostpath"
	"k8s.io/kubernetes/pkg/volume/local"
)

var _ initFunc = startPersistentVolumeController

func startPersistentVolumeController(ctx context.Context, controllerCtx controllerContext) error {
	params := persistentvolume.ControllerParameters{
		KubeClient:                controllerCtx.ClientBuilder.ClientOrDie("persistent-volume"),
		SyncPeriod:                1 * time.Second,
		VolumePlugins:             append(local.ProbeVolumePlugins(), hostpath.ProbeVolumePlugins(volume.VolumeConfig{})...),
		VolumeInformer:            controllerCtx.InformerFactory.Core().V1().PersistentVolumes(),
		ClaimInformer:             controllerCtx.InformerFactory.Core().V1().PersistentVolumeClaims(),
		ClassInformer:             controllerCtx.InformerFactory.Storage().V1().StorageClasses(),
		PodInformer:               controllerCtx.InformerFactory.Core().V1().Pods(),
		NodeInformer:              controllerCtx.InformerFactory.Core().V1().Nodes(),
		EnableDynamicProvisioning: true,
	}
	volumeController, err := persistentvolume.NewController(params)
	if err != nil {
		return xerrors.Errorf("construct persistentvolume controller: %w", err)
	}
	go volumeController.Run(ctx)

	return nil
}
