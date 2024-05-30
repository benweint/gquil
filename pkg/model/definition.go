package model

import (
	"encoding/json"
	"fmt"

	"github.com/vektah/gqlparser/v2/ast"
)

// Based on the __Type introspection type: https://spec.graphql.org/October2021/#sec-The-__Type-Type
type Definition struct {
	Kind        ast.DefinitionKind
	Name        string
	Description string
	Directives  DirectiveList

	// only set for interfaces, objects, input objects
	Fields FieldDefinitionList

	// only set for interfaces
	Interfaces []string

	// only set for interfaces & unions
	PossibleTypes []string

	// only set for enums
	EnumValues EnumValueList
}

func (d *Definition) String() string {
	return fmt.Sprintf("Def{name=%s, kind=%s}", d.Name, d.Kind)
}

func (d *Definition) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"kind": d.Kind,
		"name": d.Name,
	}

	if d.Description != "" {
		m["description"] = d.Description
	}

	if len(d.Directives) > 0 {
		m["directives"] = d.Directives
	}

	if len(d.Fields) > 0 {
		fieldsKeyName := "fields"
		if d.Kind == ast.InputObject {
			fieldsKeyName = "inputFields"
		}
		m[fieldsKeyName] = d.Fields
	}

	if len(d.Interfaces) > 0 {
		m["interfaces"] = d.Interfaces
	}

	if len(d.PossibleTypes) > 0 {
		m["possibleTypeNames"] = d.PossibleTypes
	}

	if len(d.EnumValues) > 0 {
		m["enumValues"] = d.EnumValues
	}

	return json.Marshal(m)
}

type DefinitionList []*Definition

func makeDefinition(in *ast.Definition) (*Definition, error) {
	def := &Definition{
		Kind:          in.Kind,
		Name:          in.Name,
		Description:   in.Description,
		Interfaces:    in.Interfaces,
		PossibleTypes: in.Types,
	}

	if in.Kind == ast.Object || in.Kind == ast.Interface || in.Kind == ast.InputObject {
		fields, err := makeFieldDefinitionList(in.Fields)
		if err != nil {
			return nil, err
		}
		def.Fields = fields
	}

	directives, err := makeDirectiveList(in.Directives)
	if err != nil {
		return nil, err
	}
	def.Directives = directives

	enumValues, err := makeEnumValueList(in.EnumValues)
	if err != nil {
		return nil, err
	}
	def.EnumValues = enumValues

	return def, nil
}

func maybeTypeName(in *ast.Definition) string {
	if in == nil {
		return ""
	}
	return in.Name
}

func resolveTypeKinds(typesByName map[string]*Definition, t *Type) {
	if t.OfType != nil {
		resolveTypeKinds(typesByName, t.OfType)
	} else {
		t.Kind = TypeKind(typesByName[t.Name].Kind)
	}
}
