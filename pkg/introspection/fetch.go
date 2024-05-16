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

type Client struct {
	endpoint string
	headers  http.Header
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

func responseToAst(s *Schema) (*ast.Schema, error) {
	var defs ast.DefinitionList

	for _, def := range s.Types {
		newDef := ast.Definition{
			Kind:        ast.DefinitionKind(def.Kind),
			Description: def.Description,
			Name:        def.Name,
		}

		if def.Kind == ObjectKind {
			var interfaces []string
			for _, iface := range def.Interfaces {
				interfaces = append(interfaces, iface.Name)
			}
			newDef.Interfaces = interfaces
		}

		if def.Kind == ObjectKind || def.Kind == InputObjectKind || def.Kind == InterfaceKind {
			var fields ast.FieldList
			for _, inField := range def.Fields {
				field := ast.FieldDefinition{
					Name:        inField.Name,
					Description: inField.Description,
					Type:        makeType(inField.Type),
					Arguments:   makeArgumentDefinitionList(inField.Args),
				}
				fields = append(fields, &field)
			}
			for _, inField := range def.InputFields {
				field := ast.FieldDefinition{
					Name:        inField.Name,
					Description: inField.Description,
					Type:        makeType(inField.Type),
				}

				if inField.DefaultValue != nil {
					field.DefaultValue = makeDefaultValue(inField.Type, *inField.DefaultValue)
				}

				fields = append(fields, &field)
			}
			newDef.Fields = fields
		}

		if def.Kind == UnionKind {
			var possibleTypes []string
			for _, pt := range def.PossibleTypes {
				possibleTypes = append(possibleTypes, pt.Name)
			}
			newDef.Types = possibleTypes
		}

		if def.Kind == EnumKind {
			var evs ast.EnumValueList
			for _, ev := range def.EnumValues {
				evs = append(evs, &ast.EnumValueDefinition{
					Name:        ev.Name,
					Description: ev.Description,
					// TODO: synthesize deprecated directive
				})
			}
			newDef.EnumValues = evs
		}

		defs = append(defs, &newDef)
	}

	typeMap := map[string]*ast.Definition{}
	for _, def := range defs {
		typeMap[def.Name] = def
	}

	return &ast.Schema{
		Types:        typeMap,
		Query:        typeMap[s.QueryType.Name],
		Mutation:     typeMap[s.MutationType.Name],
		Subscription: typeMap[s.SubscriptionType.Name],
		// TODO: directives
	}, nil
}

func makeType(in *Type) *ast.Type {
	if in.Kind == NonNullKind {
		wrappedType := makeType(in.OfType)
		wrappedType.NonNull = true
		return wrappedType
	}

	if in.Kind == ListKind {
		wrappedType := makeType(in.OfType)
		return &ast.Type{
			Elem: wrappedType,
		}
	}

	return &ast.Type{
		NamedType: in.Name,
	}
}

func makeArgumentDefinitionList(in []InputValue) ast.ArgumentDefinitionList {
	var result ast.ArgumentDefinitionList
	for _, inArg := range in {
		arg := &ast.ArgumentDefinition{
			Name:        inArg.Name,
			Description: inArg.Description,
			Type:        makeType(inArg.Type),
		}

		if inArg.DefaultValue != nil {
			arg.DefaultValue = makeDefaultValue(inArg.Type, *inArg.DefaultValue)
		}

		result = append(result, arg)
	}
	return result
}

func makeDefaultValue(t *Type, raw string) *ast.Value {
	switch t.Kind {
	case ScalarKind:
		switch t.Name {
		case "Int":
			return &ast.Value{
				Kind: ast.IntValue,
				Raw:  raw,
			}
		case "Float":
			return &ast.Value{
				Kind: ast.FloatValue,
				Raw:  raw,
			}
		case "Boolean":
			return &ast.Value{
				Kind: ast.BooleanValue,
				Raw:  raw,
			}
		case "String":
			return &ast.Value{
				Kind: ast.StringValue,
				Raw:  raw,
			}
		default:
			return &ast.Value{
				Kind: ast.StringValue,
				Raw:  raw,
			}
		}
	case InputObjectKind:
		return &ast.Value{
			Kind: ast.ObjectValue,
			Raw:  raw,
		}
	case ListKind:
		return &ast.Value{
			Kind: ast.ListValue,
			Raw:  raw,
		}
	case EnumKind:
		return &ast.Value{
			Kind: ast.EnumValue,
			Raw:  raw,
		}
	case NonNullKind:
		return makeDefaultValue(t.OfType, raw)
	default:
		panic(fmt.Errorf("unsupported type kind %s for default value '%s'", t.Kind, raw))
	}

}

func (c Client) fetchSchema() (*Schema, error) {
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

type GraphQLParams struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

func (c Client) issueQuery(query string, vars map[string]any, operation string) (*graphQLResponse, error) {
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
