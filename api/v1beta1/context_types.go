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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
)

// +kubebuilder:validation:XValidation:message=only one of value or valueFromSecret should be set,rule=has(self.valueFromSecret) != has(self.value)
type MountedFile struct {
	// +kubebuilder:validation:MinLength=1
	Id              string                `json:"id"`
	Value           *string               `json:"value,omitempty"`
	ValueFromSecret *v1.SecretKeySelector `json:"valueFromSecret,omitempty"`
	Secret          *bool                 `json:"secret,omitempty"`
	Description     *string               `json:"description,omitempty"`
}

// +kubebuilder:validation:XValidation:message=only one of value or valueFromSecret should be set,rule=has(self.valueFromSecret) != has(self.value)
type Environment struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=^[a-zA-Z_]+[a-zA-Z0-9_]*$
	Id              string                `json:"id"`
	Value           *string               `json:"value,omitempty"`
	ValueFromSecret *v1.SecretKeySelector `json:"valueFromSecret,omitempty"`
	Secret          *bool                 `json:"secret,omitempty"`
	Description     *string               `json:"description,omitempty"`
}

type Hooks struct {
	AfterApply    []string `json:"afterApply,omitempty"`
	AfterDestroy  []string `json:"afterDestroy,omitempty"`
	AfterInit     []string `json:"afterInit,omitempty"`
	AfterPerform  []string `json:"afterPerform,omitempty"`
	AfterPlan     []string `json:"afterPlan,omitempty"`
	AfterRun      []string `json:"afterRun,omitempty"`
	BeforeApply   []string `json:"beforeApply,omitempty"`
	BeforeDestroy []string `json:"beforeDestroy,omitempty"`
	BeforeInit    []string `json:"beforeInit,omitempty"`
	BeforePerform []string `json:"beforePerform,omitempty"`
	BeforePlan    []string `json:"beforePlan,omitempty"`
}

// +kubebuilder:validation:XValidation:message=only one of stack or stackId or moduleId should be set,rule=has(self.stack) != has(self.stackId) != has(self.moduleId)
type Attachment struct {
	// +kubebuilder:validation:MinLength=1
	ModuleId *string `json:"moduleId,omitempty"`
	// +kubebuilder:validation:MinLength=1
	StackId *string `json:"stackId,omitempty"`
	// +kubebuilder:validation:MinLength=1
	Stack    *string `json:"stack,omitempty"`
	Priority *int    `json:"priority,omitempty"`
}

// ContextSpec defines the desired state of Context
// +kubebuilder:validation:XValidation:message=only one of space or spaceId should be set,rule=has(self.spaceId) != has(self.space)
type ContextSpec struct {
	Name *string `json:"name,omitempty"`
	// +kubebuilder:validation:MinLength=1
	SpaceId *string `json:"spaceId,omitempty"`
	// +kubebuilder:validation:MinLength=1
	Space       *string  `json:"space,omitempty"`
	Description *string  `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`

	Attachments  []Attachment  `json:"attachments,omitempty"`
	Hooks        Hooks         `json:"hooks,omitempty"`
	Environment  []Environment `json:"environment,omitempty"`
	MountedFiles []MountedFile `json:"mountedFiles,omitempty"`
}

// ContextStatus defines the observed state of Context
type ContextStatus struct {
	Id string `json:"id"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Context is the Schema for the contexts API
type Context struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContextSpec   `json:"spec,omitempty"`
	Status ContextStatus `json:"status,omitempty"`
}

func (c *Context) Name() string {
	if c.Spec.Name != nil {
		return *c.Spec.Name
	}
	return c.ObjectMeta.Name
}

func (c *Context) SetContext(context *models.Context) {
	if context.Id != "" {
		c.Status.Id = context.Id
	}
}

//+kubebuilder:object:root=true

// ContextList contains a list of Context
type ContextList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Context `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Context{}, &ContextList{})
}
