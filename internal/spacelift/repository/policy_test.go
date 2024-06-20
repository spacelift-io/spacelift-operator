package repository

import (
	"context"
	"testing"

	"github.com/shurcooL/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	spaceliftclient "github.com/spacelift-io/spacelift-operator/internal/spacelift/client"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/client/mocks"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/structs"
	"github.com/spacelift-io/spacelift-operator/internal/utils"
)

func Test_policyRepository_Create(t *testing.T) {
	testCases := []struct {
		name         string
		policy       v1beta1.Policy
		expectedVars map[string]any
	}{
		{
			name: "basic policy",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Body:        "body",
					Type:        "PLAN",
					Description: utils.AddressOf("description"),
					Labels: []string{
						"label1",
						"label2",
					},
				},
			},
			expectedVars: map[string]any{
				"name":        graphql.String("name"),
				"body":        graphql.String("body"),
				"description": graphql.String("description"),
				"type":        PolicyType("PLAN"),
				"labels":      structs.GetGraphQLStrings(&[]string{"label1", "label2"}),
				"space":       (*graphql.ID)(nil),
			},
		},
		{
			name: "basic policy without description",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Body: "body",
					Type: "PLAN",
					Labels: []string{
						"label1",
						"label2",
					},
				},
			},
			expectedVars: map[string]any{
				"name":        graphql.String("name"),
				"body":        graphql.String("body"),
				"type":        PolicyType("PLAN"),
				"description": graphql.String(""),
				"labels":      structs.GetGraphQLStrings(&[]string{"label1", "label2"}),
				"space":       (*graphql.ID)(nil),
			},
		},
		{
			name: "basic policy with name",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Name:        utils.AddressOf("test name override"),
					Body:        "body",
					Description: utils.AddressOf("description"),
					Type:        "PLAN",
					Labels: []string{
						"label1",
						"label2",
					},
				},
			},
			expectedVars: map[string]any{
				"name":        graphql.String("test name override"),
				"body":        graphql.String("body"),
				"description": graphql.String("description"),
				"type":        PolicyType("PLAN"),
				"labels":      structs.GetGraphQLStrings(&[]string{"label1", "label2"}),
				"space":       (*graphql.ID)(nil),
			},
		},
		{
			name: "basic policy with space",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Body:        "body",
					Description: utils.AddressOf("description"),
					Type:        "PLAN",
					SpaceId:     utils.AddressOf("space-1"),
					Labels:      []string{},
				},
			},
			expectedVars: map[string]any{
				"name":        graphql.String("name"),
				"body":        graphql.String("body"),
				"description": graphql.String("description"),
				"type":        PolicyType("PLAN"),
				"labels":      structs.GetGraphQLStrings(&[]string{}),
				"space":       graphql.ID("space-1"),
			},
		},
		{
			name: "policy with space to attach",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Body:        "body",
					Description: utils.AddressOf("description"),
					Type:        "PLAN",
					SpaceId:     utils.AddressOf("space-1"),
					Labels:      []string{},
				},
			},
			expectedVars: map[string]any{
				"name":        graphql.String("name"),
				"body":        graphql.String("body"),
				"description": graphql.String("description"),
				"type":        PolicyType("PLAN"),
				"labels":      structs.GetGraphQLStrings(&[]string{}),
				"space":       graphql.ID("space-1"),
			},
		},
	}

	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewPolicyRepository(nil)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeClient = mocks.NewClient(t)
			var actualVars = map[string]any{}
			fakeClient.EXPECT().
				Mutate(mock.Anything, mock.AnythingOfType("*repository.policyCreateMutation"), mock.Anything).
				Run(func(_ context.Context, _ interface{}, vars map[string]interface{}, _a3 ...graphql.RequestOption) {
					actualVars = vars
				}).Return(nil)
			_, err := repo.Create(context.Background(), &testCase.policy)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedVars, actualVars)
		})
	}

}

