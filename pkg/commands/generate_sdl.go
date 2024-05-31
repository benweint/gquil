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
	Headers  []string `name:"header" short:"H" help:"Set custom headers on the introspection request, e.g. for authentication. Format: <key>: <value>. May be specified multiple times. Header values may be read from a file with the syntax @<filename>, e.g. --header @my-headers.txt."`
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

Note that since GraphQL's introspection schema does not expose information about the application sites of most directives, the generated SDL will lack any applied directives (with the exception of @deprecated, which is exposed via the introspection system).

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

	headers, err := parseHeaders(c.Headers)
	if err != nil {
		return fmt.Errorf("failed to parse custom header: %w", err)
	}

	client := introspection.NewClient(c.Endpoint, headers, sv, traceOut)
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

func parseHeaders(raw []string) (http.Header, error) {
	result := http.Header{}
	for _, rawHeader := range raw {
		parsedHeaders, err := parseHeaderValue(rawHeader)
		if err != nil {
			return nil, err
		}
		for key, vals := range parsedHeaders {
			for _, val := range vals {
				result.Add(key, val)
			}
		}
	}
	return result, nil
}

func parseHeaderValue(raw string) (http.Header, error) {
	if strings.HasPrefix(raw, "@") {
		return parseHeadersFromFile(strings.TrimPrefix(raw, "@"))
	}

	key, val, err := parseHeaderString(raw)
	if err != nil {
		return nil, err
	}
	return http.Header{
		key: []string{val},
	}, nil
}

func parseHeadersFromFile(path string) (http.Header, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result := http.Header{}
	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		key, val, err := parseHeaderString(line)
		if err != nil {
			return nil, err
		}
		result.Add(key, val)
	}
	return result, nil
}

func parseHeaderString(raw string) (string, string, error) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid header value '%s', expected format '<key>: <value>'", raw)
	}
	key := parts[0]
	value := strings.TrimLeft(parts[1], " ")
	return key, value, nil
}
