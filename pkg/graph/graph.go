package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/benweint/gquil/pkg/astutil"
	"github.com/benweint/gquil/pkg/model"
	"github.com/vektah/gqlparser/v2/ast"
)

type Graph struct {
	nodes              map[string]*node
	edges              map[string][]*edge
	interfacesAsUnions bool
	renderBuiltins     bool
	opts               []GraphOption
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

func WithBuiltins(renderBuiltins bool) GraphOption {
	return func(g *Graph) {
		g.renderBuiltins = renderBuiltins
	}
}

func MakeGraph(defs model.DefinitionList, opts ...GraphOption) *Graph {
	g := &Graph{
		nodes: map[string]*node{},
		edges: map[string][]*edge{},
		opts:  opts,
	}

	for _, opt := range opts {
		opt(g)
	}

	for _, t := range defs {
		g.nodes[t.Name] = &node{Definition: t}
	}

	for _, t := range defs {
		var typeEdges []*edge
		kind := normalizeKind(t.Kind, g.interfacesAsUnions)
		switch kind {
		case ast.Object:
			typeEdges = g.makeFieldEdges(t)
		case ast.InputObject:
			typeEdges = g.makeInputEdges(t)
		case ast.Union:
			typeEdges = g.makeUnionEdges(t)
		}
		g.edges[t.Name] = typeEdges
	}

	return g
}

func (g *Graph) makeFieldEdges(t *model.Definition) []*edge {
	var result []*edge
	for _, f := range t.Fields {
		fieldEdge := g.makeFieldEdge(t, f.Type.Unwrap(), f, nil)
		if fieldEdge == nil {
			continue
		}
		result = append(result, fieldEdge)
		for _, arg := range f.Arguments {
			argEdge := g.makeFieldEdge(t, arg.Type.Unwrap(), f, arg)
			if argEdge == nil {
				continue
			}
			result = append(result, argEdge)
		}
	}
	return result
}

func (g *Graph) makeInputEdges(t *model.Definition) []*edge {
	var result []*edge
	for _, f := range t.Fields {
		targetType := f.Type.Unwrap()
		if targetType.Kind == model.ScalarKind {
			continue
		}
		srcNode := g.nodes[t.Name]
		dstNode := g.nodes[targetType.Name]
		if srcNode == nil || dstNode == nil {
			continue
		}
		result = append(result, &edge{
			src:   srcNode,
			dst:   dstNode,
			kind:  edgeKindInputField,
			field: f,
		})
	}
	return result
}

func (g *Graph) makeUnionEdges(t *model.Definition) []*edge {
	var result []*edge
	for _, possibleType := range t.PossibleTypes {
		srcNode := g.nodes[t.Name]
		dstNode := g.nodes[possibleType]
		if srcNode == nil || dstNode == nil {
			continue
		}
		result = append(result, &edge{
			src:          srcNode,
			dst:          dstNode,
			kind:         edgeKindPossibleType,
			possibleType: possibleType,
		})
	}
	return result
}

func (g *Graph) makeFieldEdge(src *model.Definition, targetType *model.Type, f *model.FieldDefinition, arg *model.ArgumentDefinition) *edge {
	kind := edgeKindField

	if arg != nil {
		targetType = arg.Type.Unwrap()
		kind = edgeKindArgument
	}

	targetTypeKind := targetType.Kind
	if targetTypeKind == model.ScalarKind {
		return nil
	}
	srcNode := g.nodes[src.Name]
	dstNode := g.nodes[targetType.Name]
	if srcNode == nil || dstNode == nil {
		return nil
	}
	return &edge{
		src:      srcNode,
		dst:      dstNode,
		kind:     kind,
		field:    f,
		argument: arg,
	}
}

func (g *Graph) GetDefinitions() model.DefinitionList {
	var result model.DefinitionList
	for _, node := range g.nodes {
		result = append(result, node.Definition)
	}
	return result
}

func (g *Graph) ReachableFrom(roots []*model.NameReference, maxDepth int) *Graph {
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
		case ast.Object, ast.InputObject:
			for _, field := range def.Fields {
				for _, arg := range field.Arguments {
					argType := arg.Type.Unwrap()
					traverse(defMap[argType.Name], depth+1)
				}

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
	nodeDefs := g.buildNodeDefs()
	edgeDefs := g.buildEdgeDefs()
	return "digraph {\n  rankdir=LR\n  ranksep=2\n  node [shape=box fontname=Courier]\n" + strings.Join(nodeDefs, "\n") + "\n" + strings.Join(edgeDefs, "\n") + "\n}\n"
}

func (g *Graph) buildNodeDefs() []string {
	var result []string
	for _, name := range sortedKeys(g.nodes) {
		if astutil.IsBuiltinType(name) && !g.renderBuiltins {
			continue
		}
		node := g.nodes[name]
		if node.Kind == ast.Scalar {
			continue
		}
		nodeDef := fmt.Sprintf("  %s [shape=plain, label=<%s>]", node.ID(), g.makeNodeLabel(node))
		result = append(result, nodeDef)
	}
	return result
}

func (g *Graph) buildEdgeDefs() []string {

	var result []string
	for _, sourceNodeName := range sortedKeys(g.edges) {
		edges := g.edges[sourceNodeName]
		for _, edge := range edges {
			srcPortSuffix := ""
			dstPortSuffix := ":main"

			if !g.renderBuiltins {
				if astutil.IsBuiltinType(edge.src.Name) {
					continue
				}
				if edge.field != nil && astutil.IsBuiltinField(edge.field.Name) {
					continue
				}
			}

			switch edge.kind {
			case edgeKindField, edgeKindInputField:
				srcPortSuffix = ":" + portName(edge.field.Name)
			case edgeKindArgument:
				srcPortSuffix = ":" + portNameForArgument(edge.field.Name, edge.argument.Name)
			case edgeKindPossibleType:
				srcPortSuffix = ":" + portName(edge.possibleType)
			}

			result = append(result, fmt.Sprintf("  %s%s -> %s%s", edge.src.ID(), srcPortSuffix, edge.dst.ID(), dstPortSuffix))
		}
	}

	return result
}

func sortedKeys[T any](m map[string]T) []string {
	var result []string
	for k := range m {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func (g *Graph) makeNodeLabel(node *node) string {
	switch normalizeKind(node.Kind, g.interfacesAsUnions) {
	case ast.Object:
		return g.makeFieldTableNodeLabel(node)
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

func makeEnumLabel(node *node) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD PORT="main" BGCOLOR="%s">enum %s</TD></TR>`, colorForKind(node.Kind), node.Name)
	for _, val := range node.EnumValues {
		result += fmt.Sprintf(`  <TR><TD>%s</TD></TR>\n`, val.Name)
	}
	result += "</TABLE>"
	return result
}

func makePolymorphicLabel(node *node) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD PORT="main" BGCOLOR="%s">%s %s</TD></TR>`, colorForKind(node.Kind), strings.ToLower(string(node.Kind)), node.Name)
	for _, possibleType := range node.PossibleTypes {
		result += fmt.Sprintf(`  <TR><TD PORT="%s">%s</TD></TR>\n`, portName(possibleType), possibleType)
	}
	result += "</TABLE>"
	return result
}

func (g *Graph) makeFieldTableNodeLabel(node *node) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`    <TR><TD COLSPAN="3" PORT="main" BGCOLOR="%s">%s %s</TD></TR>`+"\n", colorForKind(node.Kind), strings.ToLower(string(node.Kind)), node.Name)
	for _, field := range node.Fields {
		if !g.renderBuiltins && astutil.IsBuiltinField(field.Name) {
			continue
		}
		args := field.Arguments
		result += fmt.Sprintf(`    <TR><TD ROWSPAN="%d">%s</TD><TD COLSPAN="2" PORT="%s">%s</TD></TR>`+"\n", len(args)+1, field.Name, portName(field.Name), field.Type.String())
		for _, arg := range args {
			result += fmt.Sprintf(`    <TR><TD>%s</TD><TD PORT="%s">%s</TD></TR>`+"\n", arg.Name, portNameForArgument(field.Name, arg.Name), arg.Type)
		}
	}
	result += "\n  </TABLE>"
	return result
}

func makeInputObjectNodeLabel(node *node) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD COLSPAN="2" PORT="main" BGCOLOR="%s">input %s</TD></TR>`+"\n", colorForKind(node.Kind), node.Name)
	for _, field := range node.Fields {
		result += fmt.Sprintf(`  <TR><TD>%s</TD><TD PORT="%s">%s</TD></TR>`+"\n", field.Name, portName(field.Name), field.Type)
	}
	result += "</TABLE>"
	return result
}

func makeGenericNodeLabel(node *node) string {
	return fmt.Sprintf("%s\n%s", strings.ToLower(string(node.Kind)), node.Name)
}
