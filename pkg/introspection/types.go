package introspection

type IntrospectionQueryResult struct {
	Schema Schema `json:"__schema"`
}

type Schema struct {
	Types            []Type      `json:"types"`
	QueryType        Type        `json:"queryType,omitempty"`
	MutationType     Type        `json:"mutationType,omitempty"`
	SubscriptionType Type        `json:"subscriptionType,omitempty"`
	Directives       []Directive `json:"directives,omitempty"`
}

type Type struct {
	Kind        TypeKind `json:"kind"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`

	// OBJECT and INTERFACE only
	Fields []Field `json:"fields,omitempty"`

	// OBJECT only
	Interfaces []Type `json:"interfaces,omitempty"`

	// INTERFACE and UNION only
	PossibleTypes []Type `json:"possibleTypes,omitempty"`

	// ENUM only
	EnumValues []EnumValue `json:"enumValues,omitempty"`

	// INPUT_OBJECT only
	InputFields []InputValue `json:"inputFields,omitempty"`

	// NON_NULL and LIST only
	OfType *Type `json:"ofType,omitempty"`
}

type Directive struct {
	Name         string              `json:"name"`
	Description  string              `json:"description,omitempty"`
	Locations    []DirectiveLocation `json:"locations"`
	Args         []InputValue        `json:"args"`
	IsRepeatable bool                `json:"isRepeatable"`
}

type Field struct {
	Name              string       `json:"name"`
	Description       string       `json:"description,omitempty"`
	Args              []InputValue `json:"args,omitempty"`
	Type              *Type        `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason string       `json:"deprecationReason"`
}

type EnumValue struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
}

type InputValue struct {
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	Type         *Type   `json:"type"`
	DefaultValue *string `json:"defaultValue,omitempty"`
}

type TypeKind string

const (
	ScalarKind      = TypeKind("SCALAR")
	ObjectKind      = TypeKind("OBJECT")
	InterfaceKind   = TypeKind("INTERFACE")
	UnionKind       = TypeKind("UNION")
	EnumKind        = TypeKind("ENUM")
	InputObjectKind = TypeKind("INPUT_OBJECT")
	ListKind        = TypeKind("LIST")
	NonNullKind     = TypeKind("NON_NULL")
)

type DirectiveLocation string
