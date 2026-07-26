package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func TestExportCreatesExpectedFilesAndContent(t *testing.T) {
	state := &world.State{
		Width:  100,
		Height: 100,
		Settlements: []world.Settlement{
			{Name: "Riverwatch", Type: "Town", X: 10, Y: 20, Faction: " Ironbound", Population: 500},
			{Name: "Oakhaven", Type: "City", X: 30, Y: 40, Faction: "Sylvani", Population: 1200},
		},
	}

	tmpDir, err := os.MkdirTemp("", "exporter-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

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
	if !strings.Contains(riverwatch, "subtype: Town") {
		t.Errorf("settlement file missing subtype field")
	}
	if !strings.Contains(riverwatch, "**Type:** Town") {
		t.Errorf("settlement file missing Type line in body")
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

	if !strings.Contains(oakhaven, "subtype: City") {
		t.Errorf("Oakhaven file missing subtype field")
	}
	if !strings.Contains(oakhaven, "**Type:** City") {
		t.Errorf("Oakhaven file missing Type line in body")
	}
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
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

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

func TestExportPointcrawlCreatesFiles(t *testing.T) {
	state := &world.State{Width: 10, Height: 10}
	graph := pointcrawl.NewGraph()
	graph.AddNode(&pointcrawl.Node{ID: 0, X: 1, Y: 1, Name: "Testville", Kind: "settlement", Visibility: pointcrawl.Known})
	graph.AddNode(&pointcrawl.Node{ID: 1, X: 5, Y: 5, Name: "WildPlace", Kind: "wilderness", Visibility: pointcrawl.Unknown})
	graph.AddEdge(0, 1, 3)
	graph.AddEdge(1, 0, 3)
	state.PointcrawlGraph = graph

	tmpDir := t.TempDir()
	err := ExportPointcrawl(state, tmpDir)
	if err != nil {
		t.Fatalf("ExportPointcrawl error: %v", err)
	}

	networkPath := filepath.Join(tmpDir, "pointcrawl", "Network.md")
	if _, err := os.Stat(networkPath); os.IsNotExist(err) {
		t.Fatalf("expected Network.md to exist")
	}

	nodePath := filepath.Join(tmpDir, "pointcrawl", "Testville.md")
	if _, err := os.Stat(nodePath); os.IsNotExist(err) {
		t.Fatalf("expected Testville.md to exist")
	}

	wildPath := filepath.Join(tmpDir, "pointcrawl", "WildPlace.md")
	if _, err := os.Stat(wildPath); os.IsNotExist(err) {
		t.Fatalf("expected WildPlace.md to exist")
	}
}

func TestExportTimelineCreatesFiles(t *testing.T) {
	events := []simulation.Event{
		{Year: 105, Category: "war", Description: "Battle of Testville"},
		{Year: 110, Category: "founding", Description: "Oakhaven founded"},
	}

	tmpDir := t.TempDir()
	err := ExportTimeline(events, tmpDir)
	if err != nil {
		t.Fatalf("ExportTimeline error: %v", err)
	}

	chroniclePath := filepath.Join(tmpDir, "chronicles", "Chronicle.md")
	if _, err := os.Stat(chroniclePath); os.IsNotExist(err) {
		t.Fatalf("expected Chronicle.md to exist")
	}
}

func TestExportPointcrawlNilState(t *testing.T) {
	tmpDir := t.TempDir()
	if err := ExportPointcrawl(nil, tmpDir); err != nil {
		t.Fatalf("expected nil error for nil state, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "pointcrawl")); !os.IsNotExist(err) {
		t.Fatal("expected pointcrawl directory NOT to be created for nil state")
	}
}

func TestExportPointcrawlNilGraph(t *testing.T) {
	state := &world.State{Width: 10, Height: 10}
	tmpDir := t.TempDir()
	if err := ExportPointcrawl(state, tmpDir); err != nil {
		t.Fatalf("expected nil error for nil graph, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "pointcrawl")); !os.IsNotExist(err) {
		t.Fatal("expected pointcrawl directory NOT to be created for nil graph")
	}
}

func TestExportTimelineEmptyEvents(t *testing.T) {
	tmpDir := t.TempDir()
	if err := ExportTimeline(nil, tmpDir); err != nil {
		t.Fatalf("expected nil error for nil events, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "chronicles")); !os.IsNotExist(err) {
		t.Fatal("expected chronicles directory NOT to be created for empty events")
	}

	if err := ExportTimeline([]simulation.Event{}, tmpDir); err != nil {
		t.Fatalf("expected nil error for empty events, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "chronicles")); !os.IsNotExist(err) {
		t.Fatal("expected chronicles directory NOT to be created for empty events")
	}
}
