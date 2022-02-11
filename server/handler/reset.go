package handler

import (
	"net/http"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"
)

// ResetHandler is handler for clean up resources and scheduler configuration.
type ResetHandler struct {
	service di.ResetService
}

// NewResetHandler initializes ResetHandler.
func NewResetHandler(s di.ResetService) *ResetHandler {
	return &ResetHandler{service: s}
}

func (h *ResetHandler) Reset(c echo.Context) error {
	ctx := c.Request().Context()
	if err := h.service.Reset(ctx); err != nil {
		klog.Errorf("failed to reset all resources and schediler configuration: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusAccepted)
}
