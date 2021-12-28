package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	"k8s.io/klog/v2"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

type ExportHandler struct {
	service di.ResourcesService
}

type ResourcesApplyConfiguration struct {
	Pods            []v1.PodApplyConfiguration                     `json:"pods"`
	Nodes           []v1.NodeApplyConfiguration                    `json:"nodes"`
	Pvs             []v1.PersistentVolumeApplyConfiguration        `json:"pvs"`
	Pvcs            []v1.PersistentVolumeClaimApplyConfiguration   `json:"pvcs"`
	StorageClasses  []confstoragev1.StorageClassApplyConfiguration `json:"storageClasses"`
	SchedulerConfig *v1beta2config.KubeSchedulerConfiguration      `json:"schedulerConfig"`
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

	reqResources := new(ResourcesApplyConfiguration)
	if err := c.Bind(reqResources); err != nil {
		klog.Errorf("failed to bind import resources all request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	err := h.service.Import(ctx, convertToResourcesApplyConfiguration(reqResources))
	if err != nil {
		klog.Errorf("failed to import all resources: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// convert from ResourcesApplyConfiguration to export.ResourcesApplyConfiguration.
func convertToResourcesApplyConfiguration(r *ResourcesApplyConfiguration) *export.ResourcesApplyConfiguration {
	return &export.ResourcesApplyConfiguration{
		Pods:            r.Pods,
		Nodes:           r.Nodes,
		Pvs:             r.Pvs,
		Pvcs:            r.Pvcs,
		StorageClasses:  r.StorageClasses,
		SchedulerConfig: r.SchedulerConfig,
	}
}
