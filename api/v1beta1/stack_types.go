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

// StackSpec defines the desired state of Stack
type StackSpec struct {
	Name     string     `json:"name"`
	Settings StackInput `json:"settings"`
	// +kubebuilder:validation:MinLength=1
	CommitSHA string `json:"commitSHA"`
}

type StackInput struct {
	AdditionalProjectGlobs *[]string     `json:"additionalProjectGlobs,omitempty"`
	Administrative         *bool         `json:"administrative,omitempty"`
	AfterApply             *[]string     `json:"afterApply,omitempty"`
	AfterDestroy           *[]string     `json:"afterDestroy,omitempty"`
	AfterInit              *[]string     `json:"afterInit,omitempty"`
	AfterPerform           *[]string     `json:"afterPerform,omitempty"`
	AfterPlan              *[]string     `json:"afterPlan,omitempty"`
	AfterRun               *[]string     `json:"afterRun,omitempty"`
	Autodeploy             *bool         `json:"autodeploy,omitempty"`
	Autoretry              *bool         `json:"autoretry,omitempty"`
	BeforeApply            *[]string     `json:"beforeApply,omitempty"`
	BeforeDestroy          *[]string     `json:"beforeDestroy,omitempty"`
	BeforeInit             *[]string     `json:"beforeInit,omitempty"`
	BeforePerform          *[]string     `json:"beforePerform,omitempty"`
	BeforePlan             *[]string     `json:"beforePlan,omitempty"`
	Branch                 string        `json:"branch"`
	Description            *string       `json:"description,,omitempty"`
	GitHubActionDeploy     *bool         `json:"githubActionDeploy,omitempty"`
	IsDisabled             *bool         `json:"isDisabled,omitempty"`
	Labels                 *[]string     `json:"labels,omitempty"`
	LocalPreviewEnabled    *bool         `json:"localPreviewEnabled,omitempty"`
	Namespace              *string       `json:"namespace,omitempty"`
	ProjectRoot            *string       `json:"projectRoot,omitempty"`
	ProtectFromDeletion    *bool         `json:"protectFromDeletion,omitempty"`
	Provider               *string       `json:"provider,omitempty"`
	Repository             string        `json:"repository"`
	RepositoryURL          *string       `json:"repositoryURL,omitempty"`
	RunnerImage            *string       `json:"runnerImage,omitempty"`
	Space                  *string       `json:"space,omitempty"`
	TerraformVersion       *string       `json:"terraformVersion,omitempty"`
	VCSInteragrionID       *string       `json:"vcsIntegrationId,omitempty"`
	VendorConfig           *VendorConfig `json:"vendorConfig,omitempty"`
	WorkerPool             *string       `json:"workerPool,omitempty"`

	AWSIntegration *AWSIntegration `json:"awsIntegration,omitempty"`

	// In our API managesStateFile is not part of StackInput
	ManagesStateFile *bool `json:"managesStateFile,omitempty"`
}

type VendorConfig struct {
	Ansible        *AnsibleConfig        `json:"ansible,omitempty"`
	CloudFormation *CloudFormationConfig `json:"cloudFormation,omitempty"`
	Kubernetes     *KubernetesConfig     `json:"kubernetes,omitempty"`
	Pulumi         *PulumiConfig         `json:"pulumi,omitempty"`
	Terraform      *TerraformConfig      `json:"terraform,omitempty"`
	Terragrunt     *TerragruntConfig     `json:"terragrunt,omitempty"`
}

type AnsibleConfig struct {
	Playbook string `json:"playbook"`
}

type CloudFormationConfig struct {
	EntryTemplateFile string `json:"entryTemplateFile"`
	Region            string `json:"region"`
	StackName         string `json:"stackName"`
	TemplateBucket    string `json:"templateBucket"`
}

type KubernetesConfig struct {
	Namespace      string  `json:"namespace"`
	KubectlVersion *string `json:"kubectlVersion,omitempty"`
}

type PulumiConfig struct {
	LoginURL  string `json:"loginURL"`
	StackName string `json:"stackName"`
}

type TerraformConfig struct {
	UseSmartSanitization       bool    `json:"useSmartSanitization,omitempty"`
	Version                    *string `json:"version,omitempty"`
	WorkflowTool               *string `json:"workflowTool,omitempty"`
	Workspace                  *string `json:"workspace,omitempty"`
	ExternalStateAccessEnabled bool    `json:"externalStateAccessEnabled,omitempty"`
}

type TerragruntConfig struct {
	TerraformVersion     string `json:"terraformVersion"`
	TerragruntVersion    string `json:"terragruntVersion"`
	UseRunAll            bool   `json:"useRunAll"`
	UseSmartSanitization bool   `json:"useSmartSanitization"`
}

// StackStatus defines the observed state of Stack
type StackStatus struct {
	// State is the stack state
	Id                 string  `json:"id,omitempty"`
	Url                string  `json:"url,omitempty"`
	TrackedCommit      *Commit `json:"trackedCommit,omitempty"`
	TrackedCommitSetBy *string `json:"trackedCommitSetBy,omitempty"`
	Ready              bool    `json:"ready,omitempty"`
	// Argo is a status that could be used by argo health check to sync on health
	Argo *ArgoStatus `json:"argo,omitempty"`
}

type Commit struct {
	AuthorLogin *string `json:"authorLogin,omitempty"`
	AuthorName  string  `json:"authorName"`
	Hash        string  `json:"hash"`
	Message     string  `json:"message"`
	Timestamp   uint    `json:"timestamp"`
	URL         *string `json:"url,omitempty"`
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

func (s *Stack) Ready() bool {
	return s.Status.Ready
}

// SetStack is used to sync the k8s CRD with a spacelift stack model.
// It basically takes care of updating all status fields
func (s *Stack) SetStack(stack models.Stack) {
	if stack.Id != "" {
		s.Status.Id = stack.Id
	}

	if stack.Url != "" {
		s.Status.Url = stack.Url
	}

	if stack.TrackedCommit != nil {
		s.Status.TrackedCommit = &Commit{
			AuthorLogin: stack.TrackedCommit.AuthorLogin,
			AuthorName:  stack.TrackedCommit.AuthorName,
			Hash:        stack.TrackedCommit.Hash,
			Message:     stack.TrackedCommit.Message,
			Timestamp:   stack.TrackedCommit.Timestamp,
			URL:         stack.TrackedCommit.URL,
		}
		s.Status.Ready = s.Status.TrackedCommit.Hash == s.Spec.CommitSHA
	}

	if stack.TrackedCommitSetBy != nil {
		s.Status.TrackedCommitSetBy = stack.TrackedCommitSetBy
	}
}

type AWSIntegration struct {
	Id    string `json:"id"`
	Read  bool   `json:"read"`
	Write bool   `json:"write"`
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
