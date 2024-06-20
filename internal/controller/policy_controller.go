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
	"slices"
	"time"

	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

// PolicyReconciler reconciles a Policy object
type PolicyReconciler struct {
	PolicyRepository          *repository.PolicyRepository
	SpaceRepository           *repository.SpaceRepository
	StackRepository           *repository.StackRepository
	SpaceliftPolicyRepository spaceliftRepository.PolicyRepository
}

//+kubebuilder:rbac:groups=app.spacelift.io,resources=policies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.spacelift.io,resources=policies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.spacelift.io,resources=policies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling Policy")
	policy, err := r.PolicyRepository.Get(ctx, req.NamespacedName)

	// The Policy is removed, this should not happen because we filter out deletion events.
	// This can't really hurt and makes the reconciliation logic a bit more straightforward to read
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "Unable to retrieve Policy from kube API.")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if policy.Spec.SpaceName != nil {
		logger := logger.WithValues(
			logging.SpaceName, *policy.Spec.SpaceName,
			logging.PolicyType, policy.Spec.Type,
			logging.PolicyName, policy.Spec.Name,
		)
		space, err := r.SpaceRepository.Get(ctx, types.NamespacedName{Namespace: policy.Namespace, Name: *policy.Spec.SpaceName})
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				logger.Info("Unable to find space for policy, will retry in 10 seconds")
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
			logger.Error(err, "Error fetching space for policy.")
			return ctrl.Result{}, err
		}
		// If the policy does not have owner reference let's set it
		if len(policy.OwnerReferences) == 0 {
			if err := r.PolicyRepository.SetOwner(ctx, policy, space); err != nil {
				logger.Error(err, "Error setting space owner for policy.")
				return ctrl.Result{}, err
			}
		}

		if !space.Ready() {
			logger.Info("Space is not ready, will retry in 3 seconds")
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		}
		// This set the space ID in the spec object to be reused in the graphql mutation.
		// We kind of use the policy spec as a DTO here, but since we never update the spec in the controller
		// that should be fine.
		policy.Spec.SpaceId = &space.Status.Id
	}

	if len(policy.Spec.AttachedStacksNames) > 0 {
		for _, stackName := range policy.Spec.AttachedStacksNames {
			logger := logger.WithValues(logging.StackName, stackName)
			stack, err := r.StackRepository.Get(ctx, types.NamespacedName{Namespace: policy.Namespace, Name: stackName})
			if err != nil {
				if k8sErrors.IsNotFound(err) {
					logger.Info("Unable to find attached stack for policy, will retry in 10 seconds")
					return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
				}
				logger.Error(err, "Error fetching stack for policy.")
				return ctrl.Result{}, err
			}
			if !stack.Ready() {
				logger.Info("Stack is not ready, will retry in 3 seconds")
				return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
			}
			if !slices.Contains(policy.Spec.AttachedStacksIds, stack.Status.Id) {
				policy.Spec.AttachedStacksIds = append(policy.Spec.AttachedStacksIds, stack.Status.Id)
			}
		}
	}

	_, err = r.SpaceliftPolicyRepository.Get(ctx, policy)
	if err != nil && !errors.Is(err, spaceliftRepository.ErrPolicyNotFound) {
		return ctrl.Result{}, errors.Wrap(err, "unable to retrieve policy from spacelift")
	}

	if errors.Is(err, spaceliftRepository.ErrPolicyNotFound) {
		return r.handleCreatePolicy(ctx, policy)
	}

	return r.handleUpdatePolicy(ctx, policy)
}

func (r *PolicyReconciler) handleCreatePolicy(ctx context.Context, policy *v1beta1.Policy) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	spaceliftPolicy, err := r.SpaceliftPolicyRepository.Create(ctx, policy)
	if err != nil {
		logger.Error(err, "Unable to create policy in spacelift")
		return ctrl.Result{}, nil
	}

	res, err := r.updatePolicyStatus(ctx, policy, *spaceliftPolicy)

	logger.WithValues(logging.PolicyId, spaceliftPolicy.Id).Info("Policy created")

	return res, err
}

func (r *PolicyReconciler) handleUpdatePolicy(ctx context.Context, policy *v1beta1.Policy) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	spaceliftUpdatedPolicy, err := r.SpaceliftPolicyRepository.Update(ctx, policy)
	if err != nil {
		logger.Error(err, "Unable to update the policy in spacelift")
		return ctrl.Result{}, nil
	}

	res, err := r.updatePolicyStatus(ctx, policy, *spaceliftUpdatedPolicy)

	logger.WithValues(logging.PolicyId, spaceliftUpdatedPolicy.Id).Info("Policy updated")

	return res, err
}

func (r *PolicyReconciler) updatePolicyStatus(ctx context.Context, policy *v1beta1.Policy, spaceliftPolicy models.Policy) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	policy.SetPolicy(spaceliftPolicy)
	if err := r.PolicyRepository.UpdateStatus(ctx, policy); err != nil {
		if k8sErrors.IsConflict(err) {
			logger.Info("Conflict on Policy status update, let's try again.")
			return ctrl.Result{RequeueAfter: time.Second * 3}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Policy{}).
		WithEventFilter(predicate.Funcs{
			// Always handle new resource creation
			CreateFunc: func(event.CreateEvent) bool { return true },
			// Always handle resource update
			UpdateFunc: func(e event.UpdateEvent) bool { return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() },
			// We don't care about policy removal
			DeleteFunc: func(event.DeleteEvent) bool { return false },
		}).
		Complete(r)
}
