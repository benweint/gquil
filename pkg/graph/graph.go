package graph

import (
	"fmt"
	"strings"

	"github.com/benweint/gquilt/pkg/model"
	"github.com/vektah/gqlparser/v2/ast"
)

type Graph struct {
	nodes              map[string]*Node
	edges              map[string][]*Edge
	interfacesAsUnions bool
	opts               []GraphOption
}

type Node struct {
	*model.Definition
}

func (n *Node) ID() string {
	return "n_" + n.Name
}

type EdgeKind int

const (
	EdgeKindField = iota
	EdgeKindInputField
	EdgeKindArgument
	EdgeKindPossibleType
)

type Edge struct {
	src          *Node
	dst          *Node
	kind         EdgeKind
	field        *model.FieldDefinition
	inputField   *model.InputValueDefinition
	argument     *model.InputValueDefinition
	possibleType string
}

func normalizeKind(kind ast.DefinitionKind, interfacesAsUnions bool) ast.DefinitionKind {
	if kind == ast.Interface {
		if interfacesAsUnions {
			return ast.Union
		}
		return ast.Object
	}
	return kind
}

type GraphOption func(g *Graph)

func WithInterfacesAsUnions() GraphOption {
	return func(g *Graph) {
		g.interfacesAsUnions = true
	}
}

func MakeGraph(defs model.DefinitionList, opts ...GraphOption) *Graph {
	g := &Graph{
		nodes: map[string]*Node{},
		edges: map[string][]*Edge{},
		opts:  opts,
	}

	for _, opt := range opts {
		opt(g)
	}

	for _, t := range defs {
		g.nodes[t.Name] = &Node{
			Definition: t,
		}
	}

	for _, t := range defs {
		var typeEdges []*Edge
		kind := normalizeKind(t.Kind, g.interfacesAsUnions)
		switch kind {
		case ast.Object:
			for _, f := range t.Fields {
				targetType := f.Type.Unwrap()
				targetTypeKind := targetType.Kind
				if targetTypeKind == model.ScalarKind {
					continue
				}
				srcNode := g.nodes[t.Name]
				dstNode := g.nodes[targetType.Name]
				if srcNode == nil || dstNode == nil {
					continue
				}
				edge := &Edge{
					src:   srcNode,
					dst:   dstNode,
					kind:  EdgeKindField,
					field: f,
				}
				typeEdges = append(typeEdges, edge)

				for _, arg := range f.Arguments {
					targetType := arg.Type.Unwrap()
					if targetType.Kind == model.ScalarKind {
						continue
					}
					dstNode := g.nodes[targetType.Name]
					if dstNode == nil {
						continue
					}
					typeEdges = append(typeEdges, &Edge{
						src:      srcNode,
						dst:      g.nodes[targetType.Name],
						kind:     EdgeKindArgument,
						field:    f,
						argument: arg,
					})
				}
			}
		case ast.InputObject:
			for _, f := range t.InputFields {
				targetType := f.Type.Unwrap()
				targetTypeKind := targetType.Kind
				if targetTypeKind == model.ScalarKind {
					continue
				}
				srcNode := g.nodes[t.Name]
				dstNode := g.nodes[targetType.Name]
				if srcNode == nil || dstNode == nil {
					continue
				}
				edge := &Edge{
					src:        srcNode,
					dst:        dstNode,
					kind:       EdgeKindInputField,
					inputField: f,
				}
				typeEdges = append(typeEdges, edge)
			}
		case ast.Union:
			for _, pt := range t.PossibleTypes {
				srcNode := g.nodes[t.Name]
				dstNode := g.nodes[pt]
				if srcNode == nil || dstNode == nil {
					continue
				}
				typeEdges = append(typeEdges, &Edge{
					src:          srcNode,
					dst:          dstNode,
					kind:         EdgeKindPossibleType,
					possibleType: pt,
				})
			}
		}
		g.edges[t.Name] = typeEdges
	}

	return g
}

func (g *Graph) GetDefinitions() model.DefinitionList {
	var result model.DefinitionList
	for _, node := range g.nodes {
		result = append(result, node.Definition)
	}
	return result
}

func (g *Graph) ReachableFrom(roots []string, maxDepth int) *Graph {
	var defs model.DefinitionList
	defMap := map[string]*model.Definition{}

	for _, node := range g.nodes {
		defs = append(defs, node.Definition)
		defMap[node.Name] = node.Definition
	}
	rootDefs := applyFieldFilters(defs, roots)

	seen := map[string]*model.Definition{}

	var traverse func(def *model.Definition, depth int)
	traverse = func(def *model.Definition, depth int) {
		if maxDepth > 0 && depth > maxDepth {
			return
		}
		if _, ok := seen[def.Name]; ok {
			return
		}
		if def.Kind == ast.Scalar {
			return
		}

		seen[def.Name] = def
		kind := normalizeKind(def.Kind, g.interfacesAsUnions)

		switch kind {
		case ast.Object:
			for _, field := range def.Fields {
				for _, arg := range field.Arguments {
					argType := arg.Type.Unwrap()
					traverse(defMap[argType.Name], depth+1)
				}

				underlyingType := field.Type.Unwrap()
				traverse(defMap[underlyingType.Name], depth+1)
			}
		case ast.InputObject:
			for _, field := range def.InputFields {
				underlyingType := field.Type.Unwrap()
				traverse(defMap[underlyingType.Name], depth+1)
			}
		case ast.Union:
			for _, pt := range def.PossibleTypes {
				traverse(defMap[pt], depth+1)
			}
		}
	}

	for _, root := range rootDefs {
		traverse(root, 1)
	}

	var newDefs model.DefinitionList
	for _, def := range seen {
		newDefs = append(newDefs, def)
	}

	return MakeGraph(newDefs, g.opts...)
}

