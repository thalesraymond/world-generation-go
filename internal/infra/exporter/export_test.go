package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func TestExportCreatesExpectedFilesAndContent(t *testing.T) {
	state := &world.State{
		Width:  100,
		Height: 100,
		Settlements: []world.Settlement{
			{Name: "Riverwatch", X: 10, Y: 20, Faction: " Ironbound", Population: 500},
			{Name: "Oakhaven", X: 30, Y: 40, Faction: "Sylvani", Population: 1200},
		},
	}

	tmpDir, err := os.MkdirTemp("", "exporter-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targetDir := filepath.Join(tmpDir, "vault")

	if err := Export(state, targetDir); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	basesDir := filepath.Join(targetDir, "bases")
	factionsDir := filepath.Join(targetDir, "factions")

	if _, err := os.Stat(basesDir); os.IsNotExist(err) {
		t.Errorf("bases directory does not exist")
	}
	if _, err := os.Stat(factionsDir); os.IsNotExist(err) {
		t.Errorf("factions directory does not exist")
	}

	expectedSettlements := map[string]bool{"Riverwatch.md": false, "Oakhaven.md": false}
	entries, err := os.ReadDir(basesDir)
	if err != nil {
		t.Fatalf("read bases dir: %v", err)
	}
	for _, entry := range entries {
		expectedSettlements[entry.Name()] = true
	}
	for name, found := range expectedSettlements {
		if !found {
			t.Errorf("expected settlement file %s not found", name)
		}
	}

	expectedFactions := map[string]bool{"Ironbound.md": false, "Sylvani.md": false}
	entries, err = os.ReadDir(factionsDir)
	if err != nil {
		t.Fatalf("read factions dir: %v", err)
	}
	for _, entry := range entries {
		expectedFactions[entry.Name()] = true
	}
	for name, found := range expectedFactions {
		if !found {
			t.Errorf("expected faction file %s not found", name)
		}
	}

	riverwatchBytes, err := os.ReadFile(filepath.Join(basesDir, "Riverwatch.md"))
	if err != nil {
		t.Fatalf("read Riverwatch file: %v", err)
	}
	riverwatch := string(riverwatchBytes)

	if !strings.Contains(riverwatch, "---") {
		t.Errorf("settlement file missing frontmatter")
	}
	if !strings.Contains(riverwatch, "type: settlement") {
		t.Errorf("settlement file missing type field")
	}
	if !strings.Contains(riverwatch, "name: Riverwatch") {
		t.Errorf("settlement file missing name field")
	}
	if !strings.Contains(riverwatch, "faction:  Ironbound") {
		t.Errorf("settlement file missing faction field")
	}
	if !strings.Contains(riverwatch, "[[Ironbound]]") {
		t.Errorf("settlement file missing wiki-link to faction")
	}

	oakhavenBytes, err := os.ReadFile(filepath.Join(basesDir, "Oakhaven.md"))
	if err != nil {
		t.Fatalf("read Oakhaven file: %v", err)
	}
	oakhaven := string(oakhavenBytes)

	if !strings.Contains(oakhaven, "faction: Sylvani") {
		t.Errorf("Oakhaven file missing faction field")
	}
	if !strings.Contains(oakhaven, "[[Sylvani]]") {
		t.Errorf("Oakhaven file missing wiki-link to faction")
	}

	sylvaniBytes, err := os.ReadFile(filepath.Join(factionsDir, "Sylvani.md"))
	if err != nil {
		t.Fatalf("read Sylvani file: %v", err)
	}
	sylvani := string(sylvaniBytes)

	if !strings.Contains(sylvani, "type: faction") {
		t.Errorf("faction file missing type field")
	}
	if !strings.Contains(sylvani, "name: Sylvani") {
		t.Errorf("faction file missing name field")
	}
	if !strings.Contains(sylvani, "[[Oakhaven]]") {
		t.Errorf("faction file missing wiki-link to settlement")
	}

	ironboundBytes, err := os.ReadFile(filepath.Join(factionsDir, "Ironbound.md"))
	if err != nil {
		t.Fatalf("read Ironbound file: %v", err)
	}
	ironbound := string(ironboundBytes)

	if !strings.Contains(ironbound, "[[Riverwatch]]") {
		t.Errorf("Ironbound file missing wiki-link to settlement")
	}
}

func TestExportEmptySettlements(t *testing.T) {
	state := &world.State{
		Width:       50,
		Height:      50,
		Settlements: []world.Settlement{},
	}

	tmpDir, err := os.MkdirTemp("", "exporter-empty-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targetDir := filepath.Join(tmpDir, "vault")

	if err := Export(state, targetDir); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	basesDir := filepath.Join(targetDir, "bases")
	factionsDir := filepath.Join(targetDir, "factions")

	if _, err := os.Stat(basesDir); os.IsNotExist(err) {
		t.Errorf("bases directory does not exist")
	}
	if _, err := os.Stat(factionsDir); os.IsNotExist(err) {
		t.Errorf("factions directory does not exist")
	}

	basesEntries, err := os.ReadDir(basesDir)
	if err != nil {
		t.Fatalf("read bases dir: %v", err)
	}
	if len(basesEntries) != 0 {
		t.Errorf("expected empty bases dir, got %d entries", len(basesEntries))
	}

	factionsEntries, err := os.ReadDir(factionsDir)
	if err != nil {
		t.Fatalf("read factions dir: %v", err)
	}
	if len(factionsEntries) != 0 {
		t.Errorf("expected empty factions dir, got %d entries", len(factionsEntries))
	}
}
