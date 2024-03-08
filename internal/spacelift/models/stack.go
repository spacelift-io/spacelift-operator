package models

type Stack struct {
	Id                 string
	State              string
	Url                string
	TrackedCommit      *Commit
	TrackedCommitSetBy *string
}

type Commit struct {
	AuthorLogin *string
	AuthorName  string
	Hash        string
	Message     string
	Timestamp   uint
	URL         *string
}
