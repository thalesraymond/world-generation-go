package spatial

import (
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
)

func TestEvaluateTileSuitabilityRewardsNearWaterFlatTemperateTiles(t *testing.T) {
	high := EvaluateTileSuitability(terrain.Tile{
		Elevation: 0.45,
		Biome:     terrain.BiomeGrassland,
	}, true, 0.02)

	if high <= 0.8 {
		t.Fatalf("high suitability = %f, want > 0.8", high)
	}
}

func TestEvaluateTileSuitabilityPenalizesHarshTerrain(t *testing.T) {
	low := EvaluateTileSuitability(terrain.Tile{
		Elevation: 0.93,
		Biome:     terrain.BiomeDesert,
	}, false, 0.25)

	if low >= 0.2 {
		t.Fatalf("low suitability = %f, want < 0.2", low)
	}
}

func TestCalculateSuitabilityMapReturnsOneScorePerTile(t *testing.T) {
	terrainMap := terrain.Map{
		Width:  2,
		Height: 2,
		Tiles: []terrain.Tile{
			{Elevation: 0.2, Biome: terrain.BiomeWater},
			{Elevation: 0.4, Biome: terrain.BiomeGrassland},
			{Elevation: 0.8, Biome: terrain.BiomeDesert},
			{Elevation: 0.45, Biome: terrain.BiomeForest},
		},
	}

	scores := CalculateSuitabilityMap(terrainMap)
	if len(scores) != 4 {
		t.Fatalf("score count = %d, want 4", len(scores))
	}

	for idx, score := range scores {
		if score < 0 || score > 1 {
			t.Fatalf("score[%d] out of range: %f", idx, score)
		}
	}

	if scores[1] <= scores[2] {
		t.Fatalf("expected grassland near water to beat desert: grassland=%f desert=%f", scores[1], scores[2])
	}
}
