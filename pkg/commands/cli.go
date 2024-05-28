package commands

import (
	"os"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Ls            LsCmd            `cmd:"" aliases:"list" help:"List types, fields, or directives in a GraphQL SDL document."`
	Json          JsonCmd          `cmd:"" help:"Return a JSON representation of a GraphQL SDL document."`
	Introspection IntrospectionCmd `cmd:"" help:"Interact with a GraphQL introspection endpoint over HTTP."`
	Viz           VizCmd           `cmd:"" help:"Visualize a GraphQL schema using GraphViz."`
	Merge         MergeCmd         `cmd:"" help:"Merge multiple GraphQL SDL documents into a single one."`
}

func MakeParser(opts ...kong.Option) (*kong.Kong, error) {
	var cli CLI

	defaultOptions := []kong.Option{
		kong.Name("gquil"),
		kong.Description("Inspect, visualize, and transform GraphQL schemas."),
		kong.UsageOnError(),
		kong.ExplicitGroups(Groups),
	}

	return kong.New(&cli, append(defaultOptions, opts...)...)
}

func Main() int {
	parser, err := MakeParser()
	if err != nil {
		panic(err)
	}

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	err = ctx.Run(Context{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	})
	ctx.FatalIfErrorf(err)
	return 0
}
