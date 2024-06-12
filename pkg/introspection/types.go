package introspection

// The types in this file are used to deserialize the response from a GraphQL introspection query.
// As such, they directly mirror the specified introspection schema described in the GraphQL spec
// here: https://spec.graphql.org/October2021/#sec-Schema-Introspection.Schema-Introspection-Schema

type IntrospectionQueryResult struct {
	Schema Schema `json:"__schema"`
}

// Schema represents an instance of the __Schema introspection type:
// https://spec.graphql.org/October2021/#sec-The-__Schema-Type
type Schema struct {
	Types            []Type      `json:"types"`
	QueryType        Type        `json:"queryType,omitempty"`
	MutationType     Type        `json:"mutationType,omitempty"`
	SubscriptionType Type        `json:"subscriptionType,omitempty"`
	Directives       []Directive `json:"directives,omitempty"`
}

// Type represents an instance of the __Type introspection type:
// https://spec.graphql.org/October2021/#sec-The-__Type-Type
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

// Directive represents an instance of the __Directive introspection type:
// https://spec.graphql.org/October2021/#sec-The-__Directive-Type
type Directive struct {
	Name         string              `json:"name"`
	Description  string              `json:"description,omitempty"`
	Locations    []DirectiveLocation `json:"locations"`
	Args         []InputValue        `json:"args"`
	IsRepeatable bool                `json:"isRepeatable"`
}

// Field represents an instance of the __Field introspection type:
// https://spec.graphql.org/October2021/#sec-The-__Field-Type
type Field struct {
	Name              string       `json:"name"`
	Description       string       `json:"description,omitempty"`
	Args              []InputValue `json:"args,omitempty"`
	Type              *Type        `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason string       `json:"deprecationReason"`
}

// EnumValue represents an instance of the __EnumValue introspection type:
// https://spec.graphql.org/October2021/#sec-The-__EnumValue-Type
type EnumValue struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
}

// InputValue represents an instance of the __InputValue introspection type:
// https://spec.graphql.org/October2021/#sec-The-__InputValue-Type
type InputValue struct {
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	Type         *Type   `json:"type"`
	DefaultValue *string `json:"defaultValue,omitempty"`
}

// TypeKind represents a possible value of the __TypeKind introspection enum.
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

// DirectiveLocation represents a possible value of the __DirectiveLocation introspection enum.
type DirectiveLocation string
