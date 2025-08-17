package commands

import "github.com/alecthomas/kong"

// version string will be injected by automation
// see .goreleaser.yaml
var version string = "unknown"

type versionFlag bool

func (f versionFlag) BeforeReset(ctx *kong.Context) error {
	_, _ = ctx.Stdout.Write([]byte(version + "\n"))
	return nil
}

type VersionCmd struct {
}

func (c VersionCmd) Run(ctx Context) error {
	ctx.Print(version + "\n")
	return nil
}
