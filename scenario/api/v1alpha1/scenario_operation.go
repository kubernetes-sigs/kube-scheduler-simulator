package v1alpha1

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/xerrors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	runtimeschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

func (s *ScenarioOperation) Run(ctx context.Context, cfg *rest.Config) (bool, error) {
	switch {
	case s.Create != nil:
		ope := s.Create
		gvk := ope.Object.GetObjectKind().GroupVersionKind()
		client, err := buildClient(gvk, cfg)
		if err != nil {
			return true, xerrors.Errorf("build client failed for id: %s error: %w", s.ID, err)
		}
		_, err = client.Create(ctx, ope.Object, ope.CreateOptions)
		if err != nil {
			return true, xerrors.Errorf("run create operation: id: %s error: %w", s.ID, err)
		}
	case s.Patch != nil:
		ope := s.Patch
		gvk := ope.TypeMeta.GroupVersionKind()
		client, err := buildClient(gvk, cfg)
		if err != nil {
			return true, xerrors.Errorf("build client failed for id: %s error: %w", s.ID, err)
		}
		_, err = client.Patch(ctx, ope.ObjectMeta.Name, ope.PatchType, []byte(ope.Patch), ope.PatchOptions)
		if err != nil {
			return true, xerrors.Errorf("run patch operation: id: %s error: %w", s.ID, err)
		}
	case s.Delete != nil:
		ope := s.Delete
		gvk := ope.TypeMeta.GroupVersionKind()
		client, err := buildClient(gvk, cfg)
		if err != nil {
			return true, xerrors.Errorf("build client failed for id: %s error: %w", s.ID, err)
		}
		err = client.Delete(ctx, ope.ObjectMeta.Name, ope.DeleteOptions)
		if err != nil {
			return true, xerrors.Errorf("run delete operation: id: %s error: %w", s.ID, err)
		}
	case s.Done != nil:
		return true, nil
	default:
		return true, ErrUnknownOperation
	}

	return false, nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (s *ScenarioOperation) ValidateCreate() error {
	return s.validateOperations()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (s *ScenarioOperation) ValidateUpdate(old runtime.Object) error {
	return s.validateOperations()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (s *ScenarioOperation) ValidateDelete() error {
	return nil
}

// validateOperations checks that exactly one operation is set.
func (s *ScenarioOperation) validateOperations() error {
	var count int
	if s.Create != nil {
		count++
	}
	if s.Patch != nil {
		count++
	}
	if s.Delete != nil {
		count++
	}
	if s.Done != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("validateOperation find some operations")
	}
	return nil
}

var ErrUnknownOperation = errors.New("unknown operation")

func buildClient(gvk runtimeschema.GroupVersionKind, cfg *rest.Config) (dynamic.NamespaceableResourceInterface, error) {
	cli, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, xerrors.Errorf("build dynamic client: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, xerrors.Errorf("build discovery client: %w", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, xerrors.Errorf("build mapping from RESTMapper: %w", err)
	}

	return cli.Resource(mapping.Resource), nil
}
