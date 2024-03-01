package controller_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/tests/integration"
)

type RunControllerSuite struct {
	integration.IntegrationTestSuite
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

func (s *RunControllerSuite) TestRunCreation_OK() {
	run := &v1beta1.Run{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Run",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-run",
			Namespace: "default",
		},
		Spec: v1beta1.RunSpec{
			StackName: "foobar",
		},
	}
	err := s.Client().Create(s.Context(), run)
	s.Require().NoError(err)

	// Assert that the Queued state has been applied
	s.Eventually(func() bool {
		run, err = s.RunRepo().Get(s.Context(), types.NamespacedName{
			Namespace: run.Namespace,
			Name:      run.Name,
		})
		s.Require().NoError(err)
		return run.Status.State == v1beta1.RunStateQueued
	}, 10*time.Second, 1*time.Second)
}

func TestRunController(t *testing.T) {
	suite.Run(t, new(RunControllerSuite))
}
