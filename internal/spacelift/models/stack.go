package models

type Stack struct {
	Id      string
	Url     string
	Outputs []StackOutput
}

type StackOutput struct {
	Id    string
	Value string
}
