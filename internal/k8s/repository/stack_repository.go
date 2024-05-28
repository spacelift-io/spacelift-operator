package repository

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type StackRepository struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewStackRepository(client client.Client, scheme *runtime.Scheme) *StackRepository {
	return &StackRepository{client: client, scheme: scheme}
}

func (r *StackRepository) Get(ctx context.Context, name types.NamespacedName) (*v1beta1.Stack, error) {
	var stack v1beta1.Stack
	if err := r.client.Get(ctx, name, &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (r *StackRepository) Update(ctx context.Context, stack *v1beta1.Stack) error {
	return r.client.Update(ctx, stack)
}

func (r *StackRepository) SetOwner(ctx context.Context, stack *v1beta1.Stack, space *v1beta1.Space) error {
	if err := ctrl.SetControllerReference(space, stack, r.scheme); err != nil {
		return err
	}
	return r.client.Update(ctx, stack)
}

func (r *StackRepository) UpdateStatus(ctx context.Context, stack *v1beta1.Stack) error {
	return r.client.Status().Update(ctx, stack)
}
