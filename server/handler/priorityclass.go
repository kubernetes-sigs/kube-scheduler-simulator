package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	"k8s.io/klog/v2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

// PriorityClassHandler is a handler for managing priorityClass.
//
// Deprecated: PriorityClassHandler exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
type PriorityClassHandler struct {
	service di.PriorityClassService
}

// NewPriorityClassHandler initializes PriorityClassHandler.
func NewPriorityClassHandler(s di.PriorityClassService) *PriorityClassHandler {
	return &PriorityClassHandler{service: s}
}

// ApplyPriorityClass handles the endpoint for applying priorityClass.
//
// Deprecated: ApplyPriorityClass exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PriorityClassHandler) ApplyPriorityClass(c echo.Context) error {
	ctx := c.Request().Context()

	priorityClass := new(v1.PriorityClassApplyConfiguration)
	if err := c.Bind(priorityClass); err != nil {
		klog.Errorf("failed to bind apply priorityClass request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	newsc, err := h.service.Apply(ctx, priorityClass)
	if err != nil {
		klog.Errorf("failed to apply priorityClass: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, newsc)
}

// GetPriorityClass handles the endpoint for getting priorityClass.
//
// Deprecated: GetPriorityClass exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PriorityClassHandler) GetPriorityClass(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	p, err := h.service.Get(ctx, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		klog.Errorf("failed to get priorityClass: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPriorityClass handles the endpoint for listing priorityClass.
//
// Deprecated: ListPriorityClass exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PriorityClassHandler) ListPriorityClass(c echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.service.List(ctx)
	if err != nil {
		klog.Errorf("failed to list priorityClasss: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}

// DeletePriorityClass handles the endpoint for deleting priorityClass.
//
// Deprecated: DeletePriorityClass exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PriorityClassHandler) DeletePriorityClass(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	if err := h.service.Delete(ctx, name); err != nil {
		klog.Errorf("failed to delete priorityClass: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
