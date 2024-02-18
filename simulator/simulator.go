package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/app"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/config"
)

// entry point.
func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		klog.Fatalf("failed with error on reading config: %+v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := app.StartSimulator(ctx, cfg); err != nil {
			klog.Fatalf("failed with error on running simulator: %+v", err)
		}
	}()

	// wait the signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit

	// stop simulator
	cancel()
}
