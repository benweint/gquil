package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/benweint/gquilt/pkg/model"
)

type LsDirectivesCmd struct {
	CommonOptions
	IncludeArgs bool `name:"include-args"`
}

func (c LsDirectivesCmd) Run() error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	if !c.IncludeBuiltins {
		s.FilterBuiltins()
	}

	var results model.DirectiveDefinitionList

	for _, directive := range s.Directives {
		results = append(results, directive)
	}

	if c.Json {
		j, err := json.Marshal(results)
		if err != nil {
			return err
		}
		fmt.Printf(string(j) + "\n")
	} else {
		for _, directive := range results {
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
