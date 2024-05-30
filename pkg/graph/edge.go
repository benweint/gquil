package graph

import "github.com/benweint/gquil/pkg/model"

type edgeKind int

const (
	edgeKindField edgeKind = iota
	edgeKindInputField
	edgeKindArgument
	edgeKindPossibleType
)

type edge struct {
	src          *node
	dst          *node
	kind         edgeKind
	field        *model.FieldDefinition
	inputField   *model.InputValueDefinition
	argument     *model.InputValueDefinition
	possibleType string
}
