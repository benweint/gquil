package model

import (
	"encoding/json"

	"github.com/vektah/gqlparser/v2/ast"
)

// Based on the __Field and __InputValue introspection types: https://spec.graphql.org/October2021/#sec-The-__Field-Type
//
// Notable differences from the spec:
//   - __Field and __InputValue are represented here by a single merged type, where some fields are left blank when not
//     applicable.
//   - The directives field represents information about directives attached to a given field or input value, which is
//     not present in the introspection schema.
//   - As a result of the above, the deprecated and deprecationReason fields are omitted, since they would
//     duplicate the content of the more generic directives field.
type FieldDefinition struct {
	Name         string                   `json:"name"`
	Description  string                   `json:"description,omitempty"`
	Type         *Type                    `json:"type"`
	Arguments    InputValueDefinitionList `json:"arguments,omitempty"`    // only for fields
	DefaultValue Value                    `json:"defaultValue,omitempty"` // only for input values
	Directives   DirectiveList            `json:"directives,omitempty"`
}

// FieldDefinitionList represents a set of fields definitions on the same object, interface, or input type.
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

		args, err := makeInputValueDefinitionListFromArgs(f.Arguments)
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
