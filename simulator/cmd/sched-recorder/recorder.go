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
	duration   int
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
	if duration > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(duration)*time.Second)
	}
	defer cancel()

	if err := recorder.Run(ctx); err != nil {
		return xerrors.Errorf("run recorder: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case <-ctx.Done():
		klog.Info("recording is finishing because the specified duration has elapsed")
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
	flag.IntVar(&duration, "duration", 0, "duration in seconds for the simulator to run")
	flag.Parse()

	if recordFile == "" {
		return xerrors.New("path flag is required")
	}

	if duration < 0 {
		return xerrors.New("duration must be a non-negative value")
	}

	return nil
}
