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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
	spaceliftRepository "github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
	"github.com/spacelift-io/spacelift-operator/internal/utils"
)

// ContextReconciler reconciles a Context object
type ContextReconciler struct {
	SpaceliftContextRepository spaceliftRepository.ContextRepository
	ContextRepository          *repository.ContextRepository
	StackRepository            *repository.StackRepository
	SpaceRepository            *repository.SpaceRepository
	SecretRepository           *repository.SecretRepository
}

//+kubebuilder:rbac:groups=app.spacelift.io,resources=contexts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.spacelift.io,resources=contexts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.spacelift.io,resources=contexts/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *ContextReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling Context")
	context, err := r.ContextRepository.Get(ctx, req.NamespacedName)

	// The Context is removed, this should not happen because we filter out deletion events.
	// This can't really hurt and makes the reconciliation logic a bit more straightforward to read
	if err != nil && k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "Unable to retrieve Context from kube API.")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger = logger.WithValues(logging.ContextName, context.Spec.Name)
	log.IntoContext(ctx, logger)

	// A context should always be linked to a valid space
	if context.Spec.SpaceName != nil {
		logger := logger.WithValues(logging.SpaceName, *context.Spec.SpaceName)
		space, err := r.SpaceRepository.Get(ctx, types.NamespacedName{Namespace: context.Namespace, Name: *context.Spec.SpaceName})
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				logger.V(logging.Level4).Info("Unable to find space for context, will retry in 10 seconds")
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
			logger.Error(err, "Error fetching space for context.")
			return ctrl.Result{}, err
		}
		// If the context does not have owner reference let's set it
		if len(context.OwnerReferences) == 0 {
			if err := r.ContextRepository.SetOwner(ctx, context, space); err != nil {
				logger.Error(err, "Error setting space owner for context.")
				return ctrl.Result{}, err
			}
		}

		if !space.Ready() {
			logger.Info("Space is not ready, will retry in 3 seconds")
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		}
		// This set the space ID in the spec object to be reused in the graphql mutation.
		// We kind of use the context spec as a DTO here, but since we never update the spec in the controller
		// that should be fine.
		context.Spec.SpaceId = &space.Status.Id
	}

	// For all stack attachment, ensure that all stacks are ready
	for i, attachment := range context.Spec.Attachments {
		if attachment.StackName != nil {
			logger := logger.WithValues(logging.StackName, *attachment.StackName)
			// Test if stack exists and is ready
			stack, err := r.StackRepository.Get(ctx, types.NamespacedName{
				Namespace: context.Namespace,
				Name:      *attachment.StackName,
			})
			if err != nil {
				if k8sErrors.IsNotFound(err) {
					logger.V(logging.Level4).Info("Unable to find stack for context, will retry in 10 seconds")
					return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
				}
				logger.Error(err, "Error fetching stack for context.")
				return ctrl.Result{}, err
			}
			if !stack.Ready() {
				logger.Info("Stack is not ready, will retry in 3 seconds")
				return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
			}
			// This set the stack ID in the spec object to be reused in the graphql mutation.
			context.Spec.Attachments[i].StackId = &stack.Status.Id
		}
	}

	for i, environment := range context.Spec.Environment {
		if environment.ValueFromSecret != nil {
			logger := logger.WithValues(
				logging.SecretName, environment.ValueFromSecret.Name,
				logging.EnvironmentId, environment.Id,
			)
			secret, err := r.SecretRepository.Get(ctx, types.NamespacedName{
				Namespace: context.Namespace,
				Name:      environment.ValueFromSecret.Name,
			})
			if err != nil {
				if k8sErrors.IsNotFound(err) {
					logger.Info("Unable to find secret for context environment variable, will retry in 3 seconds.")
					return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
				}
				logger.Error(err, "Error fetching secret for context environment variable.")
				return ctrl.Result{}, err
			}
			if _, keyExist := secret.Data[environment.ValueFromSecret.Key]; !keyExist {
				err := errors.New("key does not exist for secret")
				logger.WithValues(
					logging.SecretKey, environment.ValueFromSecret.Key,
				).Error(err, "Error fetching mountedFile secret for context.")
				return ctrl.Result{}, err
			}
			valueFromSecret := string(secret.Data[environment.ValueFromSecret.Key])
			context.Spec.Environment[i].Value = &valueFromSecret
			context.Spec.Environment[i].Secret = utils.AddressOf(true)
		}
	}

	for i, mountedFile := range context.Spec.MountedFiles {
		if mountedFile.ValueFromSecret != nil {
			logger := logger.WithValues(logging.SecretName, mountedFile.ValueFromSecret.Name)
			secret, err := r.SecretRepository.Get(ctx, types.NamespacedName{
				Namespace: context.Namespace,
				Name:      mountedFile.ValueFromSecret.Name,
			})
			if err != nil {
				if k8sErrors.IsNotFound(err) {
					logger.Info("Unable to find secret for context mounted file, will retry in 3 seconds.")
					return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
				}
				logger.Error(err, "Error fetching secret for context mounted file.")
				return ctrl.Result{}, err
			}
			if _, keyExist := secret.Data[mountedFile.ValueFromSecret.Key]; !keyExist {
				err := errors.New("key does not exist for secret")
				logger.WithValues(
					logging.SecretKey, mountedFile.ValueFromSecret.Key,
				).Error(err, "Error fetching mounted file secret for context.")
				return ctrl.Result{}, err
			}
			valueFromSecret := string(secret.Data[mountedFile.ValueFromSecret.Key])
			context.Spec.MountedFiles[i].Value = &valueFromSecret
			context.Spec.MountedFiles[i].Secret = utils.AddressOf(true)
		}
	}

	_, err = r.SpaceliftContextRepository.Get(ctx, context)
	if err != nil && !errors.Is(err, spaceliftRepository.ErrContextNotFound) {
		return ctrl.Result{}, errors.Wrap(err, "unable to retrieve context from spacelift")
	}

	// Context does not exist in Spacelift, let's create it
	if errors.Is(err, spaceliftRepository.ErrContextNotFound) {
		return r.handleCreateContext(ctx, context)
	}

	return r.handleUpdateContext(ctx, context)
}

