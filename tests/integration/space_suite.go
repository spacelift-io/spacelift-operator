package integration

import (
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type WithSpaceSuiteHelper struct {
	*IntegrationTestSuite
}

func (s *WithSpaceSuiteHelper) CreateTestSpace() (*v1beta1.Space, error) {
	space := &v1beta1.Space{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Space",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-space",
			Namespace: "default",
		},
		Spec: v1beta1.SpaceSpec{
			Name:        "test-space",
			ParentSpace: "root",
		},
	}
	if err := s.Client().Create(s.Context(), space); err != nil {
		return nil, err
	}

	return space, nil
}

func (s *WithSpaceSuiteHelper) CreateSpace(space *v1beta1.Space) (*v1beta1.Space, error) {
	if err := s.Client().Create(s.Context(), space); err != nil {
		return nil, err
	}

	return space, nil
}

func (s *WithSpaceSuiteHelper) CreateTestSpaceWithStatus() (*v1beta1.Space, error) {
	space, err := s.CreateTestSpace()
	if err != nil {
		return nil, err
	}

	space.Status = v1beta1.SpaceStatus{Id: "test-space-id", Ready: true}
	if err := s.Client().Status().Update(s.Context(), space); err != nil {
		return nil, err
	}

	return space, nil
}
func (s *WithSpaceSuiteHelper) DeleteSpace(space *v1beta1.Space) {
	err := s.Client().Delete(s.Context(), space)
	s.Require().NoError(err)
	s.WaitUntilSpaceRemoved(space)
}

func (s *WithSpaceSuiteHelper) WaitUntilSpaceRemoved(space *v1beta1.Space) {
	s.Eventually(func() bool {
		st := &v1beta1.Space{}
		err := s.Client().Get(s.Context(), types.NamespacedName{Name: space.Name, Namespace: space.Namespace}, st)
		return k8sErrors.IsNotFound(err)
	}, DefaultTimeout, DefaultInterval)
}
