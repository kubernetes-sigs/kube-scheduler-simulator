/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	simulationv1alpha1 "sigs.k8s.io/kube-scheduler-simulator/scenario/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var scenariolog = logf.Log.WithName("scenario-resource")

// SetupScenarioWebhookWithManager registers the webhook for Scenario in the manager.
func SetupScenarioWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&simulationv1alpha1.Scenario{}).
		WithValidator(&ScenarioCustomValidator{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-simulation-kube-scheduler-simulator-x-k8s-io-v1alpha1-scenario,mutating=false,failurePolicy=fail,sideEffects=None,groups=simulation.kube-scheduler-simulator.x-k8s.io,resources=scenarios,verbs=create;update,versions=v1alpha1,name=vscenario-v1alpha1.kb.io,admissionReviewVersions=v1

// ScenarioCustomValidator struct is responsible for validating the Scenario resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ScenarioCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &ScenarioCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Scenario.
func (v *ScenarioCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scenario, ok := obj.(*simulationv1alpha1.Scenario)
	if !ok {
		return nil, fmt.Errorf("expected a Scenario object but got %T", obj)
	}
	scenariolog.Info("Validation for Scenario upon creation", "name", scenario.GetName())

	for _, op := range scenario.Spec.Operations {
		err := op.ValidateCreate()
		if err != nil {
			return nil, xerrors.Errorf("scenario webhook ValidateCreate: %w", err)
		}
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Scenario.
func (v *ScenarioCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scenario, ok := newObj.(*simulationv1alpha1.Scenario)
	if !ok {
		return nil, fmt.Errorf("expected a Scenario object for the newObj but got %T", newObj)
	}
	scenariolog.Info("Validation for Scenario upon update", "name", scenario.GetName())

	for _, op := range scenario.Spec.Operations {
		err := op.ValidateUpdate(oldObj)
		if err != nil {
			return nil, xerrors.Errorf("scenario webhook ValidateUpdate: %w", err)
		}
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Scenario.
func (v *ScenarioCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scenario, ok := obj.(*simulationv1alpha1.Scenario)
	if !ok {
		return nil, fmt.Errorf("expected a Scenario object but got %T", obj)
	}
	scenariolog.Info("Validation for Scenario upon deletion", "name", scenario.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
