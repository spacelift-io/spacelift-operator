package integration

import (
	"log"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
)

var DefaultValidRun = v1beta1.Run{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Run",
		APIVersion: v1beta1.GroupVersion.String(),
	},
	ObjectMeta: metav1.ObjectMeta{
		GenerateName: "test-run",
		Namespace:    "default",
	},
	Spec: v1beta1.RunSpec{
		StackName: "foobar",
	},
}

type WithRunSuiteHelper struct {
	*IntegrationTestSuite
}

func (s *WithRunSuiteHelper) CreateTestRun() (*v1beta1.Run, error) {
	fakeRunULID := ulid.Make().String()
	run := DefaultValidRun
	run.ObjectMeta.Annotations = map[string]string{
		"test.id": fakeRunULID,
	}
	s.FakeSpaceliftRunRepo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(r *v1beta1.Run) bool {
			return r.ObjectMeta.Annotations["test.id"] == fakeRunULID
		})).
		Once().
		Return(&models.Run{
			Id:    fakeRunULID,
			State: string(v1beta1.RunStateQueued),
			Url:   "http://example.com/test",
		}, nil)
	if err := s.Client().Create(s.Context(), &run); err != nil {
		return nil, err
	}

	return &run, nil
}

func (s *WithRunSuiteHelper) AssertRunState(run *v1beta1.Run, status v1beta1.RunState, timeParams ...time.Duration) *v1beta1.Run {
	var refreshedRun *v1beta1.Run
	timeout := DefaultTimeout
	interval := DefaultInterval
	if len(timeParams) == 1 {
		timeout = timeParams[0]
	}
	if len(timeParams) == 2 {
		timeout = timeParams[1]
	}
	log.Printf("Waiting for run run state: %s", status)
	result := s.Eventually(func() bool {
		var err error
		refreshedRun, err = s.runRepo.Get(s.ctx, types.NamespacedName{
			Namespace: run.Namespace,
			Name:      run.Name,
		})
		s.Require().NoError(err)
		return refreshedRun.Status.State == status
	}, timeout, interval)
	if !result {
		s.FailNowf("Failed to assert run state", "Expected: %s Actual: %s", status, refreshedRun.Status.State)
	}
	log.Printf("Successfully asserted run state: %s", status)
	return refreshedRun
}
