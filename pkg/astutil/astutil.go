package astutil

import (
	"slices"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// IsBuiltinDirective returns true if the given directive name is one of the
// built-in directives specified here: https://spec.graphql.org/October2021/#sec-Type-System.Directives.Built-in-Directives
//
// @defer is also counted as built-in, despite not appearing in the spec, because it is included in the prelude used by the
// GraphQL parsing library we're using here: https://github.com/vektah/gqlparser/blob/master/validator/prelude.graphql
func IsBuiltinDirective(name string) bool {
	builtinDirectives := []string{
		"skip",
		"include",
		"deprecated",
		"specifiedBy",
		"defer", // not specified, but in the gqlparser prelude
	}

	return slices.Contains(builtinDirectives, name)
}

// IsBuiltinType returns true if the given type name is either a reserved name
// (see https://spec.graphql.org/October2021/#sec-Names.Reserved-Names) or one of the specified
// scalar types in the GraphQL spec here: https://spec.graphql.org/October2021/#sec-Scalars.Built-in-Scalars
func IsBuiltinType(name string) bool {
	if strings.HasPrefix(name, "__") {
		return true
	}

	if name == "String" || name == "ID" || name == "Boolean" || name == "Int" || name == "Float" || name == "Enum" {
		return true
	}

	return false
}

// IsBuiltinField returns true if the given field name is a reserved name (begins with '__').
func IsBuiltinField(name string) bool {
	return strings.HasPrefix(name, "__")
}

// FilterBuiltins accepts and mutates an *ast.Schema to remove all built-in types and directives.
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
