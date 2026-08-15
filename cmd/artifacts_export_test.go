package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// simulateAndExport runs simulate then export for the given output dir and
// returns the export output.
func simulateAndExport(t *testing.T, outputDir string) string {
	t.Helper()

	viper.Reset()
	viper.Set("output", outputDir)
	executeCommand(t, "simulate", "--output", outputDir, "--seed", "42", "--years", "5", "--width", "32", "--height", "32")

	viper.Reset()
	viper.Set("output", outputDir)
	output := executeCommand(t, "export", "--format", "obsidian", "--output", outputDir)
	return output
}

func TestExportArtifactsEndToEnd(t *testing.T) {
	outputDir := t.TempDir()
	output := simulateAndExport(t, outputDir)

	artifactsDir := filepath.Join(outputDir, "artifacts")
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		t.Fatalf("read artifacts dir: %v", err)
	}

	sawIndex := false
	sawArtifact := false
	for _, entry := range entries {
		if entry.Name() == "Index.md" {
			sawIndex = true
		}
		if strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "Index.md" {
			sawArtifact = true
		}
	}
	if !sawIndex {
		t.Errorf("expected artifacts/Index.md after export")
	}
	if !sawArtifact {
		t.Errorf("expected at least one artifact note after export")
	}

	if !strings.Contains(output, "artifacts") {
		t.Errorf("export summary = %q, want it to mention artifacts", output)
	}
}

func TestExportArtifactsEndToEndDeterministic(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	simulateAndExport(t, dirA)
	simulateAndExport(t, dirB)

	compareArtifactDirs(t, filepath.Join(dirA, "artifacts"), filepath.Join(dirB, "artifacts"))
}

func compareArtifactDirs(t *testing.T, dirA, dirB string) {
	t.Helper()

	entriesA, err := os.ReadDir(dirA)
	if err != nil {
		t.Fatalf("read %s: %v", dirA, err)
	}
	entriesB, err := os.ReadDir(dirB)
	if err != nil {
		t.Fatalf("read %s: %v", dirB, err)
	}

	if len(entriesA) != len(entriesB) {
		t.Fatalf("artifact file counts differ: %d vs %d", len(entriesA), len(entriesB))
	}

	for i := range entriesA {
		name := entriesA[i].Name()
		if entriesB[i].Name() != name {
			t.Fatalf("artifact file sets differ: %s vs %s", name, entriesB[i].Name())
		}
		dataA, err := os.ReadFile(filepath.Join(dirA, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		dataB, err := os.ReadFile(filepath.Join(dirB, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !bytes.Equal(dataA, dataB) {
			t.Errorf("artifact file %s differs across same-seed runs", name)
		}
	}
}
