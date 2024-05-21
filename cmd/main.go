package main

import (
	"github.com/alecthomas/kong"
	"github.com/benweint/gquilt/pkg/commands"
)

func main() {
	var cli commands.CLI
	ctx := kong.Parse(&cli,
		kong.Name("gquilt"),
		kong.Description("Interrogate and visualize GraphQL schemas."),
		kong.UsageOnError(),
		kong.ExplicitGroups(commands.Groups),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
