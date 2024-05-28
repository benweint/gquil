package commands

import (
	"github.com/benweint/gquil/pkg/introspection"
)

type IntrospectionCmd struct {
	GenerateSDL GenerateSDLCmd `cmd:"" help:"Generate GraphQL SDL from a GraphQL introspection endpoint over HTTP(S)."`
	Query       EmitQueryCmd   `cmd:"" help:"Emit the GraphQL query used for introspection."`
}

type SpecVersionOptions struct {
	SpecVersion string `name:"spec-version" help:"GraphQL spec version to use when making the introspection query. One of june2018, october2021. You may want to use a newer spec version when interacting with servers which support it, to get newer fields (like the schema description field, or the isRepeatable field on directives, which were added in the october2021 spec)." default:"june2018"`
}

type EmitQueryCmd struct {
	Json bool `name:"json" help:"Emit the query as JSON, suitable for passing to curl or similar."`
	SpecVersionOptions
}

func (c *EmitQueryCmd) Run(ctx Context) error {
	sv, err := introspection.ParseSpecVersion(c.SpecVersion)
	if err != nil {
		return err
	}

	if c.Json {
		q := introspection.GraphQLParams{
			Query:         introspection.GetQuery(sv),
			OperationName: "IntrospectionQuery",
		}
		return ctx.PrintJson(q)
	}

	ctx.Printf("%s\n", introspection.GetQuery(sv))
	return nil
}
