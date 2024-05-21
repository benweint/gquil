package model

import "github.com/vektah/gqlparser/v2/ast"

// Directive represents a specific application site / instantiation of a directive.
//
// Note that this type does *not* appear in the GraphQL introspection schema, which lacks information
// about most directive application sites (with the exception of @deprecated, which is special-cased).
type Directive struct {
	Name      string       `json:"name"`
	Arguments ArgumentList `json:"arguments,omitempty"`
}

// DirectiveList represents a list of directives all applied at the same application site.
type DirectiveList []*Directive

// DirectiveDefinition represents the definition of a directive.
// Based on the __Directive introspection type defined here: https://spec.graphql.org/October2021/#sec-The-__Directive-Type
type DirectiveDefinition struct {
	Description  string                   `json:"description"`
	Name         string                   `json:"name"`
	Arguments    InputValueDefinitionList `json:"arguments,omitempty"`
	Locations    []ast.DirectiveLocation  `json:"locations"`
	IsRepeatable bool                     `json:"repeatable"`
}

// DirectiveDefinitionList represents a list of directive definitions.
type DirectiveDefinitionList []*DirectiveDefinition

func makeDirectiveDefinition(in *ast.DirectiveDefinition) (*DirectiveDefinition, error) {
	args, err := makeInputValueDefinitionListFromArgs(in.Arguments)
	if err != nil {
		return nil, err
	}

	return &DirectiveDefinition{
		Name:         in.Name,
		Description:  in.Description,
		Arguments:    args,
		Locations:    in.Locations,
		IsRepeatable: in.IsRepeatable,
	}, nil
}

func makeDirectiveList(in ast.DirectiveList) (DirectiveList, error) {
	var out DirectiveList
	for _, d := range in {
		args, err := makeArgumentList(d.Arguments)
		if err != nil {
			return nil, err
		}
		out = append(out, &Directive{
			Name:      d.Name,
			Arguments: args,
		})
	}
	return out, nil
}
