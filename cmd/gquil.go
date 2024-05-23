package main

import (
	"github.com/alecthomas/kong"
	"github.com/benweint/gquil/pkg/commands"
)

func main() {
	var cli commands.CLI
	ctx := kong.Parse(&cli,
		kong.Name("gquil"),
		kong.Description("Interrogate and visualize GraphQL schemas."),
		kong.UsageOnError(),
		kong.ExplicitGroups(commands.Groups),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
