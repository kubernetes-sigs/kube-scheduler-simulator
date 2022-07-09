package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/simulator/server/di"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/simulator/watcher"
)

// WatcherHandler is a handler for watching the simulator's resources.
type WatcherHandler struct {
	service di.WatcherService
}

func NewWatcherHandler(s di.WatcherService) *WatcherHandler {
	return &WatcherHandler{service: s}
}

// WatchResources provides a server-pushed response.
func (h *WatcherHandler) WatchResources(c echo.Context) error {
	ctx := c.Request().Context()
	versions := &watcher.LastResourceVersions{
		Pods:  c.FormValue("podsLastResourceVersion"),
		Nodes: c.FormValue("nodesLastResourceVersion"),
		Pvs:   c.FormValue("pvsLastResourceVersion"),
		Pvcs:  c.FormValue("pvcsLastResourceVersion"),
		Scs:   c.FormValue("scsLastResourceVersion"),
		Pcs:   c.FormValue("pcsLastResourceVersion"),
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
	// Start to watch and do server push
	err := h.service.WatchResources(ctx, c.Response(), versions)
	if err != nil {
		klog.Errorf("closed to watch resources: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}
