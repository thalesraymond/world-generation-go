package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
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

	statePath := filepath.Join(tmpDir, "world_state.json")
	stateData, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("failed to read world_state.json: %v", err)
	}

	var state world.State
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatalf("failed to unmarshal world_state.json: %v", err)
	}

	timelinePath := filepath.Join(tmpDir, "timeline.json")
	timelineData, err := os.ReadFile(timelinePath)
	if err != nil {
		t.Fatalf("failed to read timeline.json: %v", err)
	}

	var events []simulation.Event
	if err := json.Unmarshal(timelineData, &events); err != nil {
		t.Fatalf("failed to unmarshal timeline.json: %v", err)
	}

	stateFigs := make(map[[2]string]struct{ deathYear int })
	for _, s := range state.Settlements {
		for _, fig := range s.Figures {
			stateFigs[[2]string{s.Name, fig.ID}] = struct{ deathYear int }{deathYear: fig.DeathYear}
		}
	}

	for _, e := range events {
		if e.Category != "Death" {
			continue
		}
		key := [2]string{e.SettlementName, e.FigureID}
		fig, ok := stateFigs[key]
		if !ok {
			t.Errorf("death event for missing figure: %v", key)
			continue
		}
		if fig.deathYear == 0 {
			t.Errorf("death event for %v (year %d) but figure deathYear is 0", key, e.Year)
		}
	}
}

