package exporter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
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
	err := ExportTimeline(nil, events, tmpDir)
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

func TestExportFiguresCreatesFiles(t *testing.T) {
	state := &world.State{
		Width:  10,
		Height: 10,
		Settlements: []world.Settlement{
			{
				Name: "Riverwatch",
				Figures: []figures.HistoricalFigure{
					{ID: "f1", Name: "Aldric Stone", BirthYear: 100, Faction: "Ironbound", Role: "Leader"},
					{ID: "f2", Name: "Bren/Guard", BirthYear: 105, Faction: "Ironbound", Role: "Guard"},
				},
			},
		},
	}

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "vault")

	if err := ExportFigures(state, nil, targetDir); err != nil {
		t.Fatalf("ExportFigures failed: %v", err)
	}

	charsDir := filepath.Join(targetDir, "characters")
	if _, err := os.Stat(charsDir); os.IsNotExist(err) {
		t.Fatal("characters directory does not exist")
	}

	entries, err := os.ReadDir(charsDir)
	if err != nil {
		t.Fatalf("read characters dir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 figure files, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Errorf("expected .md file, got directory %s", entry.Name())
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			t.Errorf("expected .md suffix, got %s", entry.Name())
		}
		if strings.ContainsAny(entry.Name(), "<>:\"/\\|?*") {
			t.Errorf("file name contains special characters: %s", entry.Name())
		}
	}
}

func TestExportFiguresFrontmatterFields(t *testing.T) {
	aliveFigure := figures.HistoricalFigure{
		ID:        "char-alive",
		Name:      "Lyra Windwhisper",
		BirthYear: 200,
		Role:      "Mage",
		Faction:   "Sylvani",
		Relationships: figures.Relationships{
			Parents:  []string{"char-parent"},
			Spouse:   []string{"char-spouse"},
			Children: []string{"char-child"},
		},
	}
	deceasedFigure := figures.HistoricalFigure{
		ID:        "char-dead",
		Name:      "Thorin Embercrown",
		BirthYear: 150,
		DeathYear: 210,
		Role:      "King",
		Faction:   "Ironbound",
	}

	state := &world.State{
		Width:  10,
		Height: 10,
		Settlements: []world.Settlement{
			{
				Name:    "Oakhaven",
				Figures: []figures.HistoricalFigure{aliveFigure, deceasedFigure},
			},
		},
	}

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "vault")

	if err := ExportFigures(state, nil, targetDir); err != nil {
		t.Fatalf("ExportFigures failed: %v", err)
	}

	aliveBytes, err := os.ReadFile(filepath.Join(targetDir, "characters", "Lyra Windwhisper.md"))
	if err != nil {
		t.Fatalf("read alive figure: %v", err)
	}
	alive := string(aliveBytes)

	for _, want := range []string{
		"id: char-alive",
		"type: character",
		"name: Lyra Windwhisper",
		"role: Mage",
		"faction: Sylvani",
		"birthYear: 200",
		`settlement: "[[Oakhaven]]"`,
		"status: alive",
	} {
		if !strings.Contains(alive, want) {
			t.Errorf("alive figure missing frontmatter field: %s", want)
		}
	}
	if strings.Contains(alive, "deathYear") {
		t.Error("alive figure should NOT have deathYear field")
	}
	if !strings.Contains(alive, `parents:`) || !strings.Contains(alive, `[[char-parent]]`) {
		t.Errorf("alive figure missing parents frontmatter field, got:\n%s", alive)
	}
	if !strings.Contains(alive, `spouse:`) || !strings.Contains(alive, `[[char-spouse]]`) {
		t.Errorf("alive figure missing spouse frontmatter field")
	}
	if !strings.Contains(alive, `children:`) || !strings.Contains(alive, `[[char-child]]`) {
		t.Errorf("alive figure missing children frontmatter field")
	}

	deadBytes, err := os.ReadFile(filepath.Join(targetDir, "characters", "Thorin Embercrown.md"))
	if err != nil {
		t.Fatalf("read deceased figure: %v", err)
	}
	dead := string(deadBytes)

	if !strings.Contains(dead, "status: deceased") {
		t.Errorf("deceased figure missing status: deceased")
	}
	if !strings.Contains(dead, "deathYear: 210") {
		t.Errorf("deceased figure missing deathYear field")
	}
}

