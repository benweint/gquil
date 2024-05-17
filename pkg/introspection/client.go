package introspection

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/benweint/gquilt/pkg/model"
	"github.com/vektah/gqlparser/v2/ast"
)

type Client struct {
	endpoint string
	headers  http.Header
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

func NewClient(endpoint string, headers http.Header) *Client {
	mergedHeaders := http.Header{
		"content-type": []string{
			"application/json",
		},
	}

	for key, vals := range headers {
		mergedHeaders[key] = vals
	}

	return &Client{
		endpoint: endpoint,
		headers:  mergedHeaders,
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
	rsp, err := c.issueQuery(Query, nil, "IntrospectionQuery")
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send introspection request to %s: %w", c.endpoint, err)
	}
	defer resp.Body.Close()

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
