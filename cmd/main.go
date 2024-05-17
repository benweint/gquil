package main

import (
	"github.com/alecthomas/kong"
	"github.com/benweint/gquilt/pkg/commands"
)

var cli struct {
	Ls            commands.LsCmd            `cmd:"" help:"List types, fields, or directives in an SDL document."`
	Merge         commands.MergeCmd         `cmd:"" help:"Merge multiple SDL documents into a single one."`
	Json          commands.JsonCmd          `cmd:"" help:"Return a JSON representation of an SDL document."`
	Introspection commands.IntrospectionCmd `cmd:"" help:"Interact with a GraphQL introspection endpoint over HTTP."`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
