package repository

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretRepository struct {
	client client.Client
}

func NewSecretRepository(client client.Client) *SecretRepository {
	return &SecretRepository{client: client}
}

func (r *SecretRepository) Get(ctx context.Context, name types.NamespacedName) (*v1.Secret, error) {
	secret := &v1.Secret{}
	if err := r.client.Get(ctx, name, secret); err != nil {
		return nil, err
	}
	return secret, nil
}
