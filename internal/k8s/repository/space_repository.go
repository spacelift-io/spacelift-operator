package repository

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type SpaceRepository struct {
	client client.Client
}

func NewSpaceRepository(client client.Client) *SpaceRepository {
	return &SpaceRepository{client: client}
}

func (r *SpaceRepository) Get(ctx context.Context, name types.NamespacedName) (*v1beta1.Space, error) {
	var space v1beta1.Space
	if err := r.client.Get(ctx, name, &space); err != nil {
		return nil, err
	}
	return &space, nil
}

func (r *SpaceRepository) Update(ctx context.Context, space *v1beta1.Space) error {
	return r.client.Update(ctx, space)
}

func (r *SpaceRepository) UpdateStatus(ctx context.Context, space *v1beta1.Space) error {
	return r.client.Status().Update(ctx, space)
}
