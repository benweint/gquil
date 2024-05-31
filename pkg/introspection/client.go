package introspection

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// Client is a client capable of issuing an introspection query against a GraphQL server over HTTP,
// and transforming the response into either an *ast.Schema.
type Client struct {
	endpoint    string
	headers     http.Header
	specVersion SpecVersion
	traceOut    io.Writer
}

type GraphQLParams struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

type graphQLError struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
	Path []string `json:"path"`
}

// NewClient returns a new GraphQL introspection client.
// HTTP requests issued by this client will use the given HTTP headers, in addition to some defaults.
// The given SpecVersion will be used to ensure that the introspection query issued by the client is
// compatible with a specific version of the GraphQL spec.
// If traceOut is non-nil, the outbound request and returned response will be dumped to it for debugging
// purposes.
func NewClient(endpoint string, headers http.Header, specVersion SpecVersion, traceOut io.Writer) *Client {
	mergedHeaders := http.Header{
		"content-type": []string{
			"application/json",
		},
	}

	for key, vals := range headers {
		mergedHeaders[key] = vals
	}

	return &Client{
		endpoint:    endpoint,
		headers:     mergedHeaders,
		specVersion: specVersion,
		traceOut:    traceOut,
	}
}

func (c *Client) FetchSchemaAst() (*ast.Schema, error) {
	rawSchema, err := c.fetchSchema()
	if err != nil {
		return nil, err
	}

	return responseToAst(rawSchema)
}

func (c *Client) fetchSchema() (*Schema, error) {
	rsp, err := c.issueQuery(GetQuery(c.specVersion), nil, "IntrospectionQuery")
	if err != nil {
		return nil, err
	}

	if len(rsp.Errors) != 0 {
		var errs []error
		for _, err := range rsp.Errors {
			forPath := ""
			if len(err.Path) != 0 {
				forPath = fmt.Sprintf(" at path %s", strings.Join(err.Path, "."))
			}
			errs = append(errs, fmt.Errorf("error executing introspection query%s: %s", forPath, err.Message))
		}
		return nil, errors.Join(errs...)
	}

	var parsed IntrospectionQueryResult
	err = json.Unmarshal(rsp.Data, &parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize introspection query result: %w", err)
	}

	return &parsed.Schema, nil
}

func (c *Client) issueQuery(query string, vars map[string]any, operation string) (*graphQLResponse, error) {
	body := GraphQLParams{
		Query:         query,
		OperationName: operation,
		Variables:     vars,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize introspection request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header = c.headers

	if c.traceOut != nil {
		requestDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, fmt.Errorf("failed to dump introspection HTTP request: %w", err)
		}
		fmt.Fprintf(c.traceOut, "---\nIntrospection request:\n%s\n", string(requestDump))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send introspection request to %s: %w", c.endpoint, err)
	}
	defer resp.Body.Close()

	if c.traceOut != nil {
		rspDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, fmt.Errorf("failed to dump introspection HTTP response: %w", err)
		}
		fmt.Fprintf(c.traceOut, "\n---\nIntrospection response:\n%s\n", string(rspDump))
	}

	rspBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read introspection query response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response to introspection query: status=%d, body=%s", resp.StatusCode, rspBody)
	}

	var graphqlResp graphQLResponse
	err = json.Unmarshal(rspBody, &graphqlResp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize introspection response body: %w", err)
	}

	return &graphqlResp, nil
}
