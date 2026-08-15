package simulation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/artifact"
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
	if len(a.Powers) != 1 {
		t.Errorf("Powers = %v, want exactly one intrinsic power", a.Powers)
	}
}

func TestGeneratePlantedRelicsIntrinsicPowers(t *testing.T) {
	wantTypes := []string{"weapon", "armor", "jewelry", "weapon", "armor", "crown", "relic", "tome"}
	wantPowerTypes := map[string]string{
		"weapon":  "combat",
		"armor":   "combat",
		"jewelry": "influence",
		"crown":   "influence",
		"relic":   "narrative",
		"tome":    "narrative",
	}

	graph := pointcrawl.NewGraph()
	for i := range wantTypes {
		graph.AddNode(&pointcrawl.Node{ID: i, Name: fmt.Sprintf("Ruin-%d-%d", i, i), Kind: "ruin"})
	}

	artifacts := GeneratePlantedRelics(graph, 0)
	if len(artifacts) != len(wantTypes) {
		t.Fatalf("artifact count = %d, want %d", len(artifacts), len(wantTypes))
	}

	for i, a := range artifacts {
		if len(a.Powers) != 1 {
			t.Errorf("artifact[%d] (%s) Powers = %v, want exactly one intrinsic power", i, a.Type, a.Powers)
			continue
		}
		if want := wantPowerTypes[a.Type]; a.Powers[0].Type() != want {
			t.Errorf("artifact[%d] (%s) power type = %q, want %q", i, a.Type, a.Powers[0].Type(), want)
		}
		if source := powerSourceForTest(a.Powers[0]); source != "intrinsic" {
			t.Errorf("artifact[%d] (%s) power source = %q, want intrinsic", i, a.Type, source)
		}
	}
}

// powerSourceForTest extracts the Source field from a power for assertions.
func powerSourceForTest(p artifact.Power) string {
	switch v := p.(type) {
	case artifact.CombatPower:
		return v.Source
	case artifact.InfluencePower:
		return v.Source
	case artifact.NarrativePower:
		return v.Source
	}
	return ""
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

// ── ArtifactRegistry ───────────────────────────────────

// registryArtifacts is a canonical fixture exercising every index key:
// held/significant by owner, lost by fallback, and destroyed by status.
func registryArtifacts() []artifact.Artifact {
	return []artifact.Artifact{
		{
			ID:     "artifact-settlement-0",
			Name:   "Crown of Deepcrest",
			Status: "held",
			Provenance: []artifact.ProvenanceEntry{
				{Year: 4, Owner: artifact.Owner{Kind: "settlement", ID: "Deepcrest"}},
			},
		},
		{
			ID:     "artifact-settlement-1",
			Name:   "Sword of the Fallen",
			Status: "significant",
			Provenance: []artifact.ProvenanceEntry{
				{Year: 9, Owner: artifact.Owner{Kind: "figure", ID: "Deepcrest-3"}},
			},
		},
		{ID: "artifact-ruin-0", Name: "Relic of Ruin", Status: "lost"},
		{
			ID:     "artifact-settlement-2",
			Name:   "Broken Crown",
			Status: "destroyed",
			Provenance: []artifact.ProvenanceEntry{
				{Year: 21, Owner: artifact.Owner{Kind: "settlement", ID: "Blackgate"}},
			},
		},
	}
}

func TestArtifactRegistryGet(t *testing.T) {
	r := NewArtifactRegistry(registryArtifacts())

	for _, want := range registryArtifacts() {
		got, ok := r.Get(want.ID)
		if !ok {
			t.Errorf("Get(%q) = not found, want the artifact", want.ID)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Get(%q) = %+v, want %+v", want.ID, got, want)
		}
	}

	if _, ok := r.Get("artifact-settlement-9"); ok {
		t.Error("Get(unknown ID) = found, want not found")
	}
}

func TestArtifactRegistryArtifactsFor(t *testing.T) {
	r := NewArtifactRegistry(registryArtifacts())

	tests := []struct {
		name     string
		kind, id string
		wantIDs  []string
	}{
		{"held settlement artifact", "settlement", "Deepcrest", []string{"artifact-settlement-0"}},
		{"figure-owned artifact", "figure", "Deepcrest-3", []string{"artifact-settlement-1"}},
		{"lost artifact by fallback", "lost", "", []string{"artifact-ruin-0"}},
		{"destroyed artifact by status", "destroyed", "", []string{"artifact-settlement-2"}},
		{"unknown owner is empty", "settlement", "Blackgate", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.ArtifactsFor(tt.kind, tt.id)
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("ArtifactsFor(%q, %q) = %d artifacts, want %d", tt.kind, tt.id, len(got), len(tt.wantIDs))
			}
			for i, want := range tt.wantIDs {
				if got[i].ID != want {
					t.Errorf("artifact[%d] = %q, want %q", i, got[i].ID, want)
				}
			}
		})
	}
}

