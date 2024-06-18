package structs

import (
	"testing"

	"github.com/shurcooL/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/utils"
)

func TestFromStackSpec(t *testing.T) {
	tests := []struct {
		name   string
		stack  v1beta1.Stack
		assert func(*testing.T, StackInput)
	}{
		{
			name: "defaults",
			stack: v1beta1.Stack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "stack-name",
				},
			},
			assert: func(t *testing.T, input StackInput) {
				assert.EqualValues(t, "stack-name", input.Name)
				assert.EqualValues(t, "main", input.Branch)
				assert.EqualValues(t, false, input.Administrative)
			},
		},
		{
			name: "stack with name",
			stack: v1beta1.Stack{
				ObjectMeta: metav1.ObjectMeta{
					Name: "stack-name",
				},
				Spec: v1beta1.StackSpec{
					Name: utils.AddressOf("new name"),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				assert.EqualValues(t, "new name", input.Name)
			},
		},
		{
			name: "branch is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					Branch: utils.AddressOf("branch"),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				assert.EqualValues(t, "branch", input.Branch)
			},
		},
		{
			name: "repository is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					Repository: "org/repo",
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.Repository)
				assert.EqualValues(t, "org", *input.Namespace)
				assert.EqualValues(t, "repo", input.Repository)
			},
		},
		{
			name: "autodeploy is enabled",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					Autodeploy: utils.AddressOf(true),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.Autodeploy)
				assert.EqualValues(t, true, *input.Autodeploy)
			},
		},
		{
			name: "autoretry is enabled",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					Autoretry: utils.AddressOf(true),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.Autoretry)
				assert.EqualValues(t, true, *input.Autoretry)
			},
		},
		{
			name: "github action deploy is enabled",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					GitHubActionDeploy: utils.AddressOf(true),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.GitHubActionDeploy)
				assert.EqualValues(t, true, *input.GitHubActionDeploy)
			},
		},
		{
			name: "local preview is enabled",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					LocalPreviewEnabled: utils.AddressOf(true),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.LocalPreviewEnabled)
				assert.EqualValues(t, true, *input.LocalPreviewEnabled)
			},
		},
		{
			name: "protect from deletion is enabled",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					ProtectFromDeletion: utils.AddressOf(true),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.ProtectFromDeletion)
				assert.EqualValues(t, true, *input.ProtectFromDeletion)
			},
		},
		{
			name: "additional project globs are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					AdditionalProjectGlobs: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.AddditionalProjectGlobs)
				require.Len(t, *input.AddditionalProjectGlobs, 1)
				assert.EqualValues(t, "test", (*input.AddditionalProjectGlobs)[0])
			},
		},
		{
			name: "AfterApply hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					AfterApply: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.AfterApply)
				require.Len(t, *input.AfterApply, 1)
				assert.EqualValues(t, "test", (*input.AfterApply)[0])
			},
		},
		{
			name: "AfterDestroy hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					AfterDestroy: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.AfterDestroy)
				require.Len(t, *input.AfterDestroy, 1)
				assert.EqualValues(t, "test", (*input.AfterDestroy)[0])
			},
		},
		{
			name: "AfterApply hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					AfterInit: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.AfterInit)
				require.Len(t, *input.AfterInit, 1)
				assert.EqualValues(t, "test", (*input.AfterInit)[0])
			},
		},
		{
			name: "AfterPerform hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					AfterPerform: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.AfterPerform)
				require.Len(t, *input.AfterPerform, 1)
				assert.EqualValues(t, "test", (*input.AfterPerform)[0])
			},
		},
		{
			name: "AfterPlan hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					AfterPlan: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.AfterPlan)
				require.Len(t, *input.AfterPlan, 1)
				assert.EqualValues(t, "test", (*input.AfterPlan)[0])
			},
		},
		{
			name: "AfterRun hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					AfterRun: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.AfterRun)
				require.Len(t, *input.AfterRun, 1)
				assert.EqualValues(t, "test", (*input.AfterRun)[0])
			},
		},
		{
			name: "BeforeApply hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					BeforeApply: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.BeforeApply)
				require.Len(t, *input.BeforeApply, 1)
				assert.EqualValues(t, "test", (*input.BeforeApply)[0])
			},
		},
		{
			name: "BeforeDestroy hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					BeforeDestroy: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.BeforeDestroy)
				require.Len(t, *input.BeforeDestroy, 1)
				assert.EqualValues(t, "test", (*input.BeforeDestroy)[0])
			},
		},
		{
			name: "BeforeInit hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					BeforeInit: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.BeforeInit)
				require.Len(t, *input.BeforeInit, 1)
				assert.EqualValues(t, "test", (*input.BeforeInit)[0])
			},
		},
		{
			name: "BeforePerform hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					BeforePerform: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.BeforePerform)
				require.Len(t, *input.BeforePerform, 1)
				assert.EqualValues(t, "test", (*input.BeforePerform)[0])
			},
		},
		{
			name: "BeforePlan hooks are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					BeforePlan: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.BeforePlan)
				require.Len(t, *input.BeforePlan, 1)
				assert.EqualValues(t, "test", (*input.BeforePlan)[0])
			},
		},
		{
			name: "description is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					Description: utils.AddressOf("test"),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.Description)
				assert.EqualValues(t, "test", *input.Description)
			},
		},
		{
			name: "provider is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					Provider: utils.AddressOf("test"),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.Provider)
				assert.EqualValues(t, "test", *input.Provider)
			},
		},
		{
			name: "labels are set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					Labels: &[]string{"test"},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.Labels)
				require.Len(t, *input.Labels, 1)
				assert.EqualValues(t, "test", (*input.Labels)[0])
			},
		},
		{
			name: "space is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					SpaceId: utils.AddressOf("test"),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.Space)
				assert.EqualValues(t, "test", *input.Space)
			},
		},
		{
			name: "project root is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					ProjectRoot: utils.AddressOf("test"),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.ProjectRoot)
				assert.EqualValues(t, "test", *input.ProjectRoot)
			},
		},
		{
			name: "runner image is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					RunnerImage: utils.AddressOf("test"),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.RunnerImage)
				assert.EqualValues(t, "test", *input.RunnerImage)
			},
		},
		{
			name: "vendor config CF is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					VendorConfig: &v1beta1.VendorConfig{
						CloudFormation: &v1beta1.CloudFormationConfig{
							EntryTemplateFile: "template",
							Region:            "region",
							StackName:         "stack name",
							TemplateBucket:    "bucket",
						},
					},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.VendorConfig)
				assert.EqualValues(t, VendorConfigInput{
					CloudFormationInput: &CloudFormationInput{
						EntryTemplateFile: "template",
						Region:            "region",
						StackName:         "stack name",
						TemplateBucket:    "bucket",
					},
				}, *input.VendorConfig)
			},
		},
		{
			name: "vendor config K8S is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					VendorConfig: &v1beta1.VendorConfig{
						Kubernetes: &v1beta1.KubernetesConfig{
							Namespace:      "namespace",
							KubectlVersion: utils.AddressOf("version"),
						},
					},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.VendorConfig)
				assert.EqualValues(t, VendorConfigInput{
					Kubernetes: &KubernetesInput{
						Namespace:      "namespace",
						KubectlVersion: (*graphql.String)(utils.AddressOf("version")),
					},
				}, *input.VendorConfig)
			},
		},
		{
			name: "vendor config pulumi is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					VendorConfig: &v1beta1.VendorConfig{
						Pulumi: &v1beta1.PulumiConfig{
							LoginURL:  "login url",
							StackName: "stack name",
						},
					},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.VendorConfig)
				assert.EqualValues(t, VendorConfigInput{
					Pulumi: &PulumiInput{
						LoginURL:  "login url",
						StackName: "stack name",
					},
				}, *input.VendorConfig)
			},
		},
		{
			name: "vendor config ansible is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					VendorConfig: &v1beta1.VendorConfig{
						Ansible: &v1beta1.AnsibleConfig{
							Playbook: "playbook",
						},
					},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.VendorConfig)
				assert.EqualValues(t, VendorConfigInput{
					AnsibleInput: &AnsibleInput{
						Playbook: "playbook",
					},
				}, *input.VendorConfig)
			},
		},
		{
			name: "vendor config terragrunt is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					VendorConfig: &v1beta1.VendorConfig{
						Terragrunt: &v1beta1.TerragruntConfig{
							TerraformVersion:     "tf version",
							TerragruntVersion:    "tg version",
							UseRunAll:            true,
							UseSmartSanitization: true,
						},
					},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.VendorConfig)
				assert.EqualValues(t, VendorConfigInput{
					TerragruntInput: &TerragruntInput{
						TerraformVersion:     "tf version",
						TerragruntVersion:    "tg version",
						UseRunAll:            true,
						UseSmartSanitization: true,
					},
				}, *input.VendorConfig)
			},
		},
		{
			name: "vendor config TF is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					VendorConfig: &v1beta1.VendorConfig{
						Terraform: &v1beta1.TerraformConfig{
							UseSmartSanitization:       true,
							Version:                    utils.AddressOf("version"),
							WorkflowTool:               utils.AddressOf("workflow"),
							Workspace:                  utils.AddressOf("workspace"),
							ExternalStateAccessEnabled: true,
						},
					},
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.VendorConfig)
				assert.EqualValues(t, VendorConfigInput{
					Terraform: &TerraformInput{
						UseSmartSanitization:       (*graphql.Boolean)(utils.AddressOf(true)),
						Version:                    (*graphql.String)(utils.AddressOf("version")),
						WorkflowTool:               (*graphql.String)(utils.AddressOf("workflow")),
						Workspace:                  (*graphql.String)(utils.AddressOf("workspace")),
						ExternalStateAccessEnabled: (*graphql.Boolean)(utils.AddressOf(true)),
					},
				}, *input.VendorConfig)
			},
		},
		{
			name: "workerpool is set",
			stack: v1beta1.Stack{
				Spec: v1beta1.StackSpec{
					WorkerPool: utils.AddressOf("workerpool"),
				},
			},
			assert: func(t *testing.T, input StackInput) {
				require.NotNil(t, input.WorkerPool)
				assert.EqualValues(t, "workerpool", *input.WorkerPool)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assert(t, FromStackSpec(&tt.stack))
		})
	}
}