func (r *ContextReconciler) handleCreateContext(ctx context.Context, context *v1beta1.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	spaceliftContext, err := r.SpaceliftContextRepository.Create(ctx, context)
	if err != nil {
		logger.Error(err, "Unable to create the context in spacelift")
		// TODO: Implement better error handling and retry errors that could be retried
		return ctrl.Result{}, nil
	}

	context.SetContext(spaceliftContext)
	res, err := r.updateContextStatus(ctx, context)

	logger.WithValues(
		logging.ContextId, spaceliftContext.Id,
	).Info("Context created")

	return res, err
}

func (r *ContextReconciler) handleUpdateContext(ctx context.Context, context *v1beta1.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	spaceliftUpdatedContext, err := r.SpaceliftContextRepository.Update(ctx, context)
	if err != nil {
		logger.Error(err, "Unable to update the context in spacelift")
		// TODO: Implement better error handling and retry errors that could be retried
		return ctrl.Result{}, nil
	}

	context.SetContext(spaceliftUpdatedContext)
	res, err := r.updateContextStatus(ctx, context)

	logger.WithValues(
		logging.ContextId, spaceliftUpdatedContext.Id,
	).Info("Context updated")

	return res, err
}

func (r *ContextReconciler) updateContextStatus(ctx context.Context, context *v1beta1.Context) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if err := r.ContextRepository.UpdateStatus(ctx, context); err != nil {
		if k8sErrors.IsConflict(err) {
			logger.Info("Conflict on Context status update, let's try again.")
			return ctrl.Result{RequeueAfter: time.Second * 3}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContextReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Context{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		WithEventFilter(predicate.Funcs{
			// Always handle new resource creation
			CreateFunc: func(event.CreateEvent) bool { return true },
			// Always handle resource update
			UpdateFunc: func(e event.UpdateEvent) bool { return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() },
			// We don't care about context removal
			DeleteFunc: func(event.DeleteEvent) bool { return false },
		}).
		Complete(r)
}
