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

var ErrSpaceNotFound = errors.New("space not found")

//go:generate mockery --with-expecter --name SpaceRepository
type SpaceRepository interface {
	Create(context.Context, *v1beta1.Space) (*models.Space, error)
	Update(context.Context, *v1beta1.Space) (*models.Space, error)
	Get(context.Context, *v1beta1.Space) (*models.Space, error)
}

type spaceRepository struct {
	client client.Client
}

func NewSpaceRepository(client client.Client) *spaceRepository {
	return &spaceRepository{client: client}
}

func (r *spaceRepository) Create(ctx context.Context, space *v1beta1.Space) (*models.Space, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, space.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating run")
	}

	var mutation struct {
		SpaceCreate struct {
			ID string `graphql:"id"`
		} `graphql:"spaceCreate(input: $input)"`
	}

	spaceCreationVars := map[string]any{"input": structs.FromSpaceSpec(space)}

	if err := c.Mutate(ctx, &mutation, spaceCreationVars); err != nil {
		return nil, errors.Wrap(err, "unable to create space")
	}

	return &models.Space{
		ID:  mutation.SpaceCreate.ID,
		URL: c.URL("/spaces/%s", mutation.SpaceCreate.ID),
	}, nil
}

func (r *spaceRepository) Update(ctx context.Context, space *v1beta1.Space) (*models.Space, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, space.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while updating space")
	}

	var spaceUpdateMutation struct {
		SpaceUpdate struct {
			ID string `graphql:"id"`
		} `graphql:"spaceUpdate(space: $space, input: $input)"`
	}

	spaceUpdateVars := map[string]any{
		"space": graphql.ID(space.Status.Id),
		"input": structs.FromSpaceSpec(space),
	}

	if err := c.Mutate(ctx, &spaceUpdateMutation, spaceUpdateVars); err != nil {
		return nil, errors.Wrap(err, "unable to update space")
	}

	return &models.Space{
		ID:  spaceUpdateMutation.SpaceUpdate.ID,
		URL: c.URL("/spaces/%s", spaceUpdateMutation.SpaceUpdate.ID),
	}, nil
}

func (r *spaceRepository) Get(ctx context.Context, space *v1beta1.Space) (*models.Space, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, space.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while getting a space")
	}

	var spaceQuery struct {
		Space *struct {
			ID              string   `graphql:"id"`
			Name            string   `graphql:"name"`
			Description     string   `graphql:"description"`
			InheritEntities bool     `graphql:"inheritEntities"`
			ParentSpace     string   `graphql:"parentSpace"`
			Labels          []string `graphql:"labels"`
		} `graphql:"space(id: $id)"`
	}

	spaceVars := map[string]any{"id": graphql.ID(space.Status.Id)}

	if err := c.Query(ctx, &spaceQuery, spaceVars); err != nil {
		return nil, errors.Wrap(err, "unable to get space")
	}

	if spaceQuery.Space == nil {
		return nil, ErrSpaceNotFound
	}

	return &models.Space{
		ID:              spaceQuery.Space.ID,
		Name:            spaceQuery.Space.Name,
		Description:     spaceQuery.Space.Description,
		InheritEntities: spaceQuery.Space.InheritEntities,
		ParentSpace:     spaceQuery.Space.ParentSpace,
		Labels:          spaceQuery.Space.Labels,
		URL:             c.URL("/spaces/%s", spaceQuery.Space.ID),
	}, nil
}