func TestExportFiguresWikiLinks(t *testing.T) {
	state := &world.State{
		Width:  10,
		Height: 10,
		Settlements: []world.Settlement{
			{
				Name:    "Oakhaven",
				Faction: "Sylvani",
				Figures: []figures.HistoricalFigure{
					{
						ID:        "char-1",
						Name:      "Elara",
						BirthYear: 200,
						Faction:   "Sylvani",
						Role:      "Leader",
						Relationships: figures.Relationships{
							Parents: []string{"char-parent"},
							Spouse:  []string{"char-spouse"},
						},
					},
				},
			},
		},
	}

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "vault")

	if err := ExportFigures(state, nil, targetDir); err != nil {
		t.Fatalf("ExportFigures failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(targetDir, "characters", "Elara.md"))
	if err != nil {
		t.Fatalf("read figure file: %v", err)
	}
	body := string(content)

	if !strings.Contains(body, "[[Sylvani]]") {
		t.Errorf("missing wiki-link to faction")
	}
	if !strings.Contains(body, "[[Oakhaven]]") {
		t.Errorf("missing wiki-link to settlement")
	}
	if !strings.Contains(body, "[[char-parent]]") {
		t.Errorf("missing wiki-link to parent")
	}
	if !strings.Contains(body, "[[char-spouse]]") {
		t.Errorf("missing wiki-link to spouse")
	}
}

func TestExportFiguresNoCharactersDirWhenEmpty(t *testing.T) {
	state := &world.State{
		Width:  10,
		Height: 10,
		Settlements: []world.Settlement{
			{Name: "Riverwatch", Figures: []figures.HistoricalFigure{}},
			{Name: "Oakhaven"},
		},
	}

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "vault")

	if err := ExportFigures(state, nil, targetDir); err != nil {
		t.Fatalf("ExportFigures failed: %v", err)
	}

	charsDir := filepath.Join(targetDir, "characters")
	if _, err := os.Stat(charsDir); !os.IsNotExist(err) {
		t.Error("characters directory should NOT exist when no figures are present")
	}
}

