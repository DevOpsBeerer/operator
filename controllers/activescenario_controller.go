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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devopsbeererv1alpha1 "github.com/devopsbeerer/operator/api/v1alpha1"
	"github.com/devopsbeerer/operator/internal/helm"
)

// ActiveScenarioReconciler reconciles a ActiveScenario object
type ActiveScenarioReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	HelmClient *helm.Client
}

//+kubebuilder:rbac:groups=devopsbeerer.io,resources=activescenarios,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=activescenarios/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=activescenarios/finalizers,verbs=update
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=scenariodefinitions,verbs=get;list;watch
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=scenariohistories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devopsbeerer.io,resources=scenariohistories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups="*",resources="*",verbs="*"

const (
	finalizerName = "devopsbeerer.io/finalizer"
)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ActiveScenarioReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the ActiveScenario instance
	activeScenario := &devopsbeererv1alpha1.ActiveScenario{}
	if err := r.Get(ctx, req.NamespacedName, activeScenario); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, could have been deleted
			log.Info("ActiveScenario resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ActiveScenario")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !activeScenario.DeletionTimestamp.IsZero() {
		log.Info("ActiveScenario resource being deleted")
		return r.handleDeletion(ctx, activeScenario)
	}

	// Do not reconcil if phase is pending
	if activeScenario.Status.Phase != "" {
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(activeScenario, finalizerName) {
		// Refresh the object before updating
		if err := r.Get(ctx, req.NamespacedName, activeScenario); err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.AddFinalizer(activeScenario, finalizerName)
		if err := r.Update(ctx, activeScenario); err != nil {
			if errors.IsConflict(err) {
				// Conflict, requeue
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
	}

	// Get the desired scenario definition
	scenarioDef := &devopsbeererv1alpha1.ScenarioDefinition{}
	if err := r.Get(ctx, types.NamespacedName{Name: activeScenario.Spec.ScenarioId}, scenarioDef); err != nil {
		if errors.IsNotFound(err) {
			if err := r.updateStatus(ctx, activeScenario, devopsbeererv1alpha1.ActiveScenarioPhaseFailed,
				fmt.Sprintf("ScenarioDefinition '%s' not found", activeScenario.Spec.ScenarioId)); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.updateStatus(ctx, activeScenario, devopsbeererv1alpha1.ActiveScenarioPhasePending,
		fmt.Sprintf("ScenarioDefinition '%s' not found", activeScenario.Spec.ScenarioId)); err != nil {
		return ctrl.Result{}, err
	}

	// Find the currently active scenario from history
	activeHistory, err := r.findActiveScenarioHistory(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Determine action based on current state
	if activeHistory == nil {
		// No active scenario - install the requested one
		log.Info("No active scenario found, installing new scenario", "scenarioId", activeScenario.Spec.ScenarioId)
		return r.installScenario(ctx, activeScenario, scenarioDef)
	}

	// Check if we need to change scenarios
	if activeHistory.Spec.ScenarioID != activeScenario.Spec.ScenarioId {
		log.Info("Scenario change detected",
			"current", activeHistory.Spec.ScenarioID,
			"desired", activeScenario.Spec.ScenarioId)

		// Update status to show we're transitioning
		if err := r.updateStatus(ctx, activeScenario, devopsbeererv1alpha1.ActiveScenarioPhaseTerminating,
			fmt.Sprintf("Uninstalling previous scenario: %s", activeHistory.Spec.ScenarioID)); err != nil {
			return ctrl.Result{}, err
		}

		// Uninstall the current scenario
		if err := r.uninstallScenario(ctx, activeHistory); err != nil {
			return ctrl.Result{}, err
		}

		// Install the new scenario
		return r.installScenario(ctx, activeScenario, scenarioDef)
	}

	// Same scenario is active - just update status if needed
	if activeScenario.Status.Phase != devopsbeererv1alpha1.ActiveScenarioPhaseRunning {
		if err := r.updateStatus(ctx, activeScenario, devopsbeererv1alpha1.ActiveScenarioPhaseRunning,
			fmt.Sprintf("Scenario '%s' is running", scenarioDef.Spec.Name)); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil

	}

	// Requeue after 5 minutes for health checks
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// handleDeletion handles the deletion of ActiveScenario
func (r *ActiveScenarioReconciler) handleDeletion(ctx context.Context, activeScenario *devopsbeererv1alpha1.ActiveScenario) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(activeScenario, finalizerName) {
		// Find and uninstall any active scenario
		activeHistory, err := r.findActiveScenarioHistory(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}

		if activeHistory != nil {
			log.Info("Uninstalling active scenario due to ActiveScenario deletion",
				"scenarioId", activeHistory.Spec.ScenarioID)
			if err := r.uninstallScenario(ctx, activeHistory); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(activeScenario, finalizerName)
		if err := r.Update(ctx, activeScenario); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// findActiveScenarioHistory finds the currently active scenario from history
func (r *ActiveScenarioReconciler) findActiveScenarioHistory(ctx context.Context) (*devopsbeererv1alpha1.ScenarioHistory, error) {
	historyList := &devopsbeererv1alpha1.ScenarioHistoryList{}
	if err := r.List(ctx, historyList); err != nil {
		return nil, err
	}

	for i := range historyList.Items {
		if historyList.Items[i].Status.Phase == devopsbeererv1alpha1.ScenarioHistoryPhaseActive {
			return &historyList.Items[i], nil
		}
	}

	return nil, nil
}

// installScenario installs a new scenario
func (r *ActiveScenarioReconciler) installScenario(ctx context.Context,
	activeScenario *devopsbeererv1alpha1.ActiveScenario,
	scenarioDef *devopsbeererv1alpha1.ScenarioDefinition) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	// Update status to Deploying
	if err := r.updateStatus(ctx, activeScenario, devopsbeererv1alpha1.ActiveScenarioPhaseDeploying,
		fmt.Sprintf("Installing scenario: %s", scenarioDef.Spec.Name)); err != nil {
		return ctrl.Result{}, err
	}

	// Create namespace
	namespace := fmt.Sprintf("devopsbeerer-%s", scenarioDef.Spec.ID)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"devopsbeerer.io/scenario": scenarioDef.Spec.ID,
				"devopsbeerer.io/managed":  "true",
			},
		},
	}

	// Check again if another reconciliation already created the history
	activeHistory, err := r.findActiveScenarioHistory(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if activeHistory != nil {
		// Another reconciliation already installed it
		log.Info("Scenario already being installed by another reconciliation")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if err := r.Create(ctx, ns); err != nil && !errors.IsAlreadyExists(err) {
		return r.updateStatusAndRequeue(ctx, activeScenario, devopsbeererv1alpha1.ActiveScenarioPhaseFailed,
			fmt.Sprintf("Failed to create namespace: %v", err))
	}

	// TODO: Install helm chart here
	// For now, we'll simulate the installation
	log.Info("Installing helm chart",
		"repo", scenarioDef.Spec.HelmChart.Link,
		"chart", scenarioDef.Spec.HelmChart.Dir,
		"namespace", namespace)

	// Install helm chart
	helmRelease := fmt.Sprintf("devopsbeerer-%s", scenarioDef.Spec.ID)
	chartPath := scenarioDef.Spec.HelmChart.Dir
	if chartPath == "" {
		chartPath = scenarioDef.Spec.ID // Default to scenario ID
	}

	if r.HelmClient != nil {
		if err := r.HelmClient.Install(ctx, helmRelease, namespace,
			scenarioDef.Spec.HelmChart.Link, chartPath, ""); err != nil {
			return r.updateStatusAndRequeue(ctx, activeScenario, devopsbeererv1alpha1.ActiveScenarioPhaseFailed,
				fmt.Sprintf("Failed to install helm chart: %v", err))
		}
	}

	// Create history entry
	history := &devopsbeererv1alpha1.ScenarioHistory{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("history-%s-%d", scenarioDef.Spec.ID, time.Now().Unix()),
		},
		Spec: devopsbeererv1alpha1.ScenarioHistorySpec{
			ScenarioID:  scenarioDef.Spec.ID,
			Namespace:   namespace,
			HelmRelease: fmt.Sprintf("devopsbeerer-%s", scenarioDef.Spec.ID),
			InstalledAt: metav1.Now(),
		},
		Status: devopsbeererv1alpha1.ScenarioHistoryStatus{
			Phase: devopsbeererv1alpha1.ScenarioHistoryPhaseActive,
		},
	}

	if err := r.Create(ctx, history); err != nil {
		return r.updateStatusAndRequeue(ctx, activeScenario, devopsbeererv1alpha1.ActiveScenarioPhaseFailed,
			fmt.Sprintf("Failed to create history: %v", err))
	}

	// Update ActiveScenario status
	activeScenario.Status.Phase = devopsbeererv1alpha1.ActiveScenarioPhaseRunning
	activeScenario.Status.ScenarioName = scenarioDef.Spec.Name
	activeScenario.Status.HelmReleaseName = history.Spec.HelmRelease
	activeScenario.Status.StartTime = &history.Spec.InstalledAt
	activeScenario.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}
	activeScenario.Status.Message = fmt.Sprintf("Scenario '%s' is running", scenarioDef.Spec.Name)

	if err := r.Status().Update(ctx, activeScenario); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// uninstallScenario uninstalls a scenario
func (r *ActiveScenarioReconciler) uninstallScenario(ctx context.Context,
	history *devopsbeererv1alpha1.ScenarioHistory) error {

	log := log.FromContext(ctx)

	// TODO: Uninstall helm chart here
	log.Info("Uninstalling helm chart",
		"release", history.Spec.HelmRelease,
		"namespace", history.Spec.Namespace)

	// Uninstall helm chart
	if r.HelmClient != nil {
		if err := r.HelmClient.Uninstall(ctx, history.Spec.HelmRelease, history.Spec.Namespace); err != nil {
			return fmt.Errorf("failed to uninstall helm chart: %w", err)
		}
	}

	// Delete namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: history.Spec.Namespace,
		},
	}

	if err := r.Delete(ctx, ns); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	// Update history to archived
	history.Status.Phase = devopsbeererv1alpha1.ScenarioHistoryPhaseArchived
	history.Status.UninstalledAt = &metav1.Time{Time: time.Now()}
	history.Status.UninstallReason = "Replaced by new scenario"

	if err := r.Status().Update(ctx, history); err != nil {
		return fmt.Errorf("failed to update history status: %w", err)
	}

	return nil
}

// updateStatus updates the ActiveScenario status
func (r *ActiveScenarioReconciler) updateStatus(ctx context.Context,
	activeScenario *devopsbeererv1alpha1.ActiveScenario,
	phase devopsbeererv1alpha1.ActiveScenarioPhase,
	message string) error {

	activeScenario.Status.Phase = phase
	activeScenario.Status.Message = message
	activeScenario.Status.LastTransitionTime = &metav1.Time{Time: time.Now()}

	return r.Status().Update(ctx, activeScenario)
}

// updateStatusAndRequeue updates status and returns a requeue result
func (r *ActiveScenarioReconciler) updateStatusAndRequeue(ctx context.Context,
	activeScenario *devopsbeererv1alpha1.ActiveScenario,
	phase devopsbeererv1alpha1.ActiveScenarioPhase,
	message string) (ctrl.Result, error) {

	if err := r.updateStatus(ctx, activeScenario, phase, message); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ActiveScenarioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsbeererv1alpha1.ActiveScenario{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1, // Process one at a time
		}).
		Complete(r)
}
