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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScenarioSpec defines the desired state of Scenario.
type ScenarioSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Scenario. Edit scenario_types.go to remove/update
	Foo string `json:"foo,omitempty"`

	// Operations is a list of ScenarioOperation that define the actions to perform.
	// +optional
	Operations []ScenarioOperation `json:"operations,omitempty"`
}

type ScenarioOperation struct {
	// ID for this operation. Normally, the system sets this field for you.
	ID string `json:"id"`
	// MajorStep indicates when the operation should be done.
	MajorStep int32 `json:"step"`

	// One of the following four fields must be specified.
	// If more than one is set or all are empty, the operation is invalid, and the scenario will fail.

	// Create is the operation to create a new resource.
	//
	// +optional
	Create *CreateOperation `json:"createOperation,omitempty"`
	// Patch is the operation to patch a resource.
	//
	// +optional
	Patch *PatchOperation `json:"patchOperation,omitempty"`
	// Delete indicates the operation to delete a resource.
	//
	// +optional
	Delete *DeleteOperation `json:"deleteOperation,omitempty"`
	// Done indicates the operation to mark the scenario as Succeeded.
	// When finish the step DoneOperation belongs, this Scenario changes its status to Succeeded.
	//
	// +optional
	Done *DoneOperation `json:"doneOperation,omitempty"`
}

type CreateOperation struct {
	// Object is the Object to be created.
	// +kubebuilder:pruning:PreserveUnknownFields
	Object *unstructured.Unstructured `json:"object"`

	// +optional
	CreateOptions metav1.CreateOptions `json:"createOptions,omitempty"`
}

type PatchOperation struct {
	TypeMeta metav1.TypeMeta `json:"typeMeta"`
	// +kubebuilder:pruning:PreserveUnknownFields
	ObjectMeta metav1.ObjectMeta `json:"objectMeta"`
	// Patch is the patch for target.
	Patch string `json:"patch"`
	// PatchType
	PatchType types.PatchType `json:"patchType"`

	// +optional
	PatchOptions metav1.PatchOptions `json:"patchOptions,omitempty"`
}

type DeleteOperation struct {
	TypeMeta   metav1.TypeMeta   `json:"typeMeta"`
	ObjectMeta metav1.ObjectMeta `json:"objectMeta"`

	// +optional
	DeleteOptions metav1.DeleteOptions `json:"deleteOptions,omitempty"`
}

type DoneOperation struct{}

// ScenarioStatus defines the observed state of Scenario.
type ScenarioStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Scenario is the Schema for the scenarios API.
type Scenario struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScenarioSpec   `json:"spec,omitempty"`
	Status ScenarioStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScenarioList contains a list of Scenario.
type ScenarioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scenario `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Scenario{}, &ScenarioList{})
}
