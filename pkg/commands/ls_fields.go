package commands

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/benweint/gquilt/pkg/graph"
	"github.com/benweint/gquilt/pkg/model"
)

type LsFieldsCmd struct {
	InputOptions
	OnType      string `name:"on-type" group:"filtering" help:"Only include fields which appear on the specified type."`
	OfType      string `name:"of-type" group:"filtering" help:"Only include fields of the specified type. List and non-null types will be treated as being of their underlying wrapped type for the purposes of this filtering."` // TODO help
	IncludeArgs bool   `name:"include-args" group:"output" help:"Include argument definitions in human-readable output. Has no effect with --json."`
	IncludeDirectivesOption
	OutputOptions
	FilteringOptions
	GraphFilteringOptions
}

func (c LsFieldsCmd) Run() error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	if len(c.From) > 0 {
		g := graph.MakeGraph(s.Types).ReachableFrom(c.From, c.Depth)
		s.Types = g.GetDefinitions()
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
			f.Name = t.Name + "." + f.Name
			fields = append(fields, f)
		}
	}

	slices.SortFunc(fields, func(a, b *model.FieldDefinition) int {
		return strings.Compare(a.Name, b.Name)
	})

	if c.Json {
		j, err := json.Marshal(fields)
		if err != nil {
			return err
		}
		fmt.Print(string(j) + "\n")
	} else {
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
			fmt.Printf("%s%s: %s%s\n", f.Name, args, f.Type, directives)
		}
	}

	return nil
}