func Test_policyRepository_Create_WithAttachedStacks(t *testing.T) {

	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewPolicyRepository(nil)

	fakeClient = mocks.NewClient(t)
	fakeClient.EXPECT().Mutate(mock.Anything, mock.AnythingOfType("*repository.policyCreateMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, _ map[string]any, _ ...graphql.RequestOption) {
			if mut, ok := mutation.(*policyCreateMutation); ok {
				*mut = policyCreateMutation{
					PolicyCreate: policyCreate{
						Id: "policy-id",
						AttachedStacks: []attachedStack{
							{StackId: "stack-id-1"},
						},
					},
				}
			}
		}).Return(nil)

	// stack-id-1 is already attached so we should only attach stack-id-2
	fakeClient.EXPECT().Mutate(
		mock.Anything,
		mock.AnythingOfType("*repository.policyAttachMutation"),
		map[string]any{
			"id":    "policy-id",
			"stack": "stack-id-2",
		},
	).Once().Return(nil)

	policy := v1beta1.Policy{
		ObjectMeta: v1.ObjectMeta{
			Name: "name",
		},
		Spec: v1beta1.PolicySpec{
			Body:              "body",
			Type:              "PLAN",
			AttachedStacksIds: []string{"stack-id-1", "stack-id-2"},
		},
	}

	_, err := repo.Create(context.Background(), &policy)
	require.NoError(t, err)
}

func Test_policyRepository_Update(t *testing.T) {
	testCases := []struct {
		name         string
		policy       v1beta1.Policy
		expectedVars map[string]any
	}{
		{
			name: "basic policy",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Body:        "body",
					Type:        "PLAN",
					Description: utils.AddressOf("description"),
					Labels: []string{
						"label1",
						"label2",
					},
				},
				Status: v1beta1.PolicyStatus{
					Id: "policy-id",
				},
			},
			expectedVars: map[string]any{
				"id":          graphql.ID("policy-id"),
				"name":        graphql.String("name"),
				"description": graphql.String("description"),
				"body":        graphql.String("body"),
				"labels":      structs.GetGraphQLStrings(&[]string{"label1", "label2"}),
				"space":       (*graphql.ID)(nil),
			},
		},
		{
			name: "basic policy without description",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Body: "body",
					Type: "PLAN",
					Labels: []string{
						"label1",
						"label2",
					},
				},
				Status: v1beta1.PolicyStatus{
					Id: "policy-id",
				},
			},
			expectedVars: map[string]any{
				"id":          graphql.ID("policy-id"),
				"name":        graphql.String("name"),
				"description": graphql.String(""),
				"body":        graphql.String("body"),
				"labels":      structs.GetGraphQLStrings(&[]string{"label1", "label2"}),
				"space":       (*graphql.ID)(nil),
			},
		},
		{
			name: "basic policy with name",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Name:        utils.AddressOf("test name override"),
					Body:        "body",
					Description: utils.AddressOf("description"),
					Type:        "PLAN",
					Labels: []string{
						"label1",
						"label2",
					},
				},
				Status: v1beta1.PolicyStatus{
					Id: "policy-id",
				},
			},
			expectedVars: map[string]any{
				"id":          graphql.ID("policy-id"),
				"name":        graphql.String("test name override"),
				"body":        graphql.String("body"),
				"description": graphql.String("description"),
				"labels":      structs.GetGraphQLStrings(&[]string{"label1", "label2"}),
				"space":       (*graphql.ID)(nil),
			},
		},
		{
			name: "basic policy with space",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Body:        "body",
					Description: utils.AddressOf("description"),
					Type:        "PLAN",
					SpaceId:     utils.AddressOf("space-1"),
					Labels:      []string{},
				},
				Status: v1beta1.PolicyStatus{
					Id: "policy-id",
				},
			},
			expectedVars: map[string]any{
				"id":          graphql.ID("policy-id"),
				"name":        graphql.String("name"),
				"body":        graphql.String("body"),
				"description": graphql.String("description"),
				"labels":      structs.GetGraphQLStrings(&[]string{}),
				"space":       graphql.ID("space-1"),
			},
		},
		{
			name: "policy with space to attach",
			policy: v1beta1.Policy{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.PolicySpec{
					Body:        "body",
					Description: utils.AddressOf("description"),
					Type:        "PLAN",
					SpaceId:     utils.AddressOf("space-1"),
					Labels:      []string{},
				},
				Status: v1beta1.PolicyStatus{
					Id: "policy-id",
				},
			},
			expectedVars: map[string]any{
				"id":          graphql.ID("policy-id"),
				"name":        graphql.String("name"),
				"body":        graphql.String("body"),
				"description": graphql.String("description"),
				"labels":      structs.GetGraphQLStrings(&[]string{}),
				"space":       graphql.ID("space-1"),
			},
		},
	}

	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewPolicyRepository(nil)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeClient = mocks.NewClient(t)
			var actualVars = map[string]any{}
			fakeClient.EXPECT().
				Mutate(mock.Anything, mock.AnythingOfType("*repository.policyUpdateMutation"), mock.Anything).
				Run(func(_ context.Context, _ interface{}, vars map[string]interface{}, _ ...graphql.RequestOption) {
					actualVars = vars
				}).Return(nil)
			_, err := repo.Update(context.Background(), &testCase.policy)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedVars, actualVars)
		})
	}
}

