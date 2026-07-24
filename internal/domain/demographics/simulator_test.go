package demographics

import (
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func TestPreGenerateSuitabilityStoresMap(t *testing.T) {
	state := world.NewState(3, 1)
	terrainMap := terrain.Map{
		Width:  3,
		Height: 1,
		Tiles: []terrain.Tile{
			{Elevation: 0.3, Biome: terrain.BiomeWater},
			{Elevation: 0.4, Biome: terrain.BiomeGrassland},
			{Elevation: 0.9, Biome: terrain.BiomeDesert},
		},
	}

	if err := PreGenerateSuitability(state, terrainMap); err != nil {
		t.Fatalf("PreGenerateSuitability() error = %v", err)
	}

	if len(state.Suitability) != 3 {
		t.Fatalf("suitability size = %d, want 3", len(state.Suitability))
	}

	if state.Suitability[1] <= state.Suitability[2] {
		t.Fatalf("expected grassland tile to beat desert tile")
	}
}

func TestSimulateDiffusesPopulationToSuitableLowerDensityTile(t *testing.T) {
	state := world.NewState(3, 1)
	state.Suitability = []float64{0.1, 1.0, 0.2}
	state.PopulationDensity = []float64{1.0, 0.0, 0.0}
	state.FactionInfluence = []string{"auric", "", ""}

	config := DefaultConfig()
	config.Iterations = 1
	config.DiffusionRate = 0.5
	config.MinPopulation = 0.01

	if err := Simulate(state, config); err != nil {
		t.Fatalf("Simulate() error = %v", err)
	}

	if state.PopulationDensity[1] <= 0 {
		t.Fatalf("expected central tile to receive migrated population")
	}

	if state.FactionInfluence[1] != "auric" {
		t.Fatalf("expected faction to spread to central tile, got %q", state.FactionInfluence[1])
	}
}

func TestSeedPopulationFromSuitabilityAssignsFactionsDeterministically(t *testing.T) {
	state := world.NewState(2, 2)
	state.Suitability = []float64{1, 0.9, 0.8, 0.1}

	config := DefaultConfig()
	config.FactionNames = []string{"one", "two"}
	config.MinPopulation = 0.2

	if err := SeedPopulationFromSuitability(state, config); err != nil {
		t.Fatalf("SeedPopulationFromSuitability() error = %v", err)
	}

	if state.FactionInfluence[0] != "one" || state.FactionInfluence[1] != "two" {
		t.Fatalf("unexpected seeded factions: %v", state.FactionInfluence)
	}

	if state.FactionInfluence[3] != "" {
		t.Fatalf("expected low-population tile to remain unclaimed")
	}
}
