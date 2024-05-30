package graph

import (
	"fmt"
	"sort"
	"testing"

	"github.com/benweint/gquil/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestApplyFieldFilters(t *testing.T) {
	exampleDefs := model.DefinitionList{
		{
			Name: "Alpha",
			Kind: ast.Object,
			Fields: model.FieldDefinitionList{
				{
					Name: "x",
				},
				{
					Name: "y",
				},
				{
					Name: "z",
				},
			},
		},
		{
			Name: "Beta",
			Kind: ast.Object,
			Fields: model.FieldDefinitionList{
				{
					Name: "z",
				},
				{
					Name: "a",
				},
			},
		},
		{
			Name: "Theta",
			Kind: ast.Object,
			Fields: model.FieldDefinitionList{
				{
					Name: "x",
				},
				{
					Name: "y",
				},
			},
		},
		{
			Name: "InputA",
			Kind: ast.InputObject,
			Fields: model.FieldDefinitionList{
				{
					Name: "a",
				},
				{
					Name: "b",
				},
			},
		},
	}

	for _, tc := range []struct {
		name           string
		defs           model.DefinitionList
		roots          []string
		expectedFields []string
	}{
		{
			name:  "single type",
			defs:  exampleDefs,
			roots: []string{"Alpha"},
			expectedFields: []string{
				"Alpha.x",
				"Alpha.y",
				"Alpha.z",
			},
		},
		{
			name:  "multiple types",
			defs:  exampleDefs,
			roots: []string{"Alpha", "Beta"},
			expectedFields: []string{
				"Alpha.x",
				"Alpha.y",
				"Alpha.z",
				"Beta.a",
				"Beta.z",
			},
		},
		{
			name:  "single field on multiple types",
			defs:  exampleDefs,
			roots: []string{"Alpha.x", "Beta.z"},
			expectedFields: []string{
				"Alpha.x",
				"Beta.z",
			},
		},
		{
			name: "multiple fields on the same type",
			defs: exampleDefs,
			roots: []string{
				"Alpha.x",
				"Alpha.z",
			},
			expectedFields: []string{
				"Alpha.x",
				"Alpha.z",
			},
		},
		{
			name: "field on type plus whole type",
			defs: exampleDefs,
			roots: []string{
				"Alpha.x",
				"Alpha",
			},
			expectedFields: []string{
				"Alpha.x",
				"Alpha.y",
				"Alpha.z",
			},
		},
		{
			name: "input field",
			defs: exampleDefs,
			roots: []string{
				"InputA.a",
			},
			expectedFields: []string{
				"InputA.a",
			},
		},
		{
			name: "input type",
			defs: exampleDefs,
			roots: []string{
				"InputA",
			},
			expectedFields: []string{
				"InputA.a",
				"InputA.b",
			},
		},
		{
			name: "type and then field",
			defs: exampleDefs,
			roots: []string{
				"Alpha",
				"Alpha.x",
			},
			expectedFields: []string{
				"Alpha.x",
				"Alpha.y",
				"Alpha.z",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &model.Schema{
				Types: exampleDefs,
			}
			roots, err := s.ResolveNames(tc.roots)
			assert.NoError(t, err)
			filtered := applyFieldFilters(exampleDefs, roots)

			var actualFieldNames []string
			for _, def := range filtered {
				for _, field := range def.Fields {
					fieldName := fmt.Sprintf("%s.%s", def.Name, field.Name)
					actualFieldNames = append(actualFieldNames, fieldName)
				}
			}
			sort.Strings(actualFieldNames)

			assert.Equal(t, tc.expectedFields, actualFieldNames)
		})
	}
}
