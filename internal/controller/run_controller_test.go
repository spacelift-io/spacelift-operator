package controller_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
	"github.com/spacelift-io/spacelift-operator/tests/integration"
)

type RunControllerSuite struct {
	integration.IntegrationTestSuite
	integration.WithRunSuiteHelper
}

func (s *RunControllerSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()
	s.WithRunSuiteHelper = integration.WithRunSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
}

func (s *RunControllerSuite) TestRunCreation_InvalidSpec() {
	cases := []struct {
		Name        string
		Spec        v1beta1.RunSpec
		ExpectedErr string
	}{
		{
			Spec:        v1beta1.RunSpec{},
			Name:        "missing stackName",
			ExpectedErr: `Run.app.spacelift.io "invalid-run" is invalid: spec.stackName: Invalid value: "": spec.stackName in body should be at least 1 chars long`,
		},
	}

	for _, c := range cases {
		s.T().Run(c.Name, func(t *testing.T) {
			newRun := &v1beta1.Run{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Run",
					APIVersion: v1beta1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-run",
					Namespace: "default",
				},
				Spec: c.Spec,
			}
			err := s.Client().Create(s.Context(), newRun)
			s.Assert().EqualError(err, c.ExpectedErr)
		})
	}
}

func (s *RunControllerSuite) TestRunCreation_UnableToCreateOnSpacelift() {
	run := integration.DefaultValidRun
	s.FakeSpaceliftRunRepo.EXPECT().Create(mock.Anything, mock.Anything).Once().
		Return(nil, fmt.Errorf("unable to create resource on spacelift"))

	s.Logs.TakeAll()
	err := s.Client().Create(s.Context(), &run)
	s.Require().NoError(err)

	// Make sure we don't update the run ID and state
	s.Require().Never(func() bool {
		run, err := s.RunRepo().Get(s.Context(), types.NamespacedName{
			Namespace: run.Namespace,
			Name:      run.Name,
		})
		s.Require().NoError(err)
		return run.Status.Id != "" || run.Status.State != ""
	}, 3*time.Second, integration.DefaultInterval)

	// Check that the error has been logged
	logs := s.Logs.FilterMessage("Unable to create the run in spacelift")
	s.Require().Equal(1, logs.Len())
	logs = s.Logs.FilterMessage("New run created")
	s.Require().Equal(0, logs.Len())
}

func (s *RunControllerSuite) TestRunCreation_OK() {
	// mocks below will mimic the following state machine from Spacelift.
	// QUEUED -> READY -> APPLYING -> FINISHED
	// This is working by matching the run state with a mock argument matcher, then returns
	// the next state.
	s.FakeSpaceliftRunRepo.EXPECT().
		Get(mock.Anything, mock.MatchedBy(func(run *v1beta1.Run) bool {
			return run.Status.State == v1beta1.RunStateQueued
		})).
		Once().
		Return(&repository.GetRunOutput{
			State: "READY",
		}, nil)
	s.FakeSpaceliftRunRepo.EXPECT().
		Get(mock.Anything, mock.MatchedBy(func(run *v1beta1.Run) bool {
			return run.Status.State == "READY"
		})).
		Once().
		Return(&repository.GetRunOutput{
			State: "APPLYING",
		}, nil)
	s.FakeSpaceliftRunRepo.EXPECT().
		Get(mock.Anything, mock.MatchedBy(func(run *v1beta1.Run) bool {
			return run.Status.State == "APPLYING"
		})).
		Once().
		Return(&repository.GetRunOutput{
			State: string(v1beta1.RunStateFinished),
		}, nil)

	run, err := s.CreateTestRun()
	s.Require().NoError(err)

	// Assert that the Queued state has been applied
	run = s.AssertRunState(run, "READY")
	s.Require().NotNil(run.Status.Argo)
	s.Assert().Equal(v1beta1.ArgoHealthProgressing, run.Status.Argo.Health)

	// Assert that the state has been changed by the watcher
	run = s.AssertRunState(run, "APPLYING")
	s.Assert().Equal(v1beta1.ArgoHealthProgressing, run.Status.Argo.Health)

	// Assert that the state has been changed by the watcher to finished
	run = s.AssertRunState(run, v1beta1.RunStateFinished)
	s.Assert().Equal(v1beta1.ArgoHealthHealthy, run.Status.Argo.Health)
}

func (s *RunControllerSuite) TestRunCreation_OK_WithErrorDuringWatch() {
	// mocks below will mimic the following state machine from Spacelift.
	// QUEUED -> Error calling spacelift backend -> FINISHED
	errCall := s.FakeSpaceliftRunRepo.EXPECT().
		Get(mock.Anything, mock.MatchedBy(func(run *v1beta1.Run) bool {
			return run.Status.State == v1beta1.RunStateQueued
		})).
		Once().
		Return(nil, fmt.Errorf("temporary error fetching spacelift backend"))

	s.FakeSpaceliftRunRepo.EXPECT().
		Get(mock.Anything, mock.MatchedBy(func(run *v1beta1.Run) bool {
			return run.Status.State == v1beta1.RunStateQueued
		})).
		Once().
		Return(&repository.GetRunOutput{
			State: string(v1beta1.RunStateFinished),
		}, nil).NotBefore(errCall)

	s.Logs.TakeAll()
	run, err := s.CreateTestRun()
	s.Require().NoError(err)

	// Assert that the Queued state has been applied
	s.AssertRunState(run, v1beta1.RunStateQueued)

	// Ensure the state is not changed during 5 seconds because of the error
	// The watcher should retry the query and succeed after 10 seconds.
	// The normal behavior sleep for 3sec when there is no error.
	// So asserting that the status is not changed during 5 seconds allows us to know that we are
	// Using the error sleep interval.
	s.Never(func() bool {
		run, err := s.RunRepo().Get(s.Context(), types.NamespacedName{
			Namespace: run.Namespace,
			Name:      run.Name,
		})
		s.Require().NoError(err)
		return run.Status.State != v1beta1.RunStateQueued
	}, 5*time.Second, integration.DefaultInterval)

	// Assert that the state has been changed by the watcher to finished
	s.AssertRunState(run, v1beta1.RunStateFinished)

	logs := s.Logs.FilterMessage("Error fetching run from spacelift API")
	s.Assert().Equal(1, logs.Len())
}

func TestRunController(t *testing.T) {
	suite.Run(t, new(RunControllerSuite))
}
