package simulation

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
)

func TestGenerateWorldHasPointcrawlGraph(t *testing.T) {
	config := WorldGenConfig{Seed: 42, Width: 32, Height: 32, Years: 100}

	worldState, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() error = %v", err)
	}

	if worldState.PointcrawlGraph == nil {
		t.Fatalf("expected PointcrawlGraph to be populated")
	}

	if worldState.PointcrawlGraph.NodeCount() == 0 {
		t.Fatalf("expected PointcrawlGraph to have nodes")
	}

	if worldState.PointcrawlGraph.EdgeCount() == 0 {
		t.Fatalf("expected PointcrawlGraph to have edges")
	}
}

func TestGenerateWorldPointcrawlGraphIsDeterministic(t *testing.T) {
	config := WorldGenConfig{Seed: 42, Width: 32, Height: 32, Years: 100}

	first, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() first run error = %v", err)
	}

	second, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() second run error = %v", err)
	}

	firstJSON, err := graphToJSON(first.PointcrawlGraph)
	if err != nil {
		t.Fatalf("graphToJSON() first error = %v", err)
	}

	secondJSON, err := graphToJSON(second.PointcrawlGraph)
	if err != nil {
		t.Fatalf("graphToJSON() second error = %v", err)
	}

	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("expected identical pointcrawl graph for same seed")
	}
}

func graphToJSON(graph *pointcrawl.Graph) ([]byte, error) {
	return pointcrawl.GraphToJSON(graph)
}

func TestGenerateWorldCreatesSettlementFigures(t *testing.T) {
	config := WorldGenConfig{Seed: 42, Width: 48, Height: 48, Years: 100}

	worldState, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() error = %v", err)
	}

	if len(worldState.Settlements) == 0 {
		t.Fatalf("expected at least one settlement to test figure generation")
	}

	for _, settlement := range worldState.Settlements {
		if len(settlement.Figures) == 0 {
			t.Errorf("settlement %q has no figures", settlement.Name)
			continue
		}

		founder := settlement.Figures[0]
		if founder.Role != "Leader" {
			t.Errorf("settlement %q first founder role = %q, want Leader", settlement.Name, founder.Role)
		}
		if founder.Faction != settlement.Faction {
			t.Errorf("settlement %q founder faction = %q, want %q", settlement.Name, founder.Faction, settlement.Faction)
		}
	}
}

func TestGenerateWorldFoundersAreDeterministic(t *testing.T) {
	config := WorldGenConfig{Seed: 42, Width: 48, Height: 48, Years: 100}

	first, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() first run error = %v", err)
	}

	second, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() second run error = %v", err)
	}

	if len(first.Settlements) != len(second.Settlements) {
		t.Fatalf("settlement counts differ: %d vs %d", len(first.Settlements), len(second.Settlements))
	}

	for i := range first.Settlements {
		if len(first.Settlements[i].Figures) != len(second.Settlements[i].Figures) {
			t.Fatalf("settlement %q figure counts differ", first.Settlements[i].Name)
		}
		for j := range first.Settlements[i].Figures {
			f1 := first.Settlements[i].Figures[j]
			f2 := second.Settlements[i].Figures[j]
			if f1.ID != f2.ID || f1.Name != f2.Name || f1.Role != f2.Role || f1.BirthYear != f2.BirthYear || f1.MaxAge != f2.MaxAge {
				t.Errorf("settlement %q figure[%d] differs", first.Settlements[i].Name, j)
			}
		}
	}
}

func TestGenerateWorldIsDeterministicForSameSeed(t *testing.T) {
	config := WorldGenConfig{Seed: 42, Width: 16, Height: 16, Years: 100}

	first, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() first run error = %v", err)
	}

	second, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() second run error = %v", err)
	}

	firstJSON, err := first.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() first error = %v", err)
	}

	secondJSON, err := second.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() second error = %v", err)
	}

	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("expected identical world output for same seed")
	}
}

func TestGenerateWorldDiffersAcrossSeeds(t *testing.T) {
	configA := WorldGenConfig{Seed: 42, Width: 16, Height: 16, Years: 100}
	configB := WorldGenConfig{Seed: 99, Width: 16, Height: 16, Years: 100}

	worldA, err := GenerateWorld(configA)
	if err != nil {
		t.Fatalf("GenerateWorld() configA error = %v", err)
	}

	worldB, err := GenerateWorld(configB)
	if err != nil {
		t.Fatalf("GenerateWorld() configB error = %v", err)
	}

	jsonA, err := worldA.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() A error = %v", err)
	}

	jsonB, err := worldB.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() B error = %v", err)
	}

	if bytes.Equal(jsonA, jsonB) {
		t.Fatalf("expected different seeds to produce different world outputs")
	}
}

func TestGenerateWorldRejectsInvalidDimensions(t *testing.T) {
	config := WorldGenConfig{Seed: 1, Width: 0, Height: 16, Years: 100}

	_, err := GenerateWorld(config)
	if err == nil {
		t.Fatalf("expected error for invalid dimensions, got nil")
	}
}

func TestGenerateWorldCreatesPlantedRelicsPerRuin(t *testing.T) {
	config := WorldGenConfig{Seed: 42, Width: 48, Height: 48, Years: 100}

	worldState, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() error = %v", err)
	}

	ruinCount := 0
	for _, node := range worldState.PointcrawlGraph.Nodes {
		if node.Kind == "ruin" {
			ruinCount++
		}
	}

	if ruinCount == 0 {
		t.Fatalf("expected the test world to contain ruin nodes")
	}

	if len(worldState.Artifacts) != ruinCount {
		t.Fatalf("artifact count = %d, want %d ruin nodes", len(worldState.Artifacts), ruinCount)
	}
}

func TestGenerateWorldArtifactsAreDeterministic(t *testing.T) {
	config := WorldGenConfig{Seed: 42, Width: 48, Height: 48, Years: 100}

	first, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() first run error = %v", err)
	}

	second, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() second run error = %v", err)
	}

	firstJSON, err := json.Marshal(first.Artifacts)
	if err != nil {
		t.Fatalf("json.Marshal() first error = %v", err)
	}

	secondJSON, err := json.Marshal(second.Artifacts)
	if err != nil {
		t.Fatalf("json.Marshal() second error = %v", err)
	}

	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("expected identical artifacts for same seed")
	}
}

func TestGenerateWorldComponentStreamsAreIsolated(t *testing.T) {
	config := WorldGenConfig{Seed: 42, Width: 16, Height: 16, Years: 100}

	world1, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() first run error = %v", err)
	}

	world2, err := GenerateWorld(config)
	if err != nil {
		t.Fatalf("GenerateWorld() second run error = %v", err)
	}

	json1, err := world1.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() world1 error = %v", err)
	}

	json2, err := world2.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() world2 error = %v", err)
	}

	if !bytes.Equal(json1, json2) {
		t.Fatalf("expected identical output for same seed")
	}

	configDiff := WorldGenConfig{Seed: 43, Width: 16, Height: 16, Years: 100}
	world3, err := GenerateWorld(configDiff)
	if err != nil {
		t.Fatalf("GenerateWorld() third run error = %v", err)
	}

	json3, err := world3.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() world3 error = %v", err)
	}

	if bytes.Equal(json1, json3) {
		t.Fatalf("expected different seeds to produce isolated results")
	}
}
