package model

import (
	"github.com/benweint/gquil/pkg/astutil"
)

func filterBuiltinTypes(defs DefinitionList) DefinitionList {
	var result DefinitionList
	for _, def := range defs {
		if astutil.IsBuiltinType(def.Name) {
			continue
		}
		result = append(result, def)
	}
	return result
}

func filterBuiltinDirectives(dirs DirectiveDefinitionList) DirectiveDefinitionList {
	var result DirectiveDefinitionList
	for _, d := range dirs {
		if astutil.IsBuiltinDirective(d.Name) {
			continue
		}
		result = append(result, d)
	}
	return result
}
