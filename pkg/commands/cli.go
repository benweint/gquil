package commands

type CLI struct {
	Ls            LsCmd            `cmd:"" aliases:"list" help:"List types, fields, or directives in an SDL document."`
	Merge         MergeCmd         `cmd:"" help:"Merge multiple SDL documents into a single one."`
	Json          JsonCmd          `cmd:"" help:"Return a JSON representation of an SDL document."`
	Introspection IntrospectionCmd `cmd:"" help:"Interact with a GraphQL introspection endpoint over HTTP."`
	Viz           VizCmd           `cmd:"" help:"Visualize schema as a GraphViz dot file."`
}
