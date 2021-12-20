package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"
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

	// backup before overwrite exist resources.
	bkp, err := h.service.Export(ctx)
	if err != nil {
		klog.Errorf("failed to backup resources before import: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	reqResources := new(export.ResourcesApplyConfiguration)
	if err := c.Bind(reqResources); err != nil {
		klog.Errorf("failed to bind import resources all request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	rs, err := h.service.Import(ctx, reqResources)
	if err != nil {
		klog.Errorf("failed to import all resources: %+v", err)
		// revocery step from backup resources
		bkpResources := new(export.ResourcesApplyConfiguration)
		bkpr, err1 := json.Marshal(bkp)
		if err1 != nil {
			klog.Errorf("failed to parse json of recovery resources: %+v", err1)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if err2 := json.Unmarshal(bkpr, &bkpResources); err2 != nil {
			klog.Errorf("failed to convert to ResourcesApplyConfiguration of recovery resources: %+v", err2)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		_, err3 := h.service.Import(ctx, bkpResources)
		if err3 != nil {
			klog.Errorf("failed to recover of backup resources: %+v", err3)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	return c.JSON(http.StatusOK, rs)
}
