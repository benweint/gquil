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
	nodes              model.DefinitionMap
	edges              map[string][]*edge
	interfacesAsUnions bool
	renderBuiltins     bool
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

func MakeGraph(s *model.Schema, opts ...GraphOption) *Graph {
	g := &Graph{
		nodes: s.Types,
		edges: map[string][]*edge{},
	}

	for _, opt := range opts {
		opt(g)
	}

	for _, t := range s.Types {
		var typeEdges []*edge
		kind := normalizeKind(t.Kind, g.interfacesAsUnions)
		switch kind {
		case ast.Object, ast.InputObject:
			typeEdges = g.makeFieldEdges(t)
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

func (g *Graph) GetDefinitions() model.DefinitionMap {
	return g.nodes
}

func (g *Graph) ReachableFrom(roots []*model.NameReference, maxDepth int) *Graph {
	seen := referenceSet{}

	var traverse func(n *model.Definition, depth int)

	traverseField := func(typeName string, f *model.FieldDefinition, depth int) {
		key := fieldRef(typeName, f.Name)
		if seen[key] {
			return
		}
		seen[key] = true

		if maxDepth > 0 && depth > maxDepth {
			return
		}

		for _, arg := range f.Arguments {
			argType := arg.Type.Unwrap()
			traverse(g.nodes[argType.Name], depth+1)
		}

		underlyingType := f.Type.Unwrap()
		traverse(g.nodes[underlyingType.Name], depth+1)
	}

	traverse = func(n *model.Definition, depth int) {
		if maxDepth > 0 && depth > maxDepth {
			return
		}
		key := typeRef(n.Name)
		if _, ok := seen[key]; ok {
			return
		}
		if n.Kind == ast.Scalar {
			return
		}

		seen[key] = true
		kind := normalizeKind(n.Kind, g.interfacesAsUnions)

		switch kind {
		case ast.Object, ast.InputObject:
			for _, field := range n.Fields {
				traverseField(n.Name, field, depth)
			}
		case ast.Union:
			for _, pt := range n.PossibleTypes {
				traverse(g.nodes[pt], depth+1)
			}
		}
	}

	for _, root := range roots {
		targetType := root.GetTargetType()
		if fieldName := root.GetFieldName(); fieldName != "" {
			traverseField(targetType.Name, targetType.Fields.Named(fieldName), 1)
		} else {
			traverse(targetType, 1)
		}
	}

	filteredNodes := model.DefinitionMap{}
	for name, node := range g.nodes {
		if seen.includesType(name) {
			filteredNodes[name] = seen.filterFields(node)
		}
	}

	filteredEdges := map[string][]*edge{}
	for from, edges := range g.edges {
		var filtered []*edge
		for _, edge := range edges {
			_, srcPresent := filteredNodes[edge.src.Name]
			_, dstPresent := filteredNodes[edge.dst.Name]
			if srcPresent && dstPresent {
				filtered = append(filtered, edge)
			}
		}
		if len(filtered) > 0 {
			filteredEdges[from] = filtered
		}
	}

	return &Graph{
		nodes:              filteredNodes,
		edges:              filteredEdges,
		interfacesAsUnions: g.interfacesAsUnions,
		renderBuiltins:     g.renderBuiltins,
	}
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
		nodeDef := fmt.Sprintf("  %s [shape=plain, label=<%s>]", nodeID(node), g.makeNodeLabel(node))
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
			case edgeKindField:
				srcPortSuffix = ":" + portName(edge.field.Name)
			case edgeKindArgument:
				srcPortSuffix = ":" + portNameForArgument(edge.field.Name, edge.argument.Name)
			case edgeKindPossibleType:
				srcPortSuffix = ":" + portName(edge.possibleType)
			}

			result = append(result, fmt.Sprintf("  %s%s -> %s%s", nodeID(edge.src), srcPortSuffix, nodeID(edge.dst), dstPortSuffix))
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

func (g *Graph) makeNodeLabel(node *model.Definition) string {
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

func nodeID(n *model.Definition) string {
	return "n_" + n.Name
}

func portName(fieldName string) string {
	return "p_" + fieldName
}

func portNameForArgument(fieldName, argName string) string {
	return "p_" + fieldName + "_" + argName
}

func makeEnumLabel(node *model.Definition) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD PORT="main" BGCOLOR="%s">enum %s</TD></TR>`, colorForKind(node.Kind), node.Name)
	for _, val := range node.EnumValues {
		result += fmt.Sprintf(`  <TR><TD>%s</TD></TR>\n`, val.Name)
	}
	result += "</TABLE>"
	return result
}

func makePolymorphicLabel(node *model.Definition) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD PORT="main" BGCOLOR="%s">%s %s</TD></TR>`, colorForKind(node.Kind), strings.ToLower(string(node.Kind)), node.Name)
	for _, possibleType := range node.PossibleTypes {
		result += fmt.Sprintf(`  <TR><TD PORT="%s">%s</TD></TR>\n`, portName(possibleType), possibleType)
	}
	result += "</TABLE>"
	return result
}

func (g *Graph) makeFieldTableNodeLabel(node *model.Definition) string {
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

func makeInputObjectNodeLabel(node *model.Definition) string {
	result := "<TABLE>\n"
	result += fmt.Sprintf(`  <TR><TD COLSPAN="2" PORT="main" BGCOLOR="%s">input %s</TD></TR>`+"\n", colorForKind(node.Kind), node.Name)
	for _, field := range node.Fields {
		result += fmt.Sprintf(`  <TR><TD>%s</TD><TD PORT="%s">%s</TD></TR>`+"\n", field.Name, portName(field.Name), field.Type)
	}
	result += "</TABLE>"
	return result
}

func makeGenericNodeLabel(node *model.Definition) string {
	return fmt.Sprintf("%s\n%s", strings.ToLower(string(node.Kind)), node.Name)
}
