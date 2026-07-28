package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func TestMilitaryTier(t *testing.T) {
	cases := []struct {
		value float64
		want  string
	}{
		{0, "Weak"},
		{99.9, "Weak"},
		{100, "Moderate"},
		{299.9, "Moderate"},
		{300, "Strong"},
		{599.9, "Strong"},
		{600, "Mighty"},
		{1000, "Mighty"},
	}
	for _, tc := range cases {
		if got := MilitaryTier(tc.value); got != tc.want {
			t.Errorf("MilitaryTier(%v) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestWealthTier(t *testing.T) {
	cases := []struct {
		value float64
		want  string
	}{
		{0, "Poor"},
		{199.9, "Poor"},
		{200, "Comfortable"},
		{499.9, "Comfortable"},
		{500, "Prosperous"},
		{999.9, "Prosperous"},
		{1000, "Rich"},
		{5000, "Rich"},
	}
	for _, tc := range cases {
		if got := WealthTier(tc.value); got != tc.want {
			t.Errorf("WealthTier(%v) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestExportIncludesAgentStateSections(t *testing.T) {
	state := &world.State{
		Width:  100,
		Height: 100,
		Settlements: []world.Settlement{
			{
				Name:             "Riverwatch",
				Type:             "City",
				X:                10,
				Y:                20,
				Faction:          "Ironbound",
				Population:       5000,
				MilitaryStrength: 500,
				Wealth:           750,
				Goals:            []string{"expand", "grow"},
				Relations: map[string]float64{
					"Oakhaven":  0.65,
					"Blackgate": -0.8,
					"Thornhold": 0.3,
					"Eastmarch": -0.2,
					"Goldfield": 0.1,
					"Westmarch": -0.55,
					"Deepmere":  -0.1,
					"Highspire": 0.05,
				},
			},
			{Name: "Oakhaven", Type: "Town", X: 30, Y: 40, Faction: "Ironbound", Population: 1200},
			{Name: "Blackgate", Type: "Town", X: 50, Y: 60, Faction: "Ashen", Population: 900},
			{Name: "Thornhold", Type: "Town", X: 15, Y: 25, Faction: "Ironbound", Population: 800},
			{Name: "Eastmarch", Type: "Town", X: 12, Y: 22, Faction: "Ashen", Population: 700},
			{Name: "Goldfield", Type: "Town", X: 11, Y: 21, Faction: "Ironbound", Population: 600},
			{Name: "Westmarch", Type: "Town", X: 13, Y: 24, Faction: "Ashen", Population: 500},
			{Name: "Deepmere", Type: "Town", X: 14, Y: 26, Faction: "Ashen", Population: 400},
			{Name: "Highspire", Type: "Town", X: 16, Y: 27, Faction: "Ironbound", Population: 300},
		},
	}

	targetDir := filepath.Join(t.TempDir(), "vault")
	if err := Export(state, targetDir); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "bases", "Riverwatch.md"))
	if err != nil {
		t.Fatalf("read settlement file: %v", err)
	}
	content := string(data)

	checks := []string{
		"## Military Strength\n\n500 (Strong)",
		"## Wealth\n\n750 (Prosperous)",
		"## Relations",
		"### Allies",
		"### Rivals",
		"## Goals",
		"- expand",
		"- grow",
		"[[Oakhaven]] (+0.65)",
		"[[Blackgate]] (-0.80)",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("settlement export missing %q\n--- content ---\n%s", want, content)
		}
	}

	// Allies: exactly the positive relations, highest first, capped at 5.
	alliesIdx := strings.Index(content, "### Allies")
	rivalsIdx := strings.Index(content, "### Rivals")
	if alliesIdx == -1 || rivalsIdx == -1 || rivalsIdx <= alliesIdx {
		t.Fatalf("expected Allies section before Rivals section\n%s", content)
	}
	allies := content[alliesIdx:rivalsIdx]
	if !strings.Contains(allies, "[[Oakhaven]] (+0.65)") ||
		!strings.Contains(allies, "[[Thornhold]] (+0.30)") ||
		!strings.Contains(allies, "[[Goldfield]] (+0.10)") ||
		!strings.Contains(allies, "[[Highspire]] (+0.05)") {
		t.Errorf("allies section missing expected entries\n%s", allies)
	}
	if strings.Contains(allies, "Blackgate") || strings.Contains(allies, "Eastmarch") {
		t.Errorf("allies section must not contain negative relations\n%s", allies)
	}

	// Rivals: exactly the negative relations, most negative first.
	rivals := content[rivalsIdx:]
	if strings.Index(rivals, "[[Blackgate]] (-0.80)") > strings.Index(rivals, "[[Westmarch]] (-0.55)") {
		t.Errorf("rivals must be sorted most negative first\n%s", rivals)
	}
	if strings.Contains(rivals, "Oakhaven") {
		t.Errorf("rivals section must not contain positive relations\n%s", rivals)
	}
}

func TestExportRelationsCappedAtFive(t *testing.T) {
	relations := map[string]float64{
		"A1": 0.9, "A2": 0.8, "A3": 0.7, "A4": 0.6, "A5": 0.5, "A6": 0.4,
		"R1": -0.9, "R2": -0.8, "R3": -0.7, "R4": -0.6, "R5": -0.5, "R6": -0.4,
	}
	settlements := []world.Settlement{
		{Name: "Hub", Type: "City", X: 1, Y: 1, Faction: "F", Population: 1000, Wealth: 100, MilitaryStrength: 100, Relations: relations},
	}
	for name := range relations {
		settlements = append(settlements, world.Settlement{Name: name, Type: "Town", X: 2, Y: 2, Faction: "G", Population: 100})
	}

	state := &world.State{Width: 10, Height: 10, Settlements: settlements}
	targetDir := filepath.Join(t.TempDir(), "vault")
	if err := Export(state, targetDir); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "bases", "Hub.md"))
	if err != nil {
		t.Fatalf("read settlement file: %v", err)
	}
	content := string(data)

	alliesIdx := strings.Index(content, "### Allies")
	rivalsIdx := strings.Index(content, "### Rivals")

	allies := content[alliesIdx:rivalsIdx]
	if strings.Contains(allies, "A6") {
		t.Errorf("allies must be capped at top 5, A6 (0.4) should be excluded\n%s", allies)
	}
	if !strings.Contains(allies, "A5") {
		t.Errorf("allies should include A5 (0.5)\n%s", allies)
	}

	rivals := content[rivalsIdx:]
	if strings.Contains(rivals, "R6") {
		t.Errorf("rivals must be capped at top 5, R6 (-0.4) should be excluded\n%s", rivals)
	}
	if !strings.Contains(rivals, "R5") {
		t.Errorf("rivals should include R5 (-0.5)\n%s", rivals)
	}
}

func TestExportSkipsAgentSectionsWhenNoAgentState(t *testing.T) {
	state := &world.State{
		Width:  10,
		Height: 10,
		Settlements: []world.Settlement{
			{Name: "Oldville", Type: "Town", X: 1, Y: 1, Faction: "F", Population: 100},
		},
	}

	targetDir := filepath.Join(t.TempDir(), "vault")
	if err := Export(state, targetDir); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "bases", "Oldville.md"))
	if err != nil {
		t.Fatalf("read settlement file: %v", err)
	}
	content := string(data)

	for _, absent := range []string{"## Military Strength", "## Wealth", "## Relations", "## Goals"} {
		if strings.Contains(content, absent) {
			t.Errorf("legacy settlement without agent state must not contain %q\n%s", absent, content)
		}
	}
}
