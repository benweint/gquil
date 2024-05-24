package model

import (
	"fmt"
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

func (s *Schema) FilterBuiltins() {
	s.Types = filterBuiltinTypes(s.Types)
	s.Directives = filterBuiltinDirectives(s.Directives)
}

func (s *Schema) ResolveNames(names []string) ([]*NameReference, error) {
	var roots []*NameReference
	var badNames []string
	for _, rootName := range names {
		root := s.ResolveName(rootName)
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

func (s *Schema) ResolveName(name string) *NameReference {
	parts := strings.SplitN(name, ".", 2)
	typePart := parts[0]
	for _, def := range s.Types {
		if def.Name == typePart {
			if len(parts) == 1 {
				return &NameReference{
					Kind:    TypeNameReference,
					typeRef: def,
				}
			}

			fieldPart := parts[1]
			for _, field := range def.Fields {
				if field.Name == fieldPart {
					return &NameReference{
						Kind:    FieldNameReference,
						typeRef: def,
						field:   field,
					}
				}
			}

			for _, field := range def.InputFields {
				if field.Name == fieldPart {
					return &NameReference{
						Kind:       InputFieldNameReference,
						typeRef:    def,
						inputField: field,
					}
				}
			}
		}
	}

	return nil
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
