package model

import "github.com/vektah/gqlparser/v2/ast"

// EnumValueDefinition represents a single possible value for a GraphQL enum.
// Based on the __EnumValue introspection type specified here: https://spec.graphql.org/October2021/#sec-The-__EnumValue-Type
// Note that the isDeprecated and deprecationReason fields are replaced by the more generic 'directives' field.
type EnumValueDefinition struct {
	Description string        `json:"description,omitempty"`
	Name        string        `json:"name"`
	Directives  DirectiveList `json:"directives,omitempty"`
}

// EnumValueList represents a set of possible enum values for a single enum.
type EnumValueList []*EnumValueDefinition

func makeEnumValueList(in ast.EnumValueList) (EnumValueList, error) {
	var result EnumValueList
	for _, ev := range in {
		val, err := makeEnumValue(ev)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
}

func makeEnumValue(in *ast.EnumValueDefinition) (*EnumValueDefinition, error) {
	directives, err := makeDirectiveList(in.Directives)
	if err != nil {
		return nil, err
	}
	return &EnumValueDefinition{
		Name:        in.Name,
		Description: in.Description,
		Directives:  directives,
	}, nil
}
