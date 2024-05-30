package commands

import (
	"github.com/benweint/gquil/pkg/graph"
)

type VizCmd struct {
	InputOptions
	FilteringOptions
	GraphFilteringOptions
	InterfacesAsUnions bool `name:"interfaces-as-unions" help:"Treat interfaces as unions rather than objects for the purposes of graph construction."`
}

func (c *VizCmd) Run(ctx Context) error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	var opts []graph.GraphOption
	if c.InterfacesAsUnions {
		opts = append(opts, graph.WithInterfacesAsUnions())
	}

	if c.IncludeBuiltins {
		opts = append(opts, graph.WithBuiltins(true))
	}

	g := graph.MakeGraph(s, opts...)

	if len(c.From) > 0 {
		roots, err := s.ResolveNames(c.From)
		if err != nil {
			return err
		}
		g = g.ReachableFrom(roots, c.Depth)
	}

	ctx.Print(g.ToDot())

	return nil
}
