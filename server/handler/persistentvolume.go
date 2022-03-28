package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

// PersistentVolumeHandler is handler for manage persistentVolume.
//
// Deprecated: PersistentVolumeHandler exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
type PersistentVolumeHandler struct {
	service di.PersistentVolumeService
}

// NewPersistentVolumeHandler initializes PersistentVolumeHandler.
func NewPersistentVolumeHandler(s di.PersistentVolumeService) *PersistentVolumeHandler {
	return &PersistentVolumeHandler{service: s}
}

// ApplyPersistentVolume handles the endpoint for applying persistentVolume.
//
// Deprecated: ApplyPersistentVolume exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PersistentVolumeHandler) ApplyPersistentVolume(c echo.Context) error {
	ctx := c.Request().Context()

	persistentVolume := new(v1.PersistentVolumeApplyConfiguration)
	if err := c.Bind(persistentVolume); err != nil {
		klog.Errorf("failed to bind apply persistentVolume request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	newpv, err := h.service.Apply(ctx, persistentVolume)
	if err != nil {
		klog.Errorf("failed to apply persistentVolume: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, newpv)
}

// GetPersistentVolume handles the endpoint for getting persistentVolume.
//
// Deprecated: GetPersistentVolume exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PersistentVolumeHandler) GetPersistentVolume(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	p, err := h.service.Get(ctx, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		klog.Errorf("failed to get persistentVolume: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPersistentVolume handles the endpoint for listing persistentVolume.
//
// Deprecated: ListPersistentVolume exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PersistentVolumeHandler) ListPersistentVolume(c echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.service.List(ctx)
	if err != nil {
		klog.Errorf("failed to list persistentVolumes: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}

// DeletePersistentVolume handles the endpoint for deleting persistentVolume.
//
// Deprecated: DeletePersistentVolume exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PersistentVolumeHandler) DeletePersistentVolume(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	if err := h.service.Delete(ctx, name); err != nil {
		klog.Errorf("failed to delete persistentVolume: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
