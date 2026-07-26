package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestInitCommandAcknowledgesInitialization(t *testing.T) {
	t.Cleanup(func() { _ = os.Remove("worldgen.yaml") })

	output := executeCommand(t, "init", "--name", "Ashtar", "--size", "large")

	if !strings.Contains(output, "Project initialized: worldgen.yaml") {
		t.Fatalf("init output = %q, want project initialized message", output)
	}
}

func TestSimulateCommandRunsSimulation(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Set("output", tmpDir)
	output := executeCommand(t, "simulate", "--output", tmpDir, "--years", "10", "--events", "normal")

	if !strings.Contains(output, "World generated") || !strings.Contains(output, "Simulation completed successfully") {
		t.Fatalf("simulate output = %q, want simulation execution output", output)
	}

	if !strings.Contains(output, "World state saved to") {
		t.Fatalf("simulate output = %q, want world state saved message", output)
	}

	if !strings.Contains(output, "Timeline saved to") {
		t.Fatalf("simulate output = %q, want timeline saved message", output)
	}

	if !strings.Contains(output, "Chronicle") {
		t.Fatalf("simulate output = %q, want chronicle output", output)
	}
}

func TestExportCommandAcknowledgesDestination(t *testing.T) {
	tmpDir := t.TempDir()

	viper.Set("output", tmpDir)
	executeCommand(t, "simulate", "--output", tmpDir, "--years", "10", "--width", "16", "--height", "16")

	viper.Set("output", tmpDir)
	output := executeCommand(t, "export", "--format", "obsidian", "--output", tmpDir)

	if !strings.Contains(output, "Export complete:") {
		t.Fatalf("export output = %q, want export complete message", output)
	}
}

func TestConfigFlagDoesNotOverrideExplicitFlag(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "worldgen.yaml")
	configContent := "output: /should/not/use/this\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	viper.Set("output", tmpDir)
	executeCommand(t, "simulate", "--config", configPath, "--output", tmpDir, "--years", "10", "--width", "16", "--height", "16")

	if _, err := os.Stat(filepath.Join(tmpDir, "world_state.json")); err != nil {
		t.Fatalf("world_state.json not found in flag-specified output dir (config tried to redirect to /should/not/use/this): %v", err)
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
