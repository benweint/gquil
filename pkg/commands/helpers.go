package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/benweint/gquil/pkg/model"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func loadSchemaModel(paths []string) (*model.Schema, error) {
	rawSchema, err := parseSchemaFromPaths(paths)
	if err != nil {
		return nil, err
	}

	s, err := model.MakeSchema(rawSchema)
	if err != nil {
		return nil, err
	}

	sort.Slice(s.Directives, func(i, j int) bool {
		return strings.Compare(s.Directives[i].Name, s.Directives[j].Name) < 0
	})

	return s, nil
}

func parseSchemaFromPaths(paths []string) (*ast.Schema, error) {
	var sources []*ast.Source
	for _, path := range paths {
		var raw []byte
		var err error
		if path == "-" {
			path = "stdin"
			raw, err = io.ReadAll(os.Stdin)
		} else {
			raw, err = os.ReadFile(path)
		}
		if err != nil {
			return nil, fmt.Errorf("could not read source SDL from %s: %w", path, err)
		}

		source := ast.Source{
			Name:  path,
			Input: string(raw),
		}
		sources = append(sources, &source)
	}
	schema, err := gqlparser.LoadSchema(sources...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source SDL: %w", err)
	}

	return schema, nil
}

func formatArgumentDefinitionList(al model.ArgumentDefinitionList) string {
	if len(al) == 0 {
		return ""
	}

	var formattedArgs []string
	for _, arg := range al {
		formatted := fmt.Sprintf("%s: %s", arg.Name, arg.Type)
		formattedArgs = append(formattedArgs, formatted)
	}
	return "(" + strings.Join(formattedArgs, ", ") + ")"
}

func formatArgumentList(al model.ArgumentList) (string, error) {
	if len(al) == 0 {
		return "", nil
	}

	var formattedArgs []string
	for _, arg := range al {
		formattedValue, err := formatValue(arg.Value)
		if err != nil {
			return "", err
		}
		formatted := fmt.Sprintf("%s: %s", arg.Name, formattedValue)
		formattedArgs = append(formattedArgs, formatted)
	}
	return "(" + strings.Join(formattedArgs, ", ") + ")", nil
}

func formatDirectiveList(dl model.DirectiveList) (string, error) {
	if len(dl) == 0 {
		return "", nil
	}

	var formattedDirectives []string
	for _, d := range dl {
		formattedArgs, err := formatArgumentList(d.Arguments)
		if err != nil {
			return "", err
		}
		formatted := fmt.Sprintf("@%s%s", d.Name, formattedArgs)
		formattedDirectives = append(formattedDirectives, formatted)
	}

	return " " + strings.Join(formattedDirectives, " "), nil
}

func formatValue(v model.Value) (string, error) {
	raw, err := json.Marshal(v)
	return string(raw), err
}
