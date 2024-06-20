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
	"github.com/spacelift-io/spacelift-operator/tests/integration"
)

type SpaceControllerSuite struct {
	integration.IntegrationTestSuite
	integration.WithSpaceSuiteHelper
}

func (s *SpaceControllerSuite) SetupSuite() {
	s.SetupManager = func(mgr manager.Manager) {
		s.FakeSpaceliftSpaceRepo = new(mocks.SpaceRepository)
		s.SpaceRepo = repository.NewSpaceRepository(mgr.GetClient())
		err := (&controller.SpaceReconciler{
			SpaceRepository:          s.SpaceRepo,
			SpaceliftSpaceRepository: s.FakeSpaceliftSpaceRepo,
		}).SetupWithManager(mgr)
		s.Require().NoError(err)
	}
	s.IntegrationTestSuite.SetupSuite()
	s.WithSpaceSuiteHelper = integration.WithSpaceSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
}

func (s *SpaceControllerSuite) SetupTest() {
	s.FakeSpaceliftSpaceRepo.Test(s.T())
	s.IntegrationTestSuite.SetupTest()
}

func (s *SpaceControllerSuite) TearDownTest() {
	s.FakeSpaceliftSpaceRepo.AssertExpectations(s.T())
	s.FakeSpaceliftSpaceRepo.Calls = nil
	s.FakeSpaceliftSpaceRepo.ExpectedCalls = nil
}

func (s *SpaceControllerSuite) TestSpaceCreation_InvalidSpec() {
	cases := []struct {
		Name        string
		Spec        v1beta1.SpaceSpec
		ExpectedErr string
	}{
		{
			Spec:        v1beta1.SpaceSpec{},
			Name:        "missing parentSpace",
			ExpectedErr: `Space.app.spacelift.io "invalid-space" is invalid: spec.parentSpace: Invalid value: "": spec.parentSpace in body should be at least 1 chars long`,
		},
	}

	for _, c := range cases {
		s.Run(c.Name, func() {
			space := &v1beta1.Space{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Space",
					APIVersion: v1beta1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-space",
					Namespace: "default",
				},
				Spec: c.Spec,
			}
			err := s.Client().Create(s.Context(), space)
			s.Assert().EqualError(err, c.ExpectedErr)
		})
	}
}

func (s *SpaceControllerSuite) TestSpaceCreation_UnableToCreateOnSpacelift() {
	s.FakeSpaceliftSpaceRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrSpaceNotFound)
	s.FakeSpaceliftSpaceRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(nil, fmt.Errorf("unable to create resource on spacelift"))

	s.Logs.TakeAll()
	space, err := s.CreateTestSpace()
	s.Require().NoError(err)
	defer s.DeleteSpace(space)

	// Make sure we don't update the space ID
	s.Require().Never(func() bool {
		space, err := s.SpaceRepo.Get(s.Context(), types.NamespacedName{
			Namespace: space.Namespace,
			Name:      space.ObjectMeta.Name,
		})
		s.Require().NoError(err)
		return space.Status.Id != ""
	}, 3*time.Second, integration.DefaultInterval)

	// Check that the error has been logged
	logs := s.Logs.FilterMessage("Unable to create space in spacelift")
	s.Require().Equal(1, logs.Len())
	logs = s.Logs.FilterMessage("Space created")
	s.Require().Equal(0, logs.Len())
}

func (s *SpaceControllerSuite) TestSpaceCreation_Success() {
	s.FakeSpaceliftSpaceRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrSpaceNotFound)
	s.FakeSpaceliftSpaceRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(&models.Space{
			ID: "test-space-generated-id",
		}, nil)

	s.Logs.TakeAll()
	space, err := s.CreateTestSpace()
	s.Require().NoError(err)
	defer s.DeleteSpace(space)

	// Make sure space status is updated
	s.Require().Eventually(func() bool {
		space, err := s.SpaceRepo.Get(s.Context(), types.NamespacedName{
			Namespace: space.Namespace,
			Name:      space.ObjectMeta.Name,
		})
		s.Require().NoError(err)
		return space.Status.Id == "test-space-generated-id"
	}, 3*time.Second, integration.DefaultInterval)

	// Make sure we log space created
	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Space created")
		return logs.Len() == 1
	}, 3*time.Second, integration.DefaultInterval)

	logContext := logs.All()[0].ContextMap()
	s.Require().Contains(logContext, logging.SpaceId)
	s.Assert().Equal(logContext[logging.SpaceId], "test-space-generated-id")
}

func (s *SpaceControllerSuite) TestSpaceUpdate_UnableToUpdateOnSpacelift() {
	s.FakeSpaceliftSpaceRepo.EXPECT().Get(mock.Anything, mock.Anything).Times(2).
		Return(nil, nil)
	failedUpdate := s.FakeSpaceliftSpaceRepo.EXPECT().Update(mock.Anything, mock.Anything).Once().
		Return(nil, fmt.Errorf("unable to update resource on spacelift"))
	s.FakeSpaceliftSpaceRepo.EXPECT().Update(mock.Anything, mock.Anything).Once().
		Return(&models.Space{
			ID: "test-space-generated-id",
		}, nil).NotBefore(failedUpdate)

	s.Logs.TakeAll()
	space, err := s.CreateTestSpace()
	s.Require().NoError(err)
	defer s.DeleteSpace(space)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Unable to update the space in spacelift")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Space updated")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-space-generated-id", logs.All()[0].ContextMap()[logging.SpaceId])

	space, err = s.SpaceRepo.Get(s.Context(), types.NamespacedName{
		Namespace: space.Namespace,
		Name:      space.ObjectMeta.Name,
	})
	s.Require().NoError(err)
	s.Assert().Equal("test-space-generated-id", space.Status.Id)
}

func (s *SpaceControllerSuite) TestSpaceUpdate_OK() {
	fakeSpace := &models.Space{
		ID: "test-space-generated-id",
	}

	s.FakeSpaceliftSpaceRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(fakeSpace, nil)
	s.FakeSpaceliftSpaceRepo.EXPECT().Update(mock.Anything, mock.Anything).Once().
		Return(fakeSpace, nil)

	s.Logs.TakeAll()
	space, err := s.CreateTestSpace()
	s.Require().NoError(err)
	defer s.DeleteSpace(space)

	// Make sure we log space created
	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Space updated")
		return logs.Len() == 1
	}, 3*time.Second, integration.DefaultInterval)

	logContext := logs.All()[0].ContextMap()
	s.Require().Contains(logContext, logging.SpaceId)
	s.Assert().Equal(logContext[logging.SpaceId], "test-space-generated-id")
}

func TestSpaceController(t *testing.T) {
	suite.Run(t, new(SpaceControllerSuite))
}
