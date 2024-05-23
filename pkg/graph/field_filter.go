package graph

import (
	"strings"

	"github.com/benweint/gquil/pkg/model"
)

func parseFilter(root string) *fieldFilter {
	parts := strings.SplitN(root, ".", 2)
	if len(parts) == 1 {
		return &fieldFilter{
			onType:     parts[0],
			includeAll: true,
		}
	}

	return &fieldFilter{
		onType: parts[0],
		includeFields: map[string]bool{
			parts[1]: true,
		},
	}
}

func makeFieldFilters(roots []string) map[string]*fieldFilter {
	filtersByType := map[string]*fieldFilter{}
	for _, root := range roots {
		filter := parseFilter(root)
		if existingFilter, ok := filtersByType[filter.onType]; ok {
			existingFilter.merge(filter)
		} else {
			filtersByType[filter.onType] = filter
		}
	}
	return filtersByType
}

func applyFieldFilters(defs model.DefinitionList, roots []string) model.DefinitionList {
	var result model.DefinitionList
	filters := makeFieldFilters(roots)

	for _, typeDef := range defs {
		filter, ok := filters[typeDef.Name]
		if !ok {
			continue
		}

		filteredDef := &model.Definition{
			Kind:          typeDef.Kind,
			Name:          typeDef.Name,
			Description:   typeDef.Description,
			Interfaces:    typeDef.Interfaces,
			PossibleTypes: typeDef.PossibleTypes,
			EnumValues:    typeDef.EnumValues,
			Fields:        applyFilter(filter, typeDef.Fields),
			InputFields:   applyFilter(filter, typeDef.InputFields),
		}

		result = append(result, filteredDef)
	}

	return result
}

type fieldFilter struct {
	onType        string
	includeAll    bool
	includeFields map[string]bool
}

type fieldLike interface {
	FieldName() string
}

func (f *fieldFilter) merge(other *fieldFilter) {
	if other.includeAll {
		f.includeAll = true
		return
	}

	for field := range other.includeFields {
		f.includeFields[field] = true
	}
}

func applyFilter[T fieldLike](f *fieldFilter, list []T) []T {
	if f.includeAll {
		return list
	}

	var result []T
	for _, field := range list {
		if f.includeFields[field.FieldName()] {
			result = append(result, field)
		}
	}
	return result
}
