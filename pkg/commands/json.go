package commands

import (
	"encoding/json"
	"fmt"

	"github.com/benweint/gquilt/pkg/graph"
)

type JsonCmd struct {
	InputOptions
	FilteringOptions
	GraphFilteringOptions
}

func (c *JsonCmd) Run() error {
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

	out, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to serialize schema to JSON: %w", err)
	}

	fmt.Print(string(out))

	return nil
}
