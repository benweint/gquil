package introspection

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/formatter"
)

func TestResponseToAst(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		response Schema
		expected string
	}{
		{
			name: "compound default values",
			response: Schema{
				Types: []Type{
					{
						Kind:        "OBJECT",
						Name:        "Query",
						Description: "The root query object",
						Fields: []Field{
							{
								Name:        "fruits",
								Description: "Get fruits",
								Args: []InputValue{
									{
										Name:        "name",
										Description: "Filter returned fruits by name",
										Type: &Type{
											Kind: ScalarKind,
											Name: "String",
										},
									},
									{
										Name: "orderBy",
										Type: &Type{
											Kind: InputObjectKind,
											Name: "FruitsOrderBy",
										},
										DefaultValue: stringp("{ direction: ASC, field: NAME }"),
									},
								},
								Type: &Type{
									Kind: ListKind,
									OfType: &Type{
										Kind: ObjectKind,
										Name: "Fruit",
									},
								},
							},
						},
					},
					{
						Kind: "OBJECT",
						Name: "Fruit",
						Fields: []Field{
							{
								Name: "name",
								Type: &Type{
									Kind: ScalarKind,
									Name: "String",
								},
							},
						},
					},
					{
						Kind: "INPUT_OBJECT",
						Name: "FruitsOrderBy",
						InputFields: []InputValue{
							{
								Name: "direction",
								Type: &Type{
									Kind: EnumKind,
									Name: "FruitsOrderByDirection",
								},
							},
							{
								Name: "field",
								Type: &Type{
									Kind: EnumKind,
									Name: "FruitsOrderByField",
								},
							},
						},
					},
					{
						Kind:        "ENUM",
						Name:        "FruitsOrderByDirection",
						Description: "Which direction to order the fruits by",
						EnumValues: []EnumValue{
							{
								Name:        "ASC",
								Description: "Ascending",
							},
							{
								Name:        "DESC",
								Description: "Descending",
							},
						},
					},
					{
						Kind: "ENUM",
						Name: "FruitsOrderByField",
						EnumValues: []EnumValue{
							{
								Name:        "NAME",
								Description: "Name",
							},
						},
					},
				},
			},
			expected: `type Fruit {
	name: String
}
input FruitsOrderBy {
	direction: FruitsOrderByDirection
	field: FruitsOrderByField
}
"""
Which direction to order the fruits by
"""
enum FruitsOrderByDirection {
	"""
	Ascending
	"""
	ASC
	"""
	Descending
	"""
	DESC
}
enum FruitsOrderByField {
	"""
	Name
	"""
	NAME
}
"""
The root query object
"""
type Query {
	"""
	Get fruits
	"""
	fruits(
		"""
		Filter returned fruits by name
		"""
		name: String
	orderBy: FruitsOrderBy = {direction:ASC,field:NAME}): [Fruit]
}
`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			s, err := responseToAst(&testCase.response)
			assert.NoError(t, err)

			var buf bytes.Buffer
			f := formatter.NewFormatter(&buf)
			f.FormatSchema(s)

			actual := buf.String()
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func stringp(s string) *string {
	return &s
}
