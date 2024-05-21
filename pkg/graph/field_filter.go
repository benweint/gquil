package graph

import (
	"strings"

	"github.com/benweint/gquilt/pkg/model"
)

func makeFieldFilters(roots []string) map[string]*fieldFilter {
	filtersByType := map[string]*fieldFilter{}
	for _, root := range roots {
		nameParts := strings.SplitN(root, ".", 2)
		typePart := nameParts[0]
		if len(nameParts) == 1 {
			filtersByType[typePart] = makeFieldFilter(typePart, true)
			continue
		}

		fieldPart := nameParts[1]
		filter, ok := filtersByType[typePart]
		if !ok {
			filter = makeFieldFilter(typePart, false)
			filtersByType[typePart] = filter
		}

		filter.allowField(fieldPart)
	}

	return filtersByType
}

func applyFieldFilters(defs model.DefinitionList, roots []string) model.DefinitionList {
	var selected model.DefinitionList
	filters := makeFieldFilters(roots)

	for _, typeDef := range defs {
		filter, ok := filters[typeDef.Name]
		if !ok {
			continue
		}

		selected = append(selected, typeDef)
		typeDef.Fields = filter.filterFieldList(typeDef.Fields)
	}

	return selected
}

type fieldFilter struct {
	onType        string
	includeAll    bool
	includeFields map[string]bool
}

func makeFieldFilter(typeName string, includeAll bool) *fieldFilter {
	return &fieldFilter{
		onType:        typeName,
		includeAll:    includeAll,
		includeFields: map[string]bool{},
	}
}

func (f *fieldFilter) allowField(fieldName string) {
	f.includeFields[fieldName] = true
}

func (f *fieldFilter) filterFieldList(fields model.FieldDefinitionList) model.FieldDefinitionList {
	if f.includeAll {
		return fields
	}

	var result model.FieldDefinitionList
	for _, field := range fields {
		if f.includeFields[field.Name] {
			result = append(result, field)
		}
	}
	return result
}