func TestSimulateCommandPopulatesSettlementFigures(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Set("output", tmpDir)
	executeCommand(t, "simulate", "--output", tmpDir, "--years", "10", "--width", "48", "--height", "48")

	statePath := filepath.Join(tmpDir, "world_state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("failed to read world_state.json: %v", err)
	}

	state, err := world.FromJSON(data)
	if err != nil {
		t.Fatalf("failed to parse world_state.json: %v", err)
	}

	if len(state.Settlements) == 0 {
		t.Fatalf("expected at least one settlement to test figure generation")
	}

	for _, settlement := range state.Settlements {
		if len(settlement.Figures) == 0 {
			t.Errorf("settlement %q has no figures", settlement.Name)
			continue
		}
		hasLeader := false
		for _, fig := range settlement.Figures {
			if fig.Role == "Leader" {
				hasLeader = true
				break
			}
		}
		if !hasLeader {
			t.Errorf("settlement %q has no Leader figure", settlement.Name)
		}
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

func TestFullPipelineDeterminism(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	viper.Set("output", dir1)
	executeCommand(t, "simulate", "--output", dir1, "--years", "10", "--width", "32", "--height", "32")
	viper.Set("output", dir2)
	executeCommand(t, "simulate", "--output", dir2, "--years", "10", "--width", "32", "--height", "32")

	state1, err := os.ReadFile(filepath.Join(dir1, "world_state.json"))
	if err != nil {
		t.Fatalf("read world_state.json (run1): %v", err)
	}
	state2, err := os.ReadFile(filepath.Join(dir2, "world_state.json"))
	if err != nil {
		t.Fatalf("read world_state.json (run2): %v", err)
	}
	if !bytes.Equal(state1, state2) {
		t.Fatalf("world_state.json differs across identical runs")
	}

	timeline1, err := os.ReadFile(filepath.Join(dir1, "timeline.json"))
	if err != nil {
		t.Fatalf("read timeline.json (run1): %v", err)
	}
	timeline2, err := os.ReadFile(filepath.Join(dir2, "timeline.json"))
	if err != nil {
		t.Fatalf("read timeline.json (run2): %v", err)
	}
	if !bytes.Equal(timeline1, timeline2) {
		t.Fatalf("timeline.json differs across identical runs")
	}

	viper.Set("output", dir1)
	executeCommand(t, "export", "--format", "obsidian", "--output", dir1)
	viper.Set("output", dir2)
	executeCommand(t, "export", "--format", "obsidian", "--output", dir2)

	compareExportDirs(t, dir1, dir2)
}

func compareExportDirs(t *testing.T, dir1, dir2 string) {
	t.Helper()

	subdirs := []string{"bases", "factions", "characters", "chronicles", "pointcrawl"}
	for _, sub := range subdirs {
		path1 := filepath.Join(dir1, sub)
		path2 := filepath.Join(dir2, sub)

		entries1, err := os.ReadDir(path1)
		if err != nil {
			if os.IsNotExist(err) {
				if _, err2 := os.Stat(path2); !os.IsNotExist(err2) {
					t.Errorf("directory %s exists in run2 but not run1", sub)
				}
				continue
			}
			t.Fatalf("read dir %s (run1): %v", sub, err)
		}

		entries2, err := os.ReadDir(path2)
		if err != nil {
			t.Fatalf("read dir %s (run2): %v", sub, err)
		}

		if len(entries1) != len(entries2) {
			t.Errorf("directory %s file count differs: %d vs %d", sub, len(entries1), len(entries2))
			continue
		}

		for i := range entries1 {
			if entries1[i].Name() != entries2[i].Name() {
				t.Errorf("directory %s file names differ: %s vs %s", sub, entries1[i].Name(), entries2[i].Name())
				continue
			}
			data1, err := os.ReadFile(filepath.Join(path1, entries1[i].Name()))
			if err != nil {
				t.Fatalf("read %s/%s (run1): %v", sub, entries1[i].Name(), err)
			}
			data2, err := os.ReadFile(filepath.Join(path2, entries2[i].Name()))
			if err != nil {
				t.Fatalf("read %s/%s (run2): %v", sub, entries2[i].Name(), err)
			}
			if !bytes.Equal(data1, data2) {
				t.Errorf("file %s/%s differs across runs", sub, entries1[i].Name())
			}
		}
	}
}

func executeCommand(t *testing.T, args ...string) string {
	t.Helper()

	output, err := executeCommandErr(t, args...)
	if err != nil {
		t.Fatalf("Execute(%v) returned error: %v", args, err)
	}
	return output
}

func executeCommandErr(t *testing.T, args ...string) (string, error) {
	t.Helper()

	rootCmd := NewRootCommand()
	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	return output.String(), err
}

// rawEventLine matches a FormatEvent line emitted by the verbose preset, e.g.
// "[3] (Economy) Deepcrest prospers." or "[3] (Raid) Deepcrest → Northhold: ...".
var rawEventLine = regexp.MustCompile(`^\[\d+\] (\([^)]*\) )?.*$`)

func TestSimulateCommandSingleChronicleStream(t *testing.T) {
	viper.Reset()
	tmpDir := t.TempDir()
	viper.Set("output", tmpDir)
	output := executeCommand(t, "simulate", "--output", tmpDir, "--years", "5", "--width", "64", "--height", "64")

	chronicleIdx := strings.Index(output, "--- Chronicle ---")
	if chronicleIdx < 0 {
		t.Fatal("simulate output missing the Chronicle header")
	}

	// With the default preset the pass-1 collector no longer prints, so no raw
	// FormatEvent lines may appear anywhere in the output.
	for i, line := range strings.Split(output, "\n") {
		if rawEventLine.MatchString(line) {
			t.Errorf("line %d = %q, want no raw FormatEvent lines in single-stream output", i, line)
		}
	}

	// The narrated stream must carry real chronicle content (settlements with
	// events), proving the Chronicle service ran and produced prose.
	if strings.Count(output[chronicleIdx:], "\n") < 3 {
		t.Errorf("chronicle section too sparse: %q", output[chronicleIdx:])
	}
}

func TestSimulateCommandVerboseEmitsRawLines(t *testing.T) {
	viper.Reset()
	tmpDir := t.TempDir()
	viper.Set("output", tmpDir)
	output := executeCommand(t, "simulate", "--output", tmpDir, "--years", "3", "--width", "64", "--height", "64", "--events", "verbose")

	rawCount := 0
	for _, line := range strings.Split(output, "\n") {
		if rawEventLine.MatchString(line) {
			rawCount++
		}
	}
	if rawCount < 10 {
		t.Fatalf("verbose output should interleave raw FormatEvent lines with narration, got %d raw lines", rawCount)
	}
}

func TestSimulateCommandInvalidPresetSurfacesError(t *testing.T) {
	viper.Reset()
	tmpDir := t.TempDir()
	viper.Set("output", tmpDir)
	_, err := executeCommandErr(t, "simulate", "--output", tmpDir, "--years", "2", "--width", "16", "--height", "16", "--events", "bogus")
	if err == nil {
		t.Fatal("expected an actionable error for an invalid --events preset")
	}
	if !strings.Contains(err.Error(), "bogus") || !strings.Contains(err.Error(), "quiet, normal, or verbose") {
		t.Errorf("error = %v, want it to name the invalid preset and list the accepted values", err)
	}
}

func TestSimulateCommandStdoutDeterminism(t *testing.T) {
	viper.Reset()
	// Both runs share one output directory: the generated files are
	// byte-identical, so the second run overwrites with identical content and
	// the printed paths match.
	dir := t.TempDir()
	run := func() string {
		viper.Set("output", dir)
		return executeCommand(t, "simulate", "--output", dir, "--years", "10", "--width", "64", "--height", "64")
	}

	first := run()
	second := run()
	if first != second {
		t.Fatalf("same seed produced different stdout:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}
