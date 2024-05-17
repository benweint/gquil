package commands

type LsCmd struct {
	Types      LsTypesCmd      `cmd:"" help:"List types in the given schema(s)."`
	Fields     LsFieldsCmd     `cmd:"" help:"List fields in the given schema(s)."`
	Directives LsDirectivesCmd `cmd:"" help:"List directive definitions in the given schema(s)."`
}
