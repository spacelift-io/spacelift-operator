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
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/structs"
)

var ErrPolicyNotFound = errors.New("policy not found")

//go:generate mockery --with-expecter --name PolicyRepository
type PolicyRepository interface {
	Create(context.Context, *v1beta1.Policy) (*models.Policy, error)
	Update(context.Context, *v1beta1.Policy) (*models.Policy, error)
	Get(context.Context, *v1beta1.Policy) (*models.Policy, error)
}

type policyRepository struct {
	client client.Client
}

type PolicyType string
type attachedStack struct {
	Id             string `graphql:"id"`
	StackId        string `graphql:"stackId"`
	IsAutoAttached bool   `graphql:"isAutoattached"`
}
type policyCreate struct {
	Id             string          `graphql:"id"`
	AttachedStacks []attachedStack `graphql:"attachedStacks"`
}
type policyCreateMutation struct {
	PolicyCreate policyCreate `graphql:"policyCreate(name: $name, body: $body, type: $type, labels: $labels, space: $space)"`
}
type policyAttach struct {
	Id string `graphql:"id"`
}
type policyAttachMutation struct {
	PolicyAttach struct {
		Id string `graphql:"id"`
	} `graphql:"policyAttach(id: $id, stack: $stack)"`
}
type policyDetachMutation struct {
	PolicyDetach struct {
		Id string `graphql:"id"`
	} `graphql:"policyDetach(id: $id)"`
}
type policyUpdate struct {
	Id             string          `graphql:"id"`
	AttachedStacks []attachedStack `graphql:"attachedStacks"`
}
type policyUpdateMutation struct {
	PolicyUpdate policyUpdate `graphql:"policyUpdate(id: $id, name: $name, body: $body, labels: $labels, space: $space)"`
}

func NewPolicyRepository(client client.Client) *policyRepository {
	return &policyRepository{client: client}
}

func (r *policyRepository) Create(ctx context.Context, policy *v1beta1.Policy) (*models.Policy, error) {
	c, err := spaceliftclient.DefaultClient(ctx, r.client, policy.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while creating policy")
	}

	var mutation policyCreateMutation
	creationVars := map[string]any{
		"name":   graphql.String(policy.Name()),
		"body":   graphql.String(policy.Spec.Body),
		"type":   PolicyType(policy.Spec.Type),
		"labels": structs.GetGraphQLStrings(&policy.Spec.Labels),
		"space":  (*graphql.ID)(nil),
	}

	if policy.Spec.SpaceId != nil && *policy.Spec.SpaceId != "" {
		creationVars["space"] = graphql.ID(*policy.Spec.SpaceId)
	}

	if err := c.Mutate(ctx, &mutation, creationVars); err != nil {
		return nil, errors.Wrap(err, "unable to create policy")
	}

	logger := log.FromContext(ctx).WithValues(logging.PolicyId, mutation.PolicyCreate.Id)

	policyId := mutation.PolicyCreate.Id

	stacksToAttach := r.findStackToAttach(policy, mutation.PolicyCreate.AttachedStacks)
	for _, stackId := range stacksToAttach {
		attachMutation := policyAttachMutation{}
		if err := c.Mutate(ctx, &attachMutation, map[string]any{
			"id":    policyId,
			"stack": stackId,
		}); err != nil {
			return nil, errors.Wrapf(err, "unable to attach stack %s to policy %s", stackId, policyId)
		}
		logger.WithValues(logging.StackId, stackId).Info("Attached stack to policy")
	}

	return &models.Policy{
		Id: policyId,
	}, nil
}

func (r *policyRepository) Update(ctx context.Context, policy *v1beta1.Policy) (*models.Policy, error) {
	c, err := spaceliftclient.DefaultClient(ctx, r.client, policy.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while updating policy")
	}

	var updateMutation policyUpdateMutation
	updateVars := map[string]any{
		"id":     graphql.ID(policy.Status.Id),
		"name":   graphql.String(policy.Name()),
		"body":   graphql.String(policy.Spec.Body),
		"labels": structs.GetGraphQLStrings(&policy.Spec.Labels),
		"space":  (*graphql.ID)(nil),
	}

	if policy.Spec.SpaceId != nil && *policy.Spec.SpaceId != "" {
		updateVars["space"] = graphql.ID(*policy.Spec.SpaceId)
	}

	if err := c.Mutate(ctx, &updateMutation, updateVars); err != nil {
		return nil, errors.Wrap(err, "unable to update policy")
	}

	logger := log.FromContext(ctx).WithValues(logging.PolicyId, updateMutation.PolicyUpdate.Id)

	policyId := updateMutation.PolicyUpdate.Id
	stacksToAttach := r.findStackToAttach(policy, updateMutation.PolicyUpdate.AttachedStacks)
	attachmentsToDetach := r.findStackToDetach(policy, updateMutation.PolicyUpdate.AttachedStacks)

	for _, attachmentId := range attachmentsToDetach {
		detachMutation := policyDetachMutation{}
		if err := c.Mutate(ctx, &detachMutation, map[string]any{
			"id": attachmentId,
		}); err != nil {
			return nil, errors.Wrapf(err, "unable to remove attachment %s to policy %s", attachmentId, policyId)
		}
		logger.WithValues(logging.PolicyAttachmentId, attachmentId).Info("Removed policy attachment")
	}

	for _, stackId := range stacksToAttach {
		attachMutation := policyAttachMutation{}
		if err := c.Mutate(ctx, &attachMutation, map[string]any{
			"id":    policyId,
			"stack": stackId,
		}); err != nil {
			return nil, errors.Wrapf(err, "unable to attach stack %s to policy %s", stackId, policyId)
		}
		logger.WithValues(logging.StackId, stackId).Info("Attached stack to policy")
	}

	return &models.Policy{
		Id: policyId,
	}, nil
}

func (r *policyRepository) Get(ctx context.Context, policy *v1beta1.Policy) (*models.Policy, error) {
	c, err := spaceliftclient.GetSpaceliftClient(ctx, r.client, policy.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch spacelift client while getting a policy")
	}

	var spaceQuery struct {
		Policy *struct {
			Id string `graphql:"id"`
		} `graphql:"policy(id: $id)"`
	}

	vars := map[string]any{"id": graphql.ID(policy.Status.Id)}

	if err := c.Query(ctx, &spaceQuery, vars); err != nil {
		return nil, errors.Wrap(err, "unable to get policy")
	}

	if spaceQuery.Policy == nil {
		return nil, ErrPolicyNotFound
	}

	return &models.Policy{
		Id: spaceQuery.Policy.Id,
	}, nil
}

func (*policyRepository) findStackToAttach(policy *v1beta1.Policy, attachedStacks []attachedStack) []string {
	var stacksToAttach []string
stacksToAttach:
	for _, stacksId := range policy.Spec.AttachedStacksIds {
		// Let's see if the stack is already attached
		if slices.ContainsFunc(
			attachedStacks,
			func(attachedStack attachedStack) bool { return attachedStack.StackId == stacksId },
		) {
			continue stacksToAttach
		}
		stacksToAttach = append(stacksToAttach, stacksId)
	}
	return stacksToAttach
}

func (*policyRepository) findStackToDetach(policy *v1beta1.Policy, attachedStacks []attachedStack) []string {
	var attachmentsToDetach []string
	for _, attachedStack := range attachedStacks {
		if attachedStack.IsAutoAttached {
			continue
		}
		if slices.Contains(policy.Spec.AttachedStacksIds, attachedStack.StackId) {
			continue
		}
		attachmentsToDetach = append(attachmentsToDetach, attachedStack.Id)
	}
	return attachmentsToDetach
}
