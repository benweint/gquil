package commands

type CommonOptions struct {
	// TODO: stdin support
	SchemaFiles     []string `arg:"" name:"schemas" help:"Path to the GraphQL SDL schema file(s) to read from."`
	IncludeBuiltins bool     `name:"include-builtins" help:"Include built-in types and directives (they're omitted by default)."`
	Json            bool     `name:"json" help:"Output results as JSON."`
}
