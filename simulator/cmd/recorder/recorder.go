package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
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

	ctx, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	if timeout > 0 {
		var cancel2 context.CancelFunc
		ctx, cancel2 = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel2()
	}

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
	flag.StringVar(&recordFile, "path", "", "path to store the recorded resources")
	flag.StringVar(&kubeConfig, "kubeconfig", "", "path to kubeconfig file")
	flag.IntVar(&timeout, "timeout", 0, "timeout in seconds for the simulator to run")

	flag.Parse()

	if recordFile == "" {
		return xerrors.New("path flag is required")
	}

	if kubeConfig == "" {
		return xerrors.New("kubeconfig flag is required")
	}

	return nil
}
