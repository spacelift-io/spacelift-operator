package repository

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type RunRepository struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewRunRepository(client client.Client, scheme *runtime.Scheme) *RunRepository {
	return &RunRepository{client: client, scheme: scheme}
}

func (r *RunRepository) Get(ctx context.Context, name types.NamespacedName) (*v1beta1.Run, error) {
	var run v1beta1.Run
	if err := r.client.Get(ctx, name, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *RunRepository) SetOwner(ctx context.Context, run *v1beta1.Run, stack *v1beta1.Stack) error {
	if err := ctrl.SetControllerReference(stack, run, r.scheme); err != nil {
		return err
	}
	return r.client.Update(ctx, run)
}

func (r *RunRepository) Update(ctx context.Context, run *v1beta1.Run) error {
	return r.client.Update(ctx, run)
}

func (r *RunRepository) UpdateStatus(ctx context.Context, run *v1beta1.Run) error {
	return r.client.Status().Update(ctx, run)
}
