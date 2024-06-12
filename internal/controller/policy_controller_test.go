package controller_test

import (
	"context"
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

type PolicyControllerSuite struct {
	integration.IntegrationTestSuite
	integration.WithPolicySuiteHelper
	integration.WithSpaceSuiteHelper
	integration.WithStackSuiteHelper
}

func (s *PolicyControllerSuite) SetupSuite() {
	s.SetupManager = func(mgr manager.Manager) {
		s.FakeSpaceliftPolicyRepo = new(mocks.PolicyRepository)
		s.SpaceRepo = repository.NewSpaceRepository(mgr.GetClient())
		s.StackRepo = repository.NewStackRepository(mgr.GetClient(), mgr.GetScheme())
		s.PolicyRepo = repository.NewPolicyRepository(mgr.GetClient(), mgr.GetScheme())
		err := (&controller.PolicyReconciler{
			PolicyRepository:          s.PolicyRepo,
			SpaceRepository:           s.SpaceRepo,
			StackRepository:           s.StackRepo,
			SpaceliftPolicyRepository: s.FakeSpaceliftPolicyRepo,
		}).SetupWithManager(mgr)
		s.Require().NoError(err)
	}
	s.IntegrationTestSuite.SetupSuite()
	s.WithPolicySuiteHelper = integration.WithPolicySuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.WithSpaceSuiteHelper = integration.WithSpaceSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.WithStackSuiteHelper = integration.WithStackSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
}

func (s *PolicyControllerSuite) SetupTest() {
	s.FakeSpaceliftPolicyRepo.Test(s.T())
	s.IntegrationTestSuite.SetupTest()
}

func (s *PolicyControllerSuite) TestPolicyCreation_InvalidSpec() {
	cases := []struct {
		Name        string
		Spec        v1beta1.PolicySpec
		ExpectedErr string
	}{
		{
			Name: "empty name",
			Spec: v1beta1.PolicySpec{
				Body: "test",
				Type: "ACCESS",
			},
			ExpectedErr: `Policy.app.spacelift.io "invalid-policy" is invalid: spec.name: Invalid value: "": spec.name in body should be at least 1 chars long`,
		},
		{
			Name: "empty body",
			Spec: v1beta1.PolicySpec{
				Name: "test",
				Type: "ACCESS",
			},
			ExpectedErr: `Policy.app.spacelift.io "invalid-policy" is invalid: spec.body: Invalid value: "": spec.body in body should be at least 1 chars long`,
		},
		{
			Name: "empty type",
			Spec: v1beta1.PolicySpec{
				Name: "test",
				Body: "test",
			},
			ExpectedErr: `Policy.app.spacelift.io "invalid-policy" is invalid: spec.type: Unsupported value: "": supported values: "ACCESS", "APPROVAL", "GIT_PUSH", "INITIALIZATION", "LOGIN", "PLAN", "TASK", "TRIGGER", "NOTIFICATION"`,
		},
		{
			Name: "invalid type",
			Spec: v1beta1.PolicySpec{
				Name: "test",
				Body: "test",
				Type: "FOOBAR",
			},
			ExpectedErr: `Policy.app.spacelift.io "invalid-policy" is invalid: spec.type: Unsupported value: "FOOBAR": supported values: "ACCESS", "APPROVAL", "GIT_PUSH", "INITIALIZATION", "LOGIN", "PLAN", "TASK", "TRIGGER", "NOTIFICATION"`,
		},
		{
			Name: "both stackName and stackId are set",
			Spec: v1beta1.PolicySpec{
				Name:      "test",
				Body:      "test",
				Type:      "ACCESS",
				SpaceId:   utils.AddressOf("space-id"),
				SpaceName: utils.AddressOf("space-name"),
			},
			ExpectedErr: `Policy.app.spacelift.io "invalid-policy" is invalid: spec: Invalid value: "object": only one of spaceName or spaceId can be set`,
		},
	}

	for _, c := range cases {
		s.Run(c.Name, func() {
			policy := &v1beta1.Policy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Policy",
					APIVersion: v1beta1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-policy",
					Namespace: "default",
				},
				Spec: c.Spec,
			}
			err := s.Client().Create(s.Context(), policy)
			s.Assert().EqualError(err, c.ExpectedErr)
		})
	}
}

func (s *PolicyControllerSuite) TestPolicyCreation_UnableToCreateOnSpacelift() {
	s.FakeSpaceliftPolicyRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(nil, fmt.Errorf("unable to create resource on spacelift"))
	s.FakeSpaceliftPolicyRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrPolicyNotFound)

	policy, err := s.CreateTestPolicy()
	s.Require().NoError(err)
	defer s.DeletePolicy(policy)

	// Make sure we don't update the policy ID
	s.Require().Never(func() bool {
		policy, err := s.PolicyRepo.Get(s.Context(), types.NamespacedName{
			Namespace: policy.Namespace,
			Name:      policy.Name,
		})
		s.Require().NoError(err)
		return policy.Status.Id != ""
	}, 3*time.Second, integration.DefaultInterval)

	// Check that the error has been logged
	logs := s.Logs.FilterMessage("Unable to create policy in spacelift")
	s.Require().Equal(1, logs.Len())
	logs = s.Logs.FilterMessage("Policy created")
	s.Require().Equal(0, logs.Len())
}

