package repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	spaceliftclient "github.com/spacelift-io/spacelift-operator/internal/spacelift/client"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/slug"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/structs"
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

	var stackCreateMutation struct {
		StackCreate struct {
			ID            string `graphql:"id"`
			State         string `graphql:"state"`
			TrackedCommit struct {
				AuthorLogin *string `graphql:"authorLogin"`
				AuthorName  string  `graphql:"authorName"`
				Hash        string  `graphql:"hash"`
				Message     string  `graphql:"message"`
				Timestamp   uint    `graphql:"timestamp"`
				URL         *string `graphql:"url"`
			} `graphql:"trackedCommit"`
			TrackedCommitSetBy *string `graphql:"trackedCommitSetBy"`
		} `graphql:"stackCreate(input: $input, manageState: $manageState)"`
	}

	stackInput := structs.FromStackSpec(stack.Spec)
	stackCreateMutationVars := map[string]interface{}{
		"input":       stackInput,
		"manageState": graphql.Boolean(true),
	}

	if stack.Spec.Settings.ManagesStateFile != nil {
		stackCreateMutationVars["manageState"] = graphql.Boolean(*stack.Spec.Settings.ManagesStateFile)
	}

	if err := c.Mutate(ctx, &stackCreateMutation, stackCreateMutationVars); err != nil {
		return nil, errors.Wrap(err, "unable to create stack")
	}
	url := c.URL("/stack/%s", stackCreateMutation.StackCreate.ID)

	stack.Status.Id = stackCreateMutation.StackCreate.ID
	if stack.Spec.Settings.AWSIntegration != nil {
		if err := r.attachAWSIntegration(ctx, stack); err != nil {
			return nil, errors.Wrap(err, "unable to attach AWS integration to stack")
		}
	}

	trackedCommit := &models.Commit{
		AuthorLogin: stackCreateMutation.StackCreate.TrackedCommit.AuthorLogin,
		AuthorName:  stackCreateMutation.StackCreate.TrackedCommit.AuthorName,
		Hash:        stackCreateMutation.StackCreate.TrackedCommit.Hash,
		Message:     stackCreateMutation.StackCreate.TrackedCommit.Message,
		Timestamp:   stackCreateMutation.StackCreate.TrackedCommit.Timestamp,
		URL:         stackCreateMutation.StackCreate.TrackedCommit.URL,
	}
	trackedCommitSetBy := stackCreateMutation.StackCreate.TrackedCommitSetBy
	if stack.Spec.CommitSHA != nil && *stack.Spec.CommitSHA != "" {
		trackedCommit, trackedCommitSetBy, err = r.setTrackedCommit(ctx, c, stackCreateMutation.StackCreate.ID, *stack.Spec.CommitSHA)
		if err != nil {
			return nil, errors.Wrap(err, "unable to set tracked commit on stack")
		}
	}

	return &models.Stack{
		Id:                 stackCreateMutation.StackCreate.ID,
		State:              stackCreateMutation.StackCreate.State,
		Url:                url,
		TrackedCommit:      trackedCommit,
		TrackedCommitSetBy: trackedCommitSetBy,
	}, nil
}

func (r *stackRepository) attachAWSIntegration(ctx context.Context, stack *v1beta1.Stack) error {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to fetch spacelift client while creating stack")
	}
	var awsIntegrationAttachMutation struct {
		AWSIntegrationAttach struct {
			Id string `graphql:"id"`
		} `graphql:"awsIntegrationAttach(id: $id, stack: $stack, read: $read, write: $write)"`
	}
	awsIntegrationAttachVars := map[string]any{
		"id":    stack.Spec.Settings.AWSIntegration.Id,
		"stack": stack.Status.Id,
		"read":  graphql.Boolean(stack.Spec.Settings.AWSIntegration.Read),
		"write": graphql.Boolean(stack.Spec.Settings.AWSIntegration.Write),
	}
	if err := c.Mutate(ctx, &awsIntegrationAttachMutation, awsIntegrationAttachVars); err != nil {
		return err
	}

	return nil
}

func (r *stackRepository) Update(ctx context.Context, stack *v1beta1.Stack) (*models.Stack, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while updating stack")
	}

	var mutation struct {
		StackUpdate struct {
			ID    string `graphql:"id"`
			State string `graphql:"state"`
		} `graphql:"stackUpdate(id: $id, input: $input)"`
	}

	stackInput := structs.FromStackSpec(stack.Spec)
	vars := map[string]interface{}{
		"id":    slug.SafeSlug(stack.Spec.Name),
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
			Id      string `graphql:"id"`
			Outputs []struct {
				Id    string `graphql:"id"`
				Value string `graphql:"value"`
			} `graphql:"outputs"`
		} `graphql:"stack(id: $stackId)"`
	}
	vars := map[string]any{
		"stackId": graphql.ID(slug.SafeSlug(stack.Spec.Name)),
	}
	if err := c.Query(ctx, &query, vars); err != nil {
		return nil, errors.Wrap(err, "unable to get stack")
	}

	if query.Stack == nil {
		return nil, ErrStackNotFound
	}

	s := &models.Stack{
		Id:      query.Stack.Id,
		Outputs: make([]models.StackOutput, 0, len(query.Stack.Outputs)),
	}

	for _, output := range query.Stack.Outputs {
		s.Outputs = append(s.Outputs, models.StackOutput{
			Id:    output.Id,
			Value: output.Value,
		})
	}

	return s, nil
}

func (r *stackRepository) setTrackedCommit(ctx context.Context, c spaceliftclient.Client, stackID, commitSHA string) (*models.Commit, *string, error) {

	var setTrackedCommitMutation struct {
		Stack struct {
			ID            string `graphql:"id"`
			State         string `graphql:"state"`
			TrackedCommit struct {
				AuthorLogin *string `graphql:"authorLogin"`
				AuthorName  string  `graphql:"authorName"`
				Hash        string  `graphql:"hash"`
				Message     string  `graphql:"message"`
				Timestamp   uint    `graphql:"timestamp"`
				URL         *string `graphql:"url"`
			} `graphql:"trackedCommit"`
			TrackedCommitSetBy *string `graphql:"trackedCommitSetBy"`
		} `graphql:"stackSetCurrentCommit(id: $id, sha: $sha)"`
	}

	setTrackedCommitMutationVars := map[string]interface{}{
		"id":  stackID,
		"sha": graphql.String(commitSHA),
	}

	if err := c.Mutate(ctx, &setTrackedCommitMutation, setTrackedCommitMutationVars); err != nil {
		return nil, nil, errors.Wrap(err, "unable to set tracked commit on stack")
	}

	return &models.Commit{
		AuthorLogin: setTrackedCommitMutation.Stack.TrackedCommit.AuthorLogin,
		AuthorName:  setTrackedCommitMutation.Stack.TrackedCommit.AuthorName,
		Hash:        setTrackedCommitMutation.Stack.TrackedCommit.Hash,
		Message:     setTrackedCommitMutation.Stack.TrackedCommit.Message,
		Timestamp:   setTrackedCommitMutation.Stack.TrackedCommit.Timestamp,
		URL:         setTrackedCommitMutation.Stack.TrackedCommit.URL,
	}, setTrackedCommitMutation.Stack.TrackedCommitSetBy, nil
}
