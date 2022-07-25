package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/di"
)

// ResourceWatcherHandler is a handler for watching the k8s resources in the simulator.
type ResourceWatcherHandler struct {
	service di.ResourceWatcherService
}

func NewResourceWatcherHandler(s di.ResourceWatcherService) *ResourceWatcherHandler {
	return &ResourceWatcherHandler{service: s}
}

// ListWatchResources provides resource updates using `server-sent events`.
func (h *ResourceWatcherHandler) ListWatchResources(c echo.Context) error {
	ctx := c.Request().Context()
	// If key is not present, FormValue returns the empty string.
	versions := &resourcewatcher.LastResourceVersions{
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
	err := h.service.ListWatch(ctx, c.Response(), versions)
	if err != nil {
		klog.Errorf("terminated to watch resources: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	// We expect this line will be called when the connection is canceled by the client.
	return c.NoContent(http.StatusOK)
}
