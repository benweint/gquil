package graph

import (
	"github.com/benweint/gquil/pkg/model"
)

func makeFilter(root *model.NameReference) *fieldFilter {
	result := &fieldFilter{
		onType:        root.GetTargetType().Name,
		includeFields: map[string]bool{},
	}

	switch root.Kind {
	case model.TypeNameReference:
		result.includeAll = true
	case model.FieldNameReference, model.InputFieldNameReference:
		result.includeFields[root.GetFieldName()] = true
	}

	return result
}

func makeFieldFilters(roots []*model.NameReference) map[string]*fieldFilter {
	filtersByType := map[string]*fieldFilter{}
	for _, root := range roots {
		filter := makeFilter(root)
		if existingFilter, ok := filtersByType[filter.onType]; ok {
			existingFilter.merge(filter)
		} else {
			filtersByType[filter.onType] = filter
		}
	}
	return filtersByType
}

func applyFieldFilters(defs model.DefinitionList, roots []*model.NameReference) model.DefinitionList {
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

func (f *fieldFilter) merge(other *fieldFilter) {
	if other.includeAll {
		f.includeAll = true
		return
	}

	for field := range other.includeFields {
		f.includeFields[field] = true
	}
}

func applyFilter(f *fieldFilter, list model.FieldDefinitionList) model.FieldDefinitionList {
	if f.includeAll {
		return list
	}

	var result model.FieldDefinitionList
	for _, field := range list {
		if f.includeFields[field.Name] {
			result = append(result, field)
		}
	}
	return result
}
