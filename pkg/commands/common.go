package commands

type InputOptions struct {
	// TODO: stdin support
	SchemaFiles []string `arg:"" name:"schemas" help:"Path to the GraphQL SDL schema file(s) to read from."`
}

type FilteringOptions struct {
	IncludeBuiltins bool `name:"include-builtins" help:"Include built-in types and directives (they're omitted by default)."`
}

type IncludeDirectivesOption struct {
	IncludeDirectives bool `name:"include-directives" help:"Include applied directives in human-readable output. Has no effect with --json."`
}

type OutputOptions struct {
	Json bool `name:"json" help:"Output results as JSON."`
}

type GraphFilteringOptions struct {
	From  []string `name:"from" help:"Only include types reachable from the specified type(s) or field(s). May be specified multiple times to use multiple roots."`
	Depth int      `name:"depth" help:"When used with --from, limit the depth of traversal."`
}
