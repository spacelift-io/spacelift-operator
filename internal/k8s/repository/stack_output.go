package repository

import (
	"context"
	"strings"

	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
)

type StackOutputRepository struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewStackOutputRepository(client client.Client, scheme *runtime.Scheme) *StackOutputRepository {
	return &StackOutputRepository{client: client, scheme: scheme}
}

func (r *StackOutputRepository) UpdateOrCreateStackOutputSecret(ctx context.Context, stack *v1beta1.Stack, outputs []models.StackOutput) (*v1.Secret, error) {
	secretName := "stack-output-" + stack.Status.Id
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: stack.Namespace,
			Name:      secretName,
		},
	}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: stack.Namespace,
		Name:      secretName,
	}, secret)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return nil, err
	}
	isNewSecret := k8sErrors.IsNotFound(err)

	if secret.Data == nil {
		secret.Data = make(map[string][]byte, len(outputs))
	}

	for _, output := range outputs {
		// TODO(eliecharra): find a way to sanitize output.Id, or ignore invalid keys and log errors?
		secret.Data[output.Id] = []byte(strings.Trim(output.Value, `"`))
	}

	if isNewSecret {
		if err := ctrl.SetControllerReference(stack, secret, r.scheme); err != nil {
			return nil, err
		}
		if err := r.client.Create(ctx, secret); err != nil {
			return nil, err
		}
	} else {
		if err := r.client.Update(ctx, secret); err != nil {
			return nil, err
		}
	}
	return secret, nil
}
