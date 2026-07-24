package world

import (
	"reflect"
	"testing"
)

func TestNewStateInitializesGridLayers(t *testing.T) {
	state := NewState(4, 3)

	if state.CellCount() != 12 {
		t.Fatalf("cell count = %d, want 12", state.CellCount())
	}

	if len(state.PopulationDensity) != 12 {
		t.Fatalf("population layer size = %d, want 12", len(state.PopulationDensity))
	}

	if len(state.FactionInfluence) != 12 {
		t.Fatalf("faction layer size = %d, want 12", len(state.FactionInfluence))
	}

	if len(state.Suitability) != 12 {
		t.Fatalf("suitability layer size = %d, want 12", len(state.Suitability))
	}
}

func TestStateJSONRoundTripIncludesDemographicsAndSettlements(t *testing.T) {
	state := NewState(3, 2)
	state.PopulationDensity[0] = 0.7
	state.PopulationDensity[1] = 0.2
	state.FactionInfluence[0] = "auric"
	state.Suitability[0] = 0.9
	state.Settlements = []Settlement{{
		Name:       "Settlement-001",
		X:          1,
		Y:          0,
		Faction:    "auric",
		Population: 0.7,
	}}

	data, err := state.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	decoded, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if !reflect.DeepEqual(state, decoded) {
		t.Fatalf("state mismatch after round trip")
	}
}

func TestFromJSONRejectsInvalidGridLengths(t *testing.T) {
	payload := []byte(`{"width":2,"height":2,"populationDensity":[0.2],"factionInfluence":["a","b","c","d"],"suitability":[0.1,0.2,0.3,0.4],"settlements":[]}`)

	if _, err := FromJSON(payload); err == nil {
		t.Fatalf("expected validation error for invalid layer length")
	}
}
