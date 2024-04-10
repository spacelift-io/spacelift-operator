//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSIntegration) DeepCopyInto(out *AWSIntegration) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSIntegration.
func (in *AWSIntegration) DeepCopy() *AWSIntegration {
	if in == nil {
		return nil
	}
	out := new(AWSIntegration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AnsibleConfig) DeepCopyInto(out *AnsibleConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AnsibleConfig.
func (in *AnsibleConfig) DeepCopy() *AnsibleConfig {
	if in == nil {
		return nil
	}
	out := new(AnsibleConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoStatus) DeepCopyInto(out *ArgoStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoStatus.
func (in *ArgoStatus) DeepCopy() *ArgoStatus {
	if in == nil {
		return nil
	}
	out := new(ArgoStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudFormationConfig) DeepCopyInto(out *CloudFormationConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudFormationConfig.
func (in *CloudFormationConfig) DeepCopy() *CloudFormationConfig {
	if in == nil {
		return nil
	}
	out := new(CloudFormationConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Commit) DeepCopyInto(out *Commit) {
	*out = *in
	if in.AuthorLogin != nil {
		in, out := &in.AuthorLogin, &out.AuthorLogin
		*out = new(string)
		**out = **in
	}
	if in.URL != nil {
		in, out := &in.URL, &out.URL
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Commit.
func (in *Commit) DeepCopy() *Commit {
	if in == nil {
		return nil
	}
	out := new(Commit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesConfig) DeepCopyInto(out *KubernetesConfig) {
	*out = *in
	if in.KubectlVersion != nil {
		in, out := &in.KubectlVersion, &out.KubectlVersion
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesConfig.
func (in *KubernetesConfig) DeepCopy() *KubernetesConfig {
	if in == nil {
		return nil
	}
	out := new(KubernetesConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulumiConfig) DeepCopyInto(out *PulumiConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulumiConfig.
func (in *PulumiConfig) DeepCopy() *PulumiConfig {
	if in == nil {
		return nil
	}
	out := new(PulumiConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Run) DeepCopyInto(out *Run) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Run.
func (in *Run) DeepCopy() *Run {
	if in == nil {
		return nil
	}
	out := new(Run)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Run) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunCreated) DeepCopyInto(out *RunCreated) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunCreated.
func (in *RunCreated) DeepCopy() *RunCreated {
	if in == nil {
		return nil
	}
	out := new(RunCreated)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunList) DeepCopyInto(out *RunList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Run, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunList.
func (in *RunList) DeepCopy() *RunList {
	if in == nil {
		return nil
	}
	out := new(RunList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RunList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunSpec) DeepCopyInto(out *RunSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunSpec.
func (in *RunSpec) DeepCopy() *RunSpec {
	if in == nil {
		return nil
	}
	out := new(RunSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunStatus) DeepCopyInto(out *RunStatus) {
	*out = *in
	if in.Argo != nil {
		in, out := &in.Argo, &out.Argo
		*out = new(ArgoStatus)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunStatus.
func (in *RunStatus) DeepCopy() *RunStatus {
	if in == nil {
		return nil
	}
	out := new(RunStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Stack) DeepCopyInto(out *Stack) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Stack.
func (in *Stack) DeepCopy() *Stack {
	if in == nil {
		return nil
	}
	out := new(Stack)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Stack) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackInput) DeepCopyInto(out *StackInput) {
	*out = *in
	if in.AdditionalProjectGlobs != nil {
		in, out := &in.AdditionalProjectGlobs, &out.AdditionalProjectGlobs
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.Administrative != nil {
		in, out := &in.Administrative, &out.Administrative
		*out = new(bool)
		**out = **in
	}
	if in.AfterApply != nil {
		in, out := &in.AfterApply, &out.AfterApply
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.AfterDestroy != nil {
		in, out := &in.AfterDestroy, &out.AfterDestroy
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.AfterInit != nil {
		in, out := &in.AfterInit, &out.AfterInit
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.AfterPerform != nil {
		in, out := &in.AfterPerform, &out.AfterPerform
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.AfterPlan != nil {
		in, out := &in.AfterPlan, &out.AfterPlan
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.AfterRun != nil {
		in, out := &in.AfterRun, &out.AfterRun
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.Autodeploy != nil {
		in, out := &in.Autodeploy, &out.Autodeploy
		*out = new(bool)
		**out = **in
	}
	if in.Autoretry != nil {
		in, out := &in.Autoretry, &out.Autoretry
		*out = new(bool)
		**out = **in
	}
	if in.BeforeApply != nil {
		in, out := &in.BeforeApply, &out.BeforeApply
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.BeforeDestroy != nil {
		in, out := &in.BeforeDestroy, &out.BeforeDestroy
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.BeforeInit != nil {
		in, out := &in.BeforeInit, &out.BeforeInit
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.BeforePerform != nil {
		in, out := &in.BeforePerform, &out.BeforePerform
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.BeforePlan != nil {
		in, out := &in.BeforePlan, &out.BeforePlan
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.Description != nil {
		in, out := &in.Description, &out.Description
		*out = new(string)
		**out = **in
	}
	if in.GitHubActionDeploy != nil {
		in, out := &in.GitHubActionDeploy, &out.GitHubActionDeploy
		*out = new(bool)
		**out = **in
	}
	if in.IsDisabled != nil {
		in, out := &in.IsDisabled, &out.IsDisabled
		*out = new(bool)
		**out = **in
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.LocalPreviewEnabled != nil {
		in, out := &in.LocalPreviewEnabled, &out.LocalPreviewEnabled
		*out = new(bool)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
	if in.ProjectRoot != nil {
		in, out := &in.ProjectRoot, &out.ProjectRoot
		*out = new(string)
		**out = **in
	}
	if in.ProtectFromDeletion != nil {
		in, out := &in.ProtectFromDeletion, &out.ProtectFromDeletion
		*out = new(bool)
		**out = **in
	}
	if in.Provider != nil {
		in, out := &in.Provider, &out.Provider
		*out = new(string)
		**out = **in
	}
	if in.RepositoryURL != nil {
		in, out := &in.RepositoryURL, &out.RepositoryURL
		*out = new(string)
		**out = **in
	}
	if in.RunnerImage != nil {
		in, out := &in.RunnerImage, &out.RunnerImage
		*out = new(string)
		**out = **in
	}
	if in.Space != nil {
		in, out := &in.Space, &out.Space
		*out = new(string)
		**out = **in
	}
	if in.TerraformVersion != nil {
		in, out := &in.TerraformVersion, &out.TerraformVersion
		*out = new(string)
		**out = **in
	}
	if in.VCSInteragrionID != nil {
		in, out := &in.VCSInteragrionID, &out.VCSInteragrionID
		*out = new(string)
		**out = **in
	}
	if in.VendorConfig != nil {
		in, out := &in.VendorConfig, &out.VendorConfig
		*out = new(VendorConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.WorkerPool != nil {
		in, out := &in.WorkerPool, &out.WorkerPool
		*out = new(string)
		**out = **in
	}
	if in.AWSIntegration != nil {
		in, out := &in.AWSIntegration, &out.AWSIntegration
		*out = new(AWSIntegration)
		**out = **in
	}
	if in.ManagesStateFile != nil {
		in, out := &in.ManagesStateFile, &out.ManagesStateFile
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackInput.
func (in *StackInput) DeepCopy() *StackInput {
	if in == nil {
		return nil
	}
	out := new(StackInput)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackList) DeepCopyInto(out *StackList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Stack, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackList.
func (in *StackList) DeepCopy() *StackList {
	if in == nil {
		return nil
	}
	out := new(StackList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackSpec) DeepCopyInto(out *StackSpec) {
	*out = *in
	in.Settings.DeepCopyInto(&out.Settings)
	if in.CommitSHA != nil {
		in, out := &in.CommitSHA, &out.CommitSHA
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackSpec.
func (in *StackSpec) DeepCopy() *StackSpec {
	if in == nil {
		return nil
	}
	out := new(StackSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackStatus) DeepCopyInto(out *StackStatus) {
	*out = *in
	if in.TrackedCommit != nil {
		in, out := &in.TrackedCommit, &out.TrackedCommit
		*out = new(Commit)
		(*in).DeepCopyInto(*out)
	}
	if in.TrackedCommitSetBy != nil {
		in, out := &in.TrackedCommitSetBy, &out.TrackedCommitSetBy
		*out = new(string)
		**out = **in
	}
	if in.Argo != nil {
		in, out := &in.Argo, &out.Argo
		*out = new(ArgoStatus)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackStatus.
func (in *StackStatus) DeepCopy() *StackStatus {
	if in == nil {
		return nil
	}
	out := new(StackStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TerraformConfig) DeepCopyInto(out *TerraformConfig) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
	if in.WorkflowTool != nil {
		in, out := &in.WorkflowTool, &out.WorkflowTool
		*out = new(string)
		**out = **in
	}
	if in.Workspace != nil {
		in, out := &in.Workspace, &out.Workspace
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TerraformConfig.
func (in *TerraformConfig) DeepCopy() *TerraformConfig {
	if in == nil {
		return nil
	}
	out := new(TerraformConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TerragruntConfig) DeepCopyInto(out *TerragruntConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TerragruntConfig.
func (in *TerragruntConfig) DeepCopy() *TerragruntConfig {
	if in == nil {
		return nil
	}
	out := new(TerragruntConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VendorConfig) DeepCopyInto(out *VendorConfig) {
	*out = *in
	if in.Ansible != nil {
		in, out := &in.Ansible, &out.Ansible
		*out = new(AnsibleConfig)
		**out = **in
	}
	if in.CloudFormation != nil {
		in, out := &in.CloudFormation, &out.CloudFormation
		*out = new(CloudFormationConfig)
		**out = **in
	}
	if in.Kubernetes != nil {
		in, out := &in.Kubernetes, &out.Kubernetes
		*out = new(KubernetesConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Pulumi != nil {
		in, out := &in.Pulumi, &out.Pulumi
		*out = new(PulumiConfig)
		**out = **in
	}
	if in.Terraform != nil {
		in, out := &in.Terraform, &out.Terraform
		*out = new(TerraformConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Terragrunt != nil {
		in, out := &in.Terragrunt, &out.Terragrunt
		*out = new(TerragruntConfig)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VendorConfig.
func (in *VendorConfig) DeepCopy() *VendorConfig {
	if in == nil {
		return nil
	}
	out := new(VendorConfig)
	in.DeepCopyInto(out)
	return out
}
