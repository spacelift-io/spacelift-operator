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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StackSpec defines the desired state of Stack
type StackSpec struct {
	// StackName is the name of the stack, this is mandatory
	// +kubebuilder:validation:MinLength=1
	// TODO add kubebuilder validations

	// AdditionalProjectGlobs []string `json:"additionalProjectGlobs"`

	// Administrative         bool     `json:"administrative"`
	// AfterApply             []string `json:"afterApply"`
	// AfterDestroy           []string `json:"afterDestroy"`
	// AfterInit              []string `json:"afterInit"`
	// AfterPerform           []string `json:"afterPerform"`
	// AfterPlan              []string `json:"afterPlan"`
	// AfterRun               []string `json:"afterRun"`
	// Autodeploy             bool     `json:"autodeploy"`
	// Autoretry              bool     `json:"autoretry"`
	// BeforeApply            []string `json:"beforeApply"`
	// BeforeDestroy          []string `json:"beforeDestroy"`
	// BeforeInit             []string `json:"beforeInit"`
	// BeforePerform          []string `json:"beforePerform"`
	// BeforePlan             []string `json:"beforePlan"`
	// Branch                 string   `json:"branch"`
	// Description            *string  `json:"description,,omitempty"`
	// GitHubActionDeploy     bool     `json:"githubActionDeploy"`
	// IsDisabled             bool     `json:"isDisabled"`
	// Labels                 []string `json:"labels"`
	// LocalPreviewEnabled    bool     `json:"localPreviewEnabled"`
	ManagesStateFile bool `json:"managesStateFile"`

	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Namespace              string   `json:"namespace"`
	// ProjectRoot            *string  `json:"projectRoot,omitempty"`
	// ProtectFromDeletion    bool     `json:"protectFromDeletion"`
	// Provider               string   `json:"provider"`
	// Repository             string   `json:"repository"`
	// RepositoryURL          *string  `json:"repositoryURL,omitempty"`
	// RunnerImage            *string  `json:"runnerImage,omitempty"`
	// Space                  string   `json:"space"`
	// TerraformVersion       *string  `json:"terraformVersion,omitempty"`
	// VCSInteragrionID       string   `json:"vcsIntegrationId"`
	// VendorConfig           struct {
	// 	Ansible struct {
	// 		Playbook string `json:"playbook"`
	// 	} `json:"ansible,omitempty"`
	// 	CloudFormation struct {
	// 		EntryTemplateName string `json:"entryTemplateFile"`
	// 		Region            string `json:"region"`
	// 		StackName         string `json:"stackName"`
	// 		TemplateBucket    string `json:"templateBucket"`
	// 	} `json:"cloudFormation,omitempty"`
	// 	Kubernetes struct {
	// 		Namespace      string  `json:"namespace"`
	// 		KubectlVersion *string `json:"kubectlVersion,omitempty"`
	// 	} `json:"kubernetes,omitempty"`
	// 	Pulumi struct {
	// 		LoginURL  string `json:"loginURL"`
	// 		StackName string `json:"stackName"`
	// 	} `json:"pulumi,omitempty"`
	// 	Terraform struct {
	// 		UseSmartSanitization       bool    `json:"useSmartSanitization"`
	// 		Version                    *string `json:"version,omitempty"`
	// 		WorkflowTool               *string `json:"workflowTool,omitempty"`
	// 		Workspace                  *string `json:"workspace,omitempty"`
	// 		ExternalStateAccessEnabled bool    `json:"externalStateAccessEnabled"`
	// 	} `json:"terraform,omitempty"`
	// } `json:"vendorConfig"`
	// WorkerPool *string `json:"workerPool,omitempty"`
}

// StackStatus defines the observed state of Stack
type StackStatus struct {
	// State is the stack state
	State string `json:"state,omitempty"`
	Name  string `json:"name,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Stack is the Schema for the stacks API
type Stack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackSpec   `json:"spec,omitempty"`
	Status StackStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StackList contains a list of Stack
type StackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Stack `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Stack{}, &StackList{})
}
