package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"
	"k8s.io/kube-scheduler/config/v1beta2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

// SchedulerConfigHandler is handler for manage scheduler config.
type SchedulerConfigHandler struct {
	service di.SchedulerService
}

func NewSchedulerConfigHandler(s di.SchedulerService) *SchedulerConfigHandler {
	return &SchedulerConfigHandler{
		service: s,
	}
}

func (h *SchedulerConfigHandler) GetSchedulerConfig(c echo.Context) error {
	cfg := h.service.GetSchedulerConfig()
	return c.JSON(http.StatusOK, cfg)
}

// ApplySchedulerConfig currently only takes profiles from the posted payload and applies them.
func (h *SchedulerConfigHandler) ApplySchedulerConfig(c echo.Context) error {
	reqSchedulerCfg := new(v1beta2.KubeSchedulerConfiguration)
	if err := c.Bind(reqSchedulerCfg); err != nil {
		klog.Errorf("failed to bind scheduler config request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	cfg := h.service.GetSchedulerConfig().DeepCopy()
	cfg.Profiles = reqSchedulerCfg.Profiles
	if err := h.service.RestartScheduler(cfg); err != nil {
		klog.Errorf("failed to restart scheduler: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusAccepted)
}
