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

func (c *VizCmd) Help() string {
	return `To render the resulting graph to a PDF, you can use the 'dot' tool that comes with GraphViz:

  gquil viz schema.graphql | dot -Tpdf >out.pdf

For GraphQL schemas with a large number of types and fields, the resulting diagram may be very large. You can trim it down to a particular region of interest using the --from and --depth flags to indicate a starting point and maximum traversal depth to use when traversing the graph.

  gquil viz --from Reviews --depth 2 schema.graphql | dot -Tpdf >out.pdf

GraphQL unions are represented as nodes in the graph with outbound edges to each member type. Interfaces are represented in the same way as object types by default, with one outbound edge per field, pointing to the type of that field. To instead render interfaces with one outbound edge per implementing type, you can use the --interfaces-as-unions flag.`
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
