package repository

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type PolicyRepository struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewPolicyRepository(client client.Client, scheme *runtime.Scheme) *PolicyRepository {
	return &PolicyRepository{client: client, scheme: scheme}
}

func (r *PolicyRepository) Get(ctx context.Context, name types.NamespacedName) (*v1beta1.Policy, error) {
	var policy v1beta1.Policy
	if err := r.client.Get(ctx, name, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *PolicyRepository) Update(ctx context.Context, policy *v1beta1.Policy) error {
	return r.client.Update(ctx, policy)
}

func (r *PolicyRepository) UpdateStatus(ctx context.Context, policy *v1beta1.Policy) error {
	return r.client.Status().Update(ctx, policy)
}

func (r *PolicyRepository) SetOwner(ctx context.Context, policy *v1beta1.Policy, space *v1beta1.Space) error {
	if err := ctrl.SetControllerReference(space, policy, r.scheme); err != nil {
		return err
	}
	return r.client.Update(ctx, policy)
}
