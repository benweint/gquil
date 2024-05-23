package introspection

import (
	"fmt"
	"strings"
)

// SpecVersion represents a version of the GraphQL specification.
// Versions are listed at https://spec.graphql.org/
type SpecVersion struct {
	name                 string
	HasSpecifiedByURL    bool
	HasIsRepeatable      bool
	HasSchemaDescription bool
}

var specVersions = map[string]SpecVersion{
	"june2018": {
		name: "june2018",
	},
	"october2021": {
		name:                 "october2021",
		HasSpecifiedByURL:    true,
		HasIsRepeatable:      true,
		HasSchemaDescription: true,
	},
}

func ParseSpecVersion(raw string) (SpecVersion, error) {
	sv, ok := specVersions[raw]
	if !ok {
		return SpecVersion{}, fmt.Errorf("invalid spec version '%s', known versions are %s", raw, strings.Join(knownVersions(), ", "))
	}

	return sv, nil
}

func knownVersions() []string {
	var result []string
	for name := range specVersions {
		result = append(result, name)
	}
	return result
}
