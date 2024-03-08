package structs

import (
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

// StackInput represents the input required to create or update a Stack.
type StackInput struct {
	AddditionalProjectGlobs *[]graphql.String  `json:"additionalProjectGlobs"`
	Administrative          graphql.Boolean    `json:"administrative"`
	AfterApply              *[]graphql.String  `json:"afterApply"`
	AfterDestroy            *[]graphql.String  `json:"afterDestroy"`
	AfterInit               *[]graphql.String  `json:"afterInit"`
	AfterPerform            *[]graphql.String  `json:"afterPerform"`
	AfterPlan               *[]graphql.String  `json:"afterPlan"`
	AfterRun                *[]graphql.String  `json:"afterRun"`
	Autodeploy              graphql.Boolean    `json:"autodeploy"`
	Autoretry               graphql.Boolean    `json:"autoretry"`
	BeforeApply             *[]graphql.String  `json:"beforeApply"`
	BeforeDestroy           *[]graphql.String  `json:"beforeDestroy"`
	BeforeInit              *[]graphql.String  `json:"beforeInit"`
	BeforePerform           *[]graphql.String  `json:"beforePerform"`
	BeforePlan              *[]graphql.String  `json:"beforePlan"`
	Branch                  graphql.String     `json:"branch"`
	Description             *graphql.String    `json:"description"`
	GitHubActionDeploy      graphql.Boolean    `json:"githubActionDeploy"`
	Labels                  *[]graphql.String  `json:"labels"`
	LocalPreviewEnabled     graphql.Boolean    `json:"localPreviewEnabled"`
	Name                    graphql.String     `json:"name"`
	Namespace               *graphql.String    `json:"namespace"`
	ProjectRoot             *graphql.String    `json:"projectRoot"`
	ProtectFromDeletion     graphql.Boolean    `json:"protectFromDeletion"`
	Provider                *graphql.String    `json:"provider"`
	Repository              graphql.String     `json:"repository"`
	RepositoryURL           *graphql.String    `json:"repositoryURL"`
	RunnerImage             *graphql.String    `json:"runnerImage"`
	Space                   *graphql.String    `json:"space"`
	VendorConfig            *VendorConfigInput `json:"vendorConfig"`
	WorkerPool              *graphql.ID        `json:"workerPool"`
}

// VendorConfigInput represents vendor-specific configuration.
type VendorConfigInput struct {
	AnsibleInput        *AnsibleInput        `json:"ansible"`
	CloudFormationInput *CloudFormationInput `json:"cloudFormation"`
	Kubernetes          *KubernetesInput     `json:"kubernetes"`
	Pulumi              *PulumiInput         `json:"pulumi"`
	Terraform           *TerraformInput      `json:"terraform"`
	TerragruntInput     *TerragruntInput     `json:"terragrunt"`
}

// AnsibleInput represents Ansible-specific configuration.
type AnsibleInput struct {
	Playbook graphql.String `json:"playbook"`
}

// CloudFormationInput represents CloudFormation-specific configuration.
type CloudFormationInput struct {
	EntryTemplateFile graphql.String `json:"entryTemplateFile"`
	Region            graphql.String `json:"region"`
	StackName         graphql.String `json:"stackName"`
	TemplateBucket    graphql.String `json:"templateBucket"`
}

// KubernetesInput represents Kubernetes-specific configuration.
type KubernetesInput struct {
	Namespace      graphql.String  `json:"namespace"`
	KubectlVersion *graphql.String `json:"kubectlVersion"`
}

// PulumiInput represents Pulumi-specific configuration.
type PulumiInput struct {
	LoginURL  graphql.String `json:"loginURL"`
	StackName graphql.String `json:"stackName"`
}

type TerragruntInput struct {
	TerraformVersion     graphql.String  `json:"terraformVersion"`
	TerragruntVersion    graphql.String  `json:"terragruntVersion"`
	UseRunAll            graphql.Boolean `json:"useRunAll"`
	UseSmartSanitization graphql.Boolean `json:"useSmartSanitization"`
}

// TerraformInput represents Terraform-specific configuration.
type TerraformInput struct {
	UseSmartSanitization       *graphql.Boolean `json:"useSmartSanitization"`
	Version                    *graphql.String  `json:"version"`
	WorkflowTool               *graphql.String  `json:"workflowTool"`
	Workspace                  *graphql.String  `json:"workspace"`
	ExternalStateAccessEnabled *graphql.Boolean `json:"externalStateAccessEnabled"`
}

func FromStackSpec(stackInput v1beta1.StackInput) StackInput {
	ret := StackInput{
		Administrative:      graphql.Boolean(stackInput.Administrative),
		Autodeploy:          graphql.Boolean(stackInput.Autodeploy),
		Autoretry:           graphql.Boolean(stackInput.Autoretry),
		Branch:              graphql.String(stackInput.Branch),
		GitHubActionDeploy:  graphql.Boolean(stackInput.GitHubActionDeploy),
		LocalPreviewEnabled: graphql.Boolean(stackInput.LocalPreviewEnabled),
		Name:                graphql.String(stackInput.Name),
		ProtectFromDeletion: graphql.Boolean(stackInput.ProtectFromDeletion),
		Repository:          graphql.String(stackInput.Repository),
	}

	ret.AddditionalProjectGlobs = getGraphQLStrings(stackInput.AdditionalProjectGlobs)
	ret.AfterApply = getGraphQLStrings(stackInput.AfterApply)
	ret.AfterDestroy = getGraphQLStrings(stackInput.AfterDestroy)
	ret.AfterInit = getGraphQLStrings(stackInput.AfterInit)
	ret.AfterPerform = getGraphQLStrings(stackInput.AfterPerform)
	ret.AfterPlan = getGraphQLStrings(stackInput.AfterPlan)
	ret.AfterRun = getGraphQLStrings(stackInput.AfterRun)
	ret.BeforeApply = getGraphQLStrings(stackInput.BeforeApply)
	ret.BeforeDestroy = getGraphQLStrings(stackInput.BeforeDestroy)
	ret.BeforeInit = getGraphQLStrings(stackInput.BeforeInit)
	ret.BeforePerform = getGraphQLStrings(stackInput.BeforePerform)
	ret.BeforePlan = getGraphQLStrings(stackInput.BeforePlan)
	ret.Description = getGraphQLString(stackInput.Description)
	ret.Provider = getGraphQLString(stackInput.Provider)
	ret.Labels = getGraphQLStrings(stackInput.Labels)
	ret.Space = getGraphQLString(stackInput.Space)
	ret.ProjectRoot = getGraphQLString(stackInput.ProjectRoot)

	ret.RunnerImage = getGraphQLString(stackInput.RunnerImage)

	ret.VendorConfig = getVendorConfig(stackInput.VendorConfig)

	ret.WorkerPool = getGraphQLID(stackInput.WorkerPool)

	return ret
}

func getGraphQLStrings(input *[]string) *[]graphql.String {
	if input == nil {
		return nil
	}

	var ret []graphql.String
	for _, s := range *input {
		ret = append(ret, graphql.String(s))
	}

	return &ret
}

func getGraphQLString(input *string) *graphql.String {
	if input == nil {
		return nil
	}

	return graphql.NewString(graphql.String(*input))
}

func getGraphQLID(input *string) *graphql.ID {
	if input == nil {
		return nil
	}

	return graphql.NewID(graphql.ID(*input))
}

func getVendorConfig(vendorConfig *v1beta1.VendorConfig) *VendorConfigInput {
	if vendorConfig == nil {
		return nil
	}

	if vendorConfig.CloudFormation != nil {
		return &VendorConfigInput{
			CloudFormationInput: &CloudFormationInput{
				EntryTemplateFile: graphql.String(vendorConfig.CloudFormation.EntryTemplateFile),
				Region:            graphql.String(vendorConfig.CloudFormation.Region),
				StackName:         graphql.String(vendorConfig.CloudFormation.StackName),
				TemplateBucket:    graphql.String(vendorConfig.CloudFormation.TemplateBucket),
			},
		}
	}

	if vendorConfig.Kubernetes != nil {
		return &VendorConfigInput{
			Kubernetes: &KubernetesInput{
				Namespace:      graphql.String(vendorConfig.Kubernetes.Namespace),
				KubectlVersion: getGraphQLString(vendorConfig.Kubernetes.KubectlVersion),
			},
		}
	}

	if vendorConfig.Pulumi != nil {
		return &VendorConfigInput{
			Pulumi: &PulumiInput{
				LoginURL:  graphql.String(vendorConfig.Pulumi.LoginURL),
				StackName: graphql.String(vendorConfig.Pulumi.StackName),
			},
		}
	}

	if vendorConfig.Ansible != nil {
		return &VendorConfigInput{
			AnsibleInput: &AnsibleInput{
				Playbook: graphql.String(vendorConfig.Ansible.Playbook),
			},
		}
	}

	if vendorConfig.Terragrunt != nil {
		return &VendorConfigInput{
			TerragruntInput: &TerragruntInput{
				TerraformVersion:     graphql.String(vendorConfig.Terragrunt.TerraformVersion),
				TerragruntVersion:    graphql.String(vendorConfig.Terragrunt.TerragruntVersion),
				UseRunAll:            graphql.Boolean(vendorConfig.Terragrunt.UseRunAll),
				UseSmartSanitization: graphql.Boolean(vendorConfig.Terragrunt.UseSmartSanitization),
			},
		}
	}

	// If nothing is specified, terraform will be the default vendor
	terraformConfig := &TerraformInput{}
	terraformConfig.Version = getGraphQLString(vendorConfig.Terraform.Version)
	terraformConfig.WorkflowTool = getGraphQLString(vendorConfig.Terraform.WorkflowTool)
	terraformConfig.Workspace = getGraphQLString(vendorConfig.Terraform.Workspace)
	terraformConfig.UseSmartSanitization = (*graphql.Boolean)(&vendorConfig.Terraform.UseSmartSanitization)
	terraformConfig.ExternalStateAccessEnabled = (*graphql.Boolean)(&vendorConfig.Terraform.ExternalStateAccessEnabled)

	return &VendorConfigInput{Terraform: terraformConfig}
}
