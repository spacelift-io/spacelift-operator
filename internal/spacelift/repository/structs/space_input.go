package structs

import (
	"github.com/shurcooL/graphql"

	"github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

type SpaceInput struct {
	Name            graphql.String    `json:"name"`
	Description     graphql.String    `json:"description"`
	InheritEntities graphql.Boolean   `json:"inheritEntities"`
	ParentSpace     graphql.String    `json:"parentSpace"`
	Labels          *[]graphql.String `json:"labels"`
}

func FromSpaceSpec(spec v1beta1.SpaceSpec) SpaceInput {
	return SpaceInput{
		Name:            graphql.String(spec.Name),
		Description:     graphql.String(spec.Description),
		InheritEntities: graphql.Boolean(spec.InheritEntities),
		ParentSpace:     graphql.String(spec.ParentSpace),
		Labels:          getGraphQLStrings(spec.Labels),
	}
}
