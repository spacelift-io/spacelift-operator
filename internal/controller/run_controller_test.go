package controller_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest/observer"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/controller"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/logging"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/repository/mocks"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/watcher"
	"github.com/spacelift-io/spacelift-operator/tests/integration"
)

type RunControllerSuite struct {
	integration.IntegrationTestSuite
	integration.WithRunSuiteHelper
	integration.WithStackSuiteHelper
}

func (s *RunControllerSuite) SetupSuite() {
	s.SetupManager = func(mgr manager.Manager) {
		stackOutputRepo := repository.NewStackOutputRepository(mgr.GetClient(), mgr.GetScheme(), mgr.GetEventRecorderFor("stack-output-repository"))
		s.RunRepo = repository.NewRunRepository(mgr.GetClient(), mgr.GetScheme())
		s.FakeSpaceliftRunRepo = new(mocks.RunRepository)
		s.FakeSpaceliftStackRepo = new(mocks.StackRepository)
		s.StackRepo = repository.NewStackRepository(mgr.GetClient())
		w := watcher.NewRunWatcher(s.RunRepo, s.FakeSpaceliftRunRepo)
		err := (&controller.RunReconciler{
			RunRepository:            s.RunRepo,
			StackRepository:          s.StackRepo,
			StackOutputRepository:    stackOutputRepo,
			SpaceliftRunRepository:   s.FakeSpaceliftRunRepo,
			SpaceliftStackRepository: s.FakeSpaceliftStackRepo,
			RunWatcher:               w,
		}).SetupWithManager(mgr)
		s.Require().NoError(err)
	}
	s.IntegrationTestSuite.SetupSuite()
	s.WithRunSuiteHelper = integration.WithRunSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.WithStackSuiteHelper = integration.WithStackSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
}

func (s *RunControllerSuite) SetupTest() {
	s.FakeSpaceliftRunRepo.Test(s.T())
	s.FakeSpaceliftStackRepo.Test(s.T())
}

func (s *RunControllerSuite) TearDownTest() {
	s.FakeSpaceliftRunRepo.AssertExpectations(s.T())
	s.FakeSpaceliftRunRepo.Calls = nil
	s.FakeSpaceliftRunRepo.ExpectedCalls = nil

	s.FakeSpaceliftStackRepo.AssertExpectations(s.T())
	s.FakeSpaceliftStackRepo.Calls = nil
	s.FakeSpaceliftStackRepo.ExpectedCalls = nil
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
		s.Run(c.Name, func() {
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

	stack, err := s.CreateTestStackWithStatus()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	s.Logs.TakeAll()
	err = s.Client().Create(s.Context(), &run)
	s.Require().NoError(err)

	// Make sure we don't update the run ID and state
	s.Require().Never(func() bool {
		run, err := s.RunRepo.Get(s.Context(), types.NamespacedName{
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
		Return(&models.Run{
			State: "READY",
		}, nil)
	s.FakeSpaceliftRunRepo.EXPECT().
		Get(mock.Anything, mock.MatchedBy(func(run *v1beta1.Run) bool {
			return run.Status.State == "READY"
		})).
		Once().
		Return(&models.Run{
			State: "APPLYING",
		}, nil)
	s.FakeSpaceliftRunRepo.EXPECT().
		Get(mock.Anything, mock.MatchedBy(func(run *v1beta1.Run) bool {
			return run.Status.State == "APPLYING"
		})).
		Once().
		Return(&models.Run{
			State: string(v1beta1.RunStateFinished),
		}, nil)

	stack, err := s.CreateTestStackWithStatus()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	s.Logs.TakeAll()
	run, err := s.CreateTestRun()
	s.Require().NoError(err)

	var logs *observer.ObservedLogs
	s.Require().Eventually(func() bool {
		logs = s.Logs.FilterMessage("New run created")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)
	logContext := logs.All()[0].ContextMap()
	s.Assert().Equal(v1beta1.RunStateQueued, logContext[logging.RunState])
	s.Assert().Equal("test-stack", logContext[logging.StackId])
	// A special test.id annotation is set by the CreateTestRun function to be
	// able to assert on that.
	s.Assert().Equal(run.Annotations["test.id"], logContext[logging.RunId])

	// Assert that the Queued state has been applied
	run = s.AssertRunState(run, "READY")
	s.Require().NotNil(run.Annotations)
	s.Assert().Equal("http://example.com/test", run.Annotations[v1beta1.ArgoExternalLink])

	// Assert that the state has been changed by the watcher
	run = s.AssertRunState(run, "APPLYING")

	// Assert that the state has been changed by the watcher to finished
	run = s.AssertRunState(run, v1beta1.RunStateFinished)

	// Make sure no secrets are created by default
	s.Require().Never(func() bool {
		secrets, err := s.GetStackOutput(stack)
		return secrets != nil || !k8sErrors.IsNotFound(err)
	}, integration.DefaultTimeout, integration.DefaultInterval)
}

func (s *RunControllerSuite) TestRunCreation_OK_WithCreateSecretFromStackOutput() {
	run := integration.DefaultValidRun
	run.Spec.CreateSecretFromStackOutput = true

	// mock below will mimic the following state machine from Spacelift.
	// QUEUED -> FINISHED
	// This is working by matching the run state with a mock argument matcher, then returns
	// the next state.
	// We don't really need a real scenario here, we just want to check that secrets are created

	s.FakeSpaceliftRunRepo.EXPECT().
		Get(mock.Anything, mock.MatchedBy(func(run *v1beta1.Run) bool {
			return run.Status.State == v1beta1.RunStateQueued
		})).
		Once().
		Return(&models.Run{
			State: string(v1beta1.RunStateFinished),
		}, nil)

	s.FakeSpaceliftStackRepo.EXPECT().Get(mock.Anything, mock.Anything).Return(&models.Stack{
		Outputs: []models.StackOutput{
			{
				Id:    "STACK_OUTPUT",
				Value: "output-value",
			},
		},
	}, nil)

	stack, err := s.CreateTestStackWithStatus()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

	s.Logs.TakeAll()
	err = s.CreateRun(&run)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		logs := s.Logs.FilterMessage("Updated stack output secret")
		return logs.Len() == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)

	secrets, err := s.GetStackOutput(stack)
	s.Require().NoError(err)
	s.Assert().Equal(map[string][]byte{
		"STACK_OUTPUT": []byte("output-value"),
	}, secrets.Data)
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
		Return(&models.Run{
			State: string(v1beta1.RunStateFinished),
		}, nil).NotBefore(errCall)

	stack, err := s.CreateTestStackWithStatus()
	s.Require().NoError(err)
	defer s.DeleteStack(stack)

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
		run, err := s.RunRepo.Get(s.Context(), types.NamespacedName{
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
