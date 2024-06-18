package controller_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest/observer"
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

type StackControllerSuite struct {
	integration.IntegrationTestSuite
	integration.WithStackSuiteHelper
	integration.WithSpaceSuiteHelper
}

func (s *StackControllerSuite) SetupSuite() {
	s.SetupManager = func(mgr manager.Manager) {
		s.FakeSpaceliftStackRepo = new(mocks.StackRepository)
		s.StackRepo = repository.NewStackRepository(mgr.GetClient(), mgr.GetScheme())
		s.SpaceRepo = repository.NewSpaceRepository(mgr.GetClient())
		err := (&controller.StackReconciler{
			StackRepository:          s.StackRepo,
			SpaceRepository:          s.SpaceRepo,
			SpaceliftStackRepository: s.FakeSpaceliftStackRepo,
		}).SetupWithManager(mgr)
		s.Require().NoError(err)
	}
	s.IntegrationTestSuite.SetupSuite()
	s.WithStackSuiteHelper = integration.WithStackSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.WithSpaceSuiteHelper = integration.WithSpaceSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
}

func (s *StackControllerSuite) SetupTest() {
	s.FakeSpaceliftStackRepo.Test(s.T())
	s.IntegrationTestSuite.SetupTest()
}

func (s *StackControllerSuite) TearDownTest() {
	s.FakeSpaceliftStackRepo.AssertExpectations(s.T())
	s.FakeSpaceliftStackRepo.Calls = nil
	s.FakeSpaceliftStackRepo.ExpectedCalls = nil
}

func (s *StackControllerSuite) TestStackCreation_InvalidSpec() {
	cases := []struct {
		Name        string
		Spec        v1beta1.StackSpec
		ExpectedErr string
	}{
		{
			Spec:        v1beta1.StackSpec{},
			Name:        "missing spaceName/spaceId",
			ExpectedErr: `Stack.app.spacelift.io "invalid-stack" is invalid: spec: Invalid value: "object": only one of spaceName or spaceId can be set`,
		},
	}

	for _, c := range cases {
		s.Run(c.Name, func() {
			stack := &v1beta1.Stack{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Stack",
					APIVersion: v1beta1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-stack",
					Namespace: "default",
				},
				Spec: c.Spec,
			}
			err := s.Client().Create(s.Context(), stack)
			s.Assert().EqualError(err, c.ExpectedErr)
		})
	}
}

func (s *StackControllerSuite) TestStackCreation_UnableToCreateOnSpacelift() {
	s.FakeSpaceliftStackRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrStackNotFound)
	s.FakeSpaceliftStackRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(nil, fmt.Errorf("unable to create resource on spacelift"))

	s.Logs.TakeAll()
	stack, err := s.CreateTestStack()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	// Make sure we don't update the stack ID
	s.Require().Never(func() bool {
		stack, err := s.StackRepo.Get(s.Context(), types.NamespacedName{
			Namespace: stack.Namespace,
			Name:      stack.ObjectMeta.Name,
		})
		s.Require().NoError(err)
		return stack.Status.Id != ""
	}, 3*time.Second, integration.DefaultInterval)

	// Check that the error has been logged
	logs := s.Logs.FilterMessage("Unable to create the stack in spacelift")
	s.Require().Equal(1, logs.Len())
	logs = s.Logs.FilterMessage("Stack created")
	s.Require().Equal(0, logs.Len())
}

func (s *StackControllerSuite) TestStackCreation_OK() {
	s.FakeSpaceliftStackRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrStackNotFound)
	s.FakeSpaceliftStackRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(&models.Stack{
			Id: "test-stack-generated-id",
		}, nil)

	s.Logs.TakeAll()
	stack, err := s.CreateTestStack()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	// Make sure stack status is updated
	s.Require().Eventually(func() bool {
		stack, err := s.StackRepo.Get(s.Context(), types.NamespacedName{
			Namespace: stack.Namespace,
			Name:      stack.ObjectMeta.Name,
		})
		s.Require().NoError(err)
		return stack.Status.Id == "test-stack-generated-id"
	}, 3*time.Second, integration.DefaultInterval)

	// Make sure we log stack created
	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Stack created")
		return logs.Len() == 1
	}, 3*time.Second, integration.DefaultInterval)

	logContext := logs.All()[0].ContextMap()
	s.Require().Contains(logContext, logging.StackId)
	s.Assert().Equal(logContext[logging.StackId], "test-stack-generated-id")
}