func Test_policyRepository_Update_WithAttachedStacks(t *testing.T) {

	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewPolicyRepository(nil)

	fakeClient = mocks.NewClient(t)
	fakeClient.EXPECT().Mutate(mock.Anything, mock.AnythingOfType("*repository.policyUpdateMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, _ map[string]any, _ ...graphql.RequestOption) {
			if mut, ok := mutation.(*policyUpdateMutation); ok {
				*mut = policyUpdateMutation{
					PolicyUpdate: policyUpdate{
						Id: "policy-id",
						AttachedStacks: []attachedStack{
							{
								Id:             "attachment-id-1",
								StackId:        "stack-id-1",
								IsAutoAttached: false,
							},
							{
								StackId:        "stack-id-2",
								IsAutoAttached: true,
							},
						},
					},
				}
			}
		}).Return(nil)

	// attachment-id-1 should be detached because stack-id-1 not specified in the spec
	// and it is not auto attached
	fakeClient.EXPECT().Mutate(
		mock.Anything,
		mock.AnythingOfType("*repository.policyDetachMutation"),
		map[string]any{
			"id": "attachment-id-1",
		},
	).Once().Return(nil)

	// stack-id-2 should not be detached because it is autoattached

	// stack-id-3 should be attached
	fakeClient.EXPECT().Mutate(
		mock.Anything,
		mock.AnythingOfType("*repository.policyAttachMutation"),
		map[string]any{
			"id":    "policy-id",
			"stack": "stack-id-3",
		},
	).Once().Return(nil)

	policy := v1beta1.Policy{
		ObjectMeta: v1.ObjectMeta{
			Name: "name",
		},
		Spec: v1beta1.PolicySpec{
			Body:              "body",
			Type:              "PLAN",
			AttachedStacksIds: []string{"stack-id-3"},
		},
	}

	_, err := repo.Update(context.Background(), &policy)
	require.NoError(t, err)
}

func Test_policyRepository_Update_DetachAllStacks(t *testing.T) {

	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewPolicyRepository(nil)

	fakeClient = mocks.NewClient(t)
	fakeClient.EXPECT().Mutate(mock.Anything, mock.AnythingOfType("*repository.policyUpdateMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, _ map[string]any, _ ...graphql.RequestOption) {
			if mut, ok := mutation.(*policyUpdateMutation); ok {
				*mut = policyUpdateMutation{
					PolicyUpdate: policyUpdate{
						Id: "policy-id",
						AttachedStacks: []attachedStack{
							{
								Id:             "attachment-id-1",
								StackId:        "stack-id-1",
								IsAutoAttached: false,
							},
							{
								StackId:        "stack-id-2",
								IsAutoAttached: true,
							},
						},
					},
				}
			}
		}).Return(nil)

	// attachment-id-1 should be detached because stack-id-1 not specified in the spec
	// and it is not auto attached
	fakeClient.EXPECT().Mutate(
		mock.Anything,
		mock.AnythingOfType("*repository.policyDetachMutation"),
		map[string]any{
			"id": "attachment-id-1",
		},
	).Once().Return(nil)

	// stack-id-2 should not be detached because it is autoattached

	policy := v1beta1.Policy{
		ObjectMeta: v1.ObjectMeta{
			Name: "name",
		},
		Spec: v1beta1.PolicySpec{
			Body:              "body",
			Type:              "PLAN",
			AttachedStacksIds: []string{},
		},
	}

	_, err := repo.Update(context.Background(), &policy)
	require.NoError(t, err)
}
