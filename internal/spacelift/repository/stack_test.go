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

func Test_stackRepository_Create(t *testing.T) {
	testCases := []struct {
		name          string
		stack         v1beta1.Stack
		assertPayload func(*testing.T, map[string]any)
	}{
		{
			name: "stack with disabled state management",
			stack: v1beta1.Stack{
				ObjectMeta: v1.ObjectMeta{
					Name: "stack-name",
				},
				Spec: v1beta1.StackSpec{
					ManagesStateFile: utils.AddressOf(false),
				},
			},
			assertPayload: func(t *testing.T, vars map[string]any) {
				assert.Equal(t, graphql.Boolean(false), vars["manageState"])
				input := vars["input"].(structs.StackInput)
				assert.Equal(t, graphql.String("stack-name"), input.Name)
			},
		},
	}
	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewStackRepository(nil)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeClient = mocks.NewClient(t)
			var actualVars = map[string]any{}
			fakeClient.EXPECT().
				Mutate(mock.Anything, mock.AnythingOfType("*repository.stackCreateMutation"), mock.Anything).
				Run(func(_ context.Context, mutation any, vars map[string]interface{}, _ ...graphql.RequestOption) {
					actualVars = vars
					createMutation := mutation.(*stackCreateMutation)
					createMutation.StackCreate.ID = "stack-id"
				}).Return(nil)
			fakeClient.EXPECT().URL("/stack/%s", "stack-id").Return("")
			_, err := repo.Create(context.Background(), &testCase.stack)
			require.NoError(t, err)
			testCase.assertPayload(t, actualVars)
		})
	}
}

func Test_stackRepository_Create_WithCommitSHA(t *testing.T) {
	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewStackRepository(nil)

	fakeClient = mocks.NewClient(t)
	fakeClient.EXPECT().
		Mutate(mock.Anything, mock.AnythingOfType("*repository.stackCreateMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, _ map[string]interface{}, _ ...graphql.RequestOption) {
			createMutation := mutation.(*stackCreateMutation)
			createMutation.StackCreate.ID = "stack-id"
		}).Return(nil)
	fakeClient.EXPECT().URL("/stack/%s", "stack-id").Return("")
	var setTrackedCommitVars = map[string]any{}
	fakeClient.EXPECT().
		Mutate(mock.Anything, mock.AnythingOfType("*repository.setTrackedCommitMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, vars map[string]interface{}, _ ...graphql.RequestOption) {
			setTrackedCommitVars = vars
		}).Return(nil)

	stack := v1beta1.Stack{
		ObjectMeta: v1.ObjectMeta{
			Name: "stack-name",
		},
		Spec: v1beta1.StackSpec{
			CommitSHA: utils.AddressOf("commit-sha"),
			SpaceId:   utils.AddressOf("space-id"),
		},
	}

	_, err := repo.Create(context.Background(), &stack)
	require.NoError(t, err)
	assert.EqualValues(t, "stack-id", setTrackedCommitVars["id"])
	assert.EqualValues(t, "commit-sha", setTrackedCommitVars["sha"])
}

func Test_stackRepository_Create_WithAWSIntegration(t *testing.T) {
	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewStackRepository(nil)

	fakeClient = mocks.NewClient(t)
	fakeClient.EXPECT().
		Mutate(mock.Anything, mock.AnythingOfType("*repository.stackCreateMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, _ map[string]interface{}, _ ...graphql.RequestOption) {
			createMutation := mutation.(*stackCreateMutation)
			createMutation.StackCreate.ID = "stack-id"
		}).Return(nil)
	fakeClient.EXPECT().URL("/stack/%s", "stack-id").Return("")
	var attachIntegrationVars = map[string]any{}
	fakeClient.EXPECT().
		Mutate(mock.Anything, mock.AnythingOfType("*repository.awsIntegrationAttachMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, vars map[string]interface{}, _ ...graphql.RequestOption) {
			attachIntegrationVars = vars
		}).Return(nil)

	stack := v1beta1.Stack{
		ObjectMeta: v1.ObjectMeta{
			Name: "stack-name",
		},
		Spec: v1beta1.StackSpec{
			SpaceId: utils.AddressOf("space-id"),
			AWSIntegration: &v1beta1.AWSIntegration{
				Id:    "integration-id",
				Read:  true,
				Write: true,
			},
		},
	}

	_, err := repo.Create(context.Background(), &stack)
	require.NoError(t, err)
	assert.EqualValues(t, "integration-id", attachIntegrationVars["id"])
	assert.EqualValues(t, "stack-id", attachIntegrationVars["stack"])
	assert.EqualValues(t, true, attachIntegrationVars["read"])
	assert.EqualValues(t, true, attachIntegrationVars["write"])
}

func Test_stackRepository_Update(t *testing.T) {
	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	fakeClient := mocks.NewClient(t)
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}

	fakeStackId := "stack-id"
	var actualVars map[string]any
	fakeClient.EXPECT().
		Mutate(mock.Anything, mock.AnythingOfType("*repository.stackUpdateMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, vars map[string]interface{}, _ ...graphql.RequestOption) {
			actualVars = vars
			updateMutation := mutation.(*stackUpdateMutation)
			updateMutation.StackUpdate.ID = fakeStackId
		}).Return(nil)
	fakeClient.EXPECT().URL("/stack/%s", fakeStackId).Return("")

	repo := NewStackRepository(nil)

	fakeStack := &v1beta1.Stack{
		ObjectMeta: v1.ObjectMeta{
			Name: "stack-name",
		},
		Spec: v1beta1.StackSpec{
			SpaceId: utils.AddressOf("space-id"),
		},
		Status: v1beta1.StackStatus{
			Id: fakeStackId,
		},
	}
	_, err := repo.Update(context.Background(), fakeStack)
	require.NoError(t, err)
	assert.Equal(t, "stack-name", actualVars["id"])
	assert.IsType(t, structs.StackInput{}, actualVars["input"])
}
