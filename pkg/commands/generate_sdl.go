package commands

import (
	"net/http"
	"os"
	"strings"

	"github.com/benweint/gquilt/pkg/astutil"
	"github.com/benweint/gquilt/pkg/introspection"
	"github.com/vektah/gqlparser/v2/formatter"
)

type GenerateSDLCmd struct {
	Endpoint        string   `arg:"" help:"The GraphQL introspection endpoint URL to fetch from."`
	Headers         []string `name:"header" short:"H" help:"Set headers on the introspection request."`
	IncludeBuiltins bool     `name:"include-builtins" help:"Include built-in definitions."`
}

func (c *GenerateSDLCmd) Run() error {
	client := introspection.NewClient(c.Endpoint, parseHeaders(c.Headers))
	s, err := client.FetchSchemaAst()
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

func parseHeaders(raw []string) http.Header {
	result := http.Header{}
	for _, rawHeader := range raw {
		parts := strings.SplitN(rawHeader, ":", 2)
		key := parts[0]
		value := parts[1]
		result[key] = append(result[key], value)
	}
	return result
}
