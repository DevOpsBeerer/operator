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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devopsbeererv1alpha1 "github.com/devopsbeerer/operator/api/v1alpha1"
)

// ActiveScenarioReconciler reconciles a ActiveScenario object
type ActiveScenarioReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devopsbeerer.io,resources=activescenarios,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=activescenarios/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=activescenarios/finalizers,verbs=update
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=scenariodefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=scenariohistories,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=scenariohistories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups="*",resources="*",verbs="*",namespace=devopsbeerer-*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ActiveScenarioReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO: Implement reconciliation logic
	log.Info("Reconciling ActiveScenario", "name", req.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ActiveScenarioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsbeererv1alpha1.ActiveScenario{}).
		Complete(r)
}
