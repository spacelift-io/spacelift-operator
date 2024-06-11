package structs

import "github.com/shurcooL/graphql"

func GetGraphQLStrings(input *[]string) *[]graphql.String {
	if input == nil {
		return nil
	}

	var ret []graphql.String
	for _, s := range *input {
		ret = append(ret, graphql.String(s))
	}

	return &ret
}
