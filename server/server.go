package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/xerrors"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/config"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/handler"
)

// SimulatorServer is server for simulator.
type SimulatorServer struct {
	e *echo.Echo
}

// NewSimulatorServer initialize SimulatorServer.
//nolint:funlen // It is okay if the definition of this function is long.
func NewSimulatorServer(cfg *config.Config, dic *di.Container) *SimulatorServer {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowCredentials: true,
	}))

	// initialize each handler
	nodeHandler := handler.NewNodeHandler(dic.NodeService())
	podHandler := handler.NewPodHandler(dic.PodService())
	pvHandler := handler.NewPersistentVolumeHandler(dic.PersistentVolumeService())
	pvcHandler := handler.NewPersistentVolumeClaimHandler(dic.PersistentVolumeClaimService())
	storageClassHandler := handler.NewStorageClassHandler(dic.StorageClassService())
	schedulercfgHandler := handler.NewSchedulerConfigHandler(dic.SchedulerService())
	priorityClassHandler := handler.NewPriorityClassHandler(dic.PriorityClassService())
	exportHandler := handler.NewExportHandler(dic.ExportService())

	// register apis
	v1 := e.Group("/api/v1")

	v1.GET("/schedulerconfiguration", schedulercfgHandler.GetSchedulerConfig)
	v1.POST("/schedulerconfiguration", schedulercfgHandler.ApplySchedulerConfig)
	v1.PUT("/schedulerconfiguration", schedulercfgHandler.ResetScheduler)

	v1.GET("/nodes", nodeHandler.ListNode)
	v1.POST("/nodes", nodeHandler.ApplyNode)
	v1.GET("/nodes/:name", nodeHandler.GetNode)
	v1.DELETE("/nodes/:name", nodeHandler.DeleteNode)

	v1.GET("/pods", podHandler.ListPod)
	v1.POST("/pods", podHandler.ApplyPod)
	v1.GET("/pods/:name", podHandler.GetPod)
	v1.DELETE("/pods/:name", podHandler.DeletePod)

	v1.GET("/persistentvolumes", pvHandler.ListPersistentVolume)
	v1.POST("/persistentvolumes", pvHandler.ApplyPersistentVolume)
	v1.GET("/persistentvolumes/:name", pvHandler.GetPersistentVolume)
	v1.DELETE("/persistentvolumes/:name", pvHandler.DeletePersistentVolume)

	v1.GET("/persistentvolumeclaims", pvcHandler.ListPersistentVolumeClaim)
	v1.POST("/persistentvolumeclaims", pvcHandler.ApplyPersistentVolumeClaim)
	v1.GET("/persistentvolumeclaims/:name", pvcHandler.GetPersistentVolumeClaim)
	v1.DELETE("/persistentvolumeclaims/:name", pvcHandler.DeletePersistentVolumeClaim)

	v1.GET("/storageclasses", storageClassHandler.ListStorageClass)
	v1.POST("/storageclasses", storageClassHandler.ApplyStorageClass)
	v1.GET("/storageclasses/:name", storageClassHandler.GetStorageClass)
	v1.DELETE("/storageclasses/:name", storageClassHandler.DeleteStorageClass)

	v1.GET("/priorityclasses", priorityClassHandler.ListPriorityClass)
	v1.POST("/priorityclasses", priorityClassHandler.ApplyPriorityClass)
	v1.GET("/priorityclasses/:name", priorityClassHandler.GetPriorityClass)
	v1.DELETE("/priorityclasses/:name", priorityClassHandler.DeletePriorityClass)

	v1.GET("/export", exportHandler.Export)
	v1.POST("/import", exportHandler.Import)
	// initialize SimulatorServer.
	s := &SimulatorServer{e: e}
	s.e.Logger.SetLevel(log.INFO)

	return s
}

// Start starts SimulatorServer.
func (s *SimulatorServer) Start(port int) (
	func(), // function for shutdown
	error,
) {
	e := s.e

	go func() {
		if err := e.Start(":" + strconv.Itoa(port)); err != nil && !xerrors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatalf("failed to start server successfully: %v", err)
		}
	}()

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Warnf("failed to shutdown simulator server successfully: %v", err)
		}
	}

	return shutdownFn, nil
}
