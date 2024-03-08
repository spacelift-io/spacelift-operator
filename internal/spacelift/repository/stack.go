package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	spaceliftclient "github.com/spacelift-io/spacelift-operator/internal/spacelift/client"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/structs"
)

var (
	ErrStackNotFound = errors.New("stack not found")
)

//go:generate mockery --with-expecter --name StackRepository
type StackRepository interface {
	Create(context.Context, *v1beta1.Stack) (*models.Stack, error)
	Update(context.Context, *v1beta1.Stack) (*models.Stack, error)
	Get(context.Context, *v1beta1.Stack) (*models.Stack, error)
}

type stackRepository struct {
	client client.Client
}

func NewStackRepository(client client.Client) *stackRepository {
	return &stackRepository{client: client}
}

type CreateStackQuery struct {
}

func (r *stackRepository) Create(ctx context.Context, stack *v1beta1.Stack) (*models.Stack, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating stack")
	}

	var mutation struct {
		StackCreate struct {
			ID    string `graphql:"id"`
			State string `graphql:"state"`
		} `graphql:"stackCreate(input: $input, manageState: $manageState, stackObjectID: $stackObjectID, slug: $slug)"`
	}

	stackInput := structs.FromStackSpec(stack.Spec)
	vars := map[string]interface{}{
		"input":         stackInput,
		"manageState":   graphql.Boolean(stack.Spec.ManagesStateFile),
		"stackObjectID": (*graphql.String)(nil),
		"slug":          (*graphql.String)(nil),
	}

	if err := c.Mutate(ctx, &mutation, vars); err != nil {
		return nil, errors.Wrap(err, "unable to create stack")
	}
	url := c.URL("/stack/%s", mutation.StackCreate.ID)
	return &models.Stack{
		Id:    mutation.StackCreate.ID,
		State: mutation.StackCreate.State,
		Url:   url,
	}, nil
}

func (r *stackRepository) Update(ctx context.Context, stack *v1beta1.Stack) (*models.Stack, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating stack")
	}

	var mutation struct {
		StackUpdate struct {
			ID    string `graphql:"id"`
			State string `graphql:"state"`
		} `graphql:"stackUpdate(id: $id, input: $input)"`
	}

	stackInput := structs.FromStackSpec(stack.Spec)
	vars := map[string]interface{}{
		"id":    stack.Status.Id,
		"input": stackInput,
	}

	if err := c.Mutate(ctx, &mutation, vars); err != nil {
		return nil, errors.Wrap(err, "unable to create stack")
	}

	// TODO(michalg): URL can never change here, should we still generate it for k8s api?
	url := c.URL("/stack/%s", mutation.StackUpdate.ID)
	return &models.Stack{
		Id:    mutation.StackUpdate.ID,
		State: mutation.StackUpdate.State,
		Url:   url,
	}, nil
}

func (r *stackRepository) Get(ctx context.Context, stack *v1beta1.Stack) (*models.Stack, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while getting a stack")
	}
	var query struct {
		Stack *struct {
			State string `graphql:"state"`
		} `graphql:"stack(id: $stackId)"`
	}
	vars := map[string]any{
		"stackId": graphql.ID(stack.Spec.Name),
	}
	if err := c.Query(ctx, &query, vars); err != nil {
		return nil, errors.Wrap(err, "unable to get stack")
	}

	if query.Stack == nil {
		return nil, ErrStackNotFound
	}

	return &models.Stack{
		Id:    stack.Spec.Name,
		State: query.Stack.State,
	}, nil
}