func (s *StackControllerSuite) TestStackCreation_OK_SpaceNotReady() {
	spaceName := "test-space1"

	stack := &v1beta1.Stack{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Stack",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-stack",
			Namespace: "default",
		},
		Spec: v1beta1.StackSpec{
			Branch:     utils.AddressOf("fake-branch"),
			Repository: "fake-repository",
			SpaceName:  utils.AddressOf(spaceName),
		},
	}

	s.FakeSpaceliftStackRepo.EXPECT().Get(mock.Anything, mock.Anything).
		Return(nil, spaceliftRepository.ErrStackNotFound)

	s.Logs.TakeAll()
	stack, err := s.CreateStack(stack)
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Unable to find space for stack")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)

	s.Logs.TakeAll()

	space := &v1beta1.Space{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Space",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      spaceName,
			Namespace: "default",
		},
		Spec: v1beta1.SpaceSpec{
			ParentSpace: "root",
		},
	}

	space, err = s.CreateSpace(space)
	s.Require().NoError(err)
	defer s.DeleteSpace(space)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Space is not ready yet, will retry in 3 seconds")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)

	s.Logs.TakeAll()

	space.Status.Id = "test-space-generated-id"
	err = s.SpaceRepo.UpdateStatus(s.Context(), space)
	s.Require().NoError(err)

	s.FakeSpaceliftStackRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(&models.Stack{
			Id: "test-stack-generated-id",
		}, nil)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Stack created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)

	stack, err = s.StackRepo.Get(s.Context(), types.NamespacedName{
		Namespace: stack.Namespace,
		Name:      stack.ObjectMeta.Name,
	})
	s.Require().NoError(err)
	s.Assert().Len(stack.OwnerReferences, 1)
	s.Assert().Equal(space.ObjectMeta.Name, stack.OwnerReferences[0].Name)
	s.Assert().Equal("Space", stack.OwnerReferences[0].Kind)
}

func (s *StackControllerSuite) TestStackCreationWithSpaceName_OK() {
	s.FakeSpaceliftStackRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrStackNotFound)
	s.FakeSpaceliftStackRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(&models.Stack{
			Id: "test-stack-generated-id",
		}, nil)

	s.Logs.TakeAll()

	space, err := s.CreateTestSpaceWithStatus()
	s.Require().NoError(err)
	defer s.DeleteSpace(space)

	stack := &v1beta1.Stack{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Stack",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-stack",
			Namespace: "default",
		},
		Spec: v1beta1.StackSpec{
			Branch:     utils.AddressOf("fake-branch"),
			Repository: "fake-repository",
			SpaceName:  utils.AddressOf(space.ObjectMeta.Name),
		},
	}

	stack, err = s.CreateStack(stack)
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	// Make sure stack status is updated
	s.Require().Eventually(func() bool {
		stack, err := s.StackRepo.Get(s.Context(), types.NamespacedName{
			Namespace: stack.Namespace,
			Name:      stack.ObjectMeta.Name,
		})
		s.Require().NoError(err)
		return stack.Status.Id == "test-stack-generated-id"
	}, 3*time.Second, integration.DefaultInterval)

	// Make sure we log stack created
	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Stack created")
		return logs.Len() == 1
	}, 3*time.Second, integration.DefaultInterval)

	logContext := logs.All()[0].ContextMap()
	s.Require().Contains(logContext, logging.StackId)
	s.Assert().Equal(logContext[logging.StackId], "test-stack-generated-id")
}

func (s *StackControllerSuite) TestStackUpdate_UnableToUpdateOnSpacelift() {
	s.FakeSpaceliftStackRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(&models.Stack{
			Id: "test-stack-generated-id",
		}, nil)
	s.FakeSpaceliftStackRepo.EXPECT().Update(mock.Anything, mock.Anything).Once().
		Return(nil, fmt.Errorf("unable to update resource on spacelift"))

	s.Logs.TakeAll()
	stack, err := s.CreateTestStack()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	// Make sure we don't update the stack ID
	s.Require().Never(func() bool {
		stack, err := s.StackRepo.Get(s.Context(), types.NamespacedName{
			Namespace: stack.Namespace,
			Name:      stack.ObjectMeta.Name,
		})
		s.Require().NoError(err)
		return stack.Status.Id != ""
	}, 3*time.Second, integration.DefaultInterval)

	// Check that the error has been logged
	logs := s.Logs.FilterMessage("Unable to update the stack in spacelift")
	s.Require().Equal(1, logs.Len())
	logs = s.Logs.FilterMessage("Stack updated")
	s.Require().Equal(0, logs.Len())
}

func (s *StackControllerSuite) TestStackUpdate_OK() {
	fakeStack := &models.Stack{
		Id: "test-stack-generated-id",
	}

	s.FakeSpaceliftStackRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(fakeStack, nil)
	s.FakeSpaceliftStackRepo.EXPECT().Update(mock.Anything, mock.Anything).Once().
		Return(fakeStack, nil)

	s.Logs.TakeAll()
	stack, err := s.CreateTestStack()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	// Make sure we log stack updated
	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Stack updated")
		return logs.Len() == 1
	}, 3*time.Second, integration.DefaultInterval)

	logContext := logs.All()[0].ContextMap()
	s.Require().Contains(logContext, logging.StackId)
	s.Assert().Equal(logContext[logging.StackId], "test-stack-generated-id")
}

func TestStackController(t *testing.T) {
	suite.Run(t, new(StackControllerSuite))
}
