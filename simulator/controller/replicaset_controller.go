package controller

import (
	"context"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller/replicaset"
)

var _ initFunc = startReplicaSetController

func startReplicaSetController(ctx context.Context, controllerCtx controllerContext) error {
	go replicaset.NewReplicaSetController(
		klog.FromContext(ctx),
		controllerCtx.InformerFactory.Apps().V1().ReplicaSets(),
		controllerCtx.InformerFactory.Core().V1().Pods(),
		controllerCtx.ClientBuilder.ClientOrDie("replicaset-controller"),
		replicaset.BurstReplicas,
	).Run(ctx, int(controllerCtx.ComponentConfig.ReplicaSetController.ConcurrentRSSyncs))
	return nil
}
