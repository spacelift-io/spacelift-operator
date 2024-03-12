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

//go:generate mockery --with-expecter --name RunRepository
type RunRepository interface {
	Create(context.Context, *v1beta1.Stack) (*models.Run, error)
	Get(context.Context, *v1beta1.Run) (*models.Run, error)
}

type runRepository struct {
	client client.Client
}

func NewRunRepository(client client.Client) *runRepository {
	return &runRepository{client: client}
}

type CreateRunQuery struct {
}

func (r *runRepository) Create(ctx context.Context, stack *v1beta1.Stack) (*models.Run, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating run")
	}
	var mutation struct {
		RunTrigger struct {
			ID    string `graphql:"id"`
			State string `graphql:"state"`
		} `graphql:"runTrigger(stack: $stack)"`
	}
	vars := map[string]any{
		"stack": graphql.ID(stack.Status.Id),
	}
	if err := c.Mutate(ctx, &mutation, vars); err != nil {
		return nil, errors.Wrap(err, "unable to create run")
	}
	url := c.URL("/stack/%s/run/%s", stack.Status.Id, mutation.RunTrigger.ID)
	return &models.Run{
		Id:      mutation.RunTrigger.ID,
		State:   mutation.RunTrigger.State,
		Url:     url,
		StackId: stack.Status.Id,
	}, nil
}

func (r *runRepository) Get(ctx context.Context, run *v1beta1.Run) (*models.Run, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, run.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating run")
	}
	var query struct {
		Stack struct {
			Run struct {
				State string `graphql:"state"`
			} `graphql:"run(id: $runId)"`
		} `graphql:"stack(id: $stackId)"`
	}
	vars := map[string]any{
		"stackId": graphql.ID(run.Status.StackId),
		"runId":   graphql.ID(run.Status.Id),
	}
	if err := c.Query(ctx, &query, vars); err != nil {
		return nil, errors.Wrap(err, "unable to get run")
	}
	return &models.Run{
		State:   query.Stack.Run.State,
		StackId: run.Status.StackId,
	}, nil
}
