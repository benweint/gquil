package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// Based on the __Schema introspection type: https://spec.graphql.org/October2021/#sec-The-__Schema-Type
//
// The QueryType, MutationType, and SubscriptionType fields have been suffixed with 'Name' and
// are represented as strings referring to named types, rather than nested objects.
type Schema struct {
	Description          string
	Types                DefinitionMap
	QueryTypeName        string
	MutationTypeName     string
	SubscriptionTypeName string
	Directives           DirectiveDefinitionList
}

func (s *Schema) FilterBuiltins() {
	s.Types = filterBuiltinTypesAndFields(s.Types)
	s.Directives = filterBuiltinDirectives(s.Directives)
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"types": s.Types.ToSortedList(),
	}

	if s.Description != "" {
		m["description"] = s.Description
	}

	if s.QueryTypeName != "" {
		m["queryTypeName"] = s.QueryTypeName
	}

	if s.MutationTypeName != "" {
		m["mutationTypeName"] = s.MutationTypeName
	}

	if s.SubscriptionTypeName != "" {
		m["subscriptionTypeName"] = s.SubscriptionTypeName
	}

	if len(s.Directives) > 0 {
		m["directives"] = s.Directives
	}

	return json.Marshal(m)
}

func (s *Schema) ResolveNames(names []string) ([]*NameReference, error) {
	var roots []*NameReference
	var badNames []string
	for _, rootName := range names {
		root := s.resolveName(rootName)
		if root == nil {
			badNames = append(badNames, rootName)
		} else {
			roots = append(roots, root)
		}
	}

	if len(badNames) > 0 {
		return nil, fmt.Errorf("unknown name(s): %s", strings.Join(badNames, ", "))
	}

	return roots, nil
}

func (s *Schema) resolveName(name string) *NameReference {
	parts := strings.SplitN(name, ".", 2)
	typePart := parts[0]
	for _, def := range s.Types {
		if def.Name == typePart {
			if len(parts) == 1 {
				return &NameReference{
					TypeName: def.Name,
				}
			}

			fieldPart := parts[1]
			for _, field := range def.Fields {
				if field.Name == fieldPart {
					return &NameReference{
						TypeName:  def.Name,
						FieldName: field.Name,
					}
				}
			}
		}
	}

	return nil
}

// MakeSchema constructs and returns a Schema from the given ast.Schema.
// The provided ast.Schema must be 'complete' in the sense that it must contain type definitions
// for all types used in the schema, including built-in types like String, Int, etc.
func MakeSchema(in *ast.Schema) (*Schema, error) {
	typesByName := DefinitionMap{}
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
	for _, t := range typesByName {
		for _, f := range t.Fields {
			if err := resolveTypeKinds(typesByName, f.Type); err != nil {
				return nil, err
			}
			for _, a := range f.Arguments {
				if err := resolveTypeKinds(typesByName, a.Type); err != nil {
					return nil, err
				}
			}
		}
	}

	for _, d := range directives {
		for _, arg := range d.Arguments {
			if err := resolveTypeKinds(typesByName, arg.Type); err != nil {
				return nil, err
			}
		}
	}

	return &Schema{
		Description:          in.Description,
		Types:                typesByName,
		QueryTypeName:        maybeTypeName(in.Query),
		MutationTypeName:     maybeTypeName(in.Mutation),
		SubscriptionTypeName: maybeTypeName(in.Subscription),
		Directives:           directives,
	}, nil
}
