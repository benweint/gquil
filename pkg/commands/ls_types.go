package commands

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/benweint/gquil/pkg/graph"
	"github.com/benweint/gquil/pkg/model"
	"github.com/vektah/gqlparser/v2/ast"
)

type ObjectFilteringOptions struct {
}

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

func (c LsTypesCmd) Run() error {
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

	slices.SortFunc(types, func(a, b *model.Definition) int {
		return strings.Compare(a.Name, b.Name)
	})

	if c.Json {
		j, err := json.Marshal(types)
		if err != nil {
			return err
		}
		fmt.Print(string(j) + "\n")
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
				fmt.Printf("%s%s\n", t.Name, directives)
			} else {
				fmt.Printf("%s\t%s%s\n", t.Kind, t.Name, directives)
			}
		}
	}

	return nil
}
