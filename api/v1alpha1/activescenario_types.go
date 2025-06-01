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

// ActiveScenarioSpec defines the desired state of ActiveScenario
type ActiveScenarioSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ScenarioId is the ID of the scenario definition to deploy
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]+(-[a-z0-9]+)*$`
	ScenarioId string `json:"scenarioId"`
}

// ActiveScenarioPhase defines the phase of scenario deployment
// +kubebuilder:validation:Enum=Pending;Deploying;Running;Failed;Terminating
type ActiveScenarioPhase string

const (
	// ActiveScenarioPhasePending means the scenario is waiting to be deployed
	ActiveScenarioPhasePending ActiveScenarioPhase = "Pending"
	// ActiveScenarioPhaseDeploying means the scenario is being deployed
	ActiveScenarioPhaseDeploying ActiveScenarioPhase = "Deploying"
	// ActiveScenarioPhaseRunning means the scenario is successfully running
	ActiveScenarioPhaseRunning ActiveScenarioPhase = "Running"
	// ActiveScenarioPhaseFailed means the scenario deployment failed
	ActiveScenarioPhaseFailed ActiveScenarioPhase = "Failed"
	// ActiveScenarioPhaseTerminating means the scenario is being terminated
	ActiveScenarioPhaseTerminating ActiveScenarioPhase = "Terminating"
)

// ActiveScenarioStatus defines the observed state of ActiveScenario
type ActiveScenarioStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase is the current phase of the scenario deployment
	// +kubebuilder:validation:Enum=Pending;Deploying;Running;Failed;Terminating
	Phase ActiveScenarioPhase `json:"phase,omitempty"`

	// ScenarioName is the name of the deployed scenario
	ScenarioName string `json:"scenarioName,omitempty"`

	// Message is a human-readable message about the current status
	Message string `json:"message,omitempty"`

	// LastTransitionTime is the last time the status changed
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// HelmReleaseName is the name of the Helm release
	HelmReleaseName string `json:"helmReleaseName,omitempty"`

	// StartTime is when the scenario was started
	StartTime *metav1.Time `json:"startTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=as;active
//+kubebuilder:printcolumn:name="Scenario",type="string",JSONPath=".spec.scenarioId"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Helm Release",type="string",JSONPath=".status.helmReleaseName"
//+kubebuilder:printcolumn:name="Started",type="date",JSONPath=".status.startTime"

// ActiveScenario is the Schema for the activescenarios API
type ActiveScenario struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActiveScenarioSpec   `json:"spec,omitempty"`
	Status ActiveScenarioStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ActiveScenarioList contains a list of ActiveScenario
type ActiveScenarioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ActiveScenario `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ActiveScenario{}, &ActiveScenarioList{})
}
