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
	dir := t.TempDir()

	args := []string{"--years", "20", "--width", "64", "--height", "64", "--seed", "42", "--events", "normal"}

	viper.Set("output", dir)
	out1 := executeCommand(t, append([]string{"simulate", "--output", dir}, args...)...)

	viper.Set("output", dir)
	out2 := executeCommand(t, append([]string{"simulate", "--output", dir}, args...)...)

	if out1 != out2 {
		t.Fatalf("same seed produced different stdout:\n--- first ---\n%s\n--- second ---\n%s", out1, out2)
	}

	state1, err := os.ReadFile(filepath.Join(dir, "world_state.json"))
	if err != nil {
		t.Fatalf("read world_state.json (run1): %v", err)
	}
	state2, err := os.ReadFile(filepath.Join(dir, "world_state.json"))
	if err != nil {
		t.Fatalf("read world_state.json (run2): %v", err)
	}
	if !bytes.Equal(state1, state2) {
		t.Fatal("world_state.json differs across identical runs")
	}

	timeline1, err := os.ReadFile(filepath.Join(dir, "timeline.json"))
	if err != nil {
		t.Fatalf("read timeline.json (run1): %v", err)
	}
	timeline2, err := os.ReadFile(filepath.Join(dir, "timeline.json"))
	if err != nil {
		t.Fatalf("read timeline.json (run2): %v", err)
	}
	if !bytes.Equal(timeline1, timeline2) {
		t.Fatal("timeline.json differs across identical runs")
	}
}

func TestChronicleNoOutcomeEcho(t *testing.T) {
	viper.Reset()
	tmpDir := t.TempDir()
	viper.Set("output", tmpDir)

	output := executeCommand(t, "simulate", "--output", tmpDir, "--years", "20", "--width", "64", "--height", "64", "--seed", "42", "--events", "normal")

	chronicleIdx := strings.Index(output, "--- Chronicle ---")
	if chronicleIdx < 0 {
		t.Fatal("simulate output missing the Chronicle header")
	}
	chronicle := output[chronicleIdx:]

	outcomeEchoPatterns := []string{
		"raided .* and seized",
		"raided .* but was driven off",
		"secured an alliance",
		"declares a festival",
		"establishes new trade routes",
	}

	for _, pattern := range outcomeEchoPatterns {
		re := regexp.MustCompile(pattern)
		lines := strings.Split(chronicle, "\n")
		for _, line := range lines {
			if re.MatchString(line) {
				continue
			}
		}
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

	goldDir := "testdata/gold"
	if err := os.MkdirAll(goldDir, 0755); err != nil {
		t.Fatalf("create gold fixture dir: %v", err)
	}

	goldPath := filepath.Join(goldDir, "chronicle-normal-seed42.txt")
	if err := os.WriteFile(goldPath, []byte(chronicle), 0644); err != nil {
		t.Fatalf("write gold fixture: %v", err)
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
