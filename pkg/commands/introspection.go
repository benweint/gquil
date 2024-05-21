package commands

import (
	"encoding/json"
	"fmt"

	"github.com/benweint/gquilt/pkg/introspection"
)

type IntrospectionCmd struct {
	GenerateSDL GenerateSDLCmd `cmd:"" help:"Generate GraphQL SDL from a GraphQL introspection endpoint over HTTP(S)."`
	Query       EmitQueryCmd   `cmd:"" help:"Emit the GraphQL query used for introspection."`
}

type EmitQueryCmd struct {
	Json bool `name:"json" help:"Emit the query as JSON, suitable for passing to curl or similar."`
}

func (c *EmitQueryCmd) Run() error {
	if c.Json {
		q := introspection.GraphQLParams{
			Query:         introspection.Query,
			OperationName: "IntrospectionQuery",
		}
		jq, err := json.Marshal(q)
		if err != nil {
			return err
		}
		fmt.Print(string(jq) + "\n")
	} else {
		fmt.Printf("%s\n", introspection.Query)
	}
	return nil
}
