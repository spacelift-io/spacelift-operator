package integration

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

var DefaultValidPolicy = v1beta1.Policy{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Policy",
		APIVersion: v1beta1.GroupVersion.String(),
	},
	ObjectMeta: metav1.ObjectMeta{
		GenerateName: "test-policy",
		Namespace:    "default",
	},
	Spec: v1beta1.PolicySpec{
		Body: "package spacelift",
		Type: "PLAN",
	},
}

type WithPolicySuiteHelper struct {
	*IntegrationTestSuite
}

func (s *WithPolicySuiteHelper) CreateTestPolicy() (*v1beta1.Policy, error) {
	policy := DefaultValidPolicy
	return &policy, s.CreatePolicy(&policy)
}

func (s *WithPolicySuiteHelper) CreatePolicy(policy *v1beta1.Policy) error {
	return s.Client().Create(s.Context(), policy)
}

func (s *WithPolicySuiteHelper) DeletePolicy(policy *v1beta1.Policy) error {
	return s.Client().Delete(s.Context(), policy)
}
