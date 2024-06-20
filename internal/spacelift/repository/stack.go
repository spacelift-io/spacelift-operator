package repository

import (
	"context"
	"slices"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
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

type stackCreateMutation struct {
	StackCreate struct {
		ID string `graphql:"id"`
	} `graphql:"stackCreate(input: $input, manageState: $manageState)"`
}

func (r *stackRepository) Create(ctx context.Context, stack *v1beta1.Stack) (*models.Stack, error) {
	c, err := spaceliftclient.DefaultClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating stack")
	}

	var mutation stackCreateMutation

	stackInput := structs.FromStackSpec(stack)
	stackCreateMutationVars := map[string]interface{}{
		"input":       stackInput,
		"manageState": graphql.Boolean(true),
	}

	if stack.Spec.ManagesStateFile != nil {
		stackCreateMutationVars["manageState"] = graphql.Boolean(*stack.Spec.ManagesStateFile)
	}

	if err := c.Mutate(ctx, &mutation, stackCreateMutationVars); err != nil {
		return nil, errors.Wrap(err, "unable to create stack")
	}
	url := c.URL("/stack/%s", mutation.StackCreate.ID)

	stack.Status.Id = mutation.StackCreate.ID
	if stack.Spec.AWSIntegration != nil {
		if err := r.attachAWSIntegration(ctx, stack); err != nil {
			return nil, errors.Wrap(err, "unable to attach AWS integration to stack")
		}
	}

	if stack.Spec.CommitSHA != nil && *stack.Spec.CommitSHA != "" {
		if err := r.setTrackedCommit(ctx, c, mutation.StackCreate.ID, *stack.Spec.CommitSHA); err != nil {
			return nil, errors.Wrap(err, "unable to set tracked commit on stack")
		}
	}

	return &models.Stack{
		Id:  mutation.StackCreate.ID,
		Url: url,
	}, nil
}

type awsIntegrationAttachMutation struct {
	AWSIntegrationAttach struct {
		Id string `graphql:"id"`
	} `graphql:"awsIntegrationAttach(id: $id, stack: $stack, read: $read, write: $write)"`
}

func (r *stackRepository) attachAWSIntegration(ctx context.Context, stack *v1beta1.Stack) error {
	c, err := spaceliftclient.DefaultClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to fetch spacelift client while creating stack")
	}
	var mutation awsIntegrationAttachMutation
	awsIntegrationAttachVars := map[string]any{
		"id":    stack.Spec.AWSIntegration.Id,
		"stack": stack.Status.Id,
		"read":  graphql.Boolean(stack.Spec.AWSIntegration.Read),
		"write": graphql.Boolean(stack.Spec.AWSIntegration.Write),
	}

	if err := c.Mutate(ctx, &mutation, awsIntegrationAttachVars); err != nil {
		return err
	}

	return nil
}

type awsIntegrationDetachMutation struct {
	AWSIntegrationDetach struct {
		ID string `graphql:"id"`
	} `graphql:"awsIntegrationDetach(id: $id)"`
}

func (r *stackRepository) detachAWSIntegration(ctx context.Context, stack *v1beta1.Stack, id string) error {
	c, err := spaceliftclient.DefaultClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to fetch spacelift client while detaching AWS integration")
	}
	var mutation awsIntegrationDetachMutation
	vars := map[string]any{
		"id": graphql.ID(id),
	}

	if err := c.Mutate(ctx, &mutation, vars); err != nil {
		return err
	}

	return nil
}

type stackUpdateMutationAWSIntegration struct {
	ID            string `graphql:"id"`
	IntegrationID string `graphql:"integrationId"`
	Read          bool   `graphql:"read"`
	Write         bool   `graphql:"write"`
}

type stackUpdateMutation struct {
	StackUpdate struct {
		ID                      string                              `graphql:"id"`
		State                   string                              `graphql:"state"`
		AttachedAWSIntegrations []stackUpdateMutationAWSIntegration `graphql:"attachedAwsIntegrations"`
	} `graphql:"stackUpdate(id: $id, input: $input)"`
}

func (r *stackRepository) Update(ctx context.Context, stack *v1beta1.Stack) (*models.Stack, error) {
	c, err := spaceliftclient.DefaultClient(ctx, r.client, stack.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while updating stack")
	}

	var mutation stackUpdateMutation

	stackInput := structs.FromStackSpec(stack)
	vars := map[string]interface{}{
		"id":    slug.SafeSlug(stack.Name()),
		"input": stackInput,
	}

	if err := c.Mutate(ctx, &mutation, vars); err != nil {
		return nil, errors.Wrap(err, "unable to create stack")
	}

	logger := log.FromContext(ctx).WithValues(logging.StackId, mutation.StackUpdate.ID)

	// First we check if there are any integrations to detach
	// Any existing attachment that does not match spec.AWSIntegration will be detached.
	attachedIntegrations := mutation.StackUpdate.AttachedAWSIntegrations
	for i, integration := range attachedIntegrations {
		if stack.Spec.AWSIntegration == nil ||
			stack.Spec.AWSIntegration.Id != integration.IntegrationID ||
			(stack.Spec.AWSIntegration.Id == integration.IntegrationID &&
				(stack.Spec.AWSIntegration.Read != integration.Read ||
					stack.Spec.AWSIntegration.Write != integration.Write)) {
			if err := r.detachAWSIntegration(ctx, stack, integration.ID); err != nil {
				return nil, errors.Wrap(err, "unable to detach AWS integration from stack")
			}
			logger.Info("Detached AWS integration from stack", logging.StackAWSIntegrationId, integration.IntegrationID)
			// If we are detaching an integration, we also reflect this change to the mutation array.
			// This allows an integration to be detached and reattached in a row if a read or write attribute has been changed.
			mutation.StackUpdate.AttachedAWSIntegrations = append(mutation.StackUpdate.AttachedAWSIntegrations[:i], mutation.StackUpdate.AttachedAWSIntegrations[i+1:]...)
		}
	}

	if stack.Spec.AWSIntegration != nil {
		shouldAttachInteration := !slices.ContainsFunc(mutation.StackUpdate.AttachedAWSIntegrations, func(s stackUpdateMutationAWSIntegration) bool {
			return s.IntegrationID == stack.Spec.AWSIntegration.Id
		})
		if shouldAttachInteration {
			if err := r.attachAWSIntegration(ctx, stack); err != nil {
				return nil, errors.Wrap(err, "unable to attach AWS integration to stack")
			}
			logger.Info("Attached AWS integration to stack", logging.StackAWSIntegrationId, stack.Spec.AWSIntegration.Id)
		}
	}

	// TODO(michalg): URL can never change here, should we still generate it for k8s api?
	url := c.URL("/stack/%s", mutation.StackUpdate.ID)
	return &models.Stack{
		Id:  mutation.StackUpdate.ID,
		Url: url,
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
		"stackId": graphql.ID(slug.SafeSlug(stack.Name())),
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

type setTrackedCommitMutation struct {
	Stack struct {
		ID string `graphql:"id"`
	} `graphql:"stackSetCurrentCommit(id: $id, sha: $sha)"`
}

func (r *stackRepository) setTrackedCommit(ctx context.Context, c spaceliftclient.Client, stackID, commitSHA string) error {

	var mutation setTrackedCommitMutation

	setTrackedCommitMutationVars := map[string]interface{}{
		"id":  stackID,
		"sha": graphql.String(commitSHA),
	}

	if err := c.Mutate(ctx, &mutation, setTrackedCommitMutationVars); err != nil {
		return errors.Wrap(err, "unable to set tracked commit on stack")
	}

	return nil
}
