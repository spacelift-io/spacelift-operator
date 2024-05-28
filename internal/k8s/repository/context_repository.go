package repository

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type ContextRepository struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewContextRepository(client client.Client, scheme *runtime.Scheme) *ContextRepository {
	return &ContextRepository{client: client, scheme: scheme}
}

func (r *ContextRepository) Get(ctx context.Context, name types.NamespacedName) (*v1beta1.Context, error) {
	var context v1beta1.Context
	if err := r.client.Get(ctx, name, &context); err != nil {
		return nil, err
	}
	return &context, nil
}

func (r *ContextRepository) UpdateStatus(ctx context.Context, context *v1beta1.Context) error {
	return r.client.Status().Update(ctx, context)
}

func (r *ContextRepository) SetOwner(ctx context.Context, context *v1beta1.Context, space *v1beta1.Space) error {
	if err := ctrl.SetControllerReference(space, context, r.scheme); err != nil {
		return err
	}
	return r.client.Update(ctx, context)
}
