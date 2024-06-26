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
	VersionFlag   versionFlag      `hidden:"" help:"Print version and exit."`
	Version       VersionCmd       `cmd:"" help:"Print the version of gquil and exit."`
}

const description = `Inspect, visualize, and transform GraphQL schemas.

For more documentation, or to report an issue:
https://github.com/benweint/gquil
`

func MakeParser(opts ...kong.Option) (*kong.Kong, error) {
	var cli CLI

	defaultOptions := []kong.Option{
		kong.Name("gquil"),
		kong.Description(description),
		kong.UsageOnError(),
		kong.ExplicitGroups(Groups),
	}

	return kong.New(&cli, append(defaultOptions, opts...)...)
}

// massageArgs munges the input args in order to translate:
// - `gquil`            -> `gquil --help`
// - `gquil help`       -> `quil --help`
// - `gquil help <cmd>` -> `gquil <cmd> --help`
func massageArgs(args []string) []string {
	args = args[1:]

	if len(args) == 0 {
		return []string{"--help"}
	}

	if args[0] == "help" {
		if len(args) == 1 {
			return []string{"--help"}
		}

		return append(args[1:], "--help")
	}

	return args
}

func Main() int {
	parser, err := MakeParser()
	if err != nil {
		panic(err)
	}

	args := massageArgs(os.Args)
	ctx, err := parser.Parse(args)
	parser.FatalIfErrorf(err)
	err = ctx.Run(Context{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	})
	ctx.FatalIfErrorf(err)
	return 0
}
