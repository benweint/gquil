package commands

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderParsing(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		values    []string
		wantError bool
		expected  http.Header
	}{
		{
			name: "single-valued header",
			values: []string{
				"foo: bar",
			},
			expected: http.Header{
				"Foo": []string{"bar"},
			},
		},
		{
			name: "multiple headers",
			values: []string{
				"foo: bar",
				"baz: qux",
			},
			expected: http.Header{
				"Foo": []string{"bar"},
				"Baz": []string{"qux"},
			},
		},
		{
			name: "single header with multiple values",
			values: []string{
				"foo: bar",
				"foo:baz",
			},
			expected: http.Header{
				"Foo": []string{"bar", "baz"},
			},
		},
		{
			name: "malformed",
			values: []string{
				"foo",
			},
			wantError: true,
		},
		{
			name: "from file",
			values: []string{
				"@testdata/headers.txt",
			},
			expected: http.Header{
				"Foo": []string{"bar"},
				"Baz": []string{"qux"},
			},
		},
		{
			name: "nonexistent file",
			values: []string{
				"@notreal.txt",
			},
			wantError: true,
		},
		{
			name: "file and inline",
			values: []string{
				"@testdata/headers.txt",
				"baz: bop",
				"other: one",
			},
			expected: http.Header{
				"Foo":   []string{"bar"},
				"Baz":   []string{"qux", "bop"},
				"Other": []string{"one"},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := parseHeaders(testCase.values)
			if testCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, result)
			}
		})
	}
}
