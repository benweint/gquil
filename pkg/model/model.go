package model

import (
	"github.com/vektah/gqlparser/v2/ast"
)

type Schema struct {
	Types      DefinitionList          `json:"types"`
	Directives DirectiveDefinitionList `json:"directives"`
}

type Definition struct {
	Kind        ast.DefinitionKind `json:"kind"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Directives  DirectiveList      `json:"directives,omitempty"`

	// only set for interfaces, objects
	Fields FieldDefinitionList `json:"fields,omitempty"`

	// only set for input objects
	InputFields FieldDefinitionList `json:"inputFields,omitempty"`

	// only set for interfaces
	Interfaces []string `json:"interfaces,omitempty"`

	// only set for interfaces & unions
	PossibleTypes []string `json:"possibleTypeNames,omitempty"`

	// only set for enums
	EnumValues EnumValueList `json:"enumValues,omitempty"`
}

type DefinitionList []*Definition

func (s *Schema) FilterBuiltins() {
	s.Types = filterBuiltinTypes(s.Types)
	s.Directives = filterBuiltinDirectives(s.Directives)
}

func MakeSchema(in *ast.Schema) (*Schema, error) {
	var types DefinitionList
	typesByName := map[string]*Definition{}
	for _, def := range in.Types {
		t, err := makeDefinition(def)
		if err != nil {
			return nil, err
		}

		if t.Kind == ast.Interface {
			var possibleTypes []string
			for _, possibleType := range in.PossibleTypes[def.Name] {
				possibleTypes = append(possibleTypes, possibleType.Name)
			}
			t.PossibleTypes = possibleTypes
		}

		types = append(types, t)
		typesByName[t.Name] = t
	}

	var directives DirectiveDefinitionList
	for _, dd := range in.Directives {
		def, err := makeDirectiveDefinition(dd)
		if err != nil {
			return nil, err
		}
		directives = append(directives, def)
	}

	// Resolve type kinds for named types by looking them up in typesByName
	for _, t := range types {
		for _, f := range t.Fields {
			resolveTypeKinds(typesByName, f.Type)
			for _, a := range f.Arguments {
				resolveTypeKinds(typesByName, a.Type)
			}
		}
		for _, f := range t.InputFields {
			resolveTypeKinds(typesByName, f.Type)
		}
	}

	for _, d := range directives {
		for _, arg := range d.Arguments {
			resolveTypeKinds(typesByName, arg.Type)
		}
	}

	return &Schema{
		Types:      types,
		Directives: directives,
	}, nil
}

func resolveTypeKinds(typesByName map[string]*Definition, t *Type) {
	if t.OfType != nil {
		resolveTypeKinds(typesByName, t.OfType)
	} else {
		t.Kind = TypeKind(typesByName[t.Name].Kind)
	}
}

func makeDefinition(in *ast.Definition) (*Definition, error) {
	def := &Definition{
		Kind:          in.Kind,
		Name:          in.Name,
		Description:   in.Description,
		Interfaces:    in.Interfaces,
		PossibleTypes: in.Types,
	}

	fields, err := makeFieldDefinitionList(in.Fields)
	if err != nil {
		return nil, err
	}
	if in.Kind == ast.Object || in.Kind == ast.Interface {
		def.Fields = fields
	} else if in.Kind == ast.InputObject {
		def.InputFields = fields
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
