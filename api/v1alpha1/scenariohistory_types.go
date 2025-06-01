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

// ScenarioHistorySpec defines the desired state of ScenarioHistory
type ScenarioHistorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ScenarioID is the ID of the installed scenario
	// +kubebuilder:validation:Required
	ScenarioID string `json:"scenarioId"`

	// Namespace is the namespace where the scenario is installed
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^devopsbeerer-[a-z0-9]+(-[a-z0-9]+)*$`
	Namespace string `json:"namespace"`

	// HelmRelease is the name of the Helm release
	// +kubebuilder:validation:Required
	HelmRelease string `json:"helmRelease"`

	// InstalledAt is the timestamp when the scenario was installed
	// +kubebuilder:validation:Required
	InstalledAt metav1.Time `json:"installedAt"`

	// Values contains the Helm values used for installation
	// +optional
	Values string `json:"values,omitempty"`

	// HelmChartVersion is the version of the helm chart used
	// +optional
	HelmChartVersion string `json:"helmChartVersion,omitempty"`

	// InstalledBy is the user/entity that triggered the installation
	// +optional
	InstalledBy string `json:"installedBy,omitempty"`
}

// ScenarioHistoryPhase defines the phase of scenario history
// +kubebuilder:validation:Enum=Active;Archived
type ScenarioHistoryPhase string

const (
	// ScenarioHistoryPhaseActive means this is the currently active scenario
	ScenarioHistoryPhaseActive ScenarioHistoryPhase = "Active"
	// ScenarioHistoryPhaseArchived means this scenario has been uninstalled
	ScenarioHistoryPhaseArchived ScenarioHistoryPhase = "Archived"
)

// ScenarioHistoryStatus defines the observed state of ScenarioHistory
type ScenarioHistoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase indicates whether this is the active scenario or archived
	// +kubebuilder:validation:Enum=Active;Archived
	// +kubebuilder:default=Active
	Phase ScenarioHistoryPhase `json:"phase,omitempty"`

	// UninstalledAt is the timestamp when the scenario was uninstalled
	// +optional
	UninstalledAt *metav1.Time `json:"uninstalledAt,omitempty"`

	// UninstallReason explains why the scenario was uninstalled
	// +optional
	UninstallReason string `json:"uninstallReason,omitempty"`

	// Message provides additional information about the current status
	// +optional
	Message string `json:"message,omitempty"`

	// Health indicates the health status of the installed scenario
	// +optional
	Health string `json:"health,omitempty"`

	// LastHealthCheck is the timestamp of the last health check
	// +optional
	LastHealthCheck *metav1.Time `json:"lastHealthCheck,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=sh;history
//+kubebuilder:printcolumn:name="Scenario",type="string",JSONPath=".spec.scenarioId"
//+kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".spec.namespace"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Installed",type="date",JSONPath=".spec.installedAt"
//+kubebuilder:printcolumn:name="Uninstalled",type="date",JSONPath=".status.uninstalledAt"

// ScenarioHistory is the Schema for the scenariohistories API
type ScenarioHistory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScenarioHistorySpec   `json:"spec,omitempty"`
	Status ScenarioHistoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScenarioHistoryList contains a list of ScenarioHistory
type ScenarioHistoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScenarioHistory `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScenarioHistory{}, &ScenarioHistoryList{})
}
