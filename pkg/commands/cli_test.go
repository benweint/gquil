package commands

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type TestCaseParams struct {
	Dir                string
	Args               []string `yaml:"args"`
	ExpectJson         bool     `yaml:"expectJson"`
	ExpectedOutput     string
	expectedOutputPath string
}

func TestCli(t *testing.T) {
	var cases []TestCaseParams

	baseDir := "testdata/cases"
	entries, err := os.ReadDir(baseDir)
	assert.NoError(t, err)

	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}

		testCaseDir := path.Join(baseDir, ent.Name())
		metaYamlPath := path.Join(testCaseDir, "meta.yaml")
		metaYamlRaw, err := os.ReadFile(metaYamlPath)
		assert.NoError(t, err)

		params := TestCaseParams{
			Dir: testCaseDir,
		}
		err = yaml.Unmarshal(metaYamlRaw, &params)
		assert.NoError(t, err)

		expectedFilename := "expected.txt"
		if params.ExpectJson {
			expectedFilename = "expected.json"
		}

		expectedOutputPath := path.Join(testCaseDir, expectedFilename)
		params.expectedOutputPath = expectedOutputPath
		expectedOutputRaw, err := os.ReadFile(expectedOutputPath)
		assert.NoError(t, err)
		params.ExpectedOutput = string(expectedOutputRaw)

		cases = append(cases, params)
	}

	for _, tc := range cases {
		t.Run(tc.Dir, func(t *testing.T) {
			parser, err := MakeParser()
			assert.NoError(t, err)

			ctx, err := parser.Parse(tc.Args)
			assert.NoError(t, err)

			var stdoutBuf, stderrBuf, stdinBuf bytes.Buffer

			err = ctx.Run(Context{
				Stdout: &stdoutBuf,
				Stderr: &stderrBuf,
				Stdin:  &stdinBuf,
			})
			assert.NoError(t, err)

			updateExpected := os.Getenv("TEST_UPDATE_EXPECTED")
			if updateExpected != "" {
				err = os.WriteFile(tc.expectedOutputPath, stdoutBuf.Bytes(), 0655)
				assert.NoError(t, err)
			}

			if tc.ExpectJson {
				assert.JSONEq(t, tc.ExpectedOutput, stdoutBuf.String())
			} else {
				assert.Equal(t, tc.ExpectedOutput, stdoutBuf.String())
			}
		})
	}
}
