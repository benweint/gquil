package commands

import (
	"fmt"

	"github.com/benweint/gquilt/pkg/graph"
)

type VizCmd struct {
	// TODO: should not include json flag
	CommonOptions
	From  []string `name:"from"`
	Depth int      `name:"depth" default:"0"`
}

func (c *VizCmd) Run() error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	g := graph.MakeGraph(s.Types)

	if len(c.From) > 0 {
		g = g.ReachableFrom(c.From, c.Depth)
	}

	fmt.Print(g.ToDot())

	return nil
}
