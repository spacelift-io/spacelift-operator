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
	"time"

	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	spaceliftRepository "github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
)

// SpaceReconciler reconciles a Space object
type SpaceReconciler struct {
	SpaceRepository          *repository.SpaceRepository
	SpaceliftSpaceRepository spaceliftRepository.SpaceRepository
}

//+kubebuilder:rbac:groups=app.spacelift.io,resources=spaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.spacelift.io,resources=spaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.spacelift.io,resources=spaces/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=create;delete;get;list;patch;update;watch

func (r *SpaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling Space")
	space, err := r.SpaceRepository.Get(ctx, req.NamespacedName)

	// The Space is removed, this should not happen because we filter out deletion events.
	// This can't really hurt and makes the reconciliation logic a bit more straightforward to read
	if err != nil && k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "Unable to retrieve Space from kube API.")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	_, err = r.SpaceliftSpaceRepository.Get(ctx, space)
	if err != nil && !errors.Is(err, spaceliftRepository.ErrSpaceNotFound) {
		return ctrl.Result{}, errors.Wrap(err, "unable to retrieve space from spacelift")
	}

	if errors.Is(err, spaceliftRepository.ErrSpaceNotFound) {
		return r.handleCreateSpace(ctx, space)
	}

	return r.handleUpdateSpace(ctx, space)
}

func (r *SpaceReconciler) handleCreateSpace(ctx context.Context, space *v1beta1.Space) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	spaceliftSpace, err := r.SpaceliftSpaceRepository.Create(ctx, space)
	if err != nil {
		logger.Error(err, "Unable to create space in spacelift")
		return ctrl.Result{}, nil
	}

	if space.Annotations == nil {
		space.Annotations = make(map[string]string, 1)
	}

	space.Annotations[v1beta1.ArgoExternalLink] = spaceliftSpace.URL

	space.SetHealthy()

	// Updating annotations will not trigger another reconciliation loop
	if err := r.SpaceRepository.Update(ctx, space); err != nil {
		if k8sErrors.IsConflict(err) {
			logger.Info("Conflict on Space update, let's try again.")
			return ctrl.Result{RequeueAfter: time.Second * 3}, nil
		}
		return ctrl.Result{}, err
	}

	res, err := r.updateSpaceStatus(ctx, space, *spaceliftSpace)

	logger.WithValues(logging.SpaceId, spaceliftSpace.ID).Info("Space created")

	return res, err
}

func (r *SpaceReconciler) handleUpdateSpace(ctx context.Context, space *v1beta1.Space) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	spaceliftUpdatedSpace, err := r.SpaceliftSpaceRepository.Update(ctx, space)
	if err != nil {
		logger.Error(err, "Unable to update the space in spacelift")
		return ctrl.Result{}, nil
	}

	res, err := r.updateSpaceStatus(ctx, space, *spaceliftUpdatedSpace)

	logger.WithValues(logging.SpaceId, spaceliftUpdatedSpace.ID).Info("Space updated")

	return res, err
}

func (r *SpaceReconciler) updateSpaceStatus(ctx context.Context, space *v1beta1.Space, spaceliftSpace models.Space) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	space.SetSpace(spaceliftSpace)
	if err := r.SpaceRepository.UpdateStatus(ctx, space); err != nil {
		if k8sErrors.IsConflict(err) {
			logger.Info("Conflict on Space status update, let's try again.")
			return ctrl.Result{RequeueAfter: time.Second * 3}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SpaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Space{}).
		WithEventFilter(predicate.Funcs{
			// Always handle new resource creation
			CreateFunc: func(event.CreateEvent) bool { return true },
			// Always handle resource update
			UpdateFunc: func(e event.UpdateEvent) bool { return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() },
			// We don't care about space removal
			DeleteFunc: func(event.DeleteEvent) bool { return false },
		}).
		Complete(r)
}
