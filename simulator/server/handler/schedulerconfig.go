package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"
	configv1 "k8s.io/kube-scheduler/config/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/di"
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
	cfg, err := h.service.GetSchedulerConfig()
	if err != nil && !errors.Is(err, scheduler.ErrServiceDisabled) {
		klog.Errorf("failed to get scheduler config: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if errors.Is(err, scheduler.ErrServiceDisabled) {
		return c.JSON(http.StatusBadRequest, "When using an external scheduler, you cannot see and edit the scheduler configuration.")
	}

	return c.JSON(http.StatusOK, cfg)
}

// ApplySchedulerConfig currently only takes profiles and extenders from the
// posted payload and applies them.
func (h *SchedulerConfigHandler) ApplySchedulerConfig(c echo.Context) error {
	reqSchedulerCfg := new(configv1.KubeSchedulerConfiguration)
	if err := c.Bind(reqSchedulerCfg); err != nil {
		klog.Errorf("failed to bind scheduler config request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	cfg, err := h.service.GetSchedulerConfig()
	if err != nil {
		klog.Errorf("failed to get scheduler config: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	cfg = cfg.DeepCopy()
	cfg.Profiles = reqSchedulerCfg.Profiles
	cfg.Extenders = reqSchedulerCfg.Extenders
	if err := h.service.RestartScheduler(cfg); err != nil {
		klog.Errorf("failed to restart scheduler: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusAccepted)
}
