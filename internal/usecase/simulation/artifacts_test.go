package simulation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/geography/pointcrawl"
)

func TestGeneratePlantedRelicsOnePerRuinInNodeIDOrder(t *testing.T) {
	tests := []struct {
		name      string
		nodes     []pointcrawl.Node
		wantIDs   []string
		wantNames []string
	}{
		{
			name: "ruins sorted by node ID, other kinds ignored",
			nodes: []pointcrawl.Node{
				{ID: 4, Name: "Ruin-40-40", Kind: "ruin"},
				{ID: 2, Name: "Ruin-20-20", Kind: "ruin"},
				{ID: 9, Name: "Ruin-90-90", Kind: "ruin"},
				{ID: 6, Name: "Wilderness-60-60", Kind: "wilderness"},
				{ID: 1, Name: "Landmark-10-10", Kind: "landmark"},
			},
			wantIDs:   []string{"artifact-ruin-0", "artifact-ruin-1", "artifact-ruin-2"},
			wantNames: []string{"Relic of Ruin-20-20", "Relic of Ruin-40-40", "Relic of Ruin-90-90"},
		},
		{
			name:      "no ruins yields empty result",
			nodes:     []pointcrawl.Node{{ID: 1, Name: "Deepcrest", Kind: "settlement"}},
			wantIDs:   []string{},
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := pointcrawl.NewGraph()
			for i := range tt.nodes {
				graph.AddNode(&tt.nodes[i])
			}

			got := GeneratePlantedRelics(graph, 0)

			if len(got) != len(tt.wantIDs) {
				t.Fatalf("artifact count = %d, want %d", len(got), len(tt.wantIDs))
			}
			for i, a := range got {
				if a.ID != tt.wantIDs[i] {
					t.Errorf("artifact[%d] ID = %q, want %q", i, a.ID, tt.wantIDs[i])
				}
				if a.Name != tt.wantNames[i] {
					t.Errorf("artifact[%d] Name = %q, want %q", i, a.Name, tt.wantNames[i])
				}
			}
		})
	}
}

func TestGeneratePlantedRelicsIntrinsicFields(t *testing.T) {
	graph := pointcrawl.NewGraph()
	graph.AddNode(&pointcrawl.Node{ID: 0, Name: "Ruin-12-34", Kind: "ruin"})

	artifacts := GeneratePlantedRelics(graph, 42)
	if len(artifacts) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(artifacts))
	}

	a := artifacts[0]
	if a.SignificanceSource != "intrinsic" {
		t.Errorf("SignificanceSource = %q, want intrinsic", a.SignificanceSource)
	}
	if !a.IsSignificant {
		t.Errorf("IsSignificant = false, want true")
	}
	if a.Status != "lost" {
		t.Errorf("Status = %q, want lost", a.Status)
	}
	if a.SignificanceScore != 3 {
		t.Errorf("SignificanceScore = %d, want 3", a.SignificanceScore)
	}
	if a.SignificanceYear != 42 {
		t.Errorf("SignificanceYear = %d, want 42", a.SignificanceYear)
	}
	if a.Description != "" {
		t.Errorf("Description = %q, want empty", a.Description)
	}
	if a.PivotalEventID != "" {
		t.Errorf("PivotalEventID = %q, want empty", a.PivotalEventID)
	}
	if len(a.Provenance) != 0 {
		t.Errorf("Provenance = %v, want empty", a.Provenance)
	}
	if a.AssociatedEventIDs != nil {
		t.Errorf("AssociatedEventIDs = %v, want nil", a.AssociatedEventIDs)
	}
	if a.Powers != nil {
		t.Errorf("Powers = %v, want nil", a.Powers)
	}
}

func TestGeneratePlantedRelicsTypeRotatesOverVocabulary(t *testing.T) {
	wantTypes := []string{"weapon", "armor", "jewelry", "weapon", "armor", "crown", "relic", "tome"}

	graph := pointcrawl.NewGraph()
	for i := range wantTypes {
		graph.AddNode(&pointcrawl.Node{ID: i, Name: fmt.Sprintf("Ruin-%d-%d", i, i), Kind: "ruin"})
	}

	artifacts := GeneratePlantedRelics(graph, 0)
	if len(artifacts) != len(wantTypes) {
		t.Fatalf("artifact count = %d, want %d", len(artifacts), len(wantTypes))
	}

	for i, a := range artifacts {
		if a.Type != wantTypes[i] {
			t.Errorf("artifact[%d] Type = %q, want %q", i, a.Type, wantTypes[i])
		}
	}
}

func TestGeneratePlantedRelicsTypeWeightsMatchRarity(t *testing.T) {
	graph := pointcrawl.NewGraph()
	for i := 0; i < 8; i++ {
		graph.AddNode(&pointcrawl.Node{ID: i, Name: fmt.Sprintf("Ruin-%d-%d", i, i), Kind: "ruin"})
	}

	artifacts := GeneratePlantedRelics(graph, 0)

	common := 0
	for _, a := range artifacts {
		switch a.Type {
		case "weapon", "armor", "jewelry":
			common++
		}
	}
	if common != 5 {
		t.Errorf("common types over one cycle = %d, want 5 (common types repeat in the pattern)", common)
	}
}

func TestGeneratePlantedRelicsIsDeterministic(t *testing.T) {
	graph := pointcrawl.NewGraph()
	graph.AddNode(&pointcrawl.Node{ID: 7, Name: "Ruin-70-80", Kind: "ruin"})
	graph.AddNode(&pointcrawl.Node{ID: 3, Name: "Ruin-30-40", Kind: "ruin"})
	graph.AddNode(&pointcrawl.Node{ID: 1, Name: "Landmark-10-10", Kind: "landmark"})

	first := GeneratePlantedRelics(graph, 7)
	second := GeneratePlantedRelics(graph, 7)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected identical artifacts for the same graph")
	}
}

func TestGeneratePlantedRelicsNilGraph(t *testing.T) {
	if got := GeneratePlantedRelics(nil, 0); got != nil {
		t.Fatalf("expected nil result for nil graph, got %v", got)
	}
}
