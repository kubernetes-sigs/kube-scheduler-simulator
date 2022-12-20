package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/server/di"
)

const indent = " "

// PodHandler is handler for pod service
type PodHandler struct {
	service di.PodService
}

func NewPodHandler(s di.PodService) *PodHandler {
	return &PodHandler{
		service: s,
	}
}

// GetPods lists all pods in all namespaces
func (h *PodHandler) GetPods(c echo.Context) error {
	ctx := c.Request().Context()
	pods, err := h.service.List(ctx, "")
	if err != nil {
		klog.Errorf("failed to get list of pods: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSONPretty(http.StatusOK, pods, indent)
}

// GetPod gets a pod using namespace and name params
func (h *PodHandler) GetPod(c echo.Context) error {
	ctx := c.Request().Context()
	name := c.Param("name")
	namespace := c.Param("namespace")
	pod, err := h.service.Get(ctx, name, namespace)
	if err != nil {
		klog.Errorf("failed to get pod %v:%v: %+v", namespace, name, err)
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return c.JSONPretty(http.StatusOK, pod, indent)
}

// GetPodMetaDataAnnotations gets a pod annotations using namespace, name params and annotation params
func (h *PodHandler) GetPodMetaDataAnnotations(c echo.Context) error {
	ctx := c.Request().Context()
	name := c.Param("name")
	namespace := c.Param("namespace")
	annotation := c.Param("annotation")
	pod, err := h.service.Get(ctx, name, namespace)
	if err != nil {
		klog.Errorf("failed to get pod annotations: %+v", err)
		return echo.NewHTTPError(http.StatusNotFound)
	}

	var ret interface{}

	switch annotation {
	case "":
		// Dump all annotations (don't unmarshall because who knows what could be in there)
		ret = pod.Annotations
	case "scheduler-simulator", "scheduler-simulator/":
		// Get and unmarshal all the scheduler-simulator annotations (prefix is "scheduler-simulator/")
		annotations := make(map[string]map[string]map[string]string)
		ret = annotations
		for k, v := range pod.Annotations {
			if strings.HasPrefix(k, "scheduler-simulator") {
				nodeAnnotations := make(map[string]map[string]string)
				err := json.Unmarshal([]byte(v), &nodeAnnotations)
				if err != nil {
					klog.Errorf("failed to unmarshal a scheduler-simulator annotation: %+v", err)
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				annotations[k] = nodeAnnotations
			}
		}
	default:
		a, ok := pod.Annotations[annotation]
		if !ok {
			klog.Errorf("annotation %v not found on %v/%v", annotation, namespace, name)
			return echo.NewHTTPError(http.StatusNotFound)
		} else {
			nodeAnnotations := make(map[string]map[string]string)
			err := json.Unmarshal([]byte(a), &nodeAnnotations)
			if err != nil {
				klog.Errorf("failed to unmarshal annotation %v returning unmarshalled: %+v", annotation, err)
				ret = a
			} else {
				ret = nodeAnnotations
			}
		}
	}

	return c.JSONPretty(http.StatusOK, ret, indent)
}
