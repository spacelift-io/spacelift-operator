package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
	"github.com/spacelift-io/spacelift-operator/internal/utils"
)

func TestFromContextSpec(t *testing.T) {
	context := &v1beta1.Context{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-context",
		},
		Spec: v1beta1.ContextSpec{
			SpaceId:     utils.AddressOf("space-id"),
			Description: utils.AddressOf("description"),
			Labels:      []string{"label1", "label2"},
			Attachments: []v1beta1.Attachment{
				{
					StackId:  utils.AddressOf("stack-id"),
					Priority: utils.AddressOf(1),
				},
				{
					ModuleId: utils.AddressOf("module-id"),
				},
			},
			Hooks: v1beta1.Hooks{
				AfterApply:    []string{"AfterApply"},
				AfterDestroy:  []string{"AfterDestroy"},
				AfterInit:     []string{"AfterInit"},
				AfterPerform:  []string{"AfterPerform"},
				AfterPlan:     []string{"AfterPlan"},
				AfterRun:      []string{"AfterRun"},
				BeforeApply:   []string{"BeforeApply"},
				BeforeDestroy: []string{"BeforeDestroy"},
				BeforeInit:    []string{"BeforeInit"},
				BeforePerform: []string{"BeforePerform"},
			},
			Environment: []v1beta1.Environment{
				{
					Id:    "env-id-1",
					Value: utils.AddressOf("env-value-1"),
				},
				{
					Id:          "env-id-2",
					Value:       utils.AddressOf("env-value-2"),
					Secret:      utils.AddressOf(true),
					Description: utils.AddressOf("description"),
				},
			},
			MountedFiles: []v1beta1.MountedFile{
				{
					Id:    "file-id-1",
					Value: utils.AddressOf("file-value-1"),
				},
				{
					Id:          "file-id-2",
					Value:       utils.AddressOf("file-value-2"),
					Secret:      utils.AddressOf(true),
					Description: utils.AddressOf("description"),
				},
			},
		},
	}
	expectedInput := ContextInput{
		Name:        "test-context",
		Space:       "space-id",
		Description: utils.AddressOf("description"),
		Labels:      []string{"label1", "label2"},
		Hooks: Hooks{
			AfterApply:    []string{"AfterApply"},
			AfterDestroy:  []string{"AfterDestroy"},
			AfterInit:     []string{"AfterInit"},
			AfterPerform:  []string{"AfterPerform"},
			AfterPlan:     []string{"AfterPlan"},
			AfterRun:      []string{"AfterRun"},
			BeforeApply:   []string{"BeforeApply"},
			BeforeDestroy: []string{"BeforeDestroy"},
			BeforeInit:    []string{"BeforeInit"},
			BeforePerform: []string{"BeforePerform"},
			BeforePlan:    []string{},
		},
		StackAttachments: []StackAttachment{
			{
				Stack:    "stack-id",
				Priority: 1,
			},
			{
				Stack:    "module-id",
				Priority: 0,
			},
		},
		ConfigAttachments: []ConfigAttachments{
			{
				Id:    "env-id-1",
				Value: "env-value-1",
				Type:  ConfigAttachmentTypeEnvVar,
			},
			{
				Id:          "env-id-2",
				Value:       "env-value-2",
				WriteOnly:   true,
				Description: utils.AddressOf("description"),
				Type:        ConfigAttachmentTypeEnvVar,
			},
			{
				Id:    "file-id-1",
				Value: "file-value-1",
				Type:  ConfigAttachmentTypeFileMount,
			},
			{
				Id:          "file-id-2",
				Value:       "file-value-2",
				Type:        ConfigAttachmentTypeFileMount,
				WriteOnly:   true,
				Description: utils.AddressOf("description"),
			},
		},
	}
	actual, err := FromContextSpec(context)
	require.NoError(t, err)
	require.NotNil(t, actual)
	assert.EqualValues(t, expectedInput, *actual)
}
