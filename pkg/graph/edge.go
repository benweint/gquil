package graph

import "github.com/benweint/gquil/pkg/model"

type edgeKind int

const (
	edgeKindField edgeKind = iota
	edgeKindArgument
	edgeKindPossibleType
)

type edge struct {
	src          *node
	dst          *node
	kind         edgeKind
	field        *model.FieldDefinition
	argument     *model.ArgumentDefinition
	possibleType string
}
