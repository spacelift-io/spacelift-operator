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

// RunSpec defines the desired state of Run
type RunSpec struct {
	// StackName is the name of the stack for this run, this is mandatory
	// +kubebuilder:validation:MinLength=1
	StackName string `json:"stackName"`
}

type RunState string

const (
	RunStateQueued      RunState = "QUEUED"
	RunStateCanceled    RunState = "CANCELED"
	RunStateFailed      RunState = "FAILED"
	RunStateFinished    RunState = "FINISHED"
	RunStateUnconfirmed RunState = "UNCONFIRMED"
	RunStateDiscarded   RunState = "DISCARDED"
	RunStateStopped     RunState = "STOPPED"
	RunStateSkipped     RunState = "SKIPPED"
)

var terminalStates = map[RunState]interface{}{
	RunStateCanceled:  nil,
	RunStateFailed:    nil,
	RunStateFinished:  nil,
	RunStateDiscarded: nil,
	RunStateStopped:   nil,
	RunStateSkipped:   nil,
}

// RunStatus defines the observed state of Run
type RunStatus struct {
	// State is the run state, see RunState for all possibles state of a run
	State RunState `json:"state,omitempty"`
	// Id is the run ULID on Spacelift
	Id string `json:"id,omitempty"`
	// Argo is a status that could be used by argo health check to sync on health
	Argo *ArgoStatus `json:"argo,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=".status.state"
//+kubebuilder:printcolumn:name="Id",type=string,JSONPath=".status.id"

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

// IsTerminated returns true if the run is in a terminal state
func (r *Run) IsTerminated() bool {
	_, found := terminalStates[r.Status.State]
	return found
}

type RunCreated struct {
	Id, Url string
	State   RunState
}

// SetRun is used to sync the k8s CRD with a spacelift run model.
// It basically takes care of updating all status fields
func (r *Run) SetRun(run *models.Run) {
	if run.Id != "" {
		r.Status.Id = run.Id
	}
	if run.State != "" {
		r.Status.State = RunState(run.State)
	}
	argoHealth := &ArgoStatus{
		Health: ArgoHealthProgressing,
	}
	if r.Status.State == RunStateFinished ||
		r.Status.State == RunStateSkipped {
		argoHealth.Health = ArgoHealthHealthy
	}
	if r.Status.State == RunStateUnconfirmed {
		argoHealth.Health = ArgoHealthSuspended
	}
	if r.Status.State == RunStateFailed ||
		r.Status.State == RunStateStopped ||
		r.Status.State == RunStateCanceled ||
		r.Status.State == RunStateDiscarded {
		argoHealth.Health = ArgoHealthDegraded
	}
	r.Status.Argo = argoHealth
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
