package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

type ExportHandler struct {
	service di.ResourcesService
}

func NewExportHandler(s di.ResourcesService) *ExportHandler {
	return &ExportHandler{service: s}
}

func (h *ExportHandler) Export(c echo.Context) error {
	ctx := c.Request().Context()

	rs, err := h.service.Export(ctx)
	if err != nil {
		klog.Errorf("failed to export all resources: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, rs)
}

func (h *ExportHandler) Import(c echo.Context) error {
	ctx := c.Request().Context()

	reqResources := new(export.ResourcesApplyConfiguration)
	if err := c.Bind(reqResources); err != nil {
		klog.Errorf("failed to bind import resources all request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	rs, err := h.service.Import(ctx, reqResources)
	if err != nil {
		klog.Errorf("failed to import all resources: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, rs)
}
