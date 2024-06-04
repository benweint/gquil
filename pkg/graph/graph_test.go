package graph

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/benweint/gquil/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type edgeSpec struct {
	srcType   string
	dstType   string
	fieldName string
	argName   string
}

func (es edgeSpec) String() string {
	return fmt.Sprintf("%s -> %s [%s].%s", es.srcType, es.dstType, es.fieldName, es.argName)
}

func TestReachableFrom(t *testing.T) {
	for _, tc := range []struct {
		name           string
		schema         string
		roots          []string
		expectedNodes  []string
		expectedFields []string
		expectedEdges  []edgeSpec
		maxDepth       int
	}{
		{
			name: "single field root",
			schema: `type Query {
				alpha: Alpha
				beta: Beta
			}

			type Alpha {
				name: String
			}

			type Beta {
				name: String
			}`,
			roots:          []string{"Query.alpha"},
			expectedNodes:  []string{"Alpha", "Query"},
			expectedFields: []string{"Alpha.name", "Query.alpha"},
			expectedEdges: []edgeSpec{
				{
					srcType:   "Query",
					dstType:   "Alpha",
					fieldName: "alpha",
				},
			},
		},
		{
			name: "multiple field roots",
			schema: `type Query {
				alpha: Alpha
				beta: Beta
				gaga: Gaga
			}

			type Alpha {
				name: String
			}

			type Beta {
				name: String
			}
			
			type Gaga {
				name: String
			}`,
			roots:          []string{"Query.alpha", "Query.beta"},
			expectedNodes:  []string{"Alpha", "Beta", "Query"},
			expectedFields: []string{"Alpha.name", "Beta.name", "Query.alpha", "Query.beta"},
			expectedEdges: []edgeSpec{
				{
					srcType:   "Query",
					dstType:   "Alpha",
					fieldName: "alpha",
				},
				{
					srcType:   "Query",
					dstType:   "Beta",
					fieldName: "beta",
				},
			},
		},
		{
			name: "root field with cycle",
			schema: `type Query {
				person(name: String): Person
				organization(name: String): Organization
			}
			
			type Person {
				name: String
				friends: [Person]
			}
			
			type Organization {
				name: String
			}`,
			roots:          []string{"Person.friends"},
			expectedNodes:  []string{"Person"},
			expectedFields: []string{"Person.friends", "Person.name"},
			expectedEdges: []edgeSpec{
				{
					srcType:   "Person",
					dstType:   "Person",
					fieldName: "friends",
				},
			},
		},
		{
			name: "unions",
			schema: `type Query {
				subject(name: String): Subject
				events: [Event]
			}
			
			union Subject = Person | Organization
			
			type Person {
				name: String
			}
			
			type Organization {
				name: String
			}
			
			type Event {
				title: String
			}`,
			roots:          []string{"Query.subject"},
			expectedNodes:  []string{"Organization", "Person", "Query", "Subject"},
			expectedFields: []string{"Organization.name", "Person.name", "Query.subject"},
			expectedEdges: []edgeSpec{
				{
					srcType:   "Query",
					dstType:   "Subject",
					fieldName: "subject",
				},
				{
					srcType: "Subject",
					dstType: "Organization",
				},
				{
					srcType: "Subject",
					dstType: "Person",
				},
			},
		},
		{
			name: "depth limited",
			schema: `type Query {
				persons(filter: PersonFilter): [Person]
				foods: [Food]
			}
			
			input PersonFilter {
				nameLike: String
				matchMode: MatchMode
			}
			
			enum MatchMode {
				CASE_SENSITIVE
				CASE_INSENSITIVE
			}
			
			type Person {
				name: String
				favoriteFoods: [Food]
			}
			
			type Food {
				name: String
			}`,
			roots:         []string{"Query.persons"},
			maxDepth:      2,
			expectedNodes: []string{"Person", "PersonFilter", "Query"},
			expectedFields: []string{
				"Person.favoriteFoods",
				"Person.name",
				"PersonFilter.matchMode",
				"PersonFilter.nameLike",
				"Query.persons",
			},
			expectedEdges: []edgeSpec{
				{
					srcType:   "Query",
					dstType:   "Person",
					fieldName: "persons",
				},
				{
					srcType:   "Query",
					dstType:   "PersonFilter",
					fieldName: "persons",
					argName:   "filter",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := ast.Source{
				Name:  "testcase",
				Input: tc.schema,
			}
			rawSchema, err := gqlparser.LoadSchema(&src)
			assert.NoError(t, err)

			s, err := model.MakeSchema(rawSchema)
			assert.NoError(t, err)

			roots, err := s.ResolveNames(tc.roots)
			assert.NoError(t, err)

			g := MakeGraph(s)
			trimmed := g.ReachableFrom(roots, tc.maxDepth)

			var actualNodes []string
			var actualEdges []edgeSpec
			var actualFields []string

			for _, node := range trimmed.nodes {
				actualNodes = append(actualNodes, node.Name)
				for _, field := range node.Fields {
					fieldId := node.Name + "." + field.Name
					actualFields = append(actualFields, fieldId)
				}
			}

			sort.Strings(actualNodes)
			sort.Strings(actualFields)

			assert.Equal(t, tc.expectedNodes, actualNodes)
			assert.Equal(t, tc.expectedFields, actualFields)

			for _, edges := range trimmed.edges {
				for _, edge := range edges {
					fieldName := ""
					if edge.field != nil {
						fieldName = edge.field.Name
					}
					argName := ""
					if edge.argument != nil {
						argName = edge.argument.Name
					}
					actualEdge := edgeSpec{
						srcType:   edge.src.Name,
						dstType:   edge.dst.Name,
						fieldName: fieldName,
						argName:   argName,
					}
					actualEdges = append(actualEdges, actualEdge)
				}
			}

			sort.Slice(actualEdges, func(i, j int) bool {
				return strings.Compare(actualEdges[i].String(), actualEdges[j].String()) < 0
			})

			assert.Equal(t, tc.expectedEdges, actualEdges)
		})
	}
}
