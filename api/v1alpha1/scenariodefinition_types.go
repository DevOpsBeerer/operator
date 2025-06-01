/*
Copyright 2024.

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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HelmChart defines the helm chart configuration
type HelmChart struct {
	// Link is the Git repository URL for helm charts
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https://.*\.git$`
	// +kubebuilder:default=`https://github.com/DevOpsBeerer/playground-scenarios-charts.git`
	Link string `json:"link"`

	// Dir is the subdirectory containing the helm chart (optional, defaults to scenario ID)
	// +optional
	Dir string `json:"dir,omitempty"`
}

// ScenarioDefinitionSpec defines the desired state of ScenarioDefinition
type ScenarioDefinitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the human-readable name of the scenario
	// +kubebuilder:validation:Required
	// +kubebuilder:example="Basic OAuth2 Beer Management"
	Name string `json:"name"`

	// ID is the unique identifier with hyphens
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]+(-[a-z0-9]+)*$`
	// +kubebuilder:example="basic-oauth2-beer-mgmt"
	ID string `json:"id"`

	// Description is the detailed description of the scenario
	// +kubebuilder:validation:Required
	Description string `json:"description"`

	// HelmChart defines the helm chart configuration
	// +kubebuilder:validation:Required
	HelmChart HelmChart `json:"helmChart"`

	// Tags for categorizing scenarios
	// +optional
	// +kubebuilder:example={"oauth2","basic","api","crud"}
	Tags []string `json:"tags,omitempty"`

	// Features is the list of features demonstrated in this scenario
	// +optional
	// +kubebuilder:example={"authorization-code-flow","refresh-tokens","rbac","api-gateway"}
	Features []string `json:"features,omitempty"`
}

// ScenarioDefinitionStatus defines the observed state of ScenarioDefinition
type ScenarioDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster,shortName=scndef;sd
//+kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.name"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".spec.id"
//+kubebuilder:printcolumn:name="Tags",type="string",JSONPath=".spec.tags"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ScenarioDefinition is the Schema for the scenariodefinitions API
type ScenarioDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScenarioDefinitionSpec   `json:"spec,omitempty"`
	Status ScenarioDefinitionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScenarioDefinitionList contains a list of ScenarioDefinition
type ScenarioDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScenarioDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScenarioDefinition{}, &ScenarioDefinitionList{})
}
