package model

import (
	"encoding/json"

	"github.com/vektah/gqlparser/v2/ast"
)

type FieldDefinition struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Type         *Type                  `json:"type"`
	Arguments    ArgumentDefinitionList `json:"arguments,omitempty"`    // only for objects
	DefaultValue Value                  `json:"defaultValue,omitempty"` // only for input objects
	Directives   DirectiveList          `json:"directives,omitempty"`
}

type FieldDefinitionList []*FieldDefinition

func (fd *FieldDefinition) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"name":               fd.Name,
		"type":               fd.Type,
		"typeName":           fd.Type.String(),
		"underlyingTypeName": fd.Type.Unwrap().Name,
	}

	if fd.Description != "" {
		m["description"] = fd.Description
	}

	if len(fd.Arguments) != 0 {
		m["arguments"] = fd.Arguments
	}

	// TODO: zerovalues
	if fd.DefaultValue != nil {
		m["defaultValue"] = fd.DefaultValue
	}

	if len(fd.Directives) != 0 {
		m["directives"] = fd.Directives
	}

	return json.Marshal(m)
}

func makeFieldDefinitionList(in ast.FieldList) (FieldDefinitionList, error) {
	var result FieldDefinitionList
	for _, f := range in {
		defaultValue, err := makeValue(f.DefaultValue)
		if err != nil {
			return nil, err
		}

		args, err := makeArgumentDefinitionList(f.Arguments)
		if err != nil {
			return nil, err
		}

		directives, err := makeDirectiveList(f.Directives)
		if err != nil {
			return nil, err
		}

		result = append(result, &FieldDefinition{
			Name:         f.Name,
			Description:  f.Description,
			Type:         makeType(f.Type),
			Arguments:    args,
			Directives:   directives,
			DefaultValue: defaultValue,
		})
	}
	return result, nil
}
