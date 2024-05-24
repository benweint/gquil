package model

type NameReferenceKind int

const (
	TypeNameReference = iota
	FieldNameReference
	InputFieldNameReference
)

type NameReference struct {
	Kind       NameReferenceKind
	typeRef    *Definition
	field      *FieldDefinition
	inputField *InputValueDefinition
}

func (n *NameReference) GetTargetType() *Definition {
	return n.typeRef
}

func (n *NameReference) GetFieldName() string {
	switch n.Kind {
	case FieldNameReference:
		return n.field.Name
	case InputFieldNameReference:
		return n.inputField.Name
	}
	return ""
}
