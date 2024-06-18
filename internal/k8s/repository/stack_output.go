package repository

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
)

type StackOutputRepository struct {
	client        client.Client
	scheme        *runtime.Scheme
	eventRecorder record.EventRecorder
}

func NewStackOutputRepository(client client.Client, scheme *runtime.Scheme, eventRecorder record.EventRecorder) *StackOutputRepository {
	return &StackOutputRepository{client: client, scheme: scheme, eventRecorder: eventRecorder}
}

func (r *StackOutputRepository) UpdateOrCreateStackOutputSecret(ctx context.Context, stack *v1beta1.Stack, outputs []models.StackOutput) (*v1.Secret, error) {
	secretName := "stack-output-" + stack.ObjectMeta.Name
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

	var invalidOutputs []string
	for _, output := range outputs {
		// If a given output is not compatible, we skip it and save it to a list of invalid outputs
		// We'll then log issues with those outputs.
		// This allows to save stack outputs to secrets in a best effort way.
		if !output.IsCompatibleWithKubeSecret() {
			invalidOutputs = append(invalidOutputs, output.Id)
			continue
		}
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

	message := fmt.Sprintf("Some stack outputs are not compatible with kubernetes secret key format: %s", strings.Join(invalidOutputs, ","))
	r.eventRecorder.Event(
		secret,
		v1.EventTypeWarning,
		v1beta1.EventReasonStackOutputCreated,
		message,
	)

	return secret, nil
}
