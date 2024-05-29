package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/benweint/gquil/pkg/astutil"
	"github.com/benweint/gquil/pkg/introspection"
	"github.com/benweint/gquil/pkg/model"
	"github.com/vektah/gqlparser/v2/formatter"
)

type GenerateSDLCmd struct {
	Endpoint string   `arg:"" help:"The GraphQL introspection endpoint URL to fetch from."`
	Headers  []string `name:"header" short:"H" help:"Set headers on the introspection request. Format: <key>: <value>."`
	Trace    bool     `name:"trace" help:"Dump the introspection HTTP request and response to stderr for debugging."`

	OutputOptions
	SpecVersionOptions
	FilteringOptions
}

func (c *GenerateSDLCmd) Help() string {
	return `Issues a GraphQL introspection query via an HTTP POST request to the specified endpoint, and uses the response to generate a GraphQL SDL document, which is emitted to stdout.

Example:

  gquil introspection generate-sdl \
    --header 'origin: https://docs.developer.yelp.com' \
    https://api.yelp.com/v3/graphql

Note that since GraphQL's introspection schema does not expose information about the application sites of most directives, the generated SDL will lack any applied directives (with the exception of @deprecated, which is exposed via the introspection system)

If your GraphQL endpoint requires authentication or other special headers, you can set custom headers on the issued request using the --header flag.`
}

func (c *GenerateSDLCmd) Run(ctx Context) error {
	sv, err := introspection.ParseSpecVersion(c.SpecVersion)
	if err != nil {
		return err
	}

	var traceOut io.Writer
	if c.Trace {
		traceOut = os.Stderr
	}

	client := introspection.NewClient(c.Endpoint, parseHeaders(c.Headers), sv, traceOut)
	s, err := client.FetchSchemaAst()
	if err != nil {
		return err
	}

	if c.Json {
		m, err := model.MakeSchema(s)
		if err != nil {
			return fmt.Errorf("failed to construct model from introspection schema AST: %w", err)
		}

		if !c.IncludeBuiltins {
			m.FilterBuiltins()
		}

		return ctx.PrintJson(m)
	} else {
		if !c.IncludeBuiltins {
			astutil.FilterBuiltins(s)
		}

		f := formatter.NewFormatter(os.Stdout)
		f.FormatSchema(s)
	}

	return nil
}

func parseHeaders(raw []string) http.Header {
	result := http.Header{}
	for _, rawHeader := range raw {
		parts := strings.SplitN(rawHeader, ":", 2)
		key := parts[0]
		value := strings.TrimLeft(parts[1], " ")
		result[key] = append(result[key], value)
	}
	return result
}
