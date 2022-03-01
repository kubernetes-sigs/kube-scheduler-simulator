package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/klog/v2"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/server/di"
)

// NodeHandler is handler for manage nodes.
//
// Deprecated: NodeHandler exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
type NodeHandler struct {
	service di.NodeService
}

// NewNodeHandler initializes NodeHandler.
func NewNodeHandler(s di.NodeService) *NodeHandler {
	return &NodeHandler{service: s}
}

// ApplyNode handles the endpoint for applying node.
//
// Deprecated: ApplyNode exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *NodeHandler) ApplyNode(c echo.Context) error {
	ctx := c.Request().Context()

	reqNode := new(v1.NodeApplyConfiguration)
	if err := c.Bind(reqNode); err != nil {
		klog.Errorf("failed to bind apply node request: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	newnode, err := h.service.Apply(ctx, reqNode)
	if err != nil {
		klog.Errorf("failed to apply node: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, newnode)
}

// GetNode handles the endpoint for getting node.
//
// Deprecated: GetNode exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *NodeHandler) GetNode(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	n, err := h.service.Get(ctx, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		klog.Errorf("failed to get node: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, n)
}

// ListNode handles the endpoint for listing node.
//
// Deprecated: ListNode exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *NodeHandler) ListNode(c echo.Context) error {
	ctx := c.Request().Context()

	ns, err := h.service.List(ctx)
	if err != nil {
		klog.Errorf("failed to list nodes: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ns)
}

// DeleteNode handles the endpoint for deleting node.
//
// Deprecated: DeleteNode exists only for backward compatibility.
// It is not maintained now and will be deleted soon.
func (h *NodeHandler) DeleteNode(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")

	if err := h.service.Delete(ctx, name); err != nil {
		klog.Errorf("failed to delete node: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
