package model

import (
	"fmt"

	"github.com/vektah/gqlparser/v2/ast"
)

// Based on the __Type introspection type: https://spec.graphql.org/October2021/#sec-The-__Type-Type
type Definition struct {
	Kind        ast.DefinitionKind `json:"kind"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Directives  DirectiveList      `json:"directives,omitempty"`

	// only set for interfaces, objects
	Fields FieldDefinitionList `json:"fields,omitempty"`

	// only set for input objects
	InputFields InputValueDefinitionList `json:"inputFields,omitempty"`

	// only set for interfaces
	Interfaces []string `json:"interfaces,omitempty"`

	// only set for interfaces & unions
	PossibleTypes []string `json:"possibleTypeNames,omitempty"`

	// only set for enums
	EnumValues EnumValueList `json:"enumValues,omitempty"`
}

func (d *Definition) String() string {
	return fmt.Sprintf("Def{name=%s, kind=%s}", d.Name, d.Kind)
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

	if in.Kind == ast.Object || in.Kind == ast.Interface {
		fields, err := makeFieldDefinitionList(in.Fields)
		if err != nil {
			return nil, err
		}
		def.Fields = fields
	} else if in.Kind == ast.InputObject {
		inputFields, err := makeInputValueDefinitionListFromFields(in.Fields)
		if err != nil {
			return nil, err
		}
		def.InputFields = inputFields
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
