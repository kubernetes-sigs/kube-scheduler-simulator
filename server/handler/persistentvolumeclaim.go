package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

// PersistentVolumeClaimHandler is handler for manage persistentVolumeClaim.
//
// Deprecated: PersistentVolumeClaimHandler exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
type PersistentVolumeClaimHandler struct {
	service di.PersistentVolumeClaimService
}

// NewPersistentVolumeClaimHandler initializes PersistentVolumeClaimHandler.
func NewPersistentVolumeClaimHandler(s di.PersistentVolumeClaimService) *PersistentVolumeClaimHandler {
	return &PersistentVolumeClaimHandler{service: s}
}

// ApplyPersistentVolumeClaim handles the endpoint for applying persistentVolumeClaim.
//
// Deprecated: ApplyPersistentVolumeClaim exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PersistentVolumeClaimHandler) ApplyPersistentVolumeClaim(c echo.Context) error {
	ctx := c.Request().Context()

	persistentVolumeClaim := new(v1.PersistentVolumeClaimApplyConfiguration)
	if err := c.Bind(persistentVolumeClaim); err != nil {
		klog.Errorf("failed to bind apply persistentVolumeClaim request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	newpvc, err := h.service.Apply(ctx, persistentVolumeClaim)
	if err != nil {
		klog.Errorf("failed to apply persistentVolumeClaim: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, newpvc)
}

// GetPersistentVolumeClaim handles the endpoint for getting persistentVolumeClaim.
//
// Deprecated: GetPersistentVolumeClaim exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PersistentVolumeClaimHandler) GetPersistentVolumeClaim(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	p, err := h.service.Get(ctx, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		klog.Errorf("failed to get persistentVolumeClaim: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, p)
}

// ListPersistentVolumeClaim handles the endpoint for listing persistentVolumeClaim.
//
// Deprecated: ListPersistentVolumeClaim exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PersistentVolumeClaimHandler) ListPersistentVolumeClaim(c echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.service.List(ctx)
	if err != nil {
		klog.Errorf("failed to list persistentVolumeClaims: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ps)
}

// DeletePersistentVolumeClaim handles the endpoint for deleting persistentVolumeClaim.
//
// Deprecated: DeletePersistentVolumeClaim exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *PersistentVolumeClaimHandler) DeletePersistentVolumeClaim(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	if err := h.service.Delete(ctx, name); err != nil {
		klog.Errorf("failed to delete persistentVolumeClaim: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
