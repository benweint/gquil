package model

import (
	"sort"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// Based on the __Schema introspection type: https://spec.graphql.org/October2021/#sec-The-__Schema-Type
//
// The QueryType, MutationType, and SubscriptionType fields have been suffixed with 'Name' and
// are represented as strings referring to named types, rather than nested objects.
type Schema struct {
	Description          string                  `json:"description,omitempty"`
	Types                DefinitionList          `json:"types"`
	QueryTypeName        string                  `json:"queryTypeName,omitempty"`
	MutationTypeName     string                  `json:"mutationTypeName,omitempty"`
	SubscriptionTypeName string                  `json:"subscriptionTypeName,omitempty"`
	Directives           DirectiveDefinitionList `json:"directives,omitempty"`
}

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

	// We sort here in order to ensure a deterministic ordering of types in the JSON representation,
	// since ast.Definition.Types is a map, which does not preserve ordering.
	sort.Slice(types, func(i, j int) bool {
		return strings.Compare(types[i].Name, types[j].Name) < 0
	})

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
		Description:          in.Description,
		Types:                types,
		QueryTypeName:        maybeTypeName(in.Query),
		MutationTypeName:     maybeTypeName(in.Mutation),
		SubscriptionTypeName: maybeTypeName(in.Subscription),
		Directives:           directives,
	}, nil
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
