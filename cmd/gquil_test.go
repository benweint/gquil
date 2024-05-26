package main_test

import (
	"os"
	"testing"

	"github.com/benweint/gquil/pkg/commands"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"gquil": commands.Main,
	}))
}

func TestGquil(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}
