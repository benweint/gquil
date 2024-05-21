package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/benweint/gquil/pkg/model"
)

type LsDirectivesCmd struct {
	InputOptions
	FilteringOptions
	OutputOptions
}

func (c LsDirectivesCmd) Run() error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	if !c.IncludeBuiltins {
		s.FilterBuiltins()
	}

	if c.Json {
		j, err := json.Marshal(s.Directives)
		if err != nil {
			return err
		}
		fmt.Printf(string(j) + "\n")
	} else {
		for _, directive := range s.Directives {
			fmt.Printf("%s\n", formatDirectiveDefinition(directive))
		}
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
