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

// PolicySpec defines the desired state of Policy
// +kubebuilder:validation:XValidation:rule="(has(self.spaceName) != has(self.spaceId)) || (!has(self.spaceName) && !has(self.spaceId))",message="only one of spaceName or spaceId can be set"
type PolicySpec struct {
	// Name of the policy - should be unique in one account
	Name *string `json:"name,omitempty"`
	// Body of the policy
	// +kubebuilder:validation:MinLength=1
	Body string `json:"body"`
	// Type of the policy. Possible values are ACCESS, APPROVAL, GIT_PUSH, INITIALIZATION, LOGIN, PLAN, TASK, TRIGGER and NOTIFICATION.
	// Deprecated values are STACK_ACCESS (use ACCESS instead), TASK_RUN (use TASK instead), and TERRAFORM_PLAN (use PLAN instead).
	// +kubebuilder:validation:Enum:=ACCESS;APPROVAL;GIT_PUSH;INITIALIZATION;LOGIN;PLAN;TASK;TRIGGER;NOTIFICATION
	Type string `json:"type"`

	// Description of the policy
	Description *string  `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`

	// SpaceName is Name of a Space kubernetes resource of the space the policy is in
	SpaceName *string `json:"spaceName,omitempty"`
	// SpaceId is ID (slug) of the space the policy is in
	SpaceId *string `json:"spaceId,omitempty"`

	AttachedStacksNames []string `json:"attachedStacks,omitempty"`
	AttachedStacksIds   []string `json:"attachedStacksIds,omitempty"`
}

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	Id string `json:"id"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Policy is the Schema for the policies API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

func (p *Policy) Name() string {
	if p.Spec.Name != nil {
		return *p.Spec.Name
	}
	return p.ObjectMeta.Name
}

func (p *Policy) SetPolicy(policy models.Policy) {
	p.Status.Id = policy.Id
}

//+kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}
