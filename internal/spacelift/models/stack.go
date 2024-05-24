package models

import (
	"regexp"
)

type Stack struct {
	Id      string
	Url     string
	Outputs []StackOutput
}

type StackOutput struct {
	Id    string
	Value string
}

// See https://kubernetes.io/docs/concepts/configuration/secret/#restriction-names-data
var validOutputId = regexp.MustCompile(`^[a-zA-Z0-9\-_.]+$`)

func (o *StackOutput) IsCompatibleWithKubeSecret() bool {
	return validOutputId.MatchString(o.Id)
}
