package structs

import (
	"github.com/pkg/errors"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type ConfigAttachmentType string

const (
	ConfigAttachmentTypeEnvVar    = "ENVIRONMENT_VARIABLE"
	ConfigAttachmentTypeFileMount = "FILE_MOUNT"
)

type Hooks struct {
	AfterApply    []string `json:"afterApply"`
	AfterDestroy  []string `json:"afterDestroy"`
	AfterInit     []string `json:"afterInit"`
	AfterPerform  []string `json:"afterPerform"`
	AfterPlan     []string `json:"afterPlan"`
	AfterRun      []string `json:"afterRun"`
	BeforeApply   []string `json:"beforeApply"`
	BeforeDestroy []string `json:"beforeDestroy"`
	BeforeInit    []string `json:"beforeInit"`
	BeforePerform []string `json:"beforePerform"`
	BeforePlan    []string `json:"beforePlan"`
}

type StackAttachment struct {
	Stack    string `json:"stack"`
	Priority int    `json:"priority"`
}

type ConfigAttachments struct {
	Description *string              `json:"description"`
	Id          string               `json:"id"`
	Type        ConfigAttachmentType `json:"type"`
	Value       string               `json:"value"`
	WriteOnly   bool                 `json:"writeOnly"`
}

type ContextInput struct {
	Name              string              `json:"name"`
	Space             string              `json:"space"`
	Description       *string             `json:"description"`
	Labels            []string            `json:"labels"`
	Hooks             Hooks               `json:"hooks"`
	StackAttachments  []StackAttachment   `json:"stackAttachments"`
	ConfigAttachments []ConfigAttachments `json:"configAttachments"`
}

func FromContextSpec(c *v1beta1.Context) (*ContextInput, error) {
	spec := c.Spec
	input := ContextInput{
		Name:  c.Name(),
		Space: *spec.SpaceId,
	}
	if spec.Description != nil {
		input.Description = spec.Description
	}
	input.Labels = spec.Labels

	for _, specAttachment := range spec.Attachments {
		attachment := StackAttachment{}
		if specAttachment.StackId != nil {
			attachment.Stack = *specAttachment.StackId
		}
		if specAttachment.ModuleId != nil {
			attachment.Stack = *specAttachment.ModuleId
		}
		if specAttachment.Priority != nil {
			attachment.Priority = *specAttachment.Priority
		}
		input.StackAttachments = append(input.StackAttachments, attachment)
	}

	input.Hooks.AfterApply = make([]string, 0)
	if spec.Hooks.AfterApply != nil {
		input.Hooks.AfterApply = spec.Hooks.AfterApply
	}

	input.Hooks.AfterDestroy = make([]string, 0)
	if spec.Hooks.AfterDestroy != nil {
		input.Hooks.AfterDestroy = spec.Hooks.AfterDestroy
	}

	input.Hooks.AfterInit = make([]string, 0)
	if spec.Hooks.AfterInit != nil {
		input.Hooks.AfterInit = spec.Hooks.AfterInit
	}

	input.Hooks.AfterPerform = make([]string, 0)
	if spec.Hooks.AfterPerform != nil {
		input.Hooks.AfterPerform = spec.Hooks.AfterPerform
	}

	input.Hooks.AfterPlan = make([]string, 0)
	if spec.Hooks.AfterPlan != nil {
		input.Hooks.AfterPlan = spec.Hooks.AfterPlan
	}

	input.Hooks.AfterRun = make([]string, 0)
	if spec.Hooks.AfterRun != nil {
		input.Hooks.AfterRun = spec.Hooks.AfterRun
	}

	input.Hooks.BeforeApply = make([]string, 0)
	if spec.Hooks.BeforeApply != nil {
		input.Hooks.BeforeApply = spec.Hooks.BeforeApply
	}

	input.Hooks.BeforeDestroy = make([]string, 0)
	if spec.Hooks.BeforeDestroy != nil {
		input.Hooks.BeforeDestroy = spec.Hooks.BeforeDestroy
	}

	input.Hooks.BeforeInit = make([]string, 0)
	if spec.Hooks.BeforeInit != nil {
		input.Hooks.BeforeInit = spec.Hooks.BeforeInit
	}

	input.Hooks.BeforePerform = make([]string, 0)
	if spec.Hooks.BeforePerform != nil {
		input.Hooks.BeforePerform = spec.Hooks.BeforePerform
	}

	input.Hooks.BeforePlan = make([]string, 0)
	if spec.Hooks.BeforePlan != nil {
		input.Hooks.BeforePlan = spec.Hooks.BeforePlan
	}

	for _, env := range spec.Environment {
		writeOnly := false
		if env.Secret != nil {
			writeOnly = *env.Secret
		}
		// This should never happen because we don't reach this code if the secret is not found.
		// Just to be a bit defensive, we return this if that happen to report it
		if env.Value == nil {
			return nil, errors.Errorf("environment value cannot be null for '%s'", env.Id)
		}
		input.ConfigAttachments = append(input.ConfigAttachments, ConfigAttachments{
			Description: env.Description,
			Id:          env.Id,
			Type:        ConfigAttachmentTypeEnvVar,
			Value:       *env.Value,
			WriteOnly:   writeOnly,
		})
	}

	for _, mountedFile := range spec.MountedFiles {
		writeOnly := false
		if mountedFile.Secret != nil {
			writeOnly = *mountedFile.Secret
		}
		// This should never happen because we don't reach this code if the secret is not found.
		// Just to be a bit defensive, we return this if that happen to report it
		if mountedFile.Value == nil {
			return nil, errors.Errorf("mounted file value cannot be null for '%s'", mountedFile.Id)
		}
		input.ConfigAttachments = append(input.ConfigAttachments, ConfigAttachments{
			Description: mountedFile.Description,
			Id:          mountedFile.Id,
			Type:        ConfigAttachmentTypeFileMount,
			Value:       *mountedFile.Value,
			WriteOnly:   writeOnly,
		})
	}

	return &input, nil
}
