package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/xerrors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/recorder"
)

var dirFlag = flag.String("dir", "", "directory to store the recorded resources")
var kubeConfigFlag = flag.String("kubeconfig", "", "path to kubeconfig file")

func main() {
	if err := startRecorder(); err != nil {
		klog.Fatalf("failed with error on running simulator: %+v", err)
	}
}

func startRecorder() error {
	flag.Parse()

	if err := validateFlags(); err != nil {
		return err
	}

	restCfg, err := clientcmd.BuildConfigFromFlags("", *kubeConfigFlag)
	if err != nil {
		return xerrors.Errorf("load kubeconfig: %w", err)
	}

	client := dynamic.NewForConfigOrDie(restCfg)

	recorderOptions := recorder.Options{RecordDir: *dirFlag}
	recorder := recorder.New(client, recorderOptions)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := recorder.Run(ctx); err != nil {
		return xerrors.Errorf("run recorder: %w", err)
	}

	// Block until signal is received
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	return nil
}

func validateFlags() error {
	if *dirFlag == "" {
		return xerrors.New("dir flag is required")
	}

	if *kubeConfigFlag == "" {
		return xerrors.New("kubeconfig flag is required")
	}

	return nil
}
