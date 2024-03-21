package integration

import (
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type WithStackSuiteHelper struct {
	*IntegrationTestSuite
}

var DefaultValidStack = v1beta1.Stack{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Stack",
		APIVersion: v1beta1.GroupVersion.String(),
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-stack",
		Namespace: "default",
	},
	Spec: v1beta1.StackSpec{
		Name:      "test-stack",
		CommitSHA: "ed56c7b20e3dd075013cf0d7ab3ce083fdb7900f",
		Settings: v1beta1.StackInput{
			ManagesStateFile: false,
			Branch:           "fake-branch",
			Repository:       "fake-repository",
		},
	},
	Status: v1beta1.StackStatus{
		Id:    "test-stack",
		Ready: true,
	},
}

func (s *WithStackSuiteHelper) CreateTestStack() (*v1beta1.Stack, error) {
	stack := DefaultValidStack
	if err := s.Client().Create(s.Context(), &stack); err != nil {
		return nil, err
	}
	stack.Status = DefaultValidStack.Status
	if err := s.Client().Status().Update(s.Context(), &stack); err != nil {
		return nil, err
	}
	return &stack, nil
}

func (s *WithStackSuiteHelper) DeleteStack(stack *v1beta1.Stack) {
	err := s.Client().Delete(s.Context(), stack)
	s.Require().NoError(err)
	s.WaitUntilStackRemoved(stack)
}

func (s *WithStackSuiteHelper) WaitUntilStackRemoved(stack *v1beta1.Stack) bool {
	return s.Eventually(func() bool {
		st := &v1beta1.Stack{}
		err := s.Client().Get(s.Context(), types.NamespacedName{Namespace: stack.Namespace, Name: stack.Name}, st)
		return k8sErrors.IsNotFound(err)
	}, DefaultTimeout, DefaultInterval)
}

func (s *WithStackSuiteHelper) GetStackOutput(stack *v1beta1.Stack) (*v1.Secret, error) {
	secret := &v1.Secret{}
	if err := s.Client().Get(s.Context(), types.NamespacedName{
		Namespace: stack.Namespace,
		Name:      "stack-output-" + stack.Status.Id,
	}, secret); err != nil {
		return nil, err
	}
	return secret, nil
}
