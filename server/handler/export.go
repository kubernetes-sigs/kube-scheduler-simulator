package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	"k8s.io/klog/v2"
	v1beta3config "k8s.io/kube-scheduler/config/v1beta3"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

type ExportHandler struct {
	service di.ExportService
}

type ResourcesForImport struct {
	Pods            []v1.PodApplyConfiguration                        `json:"pods"`
	Nodes           []v1.NodeApplyConfiguration                       `json:"nodes"`
	Pvs             []v1.PersistentVolumeApplyConfiguration           `json:"pvs"`
	Pvcs            []v1.PersistentVolumeClaimApplyConfiguration      `json:"pvcs"`
	StorageClasses  []confstoragev1.StorageClassApplyConfiguration    `json:"storageClasses"`
	PriorityClasses []schedulingcfgv1.PriorityClassApplyConfiguration `json:"priorityClasses"`
	SchedulerConfig *v1beta3config.KubeSchedulerConfiguration         `json:"schedulerConfig"`
}

func NewExportHandler(s di.ExportService) *ExportHandler {
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

	reqResources := new(ResourcesForImport)
	if err := c.Bind(reqResources); err != nil {
		klog.Errorf("failed to bind import resources all request: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	err := h.service.Import(ctx, convertToResourcesApplyConfiguration(reqResources))
	if err != nil {
		klog.Errorf("failed to import all resources: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// convertToResourcesApplyConfiguration converts from *ResourcesApplyConfiguration to *export.ResourcesApplyConfiguration.
func convertToResourcesApplyConfiguration(r *ResourcesForImport) *export.ResourcesForImport {
	return &export.ResourcesForImport{
		Pods:            r.Pods,
		Nodes:           r.Nodes,
		Pvs:             r.Pvs,
		Pvcs:            r.Pvcs,
		StorageClasses:  r.StorageClasses,
		PriorityClasses: r.PriorityClasses,
		SchedulerConfig: r.SchedulerConfig,
	}
}
