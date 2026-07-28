package pointcrawl

import (
	randv2 "math/rand/v2"
	"testing"
)

func newTestRNG() *randv2.Rand {
	return randv2.New(randv2.NewPCG(42, 7))
}

func buildGraph(nodes ...*Node) *Graph {
	g := NewGraph()
	for _, n := range nodes {
		g.AddNode(n)
	}
	return g
}

func TestFindExpansionTargetReturnsNilWhenNoTargets(t *testing.T) {
	g := buildGraph()
	sites := []SettlementSite{{Name: "Alpha", X: 0, Y: 0, Faction: "auric"}}

	got := FindExpansionTarget(g, 0, 0, "auric", sites, 10, 3, newTestRNG())
	if got != nil {
		t.Fatalf("FindExpansionTarget = %v, want nil", got)
	}
}

func TestFindExpansionTargetNilGraph(t *testing.T) {
	got := FindExpansionTarget(nil, 0, 0, "auric", nil, 10, 3, newTestRNG())
	if got != nil {
		t.Fatalf("FindExpansionTarget = %v, want nil", got)
	}
}

func TestFindExpansionTargetReturnsValidNode(t *testing.T) {
	g := buildGraph(
		&Node{ID: 0, X: 0, Y: 0, Visibility: Known, Name: "Alpha", Kind: "settlement"},
		&Node{ID: 1, X: 8, Y: 0, Visibility: Unknown, Name: "Ruins", Kind: "ruin"},
	)
	sites := []SettlementSite{{Name: "Alpha", X: 0, Y: 0, Faction: "auric"}}

	got := FindExpansionTarget(g, 0, 0, "auric", sites, 10, 3, newTestRNG())
	if got == nil {
		t.Fatal("FindExpansionTarget = nil, want node 1")
	}
	if got.ID != 1 {
		t.Fatalf("FindExpansionTarget = node %d, want node 1", got.ID)
	}
}

func TestFindExpansionTargetExcludesKnownNodes(t *testing.T) {
	g := buildGraph(
		&Node{ID: 0, X: 5, Y: 0, Visibility: Known, Name: "Beta", Kind: "settlement"},
	)
	sites := []SettlementSite{{Name: "Alpha", X: 0, Y: 0, Faction: "auric"}}

	got := FindExpansionTarget(g, 0, 0, "auric", sites, 10, 1, newTestRNG())
	if got != nil {
		t.Fatalf("FindExpansionTarget = %v, want nil (known node excluded)", got)
	}
}

func TestFindExpansionTargetExcludesTooCloseNodes(t *testing.T) {
	g := buildGraph(
		&Node{ID: 0, X: 0, Y: 0, Visibility: Known, Name: "Alpha", Kind: "settlement"},
		&Node{ID: 1, X: 2, Y: 0, Visibility: Unknown, Name: "Near", Kind: "ruin"},
		&Node{ID: 2, X: 9, Y: 0, Visibility: Unknown, Name: "Far", Kind: "ruin"},
	)
	sites := []SettlementSite{{Name: "Alpha", X: 0, Y: 0, Faction: "auric"}}

	got := FindExpansionTarget(g, 0, 0, "auric", sites, 10, 3, newTestRNG())
	if got == nil {
		t.Fatal("FindExpansionTarget = nil, want node 2")
	}
	if got.ID != 2 {
		t.Fatalf("FindExpansionTarget = node %d, want node 2 (node 1 too close)", got.ID)
	}
}

func TestFindExpansionTargetExcludesOtherFactionInfluence(t *testing.T) {
	g := buildGraph(
		&Node{ID: 0, X: 0, Y: 0, Visibility: Known, Name: "Alpha", Kind: "settlement"},
		&Node{ID: 1, X: 20, Y: 0, Visibility: Known, Name: "Beta", Kind: "settlement"},
		&Node{ID: 2, X: 18, Y: 0, Visibility: Unknown, Name: "Contested", Kind: "ruin"},
	)
	sites := []SettlementSite{
		{Name: "Alpha", X: 0, Y: 0, Faction: "auric"},
		{Name: "Beta", X: 20, Y: 0, Faction: "sylvani"},
	}

	got := FindExpansionTarget(g, 0, 0, "auric", sites, 30, 1, newTestRNG())
	if got != nil {
		t.Fatalf("FindExpansionTarget = node %d, want nil (inside sylvani influence)", got.ID)
	}
}

func TestFindExpansionTargetIndependentIgnoresInfluence(t *testing.T) {
	g := buildGraph(
		&Node{ID: 0, X: 20, Y: 0, Visibility: Known, Name: "Beta", Kind: "settlement"},
		&Node{ID: 1, X: 18, Y: 0, Visibility: Unknown, Name: "Contested", Kind: "ruin"},
	)
	sites := []SettlementSite{
		{Name: "Alpha", X: 0, Y: 0, Faction: "independent"},
		{Name: "Beta", X: 20, Y: 0, Faction: "sylvani"},
	}

	got := FindExpansionTarget(g, 0, 0, "independent", sites, 30, 1, newTestRNG())
	if got == nil {
		t.Fatal("FindExpansionTarget = nil, want node 1 (independent ignores influence)")
	}
	if got.ID != 1 {
		t.Fatalf("FindExpansionTarget = node %d, want node 1", got.ID)
	}
}

func TestFindExpansionTargetDeterministicSelection(t *testing.T) {
	build := func() *Graph {
		return buildGraph(
			&Node{ID: 0, X: 0, Y: 0, Visibility: Known, Name: "Alpha", Kind: "settlement"},
			&Node{ID: 1, X: 5, Y: 0, Visibility: Unknown, Name: "N1", Kind: "ruin"},
			&Node{ID: 2, X: 7, Y: 0, Visibility: Unknown, Name: "N2", Kind: "ruin"},
			&Node{ID: 3, X: 9, Y: 0, Visibility: Unknown, Name: "N3", Kind: "ruin"},
		)
	}
	sites := []SettlementSite{{Name: "Alpha", X: 0, Y: 0, Faction: "auric"}}

	first := FindExpansionTarget(build(), 0, 0, "auric", sites, 10, 1, newTestRNG())
	second := FindExpansionTarget(build(), 0, 0, "auric", sites, 10, 1, newTestRNG())

	if first == nil || second == nil {
		t.Fatal("FindExpansionTarget returned nil")
	}
	if first.ID != second.ID {
		t.Fatalf("non-deterministic selection: first=%d second=%d", first.ID, second.ID)
	}
}
