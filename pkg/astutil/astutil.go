package astutil

import (
	"slices"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

func IsBuiltinDirective(name string) bool {
	builtinDirectives := []string{
		"if",
		"skip",
		"include",
		"deprecated",
		"specifiedBy",
		"defer",
	}

	return slices.Contains(builtinDirectives, name)
}

func IsBuiltinType(name string) bool {
	if strings.HasPrefix(name, "__") {
		return true
	}

	if name == "String" || name == "ID" || name == "Boolean" || name == "Int" || name == "Float" || name == "Enum" {
		return true
	}

	return false
}

func FilterBuiltins(s *ast.Schema) {
	s.Types = filterBuiltinTypes(s.Types)
	s.Directives = filterBuiltinDirectives(s.Directives)
}

func filterBuiltinTypes(defs map[string]*ast.Definition) map[string]*ast.Definition {
	result := map[string]*ast.Definition{}
	for _, def := range defs {
		if IsBuiltinType(def.Name) {
			continue
		}
		result[def.Name] = def
	}
	return result
}

func filterBuiltinDirectives(dirs map[string]*ast.DirectiveDefinition) map[string]*ast.DirectiveDefinition {
	result := map[string]*ast.DirectiveDefinition{}
	for _, d := range dirs {
		if IsBuiltinDirective(d.Name) {
			continue
		}
		result[d.Name] = d
	}
	return result
}
