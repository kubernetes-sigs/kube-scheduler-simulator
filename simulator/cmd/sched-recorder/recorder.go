package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/xerrors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/recorder"
)

var (
	recordFile string
	kubeConfig string
	timeout    int
)

func main() {
	if err := startRecorder(); err != nil {
		klog.Fatalf("failed with error on running simulator: %+v", err)
	}
}

func startRecorder() error {
	if err := parseOptions(); err != nil {
		return err
	}

	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return xerrors.Errorf("load kubeconfig: %w", err)
	}

	client := dynamic.NewForConfigOrDie(restCfg)

	recorderOptions := recorder.Options{RecordFile: recordFile}
	recorder := recorder.New(client, recorderOptions)

	ctx, cancel := context.WithCancel(context.Background())
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	}
	defer cancel()

	if err := recorder.Run(ctx); err != nil {
		return xerrors.Errorf("run recorder: %w", err)
	}

	// Block until signal is received
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case <-ctx.Done():
	}

	return nil
}

func parseOptions() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return xerrors.Errorf("could not get user home directory: %w", err)
	}
	kubeConfigdefaultPath := filepath.Join(home, ".kube", "config")

	flag.StringVar(&recordFile, "path", "", "path to store the recorded resources")
	flag.StringVar(&kubeConfig, "kubeconfig", kubeConfigdefaultPath, "path to kubeconfig file")
	flag.IntVar(&timeout, "timeout", 0, "timeout in seconds for the simulator to run")
	flag.Parse()

	if recordFile == "" {
		return xerrors.New("path flag is required")
	}

	return nil
}
