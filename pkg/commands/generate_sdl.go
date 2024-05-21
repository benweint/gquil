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
	Endpoint string   `arg:"" help:"The GraphQL introspection endpoint URL to fetch from."`
	Headers  []string `name:"header" short:"H" help:"Set headers on the introspection request."`
	FilteringOptions
}

func (c *GenerateSDLCmd) Help() string {
	return `Issues a GraphQL introspection query using an HTTP POST request to the specified GraphQL endpoint, and uses the response to generate a GraphQL SDL document, which is emitted to stdout.
Note that since GraphQL's introspection schema does not expose information about the application sites of most directives, the generated SDL will lack any applied directives (with the exception of @deprecated, which is exposed via the introspection system).

If your GraphQL endpoint requires authentication, you can set custom headers on the issued request using the --headers flag.`
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