func TestExportTimelineEmptyEvents(t *testing.T) {
	tmpDir := t.TempDir()
	if err := ExportTimeline(nil, nil, tmpDir); err != nil {
		t.Fatalf("expected nil error for nil events, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "chronicles")); !os.IsNotExist(err) {
		t.Fatal("expected chronicles directory NOT to be created for empty events")
	}

	if err := ExportTimeline(nil, []simulation.Event{}, tmpDir); err != nil {
		t.Fatalf("expected nil error for empty events, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "chronicles")); !os.IsNotExist(err) {
		t.Fatal("expected chronicles directory NOT to be created for empty events")
	}
}

func TestJSONRoundTripReExportCycle(t *testing.T) {
	original := world.NewState(10, 10)
	original.Settlements = []world.Settlement{
		{
			Name: "Riverwatch", Type: "Town", X: 2, Y: 3, Faction: "Ironbound", Population: 500,
			Figures: []figures.HistoricalFigure{
				{ID: "rw-0", Name: "Aldric Stone", BirthYear: 10, Faction: "Ironbound", Role: "Leader"},
				{ID: "rw-1", Name: "Bren Moss", BirthYear: 25, Faction: "Ironbound", Role: "Explorer", DeathYear: 80, MaxAge: 70},
			},
		},
	}

	events := []simulation.Event{
		{Year: 45, Category: "politics", Description: "Aldric declares a feast", FigureID: "rw-0", SettlementName: "Riverwatch"},
	}

	dir1 := t.TempDir()
	if err := Export(original, dir1); err != nil {
		t.Fatalf("Export original: %v", err)
	}
	if err := ExportTimeline(original, events, dir1); err != nil {
		t.Fatalf("ExportTimeline original: %v", err)
	}
	if err := ExportFigures(original, events, dir1); err != nil {
		t.Fatalf("ExportFigures original: %v", err)
	}

	stateJSON, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	roundtripped, err := world.FromJSON(stateJSON)
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	dir2 := t.TempDir()
	if err := Export(roundtripped, dir2); err != nil {
		t.Fatalf("Export roundtripped: %v", err)
	}
	if err := ExportTimeline(roundtripped, events, dir2); err != nil {
		t.Fatalf("ExportTimeline roundtripped: %v", err)
	}
	if err := ExportFigures(roundtripped, events, dir2); err != nil {
		t.Fatalf("ExportFigures roundtripped: %v", err)
	}

	subdirs := []string{"bases", "factions", "characters", "chronicles"}
	for _, sub := range subdirs {
		entries1, err := os.ReadDir(filepath.Join(dir1, sub))
		if err != nil {
			t.Fatalf("read dir %s (original): %v", sub, err)
		}
		entries2, err := os.ReadDir(filepath.Join(dir2, sub))
		if err != nil {
			t.Fatalf("read dir %s (roundtripped): %v", sub, err)
		}
		if len(entries1) != len(entries2) {
			t.Errorf("dir %s file count differs: %d vs %d", sub, len(entries1), len(entries2))
			continue
		}
		for i := range entries1 {
			if entries1[i].Name() != entries2[i].Name() {
				t.Errorf("dir %s file names differ: %s vs %s", sub, entries1[i].Name(), entries2[i].Name())
				continue
			}
			data1, _ := os.ReadFile(filepath.Join(dir1, sub, entries1[i].Name()))
			data2, _ := os.ReadFile(filepath.Join(dir2, sub, entries2[i].Name()))
			if !bytes.Equal(data1, data2) {
				t.Errorf("file %s/%s differs after JSON round-trip", sub, entries1[i].Name())
			}
		}
	}
}

func TestExportFiguresRelationshipsUseDisplayNames(t *testing.T) {
	// Two founders who are spouses of each other.
	founderA := figures.HistoricalFigure{
		ID:        "Eastfield-2-0",
		Name:      "Aelar Blackwood",
		BirthYear: 100,
		Role:      "Leader",
		Faction:   "Ironbound",
		Relationships: figures.Relationships{
			Spouse: []string{"Eastfield-2-1"},
		},
	}
	founderB := figures.HistoricalFigure{
		ID:        "Eastfield-2-1",
		Name:      "Baelor Dawnwhisper",
		BirthYear: 102,
		Role:      "Healer",
		Faction:   "Ironbound",
		Relationships: figures.Relationships{
			Spouse: []string{"Eastfield-2-0"},
		},
	}
	// Newborn whose parents are the two founders.
	newborn := figures.HistoricalFigure{
		ID:        "born-2",
		Name:      "Caius Stormborn",
		BirthYear: 130,
		Role:      "Child",
		Faction:   "Ironbound",
		Relationships: figures.Relationships{
			Parents: []string{"Eastfield-2-0", "Eastfield-2-1"},
		},
	}
	// Add children back-references to founders.
	founderA.Relationships.Children = []string{"born-2"}
	founderB.Relationships.Children = []string{"born-2"}

	state := &world.State{
		Width:  10,
		Height: 10,
		Settlements: []world.Settlement{
			{
				Name:    "Eastfield-2",
				Figures: []figures.HistoricalFigure{founderA, founderB, newborn},
			},
		},
	}

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "vault")

	if err := ExportFigures(state, nil, targetDir); err != nil {
		t.Fatalf("ExportFigures failed: %v", err)
	}

	newbornBytes, err := os.ReadFile(filepath.Join(targetDir, "characters", "Caius Stormborn.md"))
	if err != nil {
		t.Fatalf("read newborn file: %v", err)
	}
	newbornContent := string(newbornBytes)

	if !strings.Contains(newbornContent, "[[Aelar Blackwood]]") {
		t.Errorf("newborn file should contain wiki-link to parent Aelar Blackwood, got:\n%s", newbornContent)
	}
	if !strings.Contains(newbornContent, "[[Baelor Dawnwhisper]]") {
		t.Errorf("newborn file should contain wiki-link to parent Baelor Dawnwhisper, got:\n%s", newbornContent)
	}
	if strings.Contains(newbornContent, "[[Eastfield-2-0]]") {
		t.Errorf("newborn file contains placeholder ID [[Eastfield-2-0]] instead of display name")
	}
	if strings.Contains(newbornContent, "[[Eastfield-2-1]]") {
		t.Errorf("newborn file contains placeholder ID [[Eastfield-2-1]] instead of display name")
	}

	parentABytes, err := os.ReadFile(filepath.Join(targetDir, "characters", "Aelar Blackwood.md"))
	if err != nil {
		t.Fatalf("read founder A file: %v", err)
	}
	parentAContent := string(parentABytes)

	if !strings.Contains(parentAContent, "[[Baelor Dawnwhisper]]") {
		t.Errorf("founder A file should contain wiki-link to spouse Baelor Dawnwhisper")
	}
	if strings.Contains(parentAContent, "[[Eastfield-2-1]]") {
		t.Errorf("founder A file contains placeholder ID [[Eastfield-2-1]] instead of display name")
	}
	if !strings.Contains(parentAContent, "[[Caius Stormborn]]") {
		t.Errorf("founder A file should contain wiki-link to child Caius Stormborn")
	}
	if strings.Contains(parentAContent, "[[born-2]]") {
		t.Errorf("founder A file contains placeholder ID [[born-2]] instead of display name")
	}

	parentBBytes, err := os.ReadFile(filepath.Join(targetDir, "characters", "Baelor Dawnwhisper.md"))
	if err != nil {
		t.Fatalf("read founder B file: %v", err)
	}
	parentBContent := string(parentBBytes)

	if !strings.Contains(parentBContent, "[[Aelar Blackwood]]") {
		t.Errorf("founder B file should contain wiki-link to spouse Aelar Blackwood")
	}
	if strings.Contains(parentBContent, "[[Eastfield-2-0]]") {
		t.Errorf("founder B file contains placeholder ID [[Eastfield-2-0]] instead of display name")
	}
	if !strings.Contains(parentBContent, "[[Caius Stormborn]]") {
		t.Errorf("founder B file should contain wiki-link to child Caius Stormborn")
	}
	if strings.Contains(parentBContent, "[[born-2]]") {
		t.Errorf("founder B file contains placeholder ID [[born-2]] instead of display name")
	}
}

func TestExportFiguresIntegration(t *testing.T) {
	state := &world.State{
		Width:  100,
		Height: 100,
		Settlements: []world.Settlement{
			{
				Name: "Riverwatch", Type: "Town", X: 10, Y: 20, Faction: "Ironbound", Population: 500,
				Figures: []figures.HistoricalFigure{
					{ID: "riverwatch-0", Name: "Aldric Bronzefist", Role: "Leader", Faction: "Ironbound", BirthYear: 10},
					{ID: "riverwatch-1", Name: "Mira Bronzefist", Role: "Explorer", Faction: "Ironbound", BirthYear: 35},
					{ID: "riverwatch-2", Name: "Beran Stonehand", Role: "Warrior", Faction: "Ironbound", BirthYear: 60},
				},
			},
			{
				Name: "Oakhaven", Type: "City", X: 30, Y: 40, Faction: "Sylvani", Population: 1200,
				Figures: []figures.HistoricalFigure{
					{ID: "oakhaven-0", Name: "Elena Silksong", Role: "Leader", Faction: "Sylvani", BirthYear: 5},
				},
			},
		},
	}

	events := []simulation.Event{
		{Year: 45, Category: "politics", Description: "Aldric declares a festival", FigureID: "riverwatch-0", SettlementName: "Riverwatch"},
		{Year: 70, Category: "discovery", Description: "Mira charts the eastern ridge", FigureID: "riverwatch-1", SettlementName: "Riverwatch"},
	}

	tmpDir := t.TempDir()

	if err := Export(state, tmpDir); err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if err := ExportTimeline(state, events, tmpDir); err != nil {
		t.Fatalf("ExportTimeline failed: %v", err)
	}
	if err := ExportFigures(state, events, tmpDir); err != nil {
		t.Fatalf("ExportFigures failed: %v", err)
	}

	charsDir := filepath.Join(tmpDir, "characters")
	if _, err := os.Stat(charsDir); os.IsNotExist(err) {
		t.Fatalf("characters directory does not exist")
	}

	entries, err := os.ReadDir(charsDir)
	if err != nil {
		t.Fatalf("read characters dir: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 character files, got %d", len(entries))
	}

	expectedFiles := map[string]bool{
		"Aldric Bronzefist.md": false,
		"Mira Bronzefist.md":   false,
		"Beran Stonehand.md":   false,
		"Elena Silksong.md":    false,
	}
	for _, entry := range entries {
		expectedFiles[entry.Name()] = true
	}
	for name, found := range expectedFiles {
		if !found {
			t.Errorf("expected character file %s not found", name)
		}
	}

	aldricBytes, err := os.ReadFile(filepath.Join(charsDir, "Aldric Bronzefist.md"))
	if err != nil {
		t.Fatalf("read Aldric file: %v", err)
	}
	aldric := string(aldricBytes)

	if !strings.Contains(aldric, "type: character") {
		t.Errorf("character file missing type field")
	}
	if !strings.Contains(aldric, "name: Aldric Bronzefist") {
		t.Errorf("character file missing name field")
	}
	if !strings.Contains(aldric, "role: Leader") {
		t.Errorf("character file missing role field")
	}
	if !strings.Contains(aldric, "settlement: \"[[Riverwatch]]\"") {
		t.Errorf("character file missing settlement wiki-link")
	}
	if !strings.Contains(aldric, "## Chronicle") {
		t.Errorf("character file missing Chronicle section")
	}
	if !strings.Contains(aldric, "Year 45: Aldric declares a festival") {
		t.Errorf("character file missing expected event")
	}

	riverwatchBytes, err := os.ReadFile(filepath.Join(tmpDir, "bases", "Riverwatch.md"))
	if err != nil {
		t.Fatalf("read Riverwatch file: %v", err)
	}
	riverwatch := string(riverwatchBytes)

	if !strings.Contains(riverwatch, "## Characters") {
		t.Errorf("settlement file missing Characters section")
	}
	if !strings.Contains(riverwatch, "### Leader") {
		t.Errorf("settlement file missing Leader subsection")
	}
	if !strings.Contains(riverwatch, "### Explorers") {
		t.Errorf("settlement file missing Explorers subsection")
	}
	if !strings.Contains(riverwatch, "### Others") {
		t.Errorf("settlement file missing Others subsection")
	}
	if !strings.Contains(riverwatch, "[[Aldric Bronzefist]] (Leader)") {
		t.Errorf("settlement file missing leader wiki-link")
	}
	if !strings.Contains(riverwatch, "[[Mira Bronzefist]] (Explorer)") {
		t.Errorf("settlement file missing explorer wiki-link")
	}
	if !strings.Contains(riverwatch, "[[Beran Stonehand]] (Warrior)") {
		t.Errorf("settlement file missing other figure wiki-link")
	}

	chronicleBytes, err := os.ReadFile(filepath.Join(tmpDir, "chronicles", "Chronicle.md"))
	if err != nil {
		t.Fatalf("read Chronicle file: %v", err)
	}
	chronicle := string(chronicleBytes)

	if !strings.Contains(chronicle, "*(by [[Aldric Bronzefist]])*") {
		t.Errorf("chronicle missing figure reference for event with FigureID")
	}
}

func TestExportTimelineUsesFigureNames(t *testing.T) {
	state := &world.State{
		Width:  10,
		Height: 10,
		Settlements: []world.Settlement{
			{
				Name: "Greenvale",
				Figures: []figures.HistoricalFigure{
					{ID: "greenvale-0", Name: "Garrick Thorne", BirthYear: 100, Role: "Leader", Faction: "Ironbound"},
					{ID: "greenvale-1", Name: "Helia Fairwind", BirthYear: 102, Role: "Healer", Faction: "Ironbound"},
				},
			},
		},
	}

	events := []simulation.Event{
		{Year: 105, Category: "politics", Description: "Garrick declares a feast", FigureID: "greenvale-0"},
		{Year: 110, Category: "discovery", Description: "Helia maps the valley", FigureID: "greenvale-1"},
	}

	tmpDir := t.TempDir()
	if err := ExportTimeline(state, events, tmpDir); err != nil {
		t.Fatalf("ExportTimeline failed: %v", err)
	}

	chronicleBytes, err := os.ReadFile(filepath.Join(tmpDir, "chronicles", "Chronicle.md"))
	if err != nil {
		t.Fatalf("read chronicle: %v", err)
	}
	chronicle := string(chronicleBytes)

	if !strings.Contains(chronicle, "*(by [[Garrick Thorne]])*") {
		t.Errorf("chronicle should use figure name [[Garrick Thorne]], got:\n%s", chronicle)
	}
	if !strings.Contains(chronicle, "*(by [[Helia Fairwind]])*") {
		t.Errorf("chronicle should use figure name [[Helia Fairwind]], got:\n%s", chronicle)
	}
	if strings.Contains(chronicle, "[[greenvale-0]]") {
		t.Errorf("chronicle contains placeholder ID [[greenvale-0]]")
	}
	if strings.Contains(chronicle, "[[greenvale-1]]") {
		t.Errorf("chronicle contains placeholder ID [[greenvale-1]]")
	}
}
