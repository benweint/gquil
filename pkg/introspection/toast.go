package introspection

import (
	"fmt"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// responseToAst converts a deserialized introspection query result into an *ast.Schema, which
// may then either be printed to GraphQL SDL, or converted into a model.Schema for further processing.
func responseToAst(s *Schema) (*ast.Schema, error) {
	var defs ast.DefinitionList

	for _, def := range s.Types {
		newDef := ast.Definition{
			Kind:        ast.DefinitionKind(def.Kind),
			Description: def.Description,
			Name:        def.Name,
		}

		if def.Kind == ObjectKind {
			var interfaces []string
			for _, iface := range def.Interfaces {
				interfaces = append(interfaces, iface.Name)
			}
			newDef.Interfaces = interfaces
		}

		if def.Kind == ObjectKind || def.Kind == InputObjectKind || def.Kind == InterfaceKind {
			var fields ast.FieldList
			for _, inField := range def.Fields {
				args, err := makeArgumentDefinitionList(inField.Args)
				if err != nil {
					return nil, err
				}
				field := ast.FieldDefinition{
					Name:        inField.Name,
					Description: inField.Description,
					Type:        makeType(inField.Type),
					Arguments:   args,
					Directives:  synthesizeDeprecationDirective(inField.IsDeprecated, inField.DeprecationReason),
				}
				fields = append(fields, &field)
			}
			for _, inField := range def.InputFields {
				field := ast.FieldDefinition{
					Name:        inField.Name,
					Description: inField.Description,
					Type:        makeType(inField.Type),
				}
				if inField.DefaultValue != nil {
					defaultValue, err := makeDefaultValue(inField.Type, *inField.DefaultValue)
					if err != nil {
						return nil, err
					}
					field.DefaultValue = defaultValue
				}
				fields = append(fields, &field)
			}
			newDef.Fields = fields
		}

		if def.Kind == UnionKind {
			var possibleTypes []string
			for _, pt := range def.PossibleTypes {
				possibleTypes = append(possibleTypes, pt.Name)
			}
			newDef.Types = possibleTypes
		}

		if def.Kind == EnumKind {
			var evs ast.EnumValueList
			for _, ev := range def.EnumValues {
				evs = append(evs, &ast.EnumValueDefinition{
					Name:        ev.Name,
					Description: ev.Description,
					Directives:  synthesizeDeprecationDirective(ev.IsDeprecated, ev.DeprecationReason),
				})
			}
			newDef.EnumValues = evs
		}

		defs = append(defs, &newDef)
	}

	typeMap := map[string]*ast.Definition{}
	for _, def := range defs {
		typeMap[def.Name] = def
	}

	directiveMap := map[string]*ast.DirectiveDefinition{}
	for _, dir := range s.Directives {
		var locations []ast.DirectiveLocation
		for _, loc := range dir.Locations {
			locations = append(locations, ast.DirectiveLocation(loc))
		}
		args, err := makeArgumentDefinitionList(dir.Args)
		if err != nil {
			return nil, err
		}
		directiveMap[dir.Name] = &ast.DirectiveDefinition{
			Name:         dir.Name,
			Description:  dir.Description,
			Arguments:    args,
			Locations:    locations,
			IsRepeatable: dir.IsRepeatable,
			Position: &ast.Position{
				Src: &ast.Source{
					BuiltIn: false,
				},
			},
		}
	}

	return &ast.Schema{
		Types:        typeMap,
		Query:        typeMap[s.QueryType.Name],
		Mutation:     typeMap[s.MutationType.Name],
		Subscription: typeMap[s.SubscriptionType.Name],
		Directives:   directiveMap,
	}, nil
}

func synthesizeDeprecationDirective(deprecated bool, deprecationReason string) ast.DirectiveList {
	if !deprecated {
		return nil
	}

	return ast.DirectiveList{
		&ast.Directive{
			Name: "deprecated",
			Arguments: ast.ArgumentList{
				&ast.Argument{
					Name: "reason",
					Value: &ast.Value{
						Kind: ast.StringValue,
						Raw:  deprecationReason,
					},
				},
			},
		},
	}
}

func makeArgumentDefinitionList(in []InputValue) (ast.ArgumentDefinitionList, error) {
	var result ast.ArgumentDefinitionList
	for _, inArg := range in {
		arg := &ast.ArgumentDefinition{
			Name:        inArg.Name,
			Description: inArg.Description,
			Type:        makeType(inArg.Type),
		}

		if inArg.DefaultValue != nil {
			var err error
			if arg.DefaultValue, err = makeDefaultValue(inArg.Type, *inArg.DefaultValue); err != nil {
				return nil, err
			}
		}

		result = append(result, arg)
	}
	return result, nil
}

func makeDefaultValue(t *Type, raw string) (*ast.Value, error) {
	var kind ast.ValueKind

	switch raw {
	case "null":
		kind = ast.NullValue
	case "true", "false":
		kind = ast.BooleanValue
	default:
		switch t.Kind {
		case ScalarKind:
			switch t.Name {
			case "Int":
				kind = ast.IntValue
			case "Float":
				kind = ast.FloatValue
			case "String":
				return makeValue(raw)
			default:
				kind = ast.StringValue
			}
		case InputObjectKind, ListKind:
			return makeValue(raw)
		case EnumKind:
			kind = ast.EnumValue
		case NonNullKind:
			return makeDefaultValue(t.OfType, raw)
		default:
			return nil, fmt.Errorf("unsupported type kind %s for default value '%s'", t.Kind, raw)
		}
	}

	return &ast.Value{
		Kind: kind,
		Raw:  raw,
	}, nil
}

// makeValue parses a raw string representing a GraphQL Value[1]
// This is useful for handling the __InputValue.defaultValue field[2] in the
// introspection schema.
//
// Since the GraphQL parser we're using doesn't support parsing a Value directly,
// we have to wrap the value in a dummy query here, providing it as an argument
// to a non-existent field.
//
// [1]: https://spec.graphql.org/October2021/#Value
// [2]: https://spec.graphql.org/October2021/#sec-The-__InputValue-Type
func makeValue(raw string) (*ast.Value, error) {
	src := &ast.Source{
		Input: fmt.Sprintf("{ f(in: %s) }", raw),
	}
	doc, err := parser.ParseQuery(src)
	if err != nil {
		return nil, err
	}

	field := doc.Operations[0].SelectionSet[0].(*ast.Field)
	return field.Arguments[0].Value, nil
}

func makeType(in *Type) *ast.Type {
	if in.Kind == NonNullKind {
		wrappedType := makeType(in.OfType)
		wrappedType.NonNull = true
		return wrappedType
	}

	if in.Kind == ListKind {
		wrappedType := makeType(in.OfType)
		return &ast.Type{
			Elem: wrappedType,
		}
	}

	return &ast.Type{
		NamedType: in.Name,
	}
}