func (g *Graph) ToDot() string {
	var nodeDefs []string
	for _, node := range g.nodes {
		nodeDef := fmt.Sprintf("  %s [shape=plain, label=<%s>]", node.ID(), g.makeNodeLabel(node))
		nodeDefs = append(nodeDefs, nodeDef)
	}

	edgeDefs := g.buildEdgeDefs()

	return "digraph {\nrankdir=LR\nranksep=2\nnode [shape=box fontname=Courier]\n" + strings.Join(nodeDefs, "\n") + "\n" + strings.Join(edgeDefs, "\n") + "\n}\n"
}

func (g *Graph) buildEdgeDefs() []string {
	var result []string

	for _, edges := range g.edges {
		for _, edge := range edges {
			srcPortSuffix := ""
			dstPortSuffix := ":main"

			switch edge.kind {
			case EdgeKindField:
				srcPortSuffix = ":" + portName(edge.field.Name)
			case EdgeKindInputField:
				srcPortSuffix = ":" + portName(edge.inputField.Name)
			case EdgeKindArgument:
				srcPortSuffix = ":" + portNameForArgument(edge.field.Name, edge.argument.Name)
			case EdgeKindPossibleType:
				srcPortSuffix = ":" + portName(edge.possibleType)
			}

			result = append(result, fmt.Sprintf("  %s%s -> %s%s", edge.src.ID(), srcPortSuffix, edge.dst.ID(), dstPortSuffix))
		}
	}

	return result
}

func (g *Graph) makeNodeLabel(node *Node) string {
	switch normalizeKind(node.Kind, g.interfacesAsUnions) {
	case ast.Object:
		return makeFieldTableNodeLabel(node)
	case ast.InputObject:
		return makeInputObjectNodeLabel(node)
	case ast.Enum:
		return makeEnumLabel(node)
	case ast.Union:
		return makePolymorphicLabel(node)
	default:
		return makeGenericNodeLabel(node)
	}
}

// From https://colorbrewer2.org/#type=qualitative&scheme=Paired&n=5
func colorForKind(kind ast.DefinitionKind) string {
	switch kind {
	case ast.Object:
		return "#fbb4ae"
	case ast.Interface:
		return "#b3cde3"
	case ast.InputObject:
		return "#ccebc5"
	case ast.Enum:
		return "#decbe4"
	case ast.Union:
		return "#fed9a6"
	default:
		return "#ffffff"
	}
}

func portName(fieldName string) string {
	return "p_" + fieldName
}

func portNameForArgument(fieldName, argName string) string {
	return "p_" + fieldName + "_" + argName
}

func makeEnumLabel(node *Node) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD PORT="main" BGCOLOR="%s">enum %s</TD></TR>`, colorForKind(node.Kind), node.Name)
	for _, val := range node.EnumValues {
		result += fmt.Sprintf(`  <TR><TD>%s</TD></TR>\n`, val.Name)
	}
	result += "</TABLE>"
	return result
}

func makePolymorphicLabel(node *Node) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD PORT="main" BGCOLOR="%s">%s %s</TD></TR>`, colorForKind(node.Kind), strings.ToLower(string(node.Kind)), node.Name)
	for _, possibleType := range node.PossibleTypes {
		result += fmt.Sprintf(`  <TR><TD PORT="%s">%s</TD></TR>\n`, portName(possibleType), possibleType)
	}
	result += "</TABLE>"
	return result
}

func makeFieldTableNodeLabel(node *Node) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`    <TR><TD COLSPAN="3" PORT="main" BGCOLOR="%s">%s %s</TD></TR>`+"\n", colorForKind(node.Kind), strings.ToLower(string(node.Kind)), node.Name)
	for _, field := range node.Fields {
		args := field.Arguments
		result += fmt.Sprintf(`    <TR><TD ROWSPAN="%d">%s</TD><TD COLSPAN="2" PORT="%s">%s</TD></TR>`+"\n", len(args)+1, field.Name, portName(field.Name), field.Type.String())
		for _, arg := range args {
			result += fmt.Sprintf(`    <TR><TD>%s</TD><TD PORT="%s">%s</TD></TR>`+"\n", arg.Name, portNameForArgument(field.Name, arg.Name), arg.Type)
		}
	}
	result += "\n  </TABLE>"
	return result
}

func makeInputObjectNodeLabel(node *Node) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD COLSPAN="2" PORT="main" BGCOLOR="%s">input %s</TD></TR>`+"\n", colorForKind(node.Kind), node.Name)
	for _, field := range node.InputFields {
		result += fmt.Sprintf(`  <TR><TD>%s</TD><TD PORT="%s">%s</TD></TR>`+"\n", field.Name, portName(field.Name), field.Type)
	}
	result += "</TABLE>"
	return result
}

func makeGenericNodeLabel(node *Node) string {
	return fmt.Sprintf("%s\n%s", strings.ToLower(string(node.Kind)), node.Name)
}
