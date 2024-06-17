package repository_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/k8s/repository"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	"github.com/spacelift-io/spacelift-operator/internal/utils"
	"github.com/spacelift-io/spacelift-operator/tests/integration"
)

type StackOutputRepositorySuite struct {
	integration.IntegrationTestSuite
	integration.WithEventHelper
	integration.WithStackSuiteHelper
	repo *repository.StackOutputRepository
}

func (s *StackOutputRepositorySuite) SetupSuite() {
	s.SetupManager = func(manager manager.Manager) {
		s.repo = repository.NewStackOutputRepository(
			manager.GetClient(),
			manager.GetScheme(),
			manager.GetEventRecorderFor("stack-output-repository"),
		)
	}
	s.WithEventHelper = integration.WithEventHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.WithStackSuiteHelper = integration.WithStackSuiteHelper{
		IntegrationTestSuite: &s.IntegrationTestSuite,
	}
	s.IntegrationTestSuite.SetupSuite()
}

func (s *StackOutputRepositorySuite) TestUpdateOrCreateStackOutputSecret_OK_WithInvalidOutputs() {
	validStack := v1beta1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foobar",
		},
		Spec: v1beta1.StackSpec{
			Name:    "foobar",
			SpaceId: utils.AddressOf("fake-space"),
		},
	}

	stack := validStack
	err := s.Client().Create(s.Context(), &stack)
	s.Require().NoError(err)
	defer s.DeleteStack(&stack)

	outputs := []models.StackOutput{
		{
			Id:    "foo",
			Value: "bar",
		},
		{
			Id:    "foo!",
			Value: "invalid special char",
		},
		{
			Id:    "",
			Value: "invalid empty",
		},
		{
			Id:    "bar",
			Value: "foo",
		},
	}

	secret, err := s.repo.UpdateOrCreateStackOutputSecret(s.Context(), &stack, outputs)
	s.Require().NoError(err)
	s.Equal("stack-output-foobar", secret.Name)

	// Ensure owner reference is set
	s.Require().Len(secret.OwnerReferences, 1)
	s.Equal("Stack", secret.OwnerReferences[0].Kind)
	s.Equal("foobar", secret.OwnerReferences[0].Name)

	// Assert that data is set even if some fields are ignored
	s.Contains(secret.Data, "bar")
	s.EqualValues(secret.Data["bar"], "foo")
	s.Contains(secret.Data, "foo")
	s.EqualValues(secret.Data["foo"], "bar")

	var events []v1.Event
	s.Eventually(func() bool {
		events, _ = s.FindEvents(types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}, v1beta1.EventReasonStackOutputCreated)
		return len(events) == 1
	}, integration.DefaultTimeout, integration.DefaultInterval)

	s.EqualValues(1, events[0].Count)
	s.Equal(`Some stack outputs are not compatible with kubernetes secret key format: foo!,`, events[0].Message)
	s.Equal(v1.EventTypeWarning, events[0].Type)
	s.Equal("Secret", events[0].InvolvedObject.Kind)
	s.Equal("stack-output-foobar", events[0].InvolvedObject.Name)

}

func TestStackOutputRepository(t *testing.T) {
	suite.Run(t, new(StackOutputRepositorySuite))
}
