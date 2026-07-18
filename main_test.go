package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/cmd"
)

func TestRootHelpListsCommands(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)
	rootCmd.SetArgs(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	help := output.String()
	for _, fragment := range []string{"Usage:", "init", "simulate", "export"} {
		if !strings.Contains(help, fragment) {
			t.Fatalf("help output missing %q in %q", fragment, help)
		}
	}
}
