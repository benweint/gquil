package commands

import (
	"fmt"

	"github.com/benweint/gquil/pkg/graph"
)

type VizCmd struct {
	// TODO: should not include json flag
	InputOptions
	GraphFilteringOptions
	InterfacesAsUnions bool `name:"interfaces-as-unions" help:"Treat interfaces as unions rather than objects for the purposes of graph construction."`
}

func (c *VizCmd) Run() error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	var opts []graph.GraphOption
	if c.InterfacesAsUnions {
		opts = append(opts, graph.WithInterfacesAsUnions())
	}

	g := graph.MakeGraph(s.Types, opts...)

	if len(c.From) > 0 {
		g = g.ReachableFrom(c.From, c.Depth)
	}

	fmt.Print(g.ToDot())

	return nil
}
