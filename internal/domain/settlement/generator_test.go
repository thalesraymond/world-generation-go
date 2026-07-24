package settlement

import (
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func TestGenerateFiltersBySuitabilityAndPopulation(t *testing.T) {
	state := world.NewState(2, 2)
	state.Suitability = []float64{0.8, 0.7, 0.5, 0.9}
	state.PopulationDensity = []float64{0.6, 0.2, 0.9, 0.7}
	state.FactionInfluence = []string{"auric", "auric", "verdant", "cinder"}

	config := DefaultConfig()
	config.MinSuitability = 0.65
	config.MinPopulation = 0.5
	config.MinDistance = 0

	if err := Generate(state, config); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(state.Settlements) != 2 {
		t.Fatalf("settlement count = %d, want 2", len(state.Settlements))
	}

	if state.Settlements[0].Faction == "" {
		t.Fatalf("expected settlement faction assignment")
	}
}

func TestGenerateEnforcesMinimumDistance(t *testing.T) {
	state := world.NewState(3, 1)
	state.Suitability = []float64{0.9, 0.8, 0.7}
	state.PopulationDensity = []float64{0.9, 0.8, 0.7}
	state.FactionInfluence = []string{"auric", "auric", "verdant"}

	config := DefaultConfig()
	config.MinSuitability = 0.6
	config.MinPopulation = 0.6
	config.MinDistance = 2.0

	if err := Generate(state, config); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(state.Settlements) != 2 {
		t.Fatalf("settlement count = %d, want 2", len(state.Settlements))
	}

	if state.Settlements[0].X == 0 && state.Settlements[1].X == 1 {
		t.Fatalf("expected spacing to reject adjacent settlement at x=1")
	}
}

func TestGenerateAssignsIndependentWhenNoDominantFaction(t *testing.T) {
	state := world.NewState(1, 1)
	state.Suitability = []float64{0.9}
	state.PopulationDensity = []float64{0.7}
	state.FactionInfluence = []string{""}

	config := DefaultConfig()
	config.MinDistance = 0

	if err := Generate(state, config); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(state.Settlements) != 1 {
		t.Fatalf("settlement count = %d, want 1", len(state.Settlements))
	}

	if state.Settlements[0].Faction != "independent" {
		t.Fatalf("faction = %q, want independent", state.Settlements[0].Faction)
	}
}
