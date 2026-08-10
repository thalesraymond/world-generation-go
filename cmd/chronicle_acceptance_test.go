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
)

func TestChronicleAcceptanceInvariants(t *testing.T) {
	seed := "42"
	years := "20"
	width := "64"
	height := "64"

	run := func(preset string) string {
		viper.Reset()
		tmpDir := t.TempDir()
		viper.Set("output", tmpDir)
		return executeCommand(t, "simulate", "--output", tmpDir, "--years", years, "--width", width, "--height", height, "--seed", seed, "--events", preset)
	}

	quietOut := run("quiet")
	normalOut := run("normal")
	verboseOut := run("verbose")

	quietLines := countChronicleLines(quietOut)
	normalLines := countChronicleLines(normalOut)
	verboseLines := countChronicleLines(verboseOut)

	if quietLines > normalLines {
		t.Errorf("quiet line count (%d) > normal (%d), want quiet ≤ normal", quietLines, normalLines)
	}
	if normalLines > verboseLines {
		t.Errorf("normal line count (%d) > verbose (%d), want normal ≤ verbose", normalLines, verboseLines)
	}

	for _, tc := range []struct {
		name string
		out  string
	}{
		{"quiet", quietOut},
		{"normal", normalOut},
		{"verbose", verboseOut},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertNoRawVariableLeaks(t, tc.out)
			assertNoEmptyRoleArtifacts(t, tc.out)
			assertNoYear0Prose(t, tc.out)
		})
	}

	if !strings.Contains(verboseOut, "[") || !rawEventLine.MatchString(extractRawLine(verboseOut)) {
		t.Error("verbose output should contain raw FormatEvent lines")
	}
}

func TestChronicleDeterminism(t *testing.T) {
	viper.Reset()
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	args := []string{"--years", "20", "--width", "64", "--height", "64", "--seed", "42", "--events", "normal"}

	viper.Set("output", dir1)
	out1 := executeCommand(t, append([]string{"simulate", "--output", dir1}, args...)...)

	viper.Set("output", dir2)
	out2 := executeCommand(t, append([]string{"simulate", "--output", dir2}, args...)...)

	chronicleIdx1 := strings.Index(out1, "--- Chronicle ---")
	chronicleIdx2 := strings.Index(out2, "--- Chronicle ---")
	if chronicleIdx1 < 0 || chronicleIdx2 < 0 {
		t.Fatal("simulate output missing the Chronicle header")
	}
	chronicle1 := out1[chronicleIdx1:]
	chronicle2 := out2[chronicleIdx2:]
	if chronicle1 != chronicle2 {
		t.Fatalf("same seed produced different chronicle:\n--- first ---\n%s\n--- second ---\n%s", chronicle1, chronicle2)
	}

	state1, err := os.ReadFile(filepath.Join(dir1, "world_state.json"))
	if err != nil {
		t.Fatalf("read world_state.json (run1): %v", err)
	}
	state2, err := os.ReadFile(filepath.Join(dir2, "world_state.json"))
	if err != nil {
		t.Fatalf("read world_state.json (run2): %v", err)
	}
	if !bytes.Equal(state1, state2) {
		t.Fatal("world_state.json differs across identical runs")
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
		t.Fatal("timeline.json differs across identical runs")
	}
}



func TestGoldFixtureChronicle(t *testing.T) {
	viper.Reset()
	tmpDir := t.TempDir()
	viper.Set("output", tmpDir)

	output := executeCommand(t, "simulate", "--output", tmpDir, "--years", "20", "--width", "64", "--height", "64", "--seed", "42", "--events", "normal")

	chronicleIdx := strings.Index(output, "--- Chronicle ---")
	if chronicleIdx < 0 {
		t.Fatal("simulate output missing the Chronicle header")
	}
	chronicle := output[chronicleIdx:]

	goldPath := filepath.Join("testdata", "gold", "chronicle-normal-seed42.txt")
	want, err := os.ReadFile(goldPath)
	if err != nil {
		t.Fatalf("read gold fixture: %v", err)
	}
	if string(want) != chronicle {
		t.Errorf("chronicle output differs from gold fixture\n--- got ---\n%s--- want ---\n%s", chronicle, string(want))
	}

	timelinePath := filepath.Join(tmpDir, "timeline.json")
	timelineData, err := os.ReadFile(timelinePath)
	if err != nil {
		t.Fatalf("read timeline.json: %v", err)
	}

	var events []simulation.Event
	if err := json.Unmarshal(timelineData, &events); err != nil {
		t.Fatalf("unmarshal timeline.json: %v", err)
	}

	for _, e := range events {
		if e.Year == 0 {
			t.Errorf("timeline contains year-0 event: %v", e)
		}
	}
}

func countChronicleLines(output string) int {
	chronicleIdx := strings.Index(output, "--- Chronicle ---")
	if chronicleIdx < 0 {
		return 0
	}
	chronicle := output[chronicleIdx:]
	lines := strings.Split(chronicle, "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "---") {
			count++
		}
	}
	return count
}

func extractRawLine(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if rawEventLine.MatchString(line) {
			return line
		}
	}
	return ""
}

func assertNoRawVariableLeaks(t *testing.T, output string) {
	t.Helper()
	leakPatterns := []string{
		`\$[A-Z][a-zA-Z]*`,
		`\$TargetSettlement`,
		`\$Outcome`,
		`\$ActionType`,
		`\$Amount`,
		`\$year`,
		`\$description`,
	}

	chronicleIdx := strings.Index(output, "--- Chronicle ---")
	if chronicleIdx < 0 {
		return
	}
	chronicle := output[chronicleIdx:]

	for _, pattern := range leakPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindAllString(chronicle, -1); len(matches) > 0 {
			t.Errorf("raw variable leak found: %q in chronicle", matches[0])
		}
	}
}

func assertNoEmptyRoleArtifacts(t *testing.T, output string) {
	t.Helper()
	emptyRolePatterns := []string{
		`the\s+of\s`,
		`the\s{2,}of`,
		`\s{2,}of\s`,
	}

	chronicleIdx := strings.Index(output, "--- Chronicle ---")
	if chronicleIdx < 0 {
		return
	}
	chronicle := output[chronicleIdx:]

	for _, pattern := range emptyRolePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindAllString(chronicle, -1); len(matches) > 0 {
			t.Errorf("empty-role artifact found: %q in chronicle", matches[0])
		}
	}
}

func assertNoYear0Prose(t *testing.T, output string) {
	t.Helper()
	year0Patterns := []string{
		`\bin\s+0\b`,
		`\byear\s+0\b`,
		`\bduring\s+0\b`,
		`\bin\s+the\s+year\s+0\b`,
	}

	chronicleIdx := strings.Index(output, "--- Chronicle ---")
	if chronicleIdx < 0 {
		return
	}
	chronicle := output[chronicleIdx:]

	for _, pattern := range year0Patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindAllString(chronicle, -1); len(matches) > 0 {
			t.Errorf("year-0 prose found: %q in chronicle", matches[0])
		}
	}
}
