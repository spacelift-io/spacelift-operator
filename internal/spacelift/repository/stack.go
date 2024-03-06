package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	spaceliftclient "github.com/spacelift-io/spacelift-operator/internal/spacelift/client"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
)

//go:generate mockery --with-expecter --name StackRepository
type StackRepository interface {
	Create(context.Context, *v1beta1.Stack) (*models.Stack, error)
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

	vars := map[string]interface{}{
		"input":         stack.Spec,
		"manageState":   stack.Spec.ManagesStateFile,
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

func (r *stackRepository) Get(ctx context.Context, stack *v1beta1.Stack) (*models.Stack, error) {
	panic("TODO implement")

	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating stack")
	}
	var query struct {
		Stack struct {
			Stack struct {
				State string `graphql:"state"`
			} `graphql:"stack(id: $stackId)"`
		} `graphql:"stack(id: $stackId)"`
	}
	vars := map[string]any{
		"stackId": graphql.ID(stack.Spec.Name),
	}
	if err := c.Query(ctx, &query, vars); err != nil {
		return nil, errors.Wrap(err, "unable to get stack")
	}
	return &models.Stack{
		State: query.Stack.Stack.State,
	}, nil
}
