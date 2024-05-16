package model

import (
	"fmt"

	"github.com/vektah/gqlparser/v2/ast"
)

type TypeKind string

// See __TypeKind in the spec:
// https://spec.graphql.org/October2021/#sel-GAJXNFAD7EAADxFAB45Y
const (
	UnknownKind   = TypeKind("")
	ScalarKind    = TypeKind("SCALAR")
	ObjectKind    = TypeKind("OBJECT")
	InterfaceKind = TypeKind("INTERFACE")
	UnionKind     = TypeKind("UNION")
	EnumKind      = TypeKind("ENUM")
	InputKind     = TypeKind("INPUT_OBJECT")
	ListKind      = TypeKind("LIST")
	NonNullKind   = TypeKind("NON_NULL")
)

type Type struct {
	Kind   TypeKind `json:"kind"`
	Name   string   `json:"name,omitempty"`
	OfType *Type    `json:"ofType,omitempty"`
}

func (t *Type) Unwrap() *Type {
	if t.OfType != nil {
		return t.OfType.Unwrap()
	}

	return t
}

func (t *Type) String() string {
	switch t.Kind {
	case NonNullKind:
		return fmt.Sprintf("%s!", t.OfType.String())
	case ListKind:
		return fmt.Sprintf("[%s]", t.OfType.String())
	default:
		return t.Name
	}
}

func makeType(in *ast.Type) *Type {
	if in == nil {
		return nil
	}

	if in.NonNull {
		return &Type{
			Kind: NonNullKind,
			OfType: makeType(&ast.Type{
				NamedType: in.NamedType,
				Elem:      in.Elem,
			}),
		}
	}

	if in.Elem != nil {
		return &Type{
			Kind:   ListKind,
			OfType: makeType(in.Elem),
		}
	}

	return &Type{
		Kind: UnknownKind,
		Name: in.NamedType,
	}
}
