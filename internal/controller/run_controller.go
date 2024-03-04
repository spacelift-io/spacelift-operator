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

package controller

import (
	"context"
	"reflect"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
	spaceliftRepository "github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/watcher"
)

// RunReconciler reconciles a Run object
type RunReconciler struct {
	RunRepository          *repository.RunRepository
	SpaceliftRunRepository spaceliftRepository.RunRepository
	RunWatcher             *watcher.RunWatcher
}

//+kubebuilder:rbac:groups=app.spacelift.io,resources=runs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.spacelift.io,resources=runs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.spacelift.io,resources=runs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *RunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	run, err := r.RunRepository.Get(ctx, req.NamespacedName)

	// The Run is removed, this should not happen because we filter out deletion events.
	// This can't really hurt and makes the reconciliation logic a bit more straightforward to read
	if err != nil && k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "Unable to retrieve Run from kube API.")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If the run is new, then create it on spacelift and update the status
	if run.IsNew() {
		return r.handleNewRun(ctx, run)
	}

	return r.handleRunUpdate(ctx, run)
}

func (r *RunReconciler) handleNewRun(ctx context.Context, run *v1beta1.Run) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	spaceliftRun, err := r.SpaceliftRunRepository.Create(ctx, run)
	if err != nil {
		logger.Error(err, "Unable to create the run in spacelift")
		// TODO(eliecharra): Implement better error handling and retry errors that could be retried
		return ctrl.Result{}, nil
	}
	run.SetState(v1beta1.RunState(spaceliftRun.State))
	run.Status.Id = spaceliftRun.RunID
	logger.WithValues(
		logging.RunState, run.Status.State,
		logging.RunId, run.Status.Id,
	).Info("New run created")
	if err := r.RunRepository.UpdateStatus(ctx, run); err != nil {
		if k8sErrors.IsConflict(err) {
			logger.Info("Conflict on Run status update, let's try again.")
			return ctrl.Result{RequeueAfter: time.Second * 3}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RunReconciler) handleRunUpdate(ctx context.Context, run *v1beta1.Run) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// If a run is not terminated and not watched it probably mean that
	// - a new run has been created
	// - the controller has crashed and is restarting
	// In that case we start a watcher on the run
	if !run.IsTerminated() && !r.RunWatcher.IsWatched(run) {
		if err := r.RunWatcher.Start(ctx, run); err != nil {
			logger.Error(err, "Cannot start run watcher")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	logger.Info("Run updated",
		logging.RunId, run.Status.Id,
		logging.RunState, run.Status.State,
		logging.ArgoHealth, run.Status.Argo.Health,
	)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Run{}).
		WithEventFilter(predicate.Funcs{
			// Always handle new resource creation
			CreateFunc: func(event.CreateEvent) bool { return true },
			// Let's consider run immutables and only care about update on the status
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldRun, _ := e.ObjectOld.(*v1beta1.Run)
				newRun, _ := e.ObjectNew.(*v1beta1.Run)
				return !reflect.DeepEqual(oldRun.Status, newRun.Status)
			},
			// We don't care about run removal
			DeleteFunc: func(event.DeleteEvent) bool { return false },
		}).
		Complete(r)
}
