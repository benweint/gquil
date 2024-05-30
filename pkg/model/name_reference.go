package model

type NameReferenceKind int

const (
	TypeNameReference = iota
	FieldNameReference
	InputFieldNameReference
)

type NameReference struct {
	Kind    NameReferenceKind
	typeRef *Definition
	field   *FieldDefinition
}

func (n *NameReference) GetTargetType() *Definition {
	return n.typeRef
}

func (n *NameReference) GetFieldName() string {
	if n.field != nil {
		return n.field.Name
	}
	return ""
}