func TestArtifactRegistryArtifactsForPreservesStateOrder(t *testing.T) {
	r := NewArtifactRegistry([]artifact.Artifact{
		{
			ID:     "artifact-a-0",
			Status: "held",
			Provenance: []artifact.ProvenanceEntry{
				{Owner: artifact.Owner{Kind: "settlement", ID: "Haven"}},
			},
		},
		{ID: "artifact-ruin-0", Status: "lost"},
		{
			ID:     "artifact-a-1",
			Status: "held",
			Provenance: []artifact.ProvenanceEntry{
				{Owner: artifact.Owner{Kind: "settlement", ID: "Haven"}},
			},
		},
	})

	got := r.ArtifactsFor("settlement", "Haven")
	if len(got) != 2 || got[0].ID != "artifact-a-0" || got[1].ID != "artifact-a-1" {
		t.Fatalf("ArtifactsFor = %v, want world-state order [artifact-a-0 artifact-a-1]", got)
	}
}

func TestArtifactRegistryEmpty(t *testing.T) {
	r := NewArtifactRegistry(nil)
	if got := r.ArtifactsFor("settlement", "Haven"); len(got) != 0 {
		t.Fatalf("ArtifactsFor on empty registry = %v, want none", got)
	}
	if _, ok := r.Get("artifact-a-0"); ok {
		t.Fatal("Get on empty registry found an artifact")
	}
}

func TestArtifactRegistryUnlose(t *testing.T) {
	r := NewArtifactRegistry([]artifact.Artifact{
		{ID: "artifact-ruin-0", Name: "Relic of Ruin", Status: "lost"},
		{ID: "artifact-ruin-1", Name: "Relic of Ruin II", Status: "lost"},
		{
			ID:     "artifact-settlement-0",
			Name:   "Crown of Deepcrest",
			Status: "held",
			Provenance: []artifact.ProvenanceEntry{
				{Owner: artifact.Owner{Kind: "settlement", ID: "Haven"}},
			},
		},
	})

	if err := r.Unlose("artifact-ruin-0", "settlement", "Haven", "event-12-3"); err != nil {
		t.Fatalf("Unlose(lost artifact) = %v, want nil", err)
	}

	a, ok := r.Get("artifact-ruin-0")
	if !ok || a.Status != "held" {
		t.Fatalf("Get after Unlose = %+v, ok=%v; want status held", a, ok)
	}

	// The artifact leaves the lost bucket — leaving the other lost relic —
	// and joins the new owner, appended after pre-existing entries so query
	// order stays deterministic.
	lost := r.ArtifactsFor("lost", "")
	if len(lost) != 1 || lost[0].ID != "artifact-ruin-1" {
		t.Fatalf("lost bucket after Unlose = %v, want [artifact-ruin-1]", lost)
	}
	held := r.ArtifactsFor("settlement", "Haven")
	if len(held) != 2 || held[0].ID != "artifact-settlement-0" || held[1].ID != "artifact-ruin-0" {
		t.Fatalf("ArtifactsFor(new owner) = %v, want [artifact-settlement-0 artifact-ruin-0]", held)
	}
}

func TestArtifactRegistryUnloseErrors(t *testing.T) {
	r := NewArtifactRegistry(registryArtifacts())

	if err := r.Unlose("artifact-settlement-0", "settlement", "Haven", "event-1"); err == nil {
		t.Error("Unlose(held artifact) = nil, want error")
	}
	if err := r.Unlose("artifact-settlement-2", "settlement", "Haven", "event-1"); err == nil {
		t.Error("Unlose(destroyed artifact) = nil, want error")
	}
	if err := r.Unlose("artifact-settlement-9", "settlement", "Haven", "event-1"); err == nil {
		t.Error("Unlose(unknown artifact) = nil, want error")
	}
}

func TestArtifactRegistryQueriesDeterministicForSameInput(t *testing.T) {
	first := NewArtifactRegistry(registryArtifacts())
	second := NewArtifactRegistry(registryArtifacts())

	for key, want := range first.byOwner {
		if got := second.byOwner[key]; !reflect.DeepEqual(got, want) {
			t.Errorf("byOwner[%+v] = %v, want %v", key, got, want)
		}
	}
	for id, want := range first.byID {
		if got := second.byID[id]; !reflect.DeepEqual(got, want) {
			t.Errorf("byID[%q] = %v, want %v", id, got, want)
		}
	}
}
