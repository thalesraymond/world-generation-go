package simulation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/artifact"
	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	dompointcrawl "github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func exportTestState() *world.State {
	state := world.NewState(8, 8)
	state.Settlements = []world.Settlement{
		{
			Name:       "Deepcrest",
			Type:       "Village",
			X:          2,
			Y:          2,
			Faction:    "Northreach",
			Population: 120,
			Figures: []figures.HistoricalFigure{
				{ID: "Deepcrest-0", Name: "Aldric", Role: "Leader"},
			},
		},
	}
	state.PointcrawlGraph = dompointcrawl.NewGraph()
	state.PointcrawlGraph.AddNode(&dompointcrawl.Node{ID: 0, X: 6, Y: 6, Visibility: dompointcrawl.Hidden, Name: "Ruin-6-6", Kind: "ruin"})
	state.Artifacts = []artifact.Artifact{
		{
			ID:                 "artifact-ruin-0",
			Name:               "Relic of Ruin-6-6",
			Type:               "weapon",
			SignificanceSource: "intrinsic",
			Status:             "lost",
			SignificanceScore:  3,
			IsSignificant:      true,
			SignificanceYear:   0,
		},
	}
	return state
}

func TestObsidianExporterWritesAllNoteTypes(t *testing.T) {
	state := exportTestState()
	events := []domsim.Event{
		{Year: 1, Category: "Birth", Description: "Someone is born.", SettlementName: "Deepcrest", FigureID: "Deepcrest-0"},
	}
	targetDir := t.TempDir()

	exporter := ObsidianExporter{}
	if err := exporter.Export(state, events, targetDir); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	wantFiles := []string{
		filepath.Join("bases", "Deepcrest.md"),
		filepath.Join("factions", "Northreach.md"),
		filepath.Join("pointcrawl", "Network.md"),
		filepath.Join("pointcrawl", "Ruin-6-6.md"),
		filepath.Join("characters", "Aldric.md"),
		filepath.Join("chronicles", "Chronicle.md"),
		filepath.Join("artifacts", "Relic of Ruin-6-6.md"),
		filepath.Join("artifacts", "Index.md"),
	}
	for _, rel := range wantFiles {
		path := filepath.Join(targetDir, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to be written: %v", rel, err)
		}
	}
}

func TestObsidianExporterNoArtifactsSkipsArtifactsDir(t *testing.T) {
	state := exportTestState()
	state.Artifacts = nil
	targetDir := t.TempDir()

	exporter := ObsidianExporter{}
	if err := exporter.Export(state, nil, targetDir); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "artifacts")); !os.IsNotExist(err) {
		t.Errorf("expected no artifacts directory when there are no artifacts")
	}
}

func TestObsidianExporterErrorPaths(t *testing.T) {
	cases := []struct {
		name      string
		conflict  string
		state     func() *world.State
		events    []domsim.Event
		wantError string
	}{
		{
			name:     "bases dir creation fails",
			state:    exportTestState,
			conflict: "bases",
		},
		{
			name:     "pointcrawl dir creation fails",
			state:    exportTestState,
			conflict: "pointcrawl",
		},
		{
			name:     "chronicles dir creation fails",
			state:    exportTestState,
			events:   []domsim.Event{{Year: 1, Category: "Birth", Description: "A birth.", SettlementName: "Deepcrest"}},
			conflict: "chronicles",
		},
		{
			name:     "characters dir creation fails",
			state:    exportTestState,
			conflict: "characters",
		},
		{
			name:     "artifacts dir creation fails",
			state:    exportTestState,
			conflict: "artifacts",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(targetDir, tc.conflict), []byte("occupied"), 0o644); err != nil {
				t.Fatalf("write conflict file: %v", err)
			}

			exporter := ObsidianExporter{}
			err := exporter.Export(tc.state(), tc.events, targetDir)
			if err == nil {
				t.Fatalf("Export() expected error for %s conflict, got nil", tc.conflict)
			}
		})
	}
}
