package integration

import (
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/utils"
)

type WithContextSuiteHelper struct {
	*IntegrationTestSuite
}

var DefaultValidContext = v1beta1.Context{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Context",
		APIVersion: v1beta1.GroupVersion.String(),
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-context",
		Namespace: "default",
	},
	Spec: v1beta1.ContextSpec{
		Name:    "test-context",
		SpaceId: utils.AddressOf("test-space-id"),
	},
}

func (s *WithContextSuiteHelper) CreateContext(context *v1beta1.Context) error {
	if err := s.Client().Create(s.Context(), context); err != nil {
		return err
	}
	s.WaitUntilContextExists(context)
	return nil
}

func (s *WithContextSuiteHelper) CreateTestContext() (*v1beta1.Context, error) {
	context := DefaultValidContext
	return &context, s.CreateContext(&context)
}

func (s *WithContextSuiteHelper) WaitUntilContextExists(context *v1beta1.Context) bool {
	return s.Eventually(func() bool {
		st := &v1beta1.Context{}
		err := s.Client().Get(s.Context(), types.NamespacedName{Namespace: context.Namespace, Name: context.Name}, st)
		return err == nil
	}, DefaultTimeout, DefaultInterval)
}

func (s *WithContextSuiteHelper) DeleteContext(context *v1beta1.Context) {
	err := s.Client().Delete(s.Context(), context)
	s.Require().NoError(err)
	s.WaitUntilContextRemoved(context)
}

func (s *WithContextSuiteHelper) WaitUntilContextRemoved(context *v1beta1.Context) bool {
	return s.Eventually(func() bool {
		st := &v1beta1.Context{}
		err := s.Client().Get(s.Context(), types.NamespacedName{Namespace: context.Namespace, Name: context.Name}, st)
		return k8sErrors.IsNotFound(err)
	}, DefaultTimeout, DefaultInterval)
}