func (s *PolicyControllerSuite) TestPolicyCreation_OK_AttachedStackNotReady() {
	p := integration.DefaultValidPolicy
	p.Spec.AttachedStacksNames = []string{"test-stack"}

	err := s.CreatePolicy(&p)
	s.Require().NoError(err)
	defer s.DeletePolicy(&p)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Unable to find attached stack for policy, will retry in 10 seconds")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-stack", logs.All()[0].ContextMap()[logging.StackName])
	s.Assert().EqualValues(logging.Level4, -logs.All()[0].Level)

	stack, err := s.CreateTestStack()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Stack is not ready, will retry in 3 seconds")
		return logs.Len() == 1
	}, 12*time.Second, integration.DefaultInterval)
	s.Assert().Equal("test-stack", logs.All()[0].ContextMap()[logging.StackName])

	s.FakeSpaceliftPolicyRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrPolicyNotFound)
	var policySpecToCreate v1beta1.PolicySpec
	s.FakeSpaceliftPolicyRepo.EXPECT().Create(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *v1beta1.Policy) {
			policySpecToCreate = c.Spec
		}).Once().
		Return(&models.Policy{Id: "test-policy-id"}, nil)

	stack.Status.Id = "test-stack-id"
	err = s.StackRepo.UpdateStatus(s.Context(), stack)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Policy created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-policy-id", logs.All()[0].ContextMap()[logging.PolicyId])

	s.Assert().Contains(policySpecToCreate.AttachedStacksIds, "test-stack-id")
}

func (s *PolicyControllerSuite) TestPolicyCreation_OK_SpaceNotReady() {
	p := integration.DefaultValidPolicy
	p.Spec.SpaceName = utils.AddressOf("test-space")

	err := s.CreatePolicy(&p)
	s.Require().NoError(err)
	defer s.DeletePolicy(&p)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Unable to find space for policy, will retry in 10 seconds")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-space", logs.All()[0].ContextMap()[logging.SpaceName])
	s.Assert().EqualValues(logging.Level4, -logs.All()[0].Level)

	space, err := s.CreateTestSpace()
	s.Require().NoError(err)
	defer s.DeleteSpace(space)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Space is not ready, will retry in 3 seconds")
		return logs.Len() == 1
	}, 12*time.Second, integration.DefaultInterval)
	s.Assert().Equal("test-space", logs.All()[0].ContextMap()[logging.SpaceName])

	s.FakeSpaceliftPolicyRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrPolicyNotFound)
	var policySpecToCreate v1beta1.PolicySpec
	s.FakeSpaceliftPolicyRepo.EXPECT().Create(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *v1beta1.Policy) {
			policySpecToCreate = c.Spec
		}).Once().
		Return(&models.Policy{Id: "test-policy-id"}, nil)

	space.Status.Id = "test-space"
	err = s.SpaceRepo.UpdateStatus(s.Context(), space)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Policy created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-policy-id", logs.All()[0].ContextMap()[logging.PolicyId])

	s.Require().NotNil(policySpecToCreate.SpaceId)
	s.Assert().Equal("test-space", *policySpecToCreate.SpaceId)
}

func (s *PolicyControllerSuite) TestPolicyCreation_OK() {

	s.FakeSpaceliftPolicyRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, spaceliftRepository.ErrPolicyNotFound)
	s.FakeSpaceliftPolicyRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(&models.Policy{Id: "test-policy-id"}, nil)

	policy, err := s.CreateTestPolicy()
	s.Require().NoError(err)
	defer s.DeletePolicy(policy)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Policy created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-policy-id", logs.All()[0].ContextMap()[logging.PolicyId])

	policy, err = s.PolicyRepo.Get(s.Context(), types.NamespacedName{
		Namespace: policy.Namespace,
		Name:      policy.Name,
	})
	s.Require().NoError(err)
	s.Assert().Equal("test-policy-id", policy.Status.Id)
}

func (s *PolicyControllerSuite) TestPolicyUpdate_OK() {

	s.FakeSpaceliftPolicyRepo.EXPECT().Get(mock.Anything, mock.Anything).Once().
		Return(nil, nil)
	s.FakeSpaceliftPolicyRepo.EXPECT().Update(mock.Anything, mock.Anything).Once().
		Return(&models.Policy{Id: "test-policy-id"}, nil)

	policy, err := s.CreateTestPolicy()
	s.Require().NoError(err)
	defer s.DeletePolicy(policy)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("Policy updated")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	s.Assert().Equal("test-policy-id", logs.All()[0].ContextMap()[logging.PolicyId])

	policy, err = s.PolicyRepo.Get(s.Context(), types.NamespacedName{
		Namespace: policy.Namespace,
		Name:      policy.Name,
	})
	s.Require().NoError(err)
	s.Assert().Equal("test-policy-id", policy.Status.Id)
}

func TestPolicyController(t *testing.T) {
	suite.Run(t, new(PolicyControllerSuite))
}
