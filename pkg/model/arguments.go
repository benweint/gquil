package model

import (
	"encoding/json"

	"github.com/vektah/gqlparser/v2/ast"
)

type ArgumentDefinition struct {
	Name         string
	Description  string
	DefaultValue Value
	Type         *Type
	Directives   DirectiveList
}

type ArgumentDefinitionList []*ArgumentDefinition

type ArgumentList []*Argument

type Argument struct {
	Name  string `json:"name"`
	Value Value  `json:"value"`
}

func (a *ArgumentDefinition) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"name":               a.Name,
		"type":               a.Type,
		"typeName":           a.Type.String(),
		"underlyingTypeName": a.Type.Unwrap().Name,
	}

	if a.Description != "" {
		m["description"] = a.Description
	}

	if a.DefaultValue != nil {
		m["defaultValue"] = a.DefaultValue
	}

	if len(a.Directives) != 0 {
		m["directives"] = a.Directives
	}

	return json.Marshal(m)
}

func makeArgumentDefinitionList(in ast.ArgumentDefinitionList) (ArgumentDefinitionList, error) {
	var result ArgumentDefinitionList
	for _, a := range in {
		argDef, err := makeArgumentDefinition(a)
		if err != nil {
			return nil, err
		}
		result = append(result, argDef)
	}
	return result, nil
}

func makeArgumentDefinition(in *ast.ArgumentDefinition) (*ArgumentDefinition, error) {
	defaultValue, err := makeValue(in.DefaultValue)
	if err != nil {
		return nil, err
	}
	directives, err := makeDirectiveList(in.Directives)
	if err != nil {
		return nil, err
	}
	return &ArgumentDefinition{
		Name:         in.Name,
		Description:  in.Description,
		DefaultValue: defaultValue,
		Type:         makeType(in.Type),
		Directives:   directives,
	}, nil
}

func makeArgumentList(in ast.ArgumentList) (ArgumentList, error) {
	var out ArgumentList
	for _, a := range in {
		val, err := makeValue(a.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, &Argument{
			Name:  a.Name,
			Value: val,
		})
	}
	return out, nil
}
