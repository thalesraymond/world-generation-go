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

	if state.Settlements[0].Type == "" {
		t.Fatalf("expected settlement type assignment")
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

	if state.Settlements[0].Type == "" {
		t.Fatalf("expected settlement type assignment")
	}
}

func TestGenerateScalesPopulation(t *testing.T) {
	state := world.NewState(1, 1)
	state.Suitability = []float64{0.9}
	state.PopulationDensity = []float64{0.7}
	state.FactionInfluence = []string{""}

	config := DefaultConfig()
	config.MinDistance = 0
	config.MaxPopulation = 1000

	if err := Generate(state, config); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(state.Settlements) != 1 {
		t.Fatalf("settlement count = %d, want 1", len(state.Settlements))
	}

	if state.Settlements[0].Population != 700 {
		t.Fatalf("population = %v, want 700", state.Settlements[0].Population)
	}

	if state.Settlements[0].Type == "" {
		t.Fatalf("expected settlement type assignment")
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

	if state.Settlements[0].Type == "" {
		t.Fatalf("expected settlement type assignment")
	}
}

func TestGenerateClassifiesSettlementTypes(t *testing.T) {
	state := world.NewState(4, 1)

	state.Suitability = []float64{0.7, 0.7, 0.7, 0.7}
	state.PopulationDensity = []float64{0.9, 0.5, 0.1, 0.01}
	state.FactionInfluence = []string{"faction", "faction", "faction", "faction"}

	config := DefaultConfig()
	config.MinSuitability = 0.1
	config.MinPopulation = 0.005
	config.MinDistance = 1

	err := Generate(state, config)
	if err != nil {
		t.Fatal(err)
	}

	if len(state.Settlements) != 4 {
		t.Fatalf("expected 4 settlements, got %d", len(state.Settlements))
	}

	expectedTypes := map[int]string{
		90000: TypeMajorCity,
		50000: TypeMajorCity,
		10000: TypeCity,
		1000:  TypeVillage,
	}

	for _, s := range state.Settlements {
		expType, ok := expectedTypes[int(s.Population)]
		if !ok {
			t.Errorf("unexpected population %.0f, got type %s", s.Population, s.Type)
			continue
		}
		if s.Type != expType {
			t.Errorf("population %.0f: expected type %s, got %s", s.Population, expType, s.Type)
		}
	}
}
