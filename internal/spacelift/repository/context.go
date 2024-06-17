package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	spaceliftclient "github.com/spacelift-io/spacelift-operator/internal/spacelift/client"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/structs"
)

var (
	ErrContextNotFound = errors.New("context not found")
)

//go:generate mockery --with-expecter --name ContextRepository
type ContextRepository interface {
	Create(context.Context, *v1beta1.Context) (*models.Context, error)
	Update(context.Context, *v1beta1.Context) (*models.Context, error)
	Get(context.Context, *v1beta1.Context) (*models.Context, error)
}

type contextRepository struct {
	client client.Client
}

func NewContextRepository(client client.Client) *contextRepository {
	return &contextRepository{client: client}
}

type contextCreateMutation struct {
	ContextCreate struct {
		Id string `graphql:"id"`
	} `graphql:"contextCreateV2(input: $input)"`
}

func (r *contextRepository) Create(ctx context.Context, context *v1beta1.Context) (*models.Context, error) {
	c, err := spaceliftclient.DefaultClient(ctx, r.client, context.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating context")
	}

	var createMutation contextCreateMutation

	contextInput, err := structs.FromContextSpec(context)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build context spec")
	}
	mutationVars := map[string]interface{}{
		"input": *contextInput,
	}

	if err := c.Mutate(ctx, &createMutation, mutationVars); err != nil {
		return nil, errors.Wrap(err, "unable to create context")
	}

	return &models.Context{
		Id: createMutation.ContextCreate.Id,
	}, nil
}

type contextUpdateMutation struct {
	ContextUpdate struct {
		Id string `graphql:"id"`
	} `graphql:"contextUpdateV2(id: $id, input: $input, replaceConfigElements: $replaceConfigElements)"`
}

func (r *contextRepository) Update(ctx context.Context, context *v1beta1.Context) (*models.Context, error) {
	c, err := spaceliftclient.DefaultClient(ctx, r.client, context.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating context")
	}

	var updateMutation contextUpdateMutation

	contextInput, err := structs.FromContextSpec(context)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build context spec")
	}
	mutationVars := map[string]interface{}{
		"id":                    context.Status.Id,
		"input":                 *contextInput,
		"replaceConfigElements": graphql.Boolean(true),
	}

	if err := c.Mutate(ctx, &updateMutation, mutationVars); err != nil {
		return nil, errors.Wrap(err, "unable to update context")
	}

	return &models.Context{
		Id: updateMutation.ContextUpdate.Id,
	}, nil
}

func (r *contextRepository) Get(ctx context.Context, context *v1beta1.Context) (*models.Context, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, context.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while getting a space")
	}
	var query struct {
		Context *struct {
			Id string `graphql:"id"`
		} `graphql:"context(id: $id)"`
	}
	queryVariables := map[string]any{"id": graphql.ID(context.Status.Id)}
	if err := c.Query(ctx, &query, queryVariables); err != nil {
		return nil, errors.Wrap(err, "unable to get context")
	}

	if query.Context == nil {
		return nil, ErrContextNotFound
	}

	return &models.Context{Id: query.Context.Id}, nil
}
