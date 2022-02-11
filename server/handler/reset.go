package handler

import (
	"net/http"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
	"github.com/labstack/echo/v4"
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
	return c.JSON(http.StatusOK, "aaa")
}
