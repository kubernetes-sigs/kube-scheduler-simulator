package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/config"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/k8sapiserver"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"
)

// start scheduler, apiserver, etcd

func startComponents() error {
	router := mux.NewRouter()
	cfg, err := config.NewConfig()
	if err != nil {
		return xerrors.Errorf("get config: %w", err)

	}

	restclientCfg, apiShutdown, err := k8sapiserver.StartAPIServer(cfg.KubeAPIServerURL, cfg.EtcdURL)
	if err != nil {
		return xerrors.Errorf("start API server: %w", err)
	}
	defer apiShutdown()

	client := clientset.NewForConfigOrDie(restclientCfg)

	existingClusterClient := &clientset.Clientset{}
	if cfg.ExternalImportEnabled {
		existingClusterClient, err = clientset.NewForConfig(cfg.ExternalKubeClientCfg)
		if err != nil {
			return xerrors.Errorf("creates a new Clientset for the ExternalKubeClientCfg: %w", err)
		}
	}

	dic := di.NewDIContainer(client, restclientCfg, cfg.InitialSchedulerCfg, cfg.ExternalImportEnabled, existingClusterClient, cfg.ExternalKubeClientCfg)

	if err := dic.SchedulerService().StartScheduler(cfg.InitialSchedulerCfg); err != nil {
		return xerrors.Errorf("start scheduler: %w", err)
	}
	defer dic.SchedulerService().ShutdownScheduler()

	// start simulator server
	s := server.NewSimulatorServer(cfg, dic)
	shutdownFn3, err := s.Start(cfg.Port)
	if err != nil {
		return xerrors.Errorf("start simulator server: %w", err)
	}
	defer shutdownFn3()

	return nil
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}

}

func TestSchedulerConfig(t *testing.T) {
	startComponents()
	t.Logf("Components created")

	req, _ := http.NewRequest("GET", "/schedulerconfiguration", nil)
	response := executeRequest(req)

	t.Logf("testing for scheduler config")

	checkResponseCode(t, http.StatusOK, response.Code)

	// Apply newscheduler config
	reqApply, _ := http.NewRequest("POST", "/schedulerconfiguration", nil)
	response = executeApplyRequest(reqApply)
}

func TestImportConfig(t *testing.T) {

}

func TestExportConifg(t *testing.T) {

}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServerHTTP(rr, req)

}

// func executeApplyRequest(req *http.Request) *httptest.ResponseRecorder {

// }

//
