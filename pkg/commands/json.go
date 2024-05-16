package commands

import (
	"encoding/json"
	"fmt"
)

type JsonCmd struct {
	CommonOptions
}

func (c *JsonCmd) Run() error {
	s, err := loadSchemaModel(c.SchemaFiles)
	if err != nil {
		return err
	}

	if !c.IncludeBuiltins {
		s.FilterBuiltins()
	}

	out, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to serialize schema to JSON: %w", err)
	}

	fmt.Print(string(out))

	return nil
}
