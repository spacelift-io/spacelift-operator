package controller_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest/observer"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/controller"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	spaceliftRepository "github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/mocks"
	"github.com/spacelift-io/spacelift-operator/internal/utils"
	"github.com/spacelift-io/spacelift-operator/tests/integration"
)

type ContextControllerTestSuite struct {
	integration.IntegrationTestSuite
	integration.WithSpaceSuiteHelper
	integration.WithStackSuiteHelper
	integration.WithContextSuiteHelper
}

func (s *ContextControllerTestSuite) SetupSuite() {
	s.SetupManager = func(mgr manager.Manager) {
		s.ContextRepo = repository.NewContextRepository(mgr.GetClient(), mgr.GetScheme())
		s.StackRepo = repository.NewStackRepository(mgr.GetClient(), mgr.GetScheme())
		s.SpaceRepo = repository.NewSpaceRepository(mgr.GetClient())
		s.SecretRepo = repository.NewSecretRepository(mgr.GetClient())
		s.FakeSpaceliftContextRepo = new(mocks.ContextRepository)
		err := (&controller.ContextReconciler{
			SpaceliftContextRepository: s.FakeSpaceliftContextRepo,
			ContextRepository:          s.ContextRepo,
			StackRepository:            s.StackRepo,
			SpaceRepository:            s.SpaceRepo,
			SecretRepository:           s.SecretRepo,
		}).SetupWithManager(mgr)
		s.Require().NoError(err)
	}
	s.WithContextSuiteHelper = integration.WithContextSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.WithSpaceSuiteHelper = integration.WithSpaceSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.WithStackSuiteHelper = integration.WithStackSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.IntegrationTestSuite.SetupSuite()
}

func (s *ContextControllerTestSuite) SetupTest() {
	s.FakeSpaceliftContextRepo.Test(s.T())
}

func (s *ContextControllerTestSuite) TearDownTest() {
	s.FakeSpaceliftContextRepo.AssertExpectations(s.T())
	s.FakeSpaceliftContextRepo.Calls = nil
	s.FakeSpaceliftContextRepo.ExpectedCalls = nil
}

