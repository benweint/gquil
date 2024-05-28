package commands

import (
	"github.com/benweint/gquil/pkg/astutil"
	"github.com/vektah/gqlparser/v2/formatter"
)

type MergeCmd struct {
	InputOptions
	FilteringOptions
}

func (c *MergeCmd) Run(ctx Context) error {
	s, err := parseSchemaFromPaths(c.SchemaFiles)
	if err != nil {
		return err
	}

	if !c.IncludeBuiltins {
		astutil.FilterBuiltins(s)
	}

	f := formatter.NewFormatter(ctx.Stdout)
	f.FormatSchema(s)
	return nil
}
