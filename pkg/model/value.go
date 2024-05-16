package model

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/vektah/gqlparser/v2/ast"
)

type Value any

func makeValue(in *ast.Value) (Value, error) {
	if in == nil {
		return nil, nil
	}

	switch in.Kind {
	case ast.IntValue:
		return strconv.ParseInt(in.Raw, 10, 64)
	case ast.FloatValue:
		return strconv.ParseFloat(in.Raw, 32)
	case ast.StringValue:
		return in.Raw, nil
	case ast.BooleanValue:
		return strconv.ParseBool(in.Raw)
	case ast.NullValue:
		return json.Marshal(nil)
	case ast.EnumValue:
		return json.Marshal(in.Raw)
	case ast.ListValue:
		var l []any
		for _, cv := range in.Children {
			entry, err := makeValue(cv.Value)
			if err != nil {
				return nil, err
			}
			l = append(l, entry)
		}
		return l, nil
	case ast.ObjectValue:
		m := map[string]any{}
		for _, cv := range in.Children {
			val, err := makeValue(cv.Value)
			if err != nil {
				return nil, err
			}
			m[cv.Name] = val
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unsupported kind: %v", in.Kind)
	}
}