func (s *ContextControllerTestSuite) TestContextCreation_InvalidSpec() {
	cases := []struct {
		Name        string
		Spec        v1beta1.ContextSpec
		ExpectedErr string
	}{
		{
			Spec:        v1beta1.ContextSpec{},
			Name:        "empty spec, missing space or spaceId",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec: Invalid value: "object": only one of space or spaceId should be set`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space:   utils.AddressOf("foobar"),
				SpaceId: utils.AddressOf("foobar"),
			},
			Name:        "both space and spaceId are set",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec: Invalid value: "object": only one of space or spaceId should be set`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf(""),
			},
			Name:        "space empty string",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.space: Invalid value: "": spec.space in body should be at least 1 chars long`,
		},
		{
			Spec: v1beta1.ContextSpec{
				SpaceId: utils.AddressOf(""),
			},
			Name:        "spaceId empty string",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.spaceId: Invalid value: "": spec.spaceId in body should be at least 1 chars long`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space:       utils.AddressOf("foobar"),
				Attachments: []v1beta1.Attachment{{}},
			},
			Name:        "empty attachment",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.attachments[0]: Invalid value: "object": only one of stack or stackId or moduleId should be set`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				Attachments: []v1beta1.Attachment{{
					Stack:   utils.AddressOf("foobar"),
					StackId: utils.AddressOf("foobar"),
				}},
			},
			Name:        "attachment with both stack and stackId",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.attachments[0]: Invalid value: "object": only one of stack or stackId or moduleId should be set`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				Attachments: []v1beta1.Attachment{{
					ModuleId: utils.AddressOf("foobar"),
					StackId:  utils.AddressOf("foobar"),
				}},
			},
			Name:        "attachment with both stackId and moduleId",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.attachments[0]: Invalid value: "object": only one of stack or stackId or moduleId should be set`,
		},

		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				Attachments: []v1beta1.Attachment{{
					Stack:    utils.AddressOf("foobar"),
					ModuleId: utils.AddressOf("foobar"),
				}},
			},
			Name:        "attachment with both stack and moduleId",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.attachments[0]: Invalid value: "object": only one of stack or stackId or moduleId should be set`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				Attachments: []v1beta1.Attachment{{
					Stack: utils.AddressOf(""),
				}},
			},
			Name:        "attachment stack empty",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.attachments[0].stack: Invalid value: "": spec.attachments[0].stack in body should be at least 1 chars long`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				Attachments: []v1beta1.Attachment{{
					StackId: utils.AddressOf(""),
				}},
			},
			Name:        "attachment stackId empty",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.attachments[0].stackId: Invalid value: "": spec.attachments[0].stackId in body should be at least 1 chars long`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				Attachments: []v1beta1.Attachment{{
					ModuleId: utils.AddressOf(""),
				}},
			},
			Name:        "attachment moduleId empty",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.attachments[0].moduleId: Invalid value: "": spec.attachments[0].moduleId in body should be at least 1 chars long`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				Environment: []v1beta1.Environment{
					{
						Id:    "",
						Value: utils.AddressOf("test"),
					},
				},
			},
			Name:        "empty environment id",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.environment[0].id: Invalid value: "": spec.environment[0].id in body should be at least 1 chars long`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				Environment: []v1beta1.Environment{
					{
						Id:              "test",
						Value:           utils.AddressOf("test"),
						ValueFromSecret: &v1.SecretKeySelector{},
					},
				},
			},
			Name:        "environment with both value and valueFromSecret set",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.environment[0]: Invalid value: "object": only one of value or valueFromSecret should be set`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				MountedFiles: []v1beta1.MountedFile{
					{
						Id:    "",
						Value: utils.AddressOf("test"),
					},
				},
			},
			Name:        "empty mounted file",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.mountedFiles[0].id: Invalid value: "": spec.mountedFiles[0].id in body should be at least 1 chars long`,
		},
		{
			Spec: v1beta1.ContextSpec{
				Space: utils.AddressOf("foobar"),
				MountedFiles: []v1beta1.MountedFile{
					{
						Id:              "test",
						Value:           utils.AddressOf("test"),
						ValueFromSecret: &v1.SecretKeySelector{},
					},
				},
			},
			Name:        "mounted file with both value and valueFromSecret set",
			ExpectedErr: `Context.app.spacelift.io "invalid-context" is invalid: spec.mountedFiles[0]: Invalid value: "object": only one of value or valueFromSecret should be set`,
		},
	}
	for _, c := range cases {
		s.Run(c.Name, func() {
			context := &v1beta1.Context{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Context",
					APIVersion: v1beta1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-context",
					Namespace: "default",
				},
				Spec: c.Spec,
			}
			err := s.Client().Create(s.Context(), context)
			s.Assert().EqualError(err, c.ExpectedErr)
		})
	}
}

func (s *ContextControllerTestSuite) TestContextCreation_UnableToCreateOnSpacelift() {
	s.FakeSpaceliftContextRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrContextNotFound)
	s.FakeSpaceliftContextRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(nil, fmt.Errorf("unable to create resource on spacelift"))

	s.Logs.TakeAll()
	context, err := s.CreateTestContext()
	s.Require().NoError(err)
	defer s.DeleteContext(context)

	// Make sure we don't update the context ID
	s.Require().Never(func() bool {
		context, err := s.ContextRepo.Get(s.Context(), types.NamespacedName{
			Namespace: context.Namespace,
			Name:      context.ObjectMeta.Name,
		})
		s.Require().NoError(err)
		return context.Status.Id != ""
	}, 3*time.Second, integration.DefaultInterval)

	// Check that the error has been logged
	logs := s.Logs.FilterMessage("Unable to create the context in spacelift")
	s.Require().Equal(1, logs.Len())
	logs = s.Logs.FilterMessage("Context created")
	s.Require().Equal(0, logs.Len())
}

