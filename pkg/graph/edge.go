package graph

import "github.com/benweint/gquil/pkg/model"

type edgeKind int

const (
	edgeKindField edgeKind = iota
	edgeKindArgument
	edgeKindPossibleType
)

type edge struct {
	kind     edgeKind
	src      *model.Definition
	dst      *model.Definition
	field    *model.FieldDefinition    // only set for fields representing eges or arguments
	argument *model.ArgumentDefinition // only set for fields representing arguments
}
