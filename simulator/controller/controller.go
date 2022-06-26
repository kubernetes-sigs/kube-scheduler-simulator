package controller

import (
	pvcontroller "github.com/kubernetes-sigs/kube-scheduler-simulator/controller/volume/persistentvolume"
	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"
)

func NewController(client clientset.Interface) (func(), error) {
	mustSetupShutdownFunc, err := mustSetupScheduler(client)
	if err != nil {
		return nil, xerrors.Errorf("call mustSetupScheduler func: %w", err)
	}

	shutdownFunc := func() {
		mustSetupShutdownFunc()
	}
	return shutdownFunc, nil
}

// mustSetupScheduler starts a controller which is required to run scheduler.
func mustSetupScheduler(client clientset.Interface) (func(), error) {
	pvshutdown, err := pvcontroller.StartPersistentVolumeController(client)
	if err != nil {
		return nil, xerrors.Errorf("start pv controller: %w", err)
	}
	return pvshutdown, nil
}