func (s *ContextControllerTestSuite) TestContextCreation_OK_SpaceNotReady() {
	c := &v1beta1.Context{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Context",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-context",
			Namespace: "default",
		},
		Spec: v1beta1.ContextSpec{
			Space: utils.AddressOf("test-space"),
		},
	}
	s.Logs.TakeAll()
	err := s.CreateContext(c)
	s.Require().NoError(err)
	defer s.DeleteContext(c)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Unable to find space for context, will retry in 10 seconds")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-space", logs.All()[0].ContextMap()[logging.SpaceName])
	s.Assert().EqualValues(logging.Level4, -logs.All()[0].Level)

	s.Logs.TakeAll()
	space, err := s.CreateTestSpace()
	s.Require().NoError(err)
	defer s.DeleteSpace(space)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Space is not ready, will retry in 3 seconds")
		return logs.Len() == 1
	}, 12*time.Second, integration.DefaultInterval)
	s.Assert().Equal("test-space", logs.All()[0].ContextMap()[logging.SpaceName])

	s.FakeSpaceliftContextRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrContextNotFound)
	var contextSpecToCreate v1beta1.ContextSpec
	s.FakeSpaceliftContextRepo.EXPECT().Create(mock.Anything, mock.Anything).
		Run(func(_a0 context.Context, c *v1beta1.Context) {
			contextSpecToCreate = c.Spec
		}).Once().
		Return(&models.Context{Id: "test-context-id"}, nil)

	space.Status.Id = "test-space"
	err = s.SpaceRepo.UpdateStatus(s.Context(), space)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Context created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-context-id", logs.All()[0].ContextMap()[logging.ContextId])

	s.Require().NotNil(contextSpecToCreate.SpaceId)
	s.Assert().Equal("test-space", *contextSpecToCreate.SpaceId)

	c, err = s.ContextRepo.Get(s.Context(), types.NamespacedName{
		Namespace: c.Namespace,
		Name:      c.ObjectMeta.Name,
	})
	s.Require().NoError(err)
	s.Assert().Len(c.OwnerReferences, 1)
	s.Assert().Equal(space.ObjectMeta.Name, c.OwnerReferences[0].Name)
	s.Assert().Equal("Space", c.OwnerReferences[0].Kind)
}

func (s *ContextControllerTestSuite) TestContextCreation_OK_AttachedStackNotReady() {
	c := &v1beta1.Context{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Context",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-context",
			Namespace: "default",
		},
		Spec: v1beta1.ContextSpec{
			SpaceId: utils.AddressOf("test-space"),
			Attachments: []v1beta1.Attachment{
				{
					Stack: utils.AddressOf("test-stack"),
				},
			},
		},
	}

	s.Logs.TakeAll()
	err := s.CreateContext(c)
	s.Require().NoError(err)
	defer s.DeleteContext(c)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Unable to find stack for context, will retry in 10 seconds")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-stack", logs.All()[0].ContextMap()[logging.StackName])
	s.Assert().EqualValues(logging.Level4, -logs.All()[0].Level)

	s.Logs.TakeAll()
	stack, err := s.CreateTestStack()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Stack is not ready, will retry in 3 seconds")
		return logs.Len() == 1
	}, 10*time.Second, integration.DefaultInterval)
	s.Assert().Equal("test-stack", logs.All()[0].ContextMap()[logging.StackName])

	s.FakeSpaceliftContextRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrContextNotFound)
	var contextSpecToCreate v1beta1.ContextSpec
	s.FakeSpaceliftContextRepo.EXPECT().Create(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *v1beta1.Context) {
			contextSpecToCreate = c.Spec
		}).Once().
		Return(&models.Context{Id: "test-context-id"}, nil)

	stack.Status.Id = "test-stack"
	err = s.StackRepo.UpdateStatus(s.Context(), stack)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Context created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-context-id", logs.All()[0].ContextMap()[logging.ContextId])

	s.Require().Len(contextSpecToCreate.Attachments, 1)
	s.Require().NotNil(contextSpecToCreate.Attachments[0].StackId)
	s.Assert().Equal("test-stack", *contextSpecToCreate.Attachments[0].StackId)
	s.Assert().Nil(contextSpecToCreate.Attachments[0].Priority)
}

