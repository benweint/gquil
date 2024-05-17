package introspection

import (
	"fmt"

	"github.com/vektah/gqlparser/v2/ast"
)

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
				field := ast.FieldDefinition{
					Name:        inField.Name,
					Description: inField.Description,
					Type:        makeType(inField.Type),
					Arguments:   makeArgumentDefinitionList(inField.Args),
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
					field.DefaultValue = makeDefaultValue(inField.Type, *inField.DefaultValue)
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
		directiveMap[dir.Name] = &ast.DirectiveDefinition{
			Name:         dir.Name,
			Description:  dir.Description,
			Arguments:    makeArgumentDefinitionList(dir.Args),
			Locations:    locations,
			IsRepeatable: dir.IsRepeatable,
		}
	}

	return &ast.Schema{
		Types:        typeMap,
		Query:        typeMap[s.QueryType.Name],
		Mutation:     typeMap[s.MutationType.Name],
		Subscription: typeMap[s.SubscriptionType.Name],
		// TODO: directives
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

func makeArgumentDefinitionList(in []InputValue) ast.ArgumentDefinitionList {
	var result ast.ArgumentDefinitionList
	for _, inArg := range in {
		arg := &ast.ArgumentDefinition{
			Name:        inArg.Name,
			Description: inArg.Description,
			Type:        makeType(inArg.Type),
		}

		if inArg.DefaultValue != nil {
			arg.DefaultValue = makeDefaultValue(inArg.Type, *inArg.DefaultValue)
		}

		result = append(result, arg)
	}
	return result
}

func makeDefaultValue(t *Type, raw string) *ast.Value {
	switch t.Kind {
	case ScalarKind:
		switch t.Name {
		case "Int":
			return &ast.Value{
				Kind: ast.IntValue,
				Raw:  raw,
			}
		case "Float":
			return &ast.Value{
				Kind: ast.FloatValue,
				Raw:  raw,
			}
		case "Boolean":
			return &ast.Value{
				Kind: ast.BooleanValue,
				Raw:  raw,
			}
		case "String":
			return &ast.Value{
				Kind: ast.StringValue,
				Raw:  raw,
			}
		default:
			return &ast.Value{
				Kind: ast.StringValue,
				Raw:  raw,
			}
		}
	case InputObjectKind:
		return &ast.Value{
			Kind: ast.ObjectValue,
			Raw:  raw,
		}
	case ListKind:
		return &ast.Value{
			Kind: ast.ListValue,
			Raw:  raw,
		}
	case EnumKind:
		return &ast.Value{
			Kind: ast.EnumValue,
			Raw:  raw,
		}
	case NonNullKind:
		return makeDefaultValue(t.OfType, raw)
	default:
		panic(fmt.Errorf("unsupported type kind %s for default value '%s'", t.Kind, raw))
	}
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
