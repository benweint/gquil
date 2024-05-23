package graph

import "github.com/benweint/gquil/pkg/model"

type node struct {
	*model.Definition
}

func (n *node) ID() string {
	return "n_" + n.Name
}
