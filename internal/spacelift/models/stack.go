package models

type Stack struct {
	Id                 string
	State              string
	Url                string
	Outputs            []StackOutput
	TrackedCommit      *Commit
	TrackedCommitSetBy *string
}

type StackOutput struct {
	Id    string
	Value string
}

type Commit struct {
	AuthorLogin *string
	AuthorName  string
	Hash        string
	Message     string
	Timestamp   uint
	URL         *string
}
