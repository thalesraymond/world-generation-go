package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestInitCommandAcknowledgesInitialization(t *testing.T) {
	output := executeCommand(t, "init", "--name", "Ashtar", "--size", "large")

	if !strings.Contains(output, "Initialization acknowledged") {
		t.Fatalf("init output = %q, want acknowledgement", output)
	}
}

func TestSimulateCommandRunsSimulation(t *testing.T) {
	output := executeCommand(t, "simulate", "--years", "10", "--events", "normal")

	if !strings.Contains(output, "World generated") || !strings.Contains(output, "Simulation completed successfully") {
		t.Fatalf("simulate output = %q, want simulation execution output", output)
	}
}

func TestExportCommandAcknowledgesDestination(t *testing.T) {
	output := executeCommand(t, "export", "--format", "obsidian", "--output", "./vault")

	if !strings.Contains(output, "./vault") {
		t.Fatalf("export output = %q, want output path", output)
	}
}

func executeCommand(t *testing.T, args ...string) string {
	t.Helper()

	rootCmd := NewRootCommand()
	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute(%v) returned error: %v", args, err)
	}

	return output.String()
}
