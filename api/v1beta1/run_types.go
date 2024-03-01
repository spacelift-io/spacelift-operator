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
)

// RunSpec defines the desired state of Run
type RunSpec struct {
	// StackName is the name of the stack for this run, this is mandatory
	// +kubebuilder:validation:MinLength=1
	StackName string `json:"stackName"`
}

type RunState string

const (
	RunStateQueued = "QUEUED"
)

// RunStatus defines the observed state of Run
type RunStatus struct {
	// State is the run state, see RunState for all possibles state of a run
	State RunState `json:"state,omitempty"`
	// Argo is a status that could be used by argo health check to sync on health
	Argo *ArgoStatus `json:"argo,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=".status.state"

// Run is the Schema for the runs API
type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunSpec   `json:"spec"`
	Status RunStatus `json:"status,omitempty"`
}

// IsNew return true if the resource has just been created.
// If status.state is nil, it means that the controller does not have handled it yet, so it mean that it's a new one
func (r *Run) IsNew() bool {
	return r.Status.State == ""
}

//+kubebuilder:object:root=true

// RunList contains a list of Run
type RunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Run `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Run{}, &RunList{})
}
