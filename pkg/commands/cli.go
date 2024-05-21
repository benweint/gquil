package commands

type CLI struct {
	Ls            LsCmd            `cmd:"" aliases:"list" help:"List types, fields, or directives in a GraphQL SDL document."`
	Merge         MergeCmd         `cmd:"" help:"Merge multiple GraphQL SDL documents into a single one."`
	Json          JsonCmd          `cmd:"" help:"Return a JSON representation of a GraphQL SDL document."`
	Introspection IntrospectionCmd `cmd:"" help:"Interact with a GraphQL introspection endpoint over HTTP."`
	Viz           VizCmd           `cmd:"" help:"Visualize a GraphQL schema using GraphViz."`
}
