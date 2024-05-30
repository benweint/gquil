package commands

import (
	"github.com/alecthomas/kong"
	"github.com/benweint/gquil/pkg/graph"
	"github.com/benweint/gquil/pkg/model"
)

var Groups = []kong.Group{
	{
		Key:   "filtering",
		Title: "Filtering options",
	},
	{
		Key:   "output",
		Title: "Output formatting options",
	},
}

type InputOptions struct {
	// TODO: stdin support
	SchemaFiles []string `arg:"" name:"schemas" help:"Path to the GraphQL SDL schema file(s) to read from."`
}

type FilteringOptions struct {
	IncludeBuiltins bool `name:"include-builtins" group:"filtering" help:"Include built-in types and directives in output (omitted by default)."`
}

type IncludeDirectivesOption struct {
	IncludeDirectives bool `name:"include-directives" group:"output" help:"Include applied directives in human-readable output. Has no effect with --json."`
}

type OutputOptions struct {
	Json bool `name:"json" group:"output" help:"Output results as JSON."`
}

type GraphFilteringOptions struct {
	From  []string `name:"from" group:"filtering" help:"Only include types reachable from the specified type(s) or field(s). May be specified multiple times to use multiple roots."`
	Depth int      `name:"depth" group:"filtering" help:"When used with --from, limit the depth of traversal."`
}

func (o GraphFilteringOptions) filterSchema(s *model.Schema) error {
	if len(o.From) == 0 {
		return nil
	}

	roots, err := s.ResolveNames(o.From)
	if err != nil {
		return err
	}

	g := graph.MakeGraph(s).ReachableFrom(roots, o.Depth)
	s.Types = g.GetDefinitions()
	return nil
}
