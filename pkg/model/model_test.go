package model

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestJSONSerialization(t *testing.T) {
	root := "testdata/cases"
	entries, err := os.ReadDir(root)
	assert.NoError(t, err)

	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}

		testDir := path.Join(root, ent.Name())
		tc := testCase{dir: testDir}
		t.Run(ent.Name(), tc.run)
	}
}

type testCase struct {
	dir string
}

func (tc *testCase) run(t *testing.T) {
	inputPath := path.Join(tc.dir, "in.graphql")
	expectedPath := path.Join(tc.dir, "expected.json")

	rawInput, err := os.ReadFile(inputPath)
	assert.NoError(t, err)

	src := ast.Source{
		Name:  "input",
		Input: string(rawInput),
	}
	s, err := gqlparser.LoadSchema(&src)
	assert.NoError(t, err)

	ss, err := MakeSchema(s)
	assert.NoError(t, err)

	ss.Types = filterBuiltinTypesAndFields(ss.Types)
	ss.Directives = filterBuiltinDirectives(ss.Directives)

	actual, err := json.MarshalIndent(ss, "", "  ")
	assert.NoError(t, err)

	updateExpected := os.Getenv("TEST_UPDATE_EXPECTED") != ""
	if updateExpected {
		err = os.WriteFile(expectedPath, actual, 0644)
		assert.NoError(t, err)
	}

	expected, err := os.ReadFile(expectedPath)
	assert.NoError(t, err)

	assert.JSONEq(t, string(expected), string(actual))
}
