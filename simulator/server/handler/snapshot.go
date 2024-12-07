package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	"k8s.io/klog/v2"
	configv1 "k8s.io/kube-scheduler/config/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/di"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/snapshot"
)

type SnapshotHandler struct {
	service di.SnapshotService
}

type ResourcesForLoad struct {
	Pods            []v1.PodApplyConfiguration                        `json:"pods"`
	Nodes           []v1.NodeApplyConfiguration                       `json:"nodes"`
	Pvs             []v1.PersistentVolumeApplyConfiguration           `json:"pvs"`
	Pvcs            []v1.PersistentVolumeClaimApplyConfiguration      `json:"pvcs"`
	StorageClasses  []confstoragev1.StorageClassApplyConfiguration    `json:"storageClasses"`
	PriorityClasses []schedulingcfgv1.PriorityClassApplyConfiguration `json:"priorityClasses"`
	SchedulerConfig *configv1.KubeSchedulerConfiguration              `json:"schedulerConfig"`
	Namespaces      []v1.NamespaceApplyConfiguration                  `json:"namespaces"`
}

func NewSnapshotHandler(s di.SnapshotService) *SnapshotHandler {
	return &SnapshotHandler{service: s}
}

func (h *SnapshotHandler) Snap(c echo.Context) error {
	ctx := c.Request().Context()

	var label metav1.LabelSelector
	rs, err := h.service.Snap(ctx, label)
	if err != nil {
		klog.Errorf("failed to save all resources: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, rs)
}

func (h *SnapshotHandler) Load(c echo.Context) error {
	ctx := c.Request().Context()

	reqResources := new(ResourcesForLoad)
	if err := c.Bind(reqResources); err != nil {
		klog.Errorf("failed to bind request: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	err := h.service.Load(ctx, convertToResourcesApplyConfiguration(reqResources))
	if err != nil {
		klog.Errorf("failed to load all resources: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// convertToResourcesApplyConfiguration converts from *ResourcesApplyConfiguration to *export.ResourcesApplyConfiguration.
func convertToResourcesApplyConfiguration(r *ResourcesForLoad) *snapshot.ResourcesForLoad {
	return &snapshot.ResourcesForLoad{
		Pods:            r.Pods,
		Nodes:           r.Nodes,
		Pvs:             r.Pvs,
		Pvcs:            r.Pvcs,
		StorageClasses:  r.StorageClasses,
		PriorityClasses: r.PriorityClasses,
		SchedulerConfig: r.SchedulerConfig,
		Namespaces:      r.Namespaces,
	}
}
