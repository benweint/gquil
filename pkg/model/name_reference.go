package model

type NameReference struct {
	TypeName  string
	FieldName string
}

func TypeNameReference(name string) NameReference {
	return NameReference{
		TypeName: name,
	}
}

func FieldNameReference(typeName, fieldName string) NameReference {
	return NameReference{
		TypeName:  typeName,
		FieldName: fieldName,
	}
}
