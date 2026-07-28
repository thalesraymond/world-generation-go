package world

import (
	"testing"
)

func TestInitRelations(t *testing.T) {
	self := Settlement{Name: "Alpha", Faction: "auric"}
	all := []Settlement{
		{Name: "Alpha", Faction: "auric"},
		{Name: "Beta", Faction: "auric"},
		{Name: "Gamma", Faction: "sylvani"},
		{Name: "Delta", Faction: "independent"},
	}

	relations := InitRelations(self, all)

	if len(relations) != 3 {
		t.Fatalf("InitRelations returned %d entries, want 3", len(relations))
	}
	if _, ok := relations["Alpha"]; ok {
		t.Fatal("InitRelations included self")
	}
	if got := relations["Beta"]; got != 0.3 {
		t.Fatalf("same-faction relation = %v, want 0.3", got)
	}
	if got := relations["Gamma"]; got != 0.0 {
		t.Fatalf("different-faction relation = %v, want 0.0", got)
	}
	if got := relations["Delta"]; got != 0.0 {
		t.Fatalf("independent-faction relation = %v, want 0.0", got)
	}
}

func TestInitRelationsIndependentSelf(t *testing.T) {
	self := Settlement{Name: "Alpha", Faction: "independent"}
	all := []Settlement{
		{Name: "Alpha", Faction: "independent"},
		{Name: "Beta", Faction: "independent"},
	}

	relations := InitRelations(self, all)

	if got := relations["Beta"]; got != 0.0 {
		t.Fatalf("independent-to-independent relation = %v, want 0.0", got)
	}
}

func approxEqual(a, b float64) bool {
	const eps = 1e-9
	d := a - b
	return d < eps && d > -eps
}

func TestShiftRelationsWithinBounds(t *testing.T) {
	s := Settlement{Name: "Alpha", Relations: map[string]float64{"Beta": 0.3}}

	ShiftRelations(&s, "Beta", 0.4)
	if got := s.Relations["Beta"]; !approxEqual(got, 0.7) {
		t.Fatalf("relations after shift = %v, want 0.7", got)
	}
}

func TestShiftRelationsClampsAtUpperBound(t *testing.T) {
	s := Settlement{Name: "Alpha", Relations: map[string]float64{"Beta": 0.8}}

	ShiftRelations(&s, "Beta", 0.5)
	if got := s.Relations["Beta"]; got != RelationMax {
		t.Fatalf("relations after clamp = %v, want %v", got, RelationMax)
	}
}

func TestShiftRelationsClampsAtLowerBound(t *testing.T) {
	s := Settlement{Name: "Alpha", Relations: map[string]float64{"Beta": -0.8}}

	ShiftRelations(&s, "Beta", -0.5)
	if got := s.Relations["Beta"]; got != RelationMin {
		t.Fatalf("relations after clamp = %v, want %v", got, RelationMin)
	}
}

func TestShiftRelationsAccumulate(t *testing.T) {
	s := Settlement{Name: "Alpha", Relations: map[string]float64{"Beta": 0.0}}

	ShiftRelations(&s, "Beta", RelationShiftRaidSuccessSelf)
	if got := s.Relations["Beta"]; !approxEqual(got, -0.4) {
		t.Fatalf("after 1 raid relations = %v, want -0.4", got)
	}
	ShiftRelations(&s, "Beta", RelationShiftRaidSuccessSelf)
	if got := s.Relations["Beta"]; !approxEqual(got, -0.8) {
		t.Fatalf("after 2 raids relations = %v, want -0.8", got)
	}
	// Third shift would pass -1.0 and must clamp.
	ShiftRelations(&s, "Beta", RelationShiftRaidSuccessSelf)
	if got := s.Relations["Beta"]; got != RelationMin {
		t.Fatalf("after 3 raids relations = %v, want %v (clamped)", got, RelationMin)
	}
}

func TestShiftRelationsInitializesNilMap(t *testing.T) {
	s := Settlement{Name: "Alpha"}

	ShiftRelations(&s, "Beta", 0.4)
	if got := s.Relations["Beta"]; got != 0.4 {
		t.Fatalf("relations on nil map = %v, want 0.4", got)
	}
}

func TestShiftRelationsIgnoresSelfAndEmpty(t *testing.T) {
	s := Settlement{Name: "Alpha", Relations: map[string]float64{}}

	ShiftRelations(&s, "Alpha", 0.4)
	ShiftRelations(&s, "", 0.4)
	if len(s.Relations) != 0 {
		t.Fatalf("relations map = %v, want empty", s.Relations)
	}
}

func TestShiftRelationsAsymmetric(t *testing.T) {
	a := Settlement{Name: "Alpha", Relations: map[string]float64{"Beta": 0.5}}
	b := Settlement{Name: "Beta", Relations: map[string]float64{"Alpha": -0.5}}

	ShiftRelations(&a, "Beta", RelationShiftRaidSuccessSelf)

	if got := a.Relations["Beta"]; !approxEqual(got, 0.1) {
		t.Fatalf("a->Beta = %v, want 0.1", got)
	}
	if got := b.Relations["Alpha"]; got != -0.5 {
		t.Fatalf("b->Alpha = %v, want -0.5 (unchanged)", got)
	}
}
