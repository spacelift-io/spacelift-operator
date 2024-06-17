package repository

import (
	"context"
	"testing"

	"github.com/shurcooL/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	spaceliftclient "github.com/spacelift-io/spacelift-operator/internal/spacelift/client"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/client/mocks"
)

func Test_runRepository_Create(t *testing.T) {
	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}

	var actualVars map[string]any
	fakeClient = mocks.NewClient(t)
	fakeClient.EXPECT().
		Mutate(mock.Anything, mock.AnythingOfType("*repository.createRunMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, vars map[string]interface{}, _ ...graphql.RequestOption) {
			actualVars = vars
			runMutation := mutation.(*createRunMutation)
			runMutation.RunTrigger.ID = "run-id"
			runMutation.RunTrigger.State = "QUEUED"
		}).Return(nil)
	fakeClient.EXPECT().URL("/stack/%s/run/%s", "stack-id", "run-id").Return("run-url")

	fakeStack := &v1beta1.Stack{
		Status: v1beta1.StackStatus{
			Id: "stack-id",
		},
	}
	repo := NewRunRepository(nil)
	run, err := repo.Create(context.Background(), fakeStack)
	assert.NoError(t, err)

	assert.Equal(t, "stack-id", actualVars["stack"])
	assert.Equal(t, "run-id", run.Id)
	assert.Equal(t, "QUEUED", run.State)
	assert.Equal(t, "run-url", run.Url)
	assert.Equal(t, "stack-id", run.StackId)
}
