package models

type Space struct {
	ID              string
	Name            string
	Description     string
	InheritEntities bool
	ParentSpace     string
	Labels          []string

	URL string
}
