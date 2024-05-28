package commands

type JsonCmd struct {
	InputOptions
	FilteringOptions
	GraphFilteringOptions
}

func (c *JsonCmd) Help() string {
	return `Print a flattened JSON representation of the given GraphQL schema, suitable for processing with jq or similar. The JSON format used is inspired by but not identical to the GraphQL introspection type system. It differs mainly in that references to named types are 'flattened' into strings, rather than being represented as recursively nested objects.

Unlike the introspection types in the GraphQL spec, the JSON output format does include information about the application sites of directives, under the 'directives' key.

The JSON format for fields and arguments also adds several convenience fields which are useful when processing the output:

  * underlyingTypeName: the underlying named type of the field, after unwrapping list and non-null wrapping types. For example, a field of type '[String!]' would have an underlyingTypeName of 'String')
  * typeName: the type of the field, represented as a string in GraphQL SDL notation (for example: '[String!]!')`
}

func (c *JsonCmd) Run(ctx Context) error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	if err = c.filterSchema(s); err != nil {
		return err
	}

	if !c.IncludeBuiltins {
		s.FilterBuiltins()
	}

	return ctx.PrintJson(s)
}