func (s *ContextControllerTestSuite) TestContextCreation_OK_SecretNotReady() {

	c := &v1beta1.Context{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Context",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-context",
			Namespace: "default",
		},
		Spec: v1beta1.ContextSpec{
			SpaceId: utils.AddressOf("test-space"),
			Environment: []v1beta1.Environment{
				{
					Id: "test_secret_id",
					ValueFromSecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "test_secret",
					},
				},
			},
			MountedFiles: []v1beta1.MountedFile{
				{
					Id: "mounted_file_id",
					ValueFromSecret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "test_mounted_file",
					},
				},
			},
		},
	}

	s.Logs.TakeAll()
	err := s.CreateContext(c)
	s.Require().NoError(err)
	defer s.DeleteContext(c)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Unable to find secret for context environment variable, will retry in 3 seconds.")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-secret", logs.All()[0].ContextMap()[logging.SecretName])

	s.FakeSpaceliftContextRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrContextNotFound)

	var contextSpecToCreate v1beta1.ContextSpec
	s.FakeSpaceliftContextRepo.EXPECT().Create(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *v1beta1.Context) {
			contextSpecToCreate = c.Spec
		}).Once().
		Return(&models.Context{Id: "test-context-id"}, nil)

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: c.Namespace,
		},
		Data: map[string][]byte{
			"test_secret":       []byte("secret_value"),
			"test_mounted_file": []byte("secret_value_for_file"),
		},
		Type: v1.SecretTypeOpaque,
	}
	err = s.Client().Create(s.Context(), secret)
	s.Require().NoError(err)
	defer s.Client().Delete(s.Context(), secret)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Context created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)

	s.Require().Len(contextSpecToCreate.Environment, 1)
	s.Assert().Equal("test_secret_id", contextSpecToCreate.Environment[0].Id)
	s.Require().NotNil(contextSpecToCreate.Environment[0].Value)
	s.Assert().Equal("secret_value", *contextSpecToCreate.Environment[0].Value)
	s.Require().NotNil(contextSpecToCreate.Environment[0].Secret)
	s.Assert().True(*contextSpecToCreate.Environment[0].Secret)

	s.Require().Len(contextSpecToCreate.MountedFiles, 1)
	s.Assert().Equal("mounted_file_id", contextSpecToCreate.MountedFiles[0].Id)
	s.Require().NotNil(contextSpecToCreate.MountedFiles[0].Value)
	s.Assert().Equal("secret_value_for_file", *contextSpecToCreate.MountedFiles[0].Value)
	s.Require().NotNil(contextSpecToCreate.MountedFiles[0].Secret)
	s.Assert().True(*contextSpecToCreate.MountedFiles[0].Secret)
}

func (s *ContextControllerTestSuite) TestContextCreation_OK() {

	s.FakeSpaceliftContextRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrContextNotFound)
	s.FakeSpaceliftContextRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(&models.Context{Id: "test-context-id"}, nil)

	context := &v1beta1.Context{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Context",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-context",
			Namespace: "default",
		},
		Spec: v1beta1.ContextSpec{
			Name:    utils.AddressOf("test context new name"),
			SpaceId: utils.AddressOf("test-space"),
		},
	}

	s.Logs.TakeAll()
	err := s.CreateContext(context)
	s.Require().NoError(err)
	defer s.DeleteContext(context)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Context created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-context-id", logs.All()[0].ContextMap()[logging.ContextId])

	context, err = s.ContextRepo.Get(s.Context(), types.NamespacedName{
		Namespace: context.Namespace,
		Name:      context.ObjectMeta.Name,
	})
	s.Require().NoError(err)
	s.Assert().Equal("test-context-id", context.Status.Id)
}

func (s *ContextControllerTestSuite) TestContextUpdate_OK() {

	s.FakeSpaceliftContextRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, nil)
	s.FakeSpaceliftContextRepo.EXPECT().Update(mock.Anything, mock.Anything).Once().
		Return(&models.Context{Id: "test-context-id"}, nil)

	s.Logs.TakeAll()
	context, err := s.CreateTestContext()
	s.Require().NoError(err)
	defer s.DeleteContext(context)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Context updated")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-context-id", logs.All()[0].ContextMap()[logging.ContextId])

	context, err = s.ContextRepo.Get(s.Context(), types.NamespacedName{
		Namespace: context.Namespace,
		Name:      context.ObjectMeta.Name,
	})
	s.Require().NoError(err)
	s.Assert().Equal("test-context-id", context.Status.Id)
}

func TestContextController(t *testing.T) {
	suite.Run(t, new(ContextControllerTestSuite))
}
