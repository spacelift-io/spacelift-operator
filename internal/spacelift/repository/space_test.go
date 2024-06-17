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

func Test_spaceRepository_Create(t *testing.T) {
	testCases := []struct {
		name          string
		space         v1beta1.Space
		assertPayload func(*testing.T, structs.SpaceInput)
	}{
		{
			name: "basic space",
			space: v1beta1.Space{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.SpaceSpec{
					ParentSpace:     "parent-space-id",
					Description:     "test description",
					InheritEntities: true,
					Labels:          &[]string{"label1", "label2"},
				},
			},
			assertPayload: func(t *testing.T, input structs.SpaceInput) {
				assert.Equal(t, graphql.String("name"), input.Name)
				assert.Equal(t, graphql.String("test description"), input.Description)
				assert.Equal(t, graphql.String("parent-space-id"), input.ParentSpace)
				assert.Equal(t, graphql.Boolean(true), input.InheritEntities)
				assert.Equal(t, &[]graphql.String{"label1", "label2"}, input.Labels)
			},
		},
		{
			name: "basic space with name",
			space: v1beta1.Space{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.SpaceSpec{
					ParentSpace: "test-space-id",
					Name:        utils.AddressOf("test name override"),
				},
			},
			assertPayload: func(t *testing.T, input structs.SpaceInput) {
				assert.Equal(t, graphql.String("test name override"), input.Name)
			},
		},
	}

	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewSpaceRepository(nil)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeClient = mocks.NewClient(t)
			var actualVars = map[string]any{}
			fakeClient.EXPECT().
				Mutate(mock.Anything, mock.AnythingOfType("*repository.spaceCreateMutation"), mock.Anything).
				Run(func(_ context.Context, mutation any, vars map[string]interface{}, _ ...graphql.RequestOption) {
					actualVars = vars
					spaceMutation := mutation.(*spaceCreateMutation)
					spaceMutation.SpaceCreate.ID = "space-id"
				}).Return(nil)
			fakeClient.EXPECT().URL("/spaces/%s", "space-id").Return("")
			_, err := repo.Create(context.Background(), &testCase.space)
			require.NoError(t, err)
			testCase.assertPayload(t, actualVars["input"].(structs.SpaceInput))
		})
	}

}

func Test_spaceRepository_Update(t *testing.T) {
	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	fakeClient := mocks.NewClient(t)
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}

	fakeSpaceId := "space-id"
	var actualVars map[string]any
	fakeClient.EXPECT().
		Mutate(mock.Anything, mock.AnythingOfType("*repository.spaceUpdateMutation"), mock.Anything).
		Run(func(_ context.Context, mutation any, vars map[string]interface{}, _ ...graphql.RequestOption) {
			actualVars = vars
			spaceMutation := mutation.(*spaceUpdateMutation)
			spaceMutation.SpaceUpdate.ID = fakeSpaceId
		}).Return(nil)
	fakeClient.EXPECT().URL("/spaces/%s", fakeSpaceId).Return("")

	repo := NewSpaceRepository(nil)

	fakeSpace := &v1beta1.Space{
		ObjectMeta: v1.ObjectMeta{
			Name: "space-name",
		},
		Spec: v1beta1.SpaceSpec{
			ParentSpace:     "parent-space-id",
			Description:     "test description",
			InheritEntities: true,
			Labels:          &[]string{"label1", "label2"},
		},
		Status: v1beta1.SpaceStatus{
			Id: fakeSpaceId,
		},
	}
	_, err := repo.Update(context.Background(), fakeSpace)
	require.NoError(t, err)
	assert.Equal(t, fakeSpaceId, actualVars["space"])
	assert.Equal(t, structs.SpaceInput{
		Name:            "space-name",
		Description:     "test description",
		InheritEntities: true,
		ParentSpace:     "parent-space-id",
		Labels:          &[]graphql.String{"label1", "label2"},
	}, actualVars["input"])
}
