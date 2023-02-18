package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/di"
)

// ExtenderHandler is a handler about scheduling.
type ExtenderHandler struct {
	service di.ExtenderService
}

func NewExtenderHandler(s di.ExtenderService) *ExtenderHandler {
	return &ExtenderHandler{
		service: s,
	}
}

// Filter request the original extender server which is specified by user,
// and return the response as is.
func (h *ExtenderHandler) Filter(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		klog.Errorf("failed to convert id to integer: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	req := new(extenderv1.ExtenderArgs)
	if err = c.Bind(req); err != nil {
		klog.Errorf("failed to bind the Filter request: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	res, err := h.service.Filter(id, *req)
	if err != nil {
		klog.Errorf("failed to Filter request to the extender's actually host server: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, res)
}

// Prioritize request the original extender server which is specified by user,
// and return the response as is.
func (h *ExtenderHandler) Prioritize(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		klog.Errorf("failed to convert id to integer: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	req := new(extenderv1.ExtenderArgs)
	if err = c.Bind(req); err != nil {
		klog.Errorf("failed to bind the Prioritize request: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	res, err := h.service.Prioritize(id, *req)
	if err != nil {
		klog.Errorf("failed to Prioritize request to the extender's actually host server: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, res)
}

// Preempt request the original extender server which is specified by user,
// and return the response as is.
func (h *ExtenderHandler) Preempt(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		klog.Errorf("failed to convert id to integer: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	req := new(extenderv1.ExtenderPreemptionArgs)
	if err = c.Bind(req); err != nil {
		klog.Errorf("failed to bind the preempt request: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	res, err := h.service.Preempt(id, *req)
	if err != nil {
		klog.Errorf("failed to Preempt request to the extender's actually host server: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, res)
}

// Bind request the original extender server which is specified by user,
// and return the response as is.
func (h *ExtenderHandler) Bind(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		klog.Errorf("failed to convert id to integer: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	req := new(extenderv1.ExtenderBindingArgs)
	if err = c.Bind(req); err != nil {
		klog.Errorf("failed to bind the bind request: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	res, err := h.service.Bind(id, *req)
	if err != nil {
		klog.Errorf("failed to bind request to the extender's actually host server: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, res)
}
