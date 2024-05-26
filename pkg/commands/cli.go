package commands

import "github.com/alecthomas/kong"

type CLI struct {
	Ls            LsCmd            `cmd:"" aliases:"list" help:"List types, fields, or directives in a GraphQL SDL document."`
	Json          JsonCmd          `cmd:"" help:"Return a JSON representation of a GraphQL SDL document."`
	Introspection IntrospectionCmd `cmd:"" help:"Interact with a GraphQL introspection endpoint over HTTP."`
	Viz           VizCmd           `cmd:"" help:"Visualize a GraphQL schema using GraphViz."`
	Merge         MergeCmd         `cmd:"" help:"Merge multiple GraphQL SDL documents into a single one."`
}

func Main() int {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("gquil"),
		kong.Description("Inspect, visualize, and transform GraphQL schemas."),
		kong.UsageOnError(),
		kong.ExplicitGroups(Groups),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
	return 0
}
