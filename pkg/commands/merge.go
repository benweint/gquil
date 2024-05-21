package commands

import (
	"os"

	"github.com/benweint/gquilt/pkg/astutil"
	"github.com/vektah/gqlparser/v2/formatter"
)

type MergeCmd struct {
	InputOptions
	FilteringOptions
}

func (c *MergeCmd) Run() error {
	s, err := parseSchemaFromPaths(c.SchemaFiles)
	if err != nil {
		return err
	}

	if !c.IncludeBuiltins {
		astutil.FilterBuiltins(s)
	}

	f := formatter.NewFormatter(os.Stdout)
	f.FormatSchema(s)
	return nil
}
