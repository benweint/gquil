package commands

import (
	"slices"
	"strings"

	"github.com/benweint/gquil/pkg/model"
	"github.com/vektah/gqlparser/v2/ast"
)

type LsTypesCmd struct {
	InputOptions
	Kind       ast.DefinitionKind `name:"kind" group:"filtering" help:"Only list types of the given kind (interface, object, union, input_object, enum, scalar)."`
	MemberOf   string             `name:"member-of" group:"filtering" help:"Only list types which are members of the given union."`
	Implements string             `name:"implements" group:"filtering" help:"Only list types which implement the given interface."`
	IncludeDirectivesOption
	FilteringOptions
	OutputOptions
	GraphFilteringOptions
}

func (c LsTypesCmd) Help() string {
	return `Types include objects types, interfaces, unions, enums, input objects, and scalars. The default output format prepends each listed type with its kind. You can filter to a specific kind using --kind, which will cause the kind to be omitted in the output. For example:

  gquil ls types --kind interface examples/github.graphql

You can also filter types based on their membership in a union type (--member-of), or based on whether they implement a specified interface (--implements). You can also filter by graph reachability using the --from and --depth options, see the help for these flags for details.

Directives are not included in the output by default, but can be added with --include-directives. You can also use --json for a JSON output format. The JSON output format matches the one used by the json subcommand.
`
}

func (c LsTypesCmd) Run(ctx Context) error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	if err = c.filterSchema(s); err != nil {
		return err
	}

	if !c.IncludeBuiltins {
		s.FilterBuiltins()
	}

	var memberTypes []string
	if c.MemberOf != "" {
		for _, t := range s.Types {
			if t.Name == c.MemberOf {
				memberTypes = t.PossibleTypes
			}
		}
	}

	var types model.DefinitionList
	normalizedKind := ast.DefinitionKind(strings.ToUpper(string(c.Kind)))
	for _, t := range s.Types {
		if normalizedKind != "" && normalizedKind != t.Kind {
			continue
		}

		if c.MemberOf != "" && !slices.Contains(memberTypes, t.Name) {
			continue
		}

		if c.Implements != "" && !slices.Contains(t.Interfaces, c.Implements) {
			continue
		}

		types = append(types, t)
	}
	types.Sort()

	if c.Json {
		return ctx.PrintJson(types)
	} else {
		for _, t := range types {
			directives := ""
			if c.IncludeDirectives {
				directives, err = formatDirectiveList(t.Directives)
				if err != nil {
					return err
				}
			}

			if c.Kind != "" {
				ctx.Printf("%s%s\n", t.Name, directives)
			} else {
				ctx.Printf("%s %s%s\n", t.Kind, t.Name, directives)
			}
		}
	}

	return nil
}
