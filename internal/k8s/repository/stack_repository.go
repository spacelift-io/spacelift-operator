package repository

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type StackRepository struct {
	client client.Client
}

func NewStackRepository(client client.Client) *StackRepository {
	return &StackRepository{client: client}
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

func (r *StackRepository) UpdateStatus(ctx context.Context, stack *v1beta1.Stack) error {
	return r.client.Status().Update(ctx, stack)
}
