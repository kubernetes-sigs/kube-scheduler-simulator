package controller

import (
	"context"

	"golang.org/x/xerrors"
	"k8s.io/kubernetes/pkg/controller/deployment"
)

var _ initFunc = startDeploymentController

func startDeploymentController(ctx context.Context, controllerCtx controllerContext) error {
	dc, err := deployment.NewDeploymentController(
		ctx,
		controllerCtx.InformerFactory.Apps().V1().Deployments(),
		controllerCtx.InformerFactory.Apps().V1().ReplicaSets(),
		controllerCtx.InformerFactory.Core().V1().Pods(),
		controllerCtx.ClientBuilder.ClientOrDie("deployment-controller"),
	)
	if err != nil {
		return xerrors.Errorf("error creating Deployment controller: %v", err)
	}
	go dc.Run(ctx, int(controllerCtx.ComponentConfig.DeploymentController.ConcurrentDeploymentSyncs))
	return nil
}
