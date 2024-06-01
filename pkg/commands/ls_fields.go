package commands

import (
	"slices"

	"github.com/benweint/gquil/pkg/model"
)

type LsFieldsCmd struct {
	InputOptions
	OnType        string `name:"on-type" group:"filtering" help:"Only include fields which appear on the specified type."`
	OfType        string `name:"of-type" group:"filtering" help:"Only include fields of the specified type. List and non-null types will be treated as being of their underlying wrapped type for the purposes of this filtering."`
	ReturningType string `name:"returning-type" group:"filtering" help:"Only include fields which may return the specified type. Interface or union-typed fields may possibly return their implementing or member types. List and non-null fields are unwrapped for the purposes of this filtering."`
	Named         string `name:"named" group:"filtering" help:"Only include fields with the given name (matched against the field name only, not including type name)."`
	IncludeArgs   bool   `name:"include-args" group:"output" help:"Include argument definitions in human-readable output. Has no effect with --json."`
	IncludeDirectivesOption
	OutputOptions
	FilteringOptions
	GraphFilteringOptions
}

func (c LsFieldsCmd) Help() string {
	return `Fields are identified as <type>.<fieldname>, where <type> is the host type on which they are defined, and are emitted in sorted order by these identifiers.

You can use the --on-type, --of-type, --returning-type, and --named arguments to filter the set of returned fields. You can also filter by graph reachability using the --from and --depth options, see the help for these flags for details.

Field arguments and directives are not included in the output by default (only names and types), but can be added with --include-args and --include-directives, respectivesly. You can also use --json for a JSON output format. The JSON output format matches the one used by the json subcommand, with the exception that field names will include the host type as a prefix (e.g. 'Query.search' instead of just 'search').`
}

func (c LsFieldsCmd) Run(ctx Context) error {
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

	var fields model.FieldDefinitionList
	for _, t := range s.Types {
		if c.OnType != "" && c.OnType != t.Name {
			continue
		}
		for _, f := range t.Fields {
			if c.OfType != "" && c.OfType != f.Type.Unwrap().String() {
				continue
			}
			if c.ReturningType != "" && !fieldMightReturn(s, f, c.ReturningType) {
				continue
			}
			if c.Named != "" && c.Named != f.Name {
				continue
			}
			f.Name = t.Name + "." + f.Name
			fields = append(fields, f)
		}
	}
	fields.Sort()

	if c.Json {
		return ctx.PrintJson(fields)
	}

	for _, f := range fields {
		args := ""
		if c.IncludeArgs {
			args = formatArgumentDefinitionList(f.Arguments)
		}
		directives := ""
		if c.IncludeDirectives {
			directives, err = formatDirectiveList(f.Directives)
			if err != nil {
				return err
			}
		}
		ctx.Printf("%s%s: %s%s\n", f.Name, args, f.Type, directives)
	}

	return nil
}

func fieldMightReturn(s *model.Schema, field *model.FieldDefinition, typeName string) bool {
	underlyingType := field.Type.Unwrap()
	if underlyingType.Name == typeName {
		return true
	}

	referencedType := s.Types[field.Type.Unwrap().Name]
	if referencedType != nil && slices.Contains(referencedType.PossibleTypes, typeName) {
		return true
	}

	return false
}
