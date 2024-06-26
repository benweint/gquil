package graph

import "github.com/benweint/gquil/pkg/model"

// referenceSet captures the set of types & fields which have been encountered when traversing a GraphQL schema.
type referenceSet map[model.NameReference]bool

// includesType returns true if the target referenceSet includes at least one field on the given type name,
// or a key representing the entire type.
func (s referenceSet) includesType(name string) bool {
	for key := range s {
		if key.TypeName == name {
			return true
		}
	}
	return false
}

// includesField returns true if the given referenceSet includes a key representing the given field on the given type.
func (s referenceSet) includesField(typeName, fieldName string) bool {
	return s[model.FieldNameReference(typeName, fieldName)]
}

// filterFields returns a copy of the given definition, where the field list has been filtered to only include
// fields which were included in the referenceSet. The original def is not modified by this method.
func (s referenceSet) filterFields(def *model.Definition) *model.Definition {
	var filteredFields []*model.FieldDefinition
	for _, field := range def.Fields {
		if s.includesField(def.Name, field.Name) {
			filteredFields = append(filteredFields, field)
		}
	}

	return &model.Definition{
		Kind:          def.Kind,
		Name:          def.Name,
		Description:   def.Description,
		Directives:    def.Directives,
		Interfaces:    def.Interfaces,
		PossibleTypes: def.PossibleTypes,
		EnumValues:    def.EnumValues,
		Fields:        filteredFields,
	}
}
