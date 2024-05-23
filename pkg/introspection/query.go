package introspection

import (
	"bytes"
	"text/template"
)

// GetQuery returns a GraphQL introspection query that is compatible with the given version
// of the GraphQL spec.
func GetQuery(sv SpecVersion) string {
	t, err := template.New("QueryTemplate").Parse(queryTemplate)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, sv); err != nil {
		panic(err)
	}

	return buf.String()
}

const queryTemplate = `
query IntrospectionQuery {
  __schema {
    {{ if eq .HasSchemaDescription true }}description{{end}}
    queryType {
      name
    }
    mutationType {
      name
    }
    subscriptionType {
      name
    }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
      {{ if eq .HasIsRepeatable true }}isRepeatable{{end}}
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
    isDeprecated
    deprecationReason
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes {
    ...TypeRef
  }
  {{ if eq .HasSpecifiedByURL true }}specifiedByURL{{end}}
}

fragment InputValue on __InputValue {
  name
  description
  type {
    ...TypeRef
  }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
      }
    }
  }
}
`
