package world

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
)

func TestStateJSONRoundTripIncludesPointcrawlGraph(t *testing.T) {
	state := NewState(3, 2)
	state.PopulationDensity[0] = 0.7
	state.FactionInfluence[0] = "auric"
	state.Suitability[0] = 0.9
	state.Settlements = []Settlement{{
		Name:       "Settlement-001",
		X:          1,
		Y:          0,
		Faction:    "auric",
		Population: 0.7,
	}}

	graph := pointcrawl.NewGraph()
	graph.AddNode(&pointcrawl.Node{ID: 0, X: 1, Y: 0, Name: "Settlement-001", Kind: "settlement", Visibility: pointcrawl.Known})
	state.PointcrawlGraph = graph

	data, err := state.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	decoded, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if decoded.PointcrawlGraph == nil {
		t.Fatal("PointcrawlGraph lost during round trip")
	}

	if decoded.PointcrawlGraph.NodeCount() != 1 {
		t.Fatalf("decoded graph NodeCount = %d, want 1", decoded.PointcrawlGraph.NodeCount())
	}

	if !reflect.DeepEqual(state, decoded) {
		t.Fatalf("state mismatch after round trip")
	}
}

func TestStateJSONRoundTripWithNilPointcrawlGraph(t *testing.T) {
	state := NewState(2, 2)

	data, err := state.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	decoded, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if decoded.PointcrawlGraph != nil {
		t.Fatalf("expected nil PointcrawlGraph, got %+v", decoded.PointcrawlGraph)
	}
}

func TestStateCellCountInvalidDimensions(t *testing.T) {
	state := NewState(0, 5)
	if state.CellCount() != 0 {
		t.Fatalf("CellCount = %d, want 0", state.CellCount())
	}
}

func TestStateIndex(t *testing.T) {
	state := NewState(4, 3)

	cases := []struct {
		x, y    int
		wantIdx int
		wantOK  bool
	}{
		{0, 0, 0, true},
		{3, 0, 3, true},
		{0, 2, 8, true},
		{3, 2, 11, true},
		{-1, 0, 0, false},
		{0, -1, 0, false},
		{4, 0, 0, false},
		{0, 3, 0, false},
	}

	for _, tc := range cases {
		idx, ok := state.Index(tc.x, tc.y)
		if ok != tc.wantOK {
			t.Fatalf("Index(%d,%d) ok = %v, want %v", tc.x, tc.y, ok, tc.wantOK)
		}
		if ok && idx != tc.wantIdx {
			t.Fatalf("Index(%d,%d) = %d, want %d", tc.x, tc.y, idx, tc.wantIdx)
		}
	}
}

func TestStateSetSuitability(t *testing.T) {
	state := NewState(2, 2)
	scores := []float64{0.1, 0.2, 0.3, 0.4}

	if err := state.SetSuitability(scores); err != nil {
		t.Fatalf("SetSuitability() error = %v", err)
	}

	for i, want := range scores {
		if state.Suitability[i] != want {
			t.Fatalf("Suitability[%d] = %v, want %v", i, state.Suitability[i], want)
		}
	}

	if err := state.SetSuitability([]float64{0.1}); err == nil {
		t.Fatal("expected error for mismatched suitability size")
	}
}

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

func TestStateValidateRejectsInvalidLayerSizes(t *testing.T) {
	cases := []struct {
		name      string
		mutate    func(*State)
		wantError string
	}{
		{
			name:      "population density",
			mutate:    func(s *State) { s.PopulationDensity = []float64{0.1} },
			wantError: "population density",
		},
		{
			name:      "faction influence",
			mutate:    func(s *State) { s.FactionInfluence = []string{"a"} },
			wantError: "faction influence",
		},
		{
			name:      "suitability",
			mutate:    func(s *State) { s.Suitability = []float64{0.1} },
			wantError: "suitability",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewState(2, 2)
			state.PopulationDensity = make([]float64, state.CellCount())
			state.FactionInfluence = make([]string, state.CellCount())
			state.Suitability = make([]float64, state.CellCount())
			tc.mutate(state)

			if err := state.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestFromJSONRejectsInvalidGridLengths(t *testing.T) {
	payload := []byte(`{"width":2,"height":2,"populationDensity":[0.2],"factionInfluence":["a","b","c","d"],"suitability":[0.1,0.2,0.3,0.4],"settlements":[]}`)

	if _, err := FromJSON(payload); err == nil {
		t.Fatalf("expected validation error for invalid layer length")
	}
}

func TestSettlementJSONRoundTripWithFigures(t *testing.T) {
	settlement := Settlement{
		Name:       "Riverwatch",
		Type:       "Town",
		X:          2,
		Y:          3,
		Faction:    "ind",
		Population: 1200,
		Figures: []figures.HistoricalFigure{
			{
				ID:        "fig-001",
				Name:      "Alden the Bold",
				BirthYear: 100,
				MaxAge:    60,
				Role:      "leader",
				Faction:   "ind",
				Relationships: figures.Relationships{
					Parents:  []string{"fig-p1"},
					Children: []string{"fig-c1"},
					Spouse:   []string{"fig-s1"},
				},
			},
			{
				ID:        "fig-002",
				Name:      "Mira the Wise",
				BirthYear: 120,
				DeathYear: 180,
				MaxAge:    70,
				Role:      "explorer",
				Faction:   "ind",
			},
		},
	}

	data, err := json.Marshal(settlement)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Settlement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(settlement, decoded) {
		t.Fatalf("settlement mismatch after round trip: got %+v, want %+v", decoded, settlement)
	}
}

func TestSettlementEmptyFiguresSerialization(t *testing.T) {
	settlement := Settlement{
		Name:       "Empty Hollow",
		Type:       "Village",
		X:          0,
		Y:          0,
		Faction:    "ind",
		Population: 500,
		Figures:    []figures.HistoricalFigure{},
	}

	data, err := json.Marshal(settlement)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Settlement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(settlement, decoded) {
		t.Fatalf("settlement mismatch after round trip: got %+v, want %+v", decoded, settlement)
	}
}

func TestSettlementBackwardCompat(t *testing.T) {
	payload := []byte(`{"name":"Test","type":"Village","x":0,"y":0,"faction":"ind","population":500}`)

	var decoded Settlement
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Name != "Test" || decoded.Type != "Village" || decoded.Population != 500 {
		t.Fatalf("decoded settlement mismatch: got %+v", decoded)
	}

	if len(decoded.Figures) != 0 {
		t.Fatalf("expected Figures nil or empty, got %v", decoded.Figures)
	}
}
