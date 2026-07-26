package simulation

import (
	"bytes"
	"testing"
)

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