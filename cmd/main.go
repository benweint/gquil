package main

import (
	"github.com/alecthomas/kong"
	"github.com/benweint/gquilt/pkg/commands"
)

var cli struct {
	Ls            commands.LsCmd            `cmd:"" help:"List types from the given GraphQL SDL file."`
	Json          commands.JsonCmd          `cmd:"" help:"Return a JSON representation of the given GraphQL SDL file."`
	Introspection commands.IntrospectionCmd `cmd:"" help:"Interact with a GraphQL introspection endpoint over HTTP."`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
