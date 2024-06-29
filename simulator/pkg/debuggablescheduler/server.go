package debuggablescheduler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/extender"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/handler"
)

// ExtenderServer is proxy server for extender.
type ExtenderServer struct {
	e *echo.Echo
}

// NewExtenderServer initialize ExtenderServer.
// This server is used as a proxy server to store Extender results.
func NewExtenderServer(service *extender.Service) ExtenderServer {
	e := echo.New()
	e.Use(middleware.Logger())

	extenderHandler := handler.NewExtenderHandler(service)
	// register apis
	v1 := e.Group("/api/v1")
	server.RouteExtender(v1, extenderHandler)
	s := ExtenderServer{e: e}
	s.e.Logger.SetLevel(log.INFO)
	return s
}

// Start starts ExtenderServer.
func (s *ExtenderServer) Start(port int) (
	func(), // function for shutdown
	error,
) {
	e := s.e

	go func() {
		if err := e.Start(":" + strconv.Itoa(port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
