package commands

import (
	"encoding/json"
	"fmt"
	"io"
)

type Context struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func (c Context) Print(s string) {
	_, _ = c.Stdout.Write([]byte(s))
}

func (c Context) Printf(fs string, args ...any) {
	_, _ = fmt.Fprintf(c.Stdout, fs, args...)
}

func (c Context) PrintJson(v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = c.Stdout.Write(out)
	return err
}
