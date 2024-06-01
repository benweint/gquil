package commands

import (
	"fmt"
	"strings"

	"github.com/benweint/gquil/pkg/model"
)

type LsDirectivesCmd struct {
	InputOptions
	FilteringOptions
	OutputOptions
}

func (c LsDirectivesCmd) Help() string {
	return `List all directive definitions in the given GraphQL SDL file(s).

By default, directives are emitted with their argument definitions and valid application locations, in a format that mirrors the SDL for defining them. You can emit JSON representations of them instead with the --json flag.`
}

func (c LsDirectivesCmd) Run(ctx Context) error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	if !c.IncludeBuiltins {
		s.FilterBuiltins()
	}

	if c.Json {
		return ctx.PrintJson(s.Directives)
	}

	for _, directive := range s.Directives {
		ctx.Printf("%s\n", formatDirectiveDefinition(directive))
	}

	return nil
}

func formatDirectiveDefinition(d *model.DirectiveDefinition) string {
	var locations []string
	for _, kind := range d.Locations {
		locations = append(locations, string(kind))
	}
	return fmt.Sprintf("@%s%s on %s", d.Name, formatArgumentDefinitionList(d.Arguments), strings.Join(locations, " | "))
}
