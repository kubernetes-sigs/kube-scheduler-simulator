package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

// PodHandler is handler for manage pod.
//
// Deprecated: PodHandler exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
type PodHandler struct {
	service di.PodService
}

// NewPodHandler initializes PodHandler.
func NewPodHandler(s di.PodService) *PodHandler {
	return &PodHandler{service: s}
}

// ApplyPod handles the endpoint for applying pod.
//
// Deprecated: ApplyPod exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PodHandler) ApplyPod(c echo.Context) error {
	ctx := c.Request().Context()

	pod := new(v1.PodApplyConfiguration)
	if err := c.Bind(pod); err != nil {
		klog.Errorf("failed to bind apply pod request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	newpod, err := h.service.Apply(ctx, pod)
	if err != nil {
		klog.Errorf("failed to apply pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, newpod)
}

// GetPod handles the endpoint for getting pod.
//
// Deprecated: GetPod exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PodHandler) GetPod(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	p, err := h.service.Get(ctx, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		klog.Errorf("failed to get pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPod handles the endpoint for listing pod.
//
// Deprecated: ListPod exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PodHandler) ListPod(c echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.service.List(ctx)
	if err != nil {
		klog.Errorf("failed to list pods: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}

// DeletePod handles the endpoint for deleting pod.
//
// Deprecated: DeletePod exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PodHandler) DeletePod(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	if err := h.service.Delete(ctx, name); err != nil {
		klog.Errorf("failed to delete pod: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
