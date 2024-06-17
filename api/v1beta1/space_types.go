/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
)

// SpaceSpec defines the desired state of space
type SpaceSpec struct {
	// +kubebuilder:validation:MinLength=1
	ParentSpace     string    `json:"parentSpace"`
	Name            *string   `json:"name,omitempty"`
	Description     string    `json:"description,omitempty"`
	InheritEntities bool      `json:"inheritEntities,omitempty"`
	Labels          *[]string `json:"labels,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Space is the Schema for the Spaces API
type Space struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpaceSpec   `json:"spec,omitempty"`
	Status SpaceStatus `json:"status,omitempty"`
}

func (s *Space) Name() string {
	if s.Spec.Name != nil {
		return *s.Spec.Name
	}
	return s.ObjectMeta.Name
}

func (s *Space) Ready() bool {
	return s.Status.Id != ""
}

type SpaceStatus struct {
	Id string `json:"id,omitempty"`
}

func (s *Space) SetSpace(space models.Space) {
	if space.ID != "" {
		s.Status.Id = space.ID
	}
}

//+kubebuilder:object:root=true

// SpaceList contains a list of Space
type SpaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Space `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Space{}, &SpaceList{})
}
