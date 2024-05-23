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

	"github.com/benweint/gquil/pkg/model"
	"github.com/vektah/gqlparser/v2/ast"
)

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

func (c *Client) FetchSchemaModel() (*model.Schema, error) {
	schemaAst, err := c.FetchSchemaAst()
	if err != nil {
		return nil, err
	}
	return model.MakeSchema(schemaAst)
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
		return nil, err
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

	for k, vs := range c.headers {
		for _, v := range vs {
			req.Header.Set(k, v)
		}
	}

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
