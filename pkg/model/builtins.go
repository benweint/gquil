package model

import (
	"github.com/benweint/gquil/pkg/astutil"
)

func filterBuiltinTypesAndFields(defs DefinitionList) DefinitionList {
	var result DefinitionList
	for _, def := range defs {
		if astutil.IsBuiltinType(def.Name) {
			continue
		}
		def.Fields = filterBuiltinFields(def.Fields)
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

func filterBuiltinFields(defs FieldDefinitionList) FieldDefinitionList {
	var result FieldDefinitionList
	for _, fd := range defs {
		if astutil.IsBuiltinField(fd.Name) {
			continue
		}
		result = append(result, fd)
	}
	return result
}
