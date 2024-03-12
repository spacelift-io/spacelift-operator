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

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/pkg/errors"
	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	spaceliftRepository "github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
)

// StackReconciler reconciles a Stack object
type StackReconciler struct {
	StackRepository          *repository.StackRepository
	SpaceliftStackRepository spaceliftRepository.StackRepository
}

//+kubebuilder:rbac:groups=app.spacelift.io,resources=stacks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.spacelift.io,resources=stacks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.spacelift.io,resources=stacks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Stack object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *StackReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling Stack")
	stack, err := r.StackRepository.Get(ctx, req.NamespacedName)

	// The Stack is removed, this should not happen because we filter out deletion events.
	// This can't really hurt and makes the reconciliation logic a bit more straightforward to read
	if err != nil && k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "Unable to retrieve Stack from kube API.")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if stack.IsNew() {
		return r.handleCreateOrUpdateStack(ctx, stack)
	}

	return ctrl.Result{}, nil
}

func (r *StackReconciler) handleCreateOrUpdateStack(ctx context.Context, stack *v1beta1.Stack) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	retStack, err := r.SpaceliftStackRepository.Get(ctx, stack)
	logger.Info("Stack is new", "stack", stack, "retStack", retStack, "err", err)

	if err != nil {
		if err == spaceliftRepository.ErrStackNotFound {
			return r.handleCreateStack(ctx, stack)
		}
		return ctrl.Result{}, errors.Wrap(err, "unable to retrieve stack from spacelift")
	}

	return r.handleUpdateStack(ctx, stack)

}

func (r *StackReconciler) handleUpdateStack(ctx context.Context, stack *v1beta1.Stack) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	spaceliftStack, err := r.SpaceliftStackRepository.Update(ctx, stack)
	if err != nil {
		logger.Error(err, "Unable to update the stack in spacelift")
		// TODO(eliecharra): Implement better error handling and retry errors that could be retried
		return ctrl.Result{}, nil
	}

	// just being defensive here, this should never happen
	if spaceliftStack == nil {
		return ctrl.Result{}, errors.New("returned spacelift stack is nil when updating stack in spacelift")
	}

	logger.WithValues(
		logging.StackId, spaceliftStack.Id,
	).Info("Stack updated")

	return r.updateK8sStackCRD(ctx, stack, *spaceliftStack)
}

func (r *StackReconciler) handleCreateStack(ctx context.Context, stack *v1beta1.Stack) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	spaceliftStack, err := r.SpaceliftStackRepository.Create(ctx, stack)
	if err != nil {
		logger.Error(err, "Unable to create the stack in spacelift")
		// TODO(eliecharra): Implement better error handling and retry errors that could be retried
		return ctrl.Result{}, nil
	}

	// just being defensive here, this should never happen
	if spaceliftStack == nil {
		return ctrl.Result{}, errors.New("returned spacelift stack is nil when creating stack in spacelift")
	}

	logger.WithValues(
		logging.StackId, spaceliftStack.Id,
	).Info("Stack created")

	return r.updateK8sStackCRD(ctx, stack, *spaceliftStack)
}

func (r *StackReconciler) updateK8sStackCRD(ctx context.Context, stack *v1beta1.Stack, spaceliftStack models.Stack) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Set initial annotations when stack is created
	if stack.Annotations == nil {
		stack.Annotations = make(map[string]string, 1)
	}

	stack.Annotations[v1beta1.ArgoExternalLink] = spaceliftStack.Url

	// Updating annotations will not trigger another reconciliation loop
	if err := r.StackRepository.Update(ctx, stack); err != nil {
		if k8sErrors.IsConflict(err) {
			logger.Info("Conflict on Stack update, let's try again.")
			return ctrl.Result{RequeueAfter: time.Second * 3}, nil
		}
		return ctrl.Result{}, err
	}

	stack.SetStack(spaceliftStack)
	if err := r.StackRepository.UpdateStatus(ctx, stack); err != nil {
		if k8sErrors.IsConflict(err) {
			logger.Info("Conflict on Stack status update, let's try again.")
			return ctrl.Result{RequeueAfter: time.Second * 3}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Stack{}).
		WithEventFilter(predicate.Funcs{
			// Always handle new resource creation
			CreateFunc: func(event.CreateEvent) bool { return true },
			// Always handle resource update
			UpdateFunc: func(e event.UpdateEvent) bool { return true },
			// We don't care about stack removal
			DeleteFunc: func(event.DeleteEvent) bool { return false },
		}).
		Complete(r)
}
