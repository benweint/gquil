package model

import (
	"encoding/json"

	"github.com/vektah/gqlparser/v2/ast"
)

// Based on the __InputValue introspection type.
type InputValueDefinition struct {
	Name         string
	Description  string
	DefaultValue Value
	Type         *Type
	Directives   DirectiveList
}

type InputValueDefinitionList []*InputValueDefinition

type ArgumentList []*Argument

type Argument struct {
	Name  string `json:"name"`
	Value Value  `json:"value"`
}

func (a *InputValueDefinition) FieldName() string {
	return a.Name
}

func (a *InputValueDefinition) MarshalJSON() ([]byte, error) {
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

func makeInputValueDefinitionListFromArgs(in ast.ArgumentDefinitionList) (InputValueDefinitionList, error) {
	var result InputValueDefinitionList
	for _, a := range in {
		argDef, err := makeInputValueDefinition(a.Name, a.Description, a.Type, a.Directives, a.DefaultValue)
		if err != nil {
			return nil, err
		}
		result = append(result, argDef)
	}
	return result, nil
}

func makeInputValueDefinitionListFromFields(in ast.FieldList) (InputValueDefinitionList, error) {
	var result InputValueDefinitionList
	for _, f := range in {
		inputValueDef, err := makeInputValueDefinition(f.Name, f.Description, f.Type, f.Directives, f.DefaultValue)
		if err != nil {
			return nil, err
		}

		result = append(result, inputValueDef)
	}
	return result, nil
}

func makeInputValueDefinition(name, description string, inType *ast.Type, inDirectives ast.DirectiveList, inDefaultValue *ast.Value) (*InputValueDefinition, error) {
	defaultValue, err := makeValue(inDefaultValue)
	if err != nil {
		return nil, err
	}
	directives, err := makeDirectiveList(inDirectives)
	if err != nil {
		return nil, err
	}
	return &InputValueDefinition{
		Name:         name,
		Description:  description,
		Type:         makeType(inType),
		DefaultValue: defaultValue,
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
