package world

import (
	randv2 "math/rand/v2"
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

func TestApplyCrossFactionFriction(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(42, 99))
	settlements := []Settlement{
		{Name: "Alpha", Faction: "auric", Relations: map[string]float64{}},
		{Name: "Beta", Faction: "auric", Relations: map[string]float64{}},
		{Name: "Gamma", Faction: "sylvani", Relations: map[string]float64{}},
		{Name: "Delta", Faction: "verdant", Relations: map[string]float64{}},
	}

	// Init relations first.
	for i := range settlements {
		settlements[i].Relations = InitRelations(settlements[i], settlements)
	}

	ApplyCrossFactionFriction(settlements, rng)

	// Same-faction pair: Alpha <-> Beta must remain +0.3.
	if got := settlements[0].Relations["Beta"]; got != RelationShiftSameFactionBaseline {
		t.Errorf("Alpha->Beta (same faction) = %v, want %v", got, RelationShiftSameFactionBaseline)
	}
	if got := settlements[1].Relations["Alpha"]; got != RelationShiftSameFactionBaseline {
		t.Errorf("Beta->Alpha (same faction) = %v, want %v", got, RelationShiftSameFactionBaseline)
	}

	// Cross-faction pairs must be negative.
	checkNegative := func(a, b int) {
		got := settlements[a].Relations[settlements[b].Name]
		if got >= 0 {
			t.Errorf("%s->%s (cross-faction) = %v, want negative", settlements[a].Name, settlements[b].Name, got)
		}
	}
	checkNegative(0, 2) // Alpha -> Gamma
	checkNegative(0, 3) // Alpha -> Delta
	checkNegative(2, 3) // Gamma -> Delta
}

func TestApplyCrossFactionFrictionExcludesSameFaction(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(77, 13))
	settlements := []Settlement{
		{Name: "Alpha", Faction: "auric", Relations: map[string]float64{}},
		{Name: "Beta", Faction: "auric", Relations: map[string]float64{}},
	}

	for i := range settlements {
		settlements[i].Relations = InitRelations(settlements[i], settlements)
	}

	ApplyCrossFactionFriction(settlements, rng)

	// Same-faction relations must remain unchanged at +0.3.
	if got := settlements[0].Relations["Beta"]; got != RelationShiftSameFactionBaseline {
		t.Errorf("same-faction relation modified by friction: %v, want %v", got, RelationShiftSameFactionBaseline)
	}
}

func TestApplyCrossFactionFrictionExcludesIndependent(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(77, 13))
	settlements := []Settlement{
		{Name: "Alpha", Faction: "auric", Relations: map[string]float64{}},
		{Name: "Beta", Faction: "independent", Relations: map[string]float64{}},
	}

	for i := range settlements {
		settlements[i].Relations = InitRelations(settlements[i], settlements)
	}

	ApplyCrossFactionFriction(settlements, rng)

	// Independent faction must not be modified by friction.
	if got := settlements[0].Relations["Beta"]; got != 0.0 {
		t.Errorf("independent relation modified by friction: %v, want 0.0", got)
	}
	if got := settlements[1].Relations["Alpha"]; got != 0.0 {
		t.Errorf("independent relation modified by friction: %v, want 0.0", got)
	}
}

func TestApplyCrossFactionFrictionDeterministic(t *testing.T) {
	seed := uint64(42)
	run := func() float64 {
		rng := randv2.New(randv2.NewPCG(seed, seed^0x9e3779b9))
		settlements := []Settlement{
			{Name: "Alpha", Faction: "auric", Relations: map[string]float64{}},
			{Name: "Beta", Faction: "sylvani", Relations: map[string]float64{}},
		}
		for i := range settlements {
			settlements[i].Relations = InitRelations(settlements[i], settlements)
		}
		ApplyCrossFactionFriction(settlements, rng)
		return settlements[0].Relations["Beta"]
	}

	first := run()
	for i := 0; i < 10; i++ {
		if got := run(); got != first {
			t.Fatalf("non-deterministic friction at iteration %d: %v vs %v", i, got, first)
		}
	}
}

func TestApplySettlementCrossFactionFriction(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(42, 99))
	existing := []Settlement{
		{Name: "Alpha", Faction: "auric", Relations: map[string]float64{}},
		{Name: "Beta", Faction: "sylvani", Relations: map[string]float64{}},
		{Name: "Gamma", Faction: "verdant", Relations: map[string]float64{}},
	}

	// Init relations for existing settlements.
	for i := range existing {
		existing[i].Relations = InitRelations(existing[i], existing)
	}
	// Apply bulk friction.
	ApplyCrossFactionFriction(existing, rng)

	// Add a new settlement mid-simulation (like ExpandAction does).
	childRNG := randv2.New(randv2.NewPCG(1, 1))
	child := Settlement{Name: "Newhold", Faction: "auric", Relations: map[string]float64{}}
	child.Relations = InitRelations(child, existing)
	ApplySettlementCrossFactionFriction(&child, existing, childRNG)

	// Same-faction: Alpha (auric) <-> Newhold (auric) should be +0.3.
	if got := child.Relations["Alpha"]; got != RelationShiftSameFactionBaseline {
		t.Errorf("Newhold->Alpha (same faction) = %v, want %v", got, RelationShiftSameFactionBaseline)
	}

	// Cross-faction: Newhold -> Beta/Gamma must be negative.
	for _, name := range []string{"Beta", "Gamma"} {
		if got := child.Relations[name]; got >= 0 {
			t.Errorf("Newhold->%s (cross-faction) = %v, want negative", name, got)
		}
	}

	// Symmetry: existing settlements should have friction toward Newhold.
	for _, name := range []string{"Beta", "Gamma"} {
		for _, other := range existing {
			if other.Name == name {
				if got := other.Relations["Newhold"]; got >= 0 {
					t.Errorf("%s->Newhold (cross-faction) = %v, want negative", name, got)
				}
			}
		}
	}
}
