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
func NewSimulatorServer(cfg *config.Config, dic *di.Container) *SimulatorServer {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowCredentials: true,
	}))

	// initialize each handler
	schedulercfgHandler := handler.NewSchedulerConfigHandler(dic.SchedulerService())
	exportHandler := handler.NewExportHandler(dic.ExportService())
	resetHandler := handler.NewResetHandler(dic.ResetService())

	// register apis
	v1 := e.Group("/api/v1")

	v1.GET("/schedulerconfiguration", schedulercfgHandler.GetSchedulerConfig)
	v1.POST("/schedulerconfiguration", schedulercfgHandler.ApplySchedulerConfig)

	v1.PUT("/reset", resetHandler.Reset)

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
