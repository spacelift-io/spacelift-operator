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

func Test_contextRepository_Create(t *testing.T) {
	testCases := []struct {
		name          string
		context       v1beta1.Context
		assertPayload func(*testing.T, structs.ContextInput)
	}{
		{
			name: "basic context",
			context: v1beta1.Context{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.ContextSpec{
					SpaceId:     utils.AddressOf("test-space-id"),
					Description: utils.AddressOf("test description"),
					Labels:      []string{"label1", "label2"},
					Attachments: []v1beta1.Attachment{
						{
							StackId: utils.AddressOf("test-attached-stack-id"),
						},
					},
					Hooks: v1beta1.Hooks{
						AfterApply: []string{"after", "apply"},
					},
					Environment: []v1beta1.Environment{
						{
							Id:          "id",
							Value:       utils.AddressOf("secret"),
							Secret:      utils.AddressOf(true),
							Description: utils.AddressOf("test description"),
						},
					},
					MountedFiles: []v1beta1.MountedFile{
						{
							Id:          "id-file",
							Value:       utils.AddressOf("secret file"),
							Secret:      utils.AddressOf(true),
							Description: utils.AddressOf("test description"),
						},
					},
				},
			},
			assertPayload: func(t *testing.T, input structs.ContextInput) {
				assert.Equal(t, "name", input.Name)
				assert.Equal(t, "test description", *input.Description)
				assert.Equal(t, []string{"label1", "label2"}, input.Labels)
				assert.Equal(t, []string{"after", "apply"}, input.Hooks.AfterApply)
				assert.Equal(t, []structs.StackAttachment{
					{
						Stack: "test-attached-stack-id",
					},
				}, input.StackAttachments)
				assert.Equal(t, []structs.ConfigAttachments{
					{
						Description: utils.AddressOf("test description"),
						Id:          "id",
						Type:        "ENVIRONMENT_VARIABLE",
						Value:       "secret",
						WriteOnly:   true,
					},
					{
						Description: utils.AddressOf("test description"),
						Id:          "id-file",
						Type:        "FILE_MOUNT",
						Value:       "secret file",
						WriteOnly:   true,
					},
				}, input.ConfigAttachments)
			},
		},
		{
			name: "basic context with name",
			context: v1beta1.Context{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.ContextSpec{
					SpaceId: utils.AddressOf("test-space-id"),
					Name:    utils.AddressOf("test name override"),
				},
			},
			assertPayload: func(t *testing.T, input structs.ContextInput) {
				assert.Equal(t, "test name override", input.Name)
			},
		},
	}

	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewContextRepository(nil)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeClient = mocks.NewClient(t)
			var actualVars = map[string]any{}
			fakeClient.EXPECT().
				Mutate(mock.Anything, mock.AnythingOfType("*repository.contextCreateMutation"), mock.Anything).
				Run(func(_ context.Context, _ interface{}, vars map[string]interface{}, _a3 ...graphql.RequestOption) {
					actualVars = vars
				}).Return(nil)
			_, err := repo.Create(context.Background(), &testCase.context)
			require.NoError(t, err)
			testCase.assertPayload(t, actualVars["input"].(structs.ContextInput))
		})
	}

}

func Test_contextRepository_Update(t *testing.T) {
	testCases := []struct {
		name          string
		context       v1beta1.Context
		assertPayload func(*testing.T, map[string]any)
	}{
		{
			name: "basic context",
			context: v1beta1.Context{
				ObjectMeta: v1.ObjectMeta{
					Name: "name",
				},
				Spec: v1beta1.ContextSpec{
					SpaceId: utils.AddressOf("test-space-id"),
				},
				Status: v1beta1.ContextStatus{
					Id: "test-context-id",
				},
			},
			assertPayload: func(t *testing.T, input map[string]any) {
				assert.Equal(t, "test-context-id", input["id"])
				assert.Equal(t, graphql.Boolean(true), input["replaceConfigElements"])
				// No need to assert on input details since we use the same code than the Create function
				// and this is already covered in Test_contextRepository_Create
				assert.IsType(t, structs.ContextInput{}, input["input"])
			},
		},
	}

	originalClient := spaceliftclient.DefaultClient
	defer func() { spaceliftclient.DefaultClient = originalClient }()
	var fakeClient *mocks.Client
	spaceliftclient.DefaultClient = func(_ context.Context, _ client.Client, _ string) (spaceliftclient.Client, error) {
		return fakeClient, nil
	}
	repo := NewContextRepository(nil)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeClient = mocks.NewClient(t)
			var actualVars = map[string]any{}
			fakeClient.EXPECT().
				Mutate(mock.Anything, mock.AnythingOfType("*repository.contextUpdateMutation"), mock.Anything).
				Run(func(_ context.Context, _ interface{}, vars map[string]interface{}, _a3 ...graphql.RequestOption) {
					actualVars = vars
				}).Return(nil)
			_, err := repo.Update(context.Background(), &testCase.context)
			require.NoError(t, err)
			testCase.assertPayload(t, actualVars)
		})
	}

}
