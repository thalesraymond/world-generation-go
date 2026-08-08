package figures

import (
	randv2 "math/rand/v2"
	"reflect"
	"testing"
)

func newTestRNG(seed uint64) *randv2.Rand {
	return randv2.New(randv2.NewPCG(seed, seed+1))
}

func TestGenerateFounders_Count(t *testing.T) {
	for seed := uint64(0); seed < 100; seed++ {
		rng := newTestRNG(seed)
		founders := GenerateFounders(rng, "Test", "faction", 0)
		if len(founders) < 3 || len(founders) > 5 {
			t.Fatalf("expected 3-5 founders, got %d for seed %d", len(founders), seed)
		}
	}
}

func TestGenerateFounders_FirstIsLeader(t *testing.T) {
	rng := newTestRNG(42)
	founders := GenerateFounders(rng, "Test", "faction", 0)
	if founders[0].Role != "Leader" {
		t.Fatalf("expected first founder to be Leader, got %q", founders[0].Role)
	}
}

func TestGenerateFounders_Determinism(t *testing.T) {
	seed := uint64(123)
	rng1 := newTestRNG(seed)
	rng2 := newTestRNG(seed)
	f1 := GenerateFounders(rng1, "Test", "faction", 100)
	f2 := GenerateFounders(rng2, "Test", "faction", 100)
	if !reflect.DeepEqual(f1, f2) {
		t.Fatalf("expected deterministic founders, got different results")
	}
}

func TestCheckDeaths_MaxAge(t *testing.T) {
	currentYear := 100
	figures := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: currentYear - 70, MaxAge: 70, Faction: "f"},
	}
	rng := newTestRNG(1)
	events := CheckDeaths(figures, currentYear, rng)
	if figures[0].DeathYear != currentYear {
		t.Fatalf("expected figure to die at current year, got %d", figures[0].DeathYear)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 death event, got %d", len(events))
	}
}

func TestCheckDeaths_UnderMaxAge_Alive(t *testing.T) {
	currentYear := 100
	figures := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: currentYear - 20, MaxAge: 70, Faction: "f"},
	}
	rng := newTestRNG(1)
	events := CheckDeaths(figures, currentYear, rng)
	if !figures[0].IsAlive() {
		t.Fatalf("expected figure to stay alive")
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestCheckDeaths_ProducesDeathEvent(t *testing.T) {
	currentYear := 100
	figures := []HistoricalFigure{
		{ID: "a", Name: "Aldric", BirthYear: currentYear - 80, MaxAge: 70, Faction: "faction", Role: "Leader"},
	}
	rng := newTestRNG(1)
	events := CheckDeaths(figures, currentYear, rng)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	e := events[0]
	if e.Year != currentYear {
		t.Errorf("expected year %d, got %d", currentYear, e.Year)
	}
	if e.Category != "Death" {
		t.Errorf("expected category Death, got %q", e.Category)
	}
	if e.FigureID != "a" {
		t.Errorf("expected FigureID a, got %q", e.FigureID)
	}
	if e.SettlementName != "faction" {
		t.Errorf("expected SettlementName faction, got %q", e.SettlementName)
	}
}

func TestCheckBirths_UnderCap(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: 0, MaxAge: 70},
	}
	rng := newTestRNG(1)
	child := CheckBirths(figures, 20000, 100, "Haven", rng)
	if child == nil {
		t.Fatalf("expected a birth under cap")
	}
}

func TestCheckBirths_AtCap(t *testing.T) {
	figures := make([]HistoricalFigure, maxFiguresPerSettlement)
	for i := range figures {
		figures[i] = HistoricalFigure{ID: "alive", Name: "A", BirthYear: 0, MaxAge: 70}
	}
	rng := newTestRNG(1)
	child := CheckBirths(figures, 99999, 100, "Haven", rng)
	if child != nil {
		t.Fatalf("expected no birth when at cap")
	}
}

func TestCheckBirths_NewFigureHasCorrectYear(t *testing.T) {
	rng := newTestRNG(1)
	currentYear := 100
	child := CheckBirths(nil, 20000, currentYear, "Haven", rng)
	if child == nil {
		t.Fatalf("expected a birth")
	}
	if child.BirthYear != currentYear {
		t.Fatalf("expected birth year %d, got %d", currentYear, child.BirthYear)
	}
}

func TestCheckBirths_PerSettlementOrdinalID(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "Haven-0", Name: "Founder", BirthYear: 0, MaxAge: 70},
	}
	rng := newTestRNG(1)
	child := CheckBirths(figures, 20000, 100, "Haven", rng)
	if child == nil {
		t.Fatalf("expected a birth")
	}
	if child.ID != "Haven-1" {
		t.Fatalf("expected ID %q, got %q", "Haven-1", child.ID)
	}
}

func TestCheckBirths_IDUniquenessAcrossPipeline(t *testing.T) {
	const settlement = "Ironhold"
	const faction = "Ironbound"
	const foundingYear = 100

	rng := newTestRNG(42)
	figs := GenerateFounders(rng, settlement, faction, foundingYear)
	seen := make(map[string]bool)
	for _, f := range figs {
		seen[f.ID] = true
	}

	for year := foundingYear + 1; year <= foundingYear+100; year++ {
		CheckDeaths(figs, year, rng)
		child := CheckBirths(figs, 15000, year, settlement, rng)
		if child != nil {
			if seen[child.ID] {
				t.Fatalf("duplicate figure ID %q", child.ID)
			}
			seen[child.ID] = true
			figs = append(figs, *child)
		}
	}
}

func TestAssignRoles_NoLeader_Assigns(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: 0, MaxAge: 70},
	}
	rng := newTestRNG(1)
	events := AssignRoles(figures, nil, 0, 0, rng)
	if figures[0].Role != "Leader" {
		t.Fatalf("expected figure to be assigned Leader, got %q", figures[0].Role)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestFullPipelineDeterminism(t *testing.T) {
	const seed = uint64(42)
	const settlement = "Ironhold"
	const faction = "Ironbound"
	const foundingYear = 100

	rng1 := newTestRNG(seed)
	rng2 := newTestRNG(seed)

	runPipeline := func(rng *randv2.Rand) []HistoricalFigure {
		figs := GenerateFounders(rng, settlement, faction, foundingYear)
		for year := foundingYear + 1; year <= foundingYear+50; year++ {
			CheckDeaths(figs, year, rng)
			child := CheckBirths(figs, 15000, year, settlement, rng)
			if child != nil {
				figs = append(figs, *child)
			}
			AssignRoles(figs, nil, 0, 0, rng)
		}
		return figs
	}

	result1 := runPipeline(rng1)
	result2 := runPipeline(rng2)

	if len(result1) != len(result2) {
		t.Fatalf("figure count differs: %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		f1 := result1[i]
		f2 := result2[i]
		if f1.ID != f2.ID || f1.Name != f2.Name || f1.Role != f2.Role ||
			f1.BirthYear != f2.BirthYear || f1.DeathYear != f2.DeathYear ||
			f1.MaxAge != f2.MaxAge || f1.Faction != f2.Faction {
			t.Errorf("figure[%d] differs: %+v vs %+v", i, f1, f2)
		}
	}
}

func TestAssignRoles_HasLeader_NoChange(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "leader", Name: "L", BirthYear: 0, MaxAge: 70, Role: "Leader"},
		{ID: "b", Name: "B", BirthYear: 0, MaxAge: 70},
	}
	rng := newTestRNG(1)
	events := AssignRoles(figures, nil, 0, 0, rng)
	if figures[1].Role != "" {
		t.Fatalf("expected roleless figure to remain unchanged, got %q", figures[1].Role)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

// 15.3: Leader availability — after leader dies and succession, new leader exists.
func TestLeaderAvailability_AfterSuccession(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "leader", Name: "A", BirthYear: 0, MaxAge: 50, Role: "Leader"},
		{ID: "b", Name: "B", BirthYear: 10, MaxAge: 70},
		{ID: "c", Name: "C", BirthYear: 10, MaxAge: 70},
	}
	currentYear := 60
	rng := newTestRNG(42)

	deathEvents := CheckDeaths(figures, currentYear, rng)
	_ = deathEvents
	AssignRoles(figures, nil, 0, 0, rng)

	leaderCount := 0
	for _, f := range figures {
		if f.IsAlive() && f.Role == "Leader" {
			leaderCount++
		}
	}
	if leaderCount != 1 {
		t.Fatalf("expected exactly 1 leader after succession, got %d", leaderCount)
	}
}

// 15.4: Empty figures — nil and empty slices handled without panics.
func TestEmptyFigures_Nil(t *testing.T) {
	rng := newTestRNG(1)

	deathEvents := CheckDeaths(nil, 100, rng)
	if len(deathEvents) != 0 {
		t.Fatalf("expected 0 death events for nil figures, got %d", len(deathEvents))
	}

	// CheckBirths is probabilistic — with nil figures and high population
	// a birth is likely but not guaranteed. The key assertion is no panic.
	CheckBirths(nil, 10000, 100, "Haven", rng)

	assignEvents := AssignRoles(nil, nil, 0, 0, rng)
	if len(assignEvents) != 0 {
		t.Fatalf("expected 0 assign events for nil figures, got %d", len(assignEvents))
	}
}

func TestEmptyFigures_EmptySlice(t *testing.T) {
	rng := newTestRNG(1)
	empty := []HistoricalFigure{}

	deathEvents := CheckDeaths(empty, 100, rng)
	if len(deathEvents) != 0 {
		t.Fatalf("expected 0 death events for empty figures, got %d", len(deathEvents))
	}

	// CheckBirths is probabilistic — the key assertion is no panic.
	CheckBirths(empty, 10000, 100, "Haven", rng)

	assignEvents := AssignRoles(empty, nil, 0, 0, rng)
	if len(assignEvents) != 0 {
		t.Fatalf("expected 0 assign events for empty figures, got %d", len(assignEvents))
	}
}

// 15.5: Death triggers succession — leader dies → new leader exists → succession event emitted.
func TestDeathTriggersSuccession(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "leader", Name: "A", BirthYear: 0, MaxAge: 50, Role: "Leader"},
		{ID: "b", Name: "B", BirthYear: 10, MaxAge: 70},
		{ID: "c", Name: "C", BirthYear: 10, MaxAge: 70},
	}
	currentYear := 60
	rng := newTestRNG(42)

	deathEvents := CheckDeaths(figures, currentYear, rng)
	if len(deathEvents) != 1 {
		t.Fatalf("expected 1 death event, got %d", len(deathEvents))
	}
	if deathEvents[0].Category != "Death" {
		t.Fatalf("expected Death event, got %q", deathEvents[0].Category)
	}

	successionEvents := AssignRoles(figures, nil, 0, 0, rng)
	if len(successionEvents) != 1 {
		t.Fatalf("expected 1 succession event, got %d", len(successionEvents))
	}
	if successionEvents[0].Category != "Politics" {
		t.Fatalf("expected Politics event, got %q", successionEvents[0].Category)
	}

	leaderCount := 0
	for _, f := range figures {
		if f.IsAlive() && f.Role == "Leader" {
			leaderCount++
		}
	}
	if leaderCount != 1 {
		t.Fatalf("expected exactly 1 leader after death+succession, got %d", leaderCount)
	}
}

func TestCheckDeaths_SuccessionWithHeir(t *testing.T) {
	parent := HistoricalFigure{ID: "leader-1", Name: "King Alaric", BirthYear: 0, MaxAge: 50, Faction: "f", Role: "Leader", Stats: Stats{Martial: 10, Diplomatic: 10, Infamy: 5}}
	child := HistoricalFigure{ID: "heir-1", Name: "Prince Aldric", BirthYear: 20, MaxAge: 70, Faction: "f", Stats: Stats{Martial: 5, Diplomatic: 5, Infamy: 1}}
	AddParentChild(&parent, &child)

	figures := []HistoricalFigure{parent, child}
	events := CheckDeaths(figures, 60, newTestRNG(1))

	hasDeath := false
	hasSuccession := false
	for _, e := range events {
		if e.Category == "Death" {
			hasDeath = true
		}
		if e.Category == "Succession" {
			hasSuccession = true
		}
	}
	if !hasDeath {
		t.Error("expected Death event")
	}
	if !hasSuccession {
		t.Error("expected Succession event")
	}

	if figures[1].Role != "Leader" {
		t.Errorf("heir Role = %q, want Leader", figures[1].Role)
	}
	if figures[1].Stats.Martial != 6 || figures[1].Stats.Diplomatic != 6 || figures[1].Stats.Infamy != 2 {
		t.Errorf("heir stats = %+v, want Martial=6 Diplomatic=6 Infamy=2", figures[1].Stats)
	}
	if figures[1].ParentID != parent.ID {
		t.Errorf("heir ParentID = %q, want %q", figures[1].ParentID, parent.ID)
	}
}

func TestCheckDeaths_NoHeirFallback(t *testing.T) {
	parent := HistoricalFigure{ID: "leader-2", Name: "Queen Elara", BirthYear: 0, MaxAge: 50, Faction: "f", Role: "Leader"}
	other := HistoricalFigure{ID: "oth-1", Name: "Stranger", BirthYear: 10, MaxAge: 70, Faction: "f", Role: "Explorer"}

	figures := []HistoricalFigure{parent, other}
	events := CheckDeaths(figures, 60, newTestRNG(1))

	hasSuccession := false
	for _, e := range events {
		if e.Category == "Succession" {
			hasSuccession = true
		}
	}
	if hasSuccession {
		t.Error("should not have succession when leader had no heir")
	}
}

func TestCheckDeaths_StatInheritanceCap(t *testing.T) {
	parent := HistoricalFigure{ID: "leader-3", Name: "Legend", BirthYear: 0, MaxAge: 50, Faction: "f", Role: "Leader", Stats: Stats{Martial: 20, Diplomatic: 20, Infamy: 20}}
	child := HistoricalFigure{ID: "heir-3", Name: "Prodigy", BirthYear: 20, MaxAge: 70, Faction: "f", Stats: Stats{Martial: 19, Diplomatic: 19, Infamy: 19}}
	AddParentChild(&parent, &child)

	figures := []HistoricalFigure{parent, child}
	CheckDeaths(figures, 60, newTestRNG(1))

	if figures[1].Stats.Martial != 20 {
		t.Errorf("heir stats capped at 20, got %d", figures[1].Stats.Martial)
	}
}

func TestCheckMarriages_SameFaction(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: 0, MaxAge: 70, Faction: "Iron"},
		{ID: "b", Name: "B", BirthYear: 0, MaxAge: 70, Faction: "Iron"},
	}
	rng := newTestRNG(42)
	events := CheckMarriages(figures, "settle", "Iron", 25, rng)
	for _, e := range events {
		if e.Category != "Marriage" {
			t.Errorf("unexpected event category: %q", e.Category)
		}
	}
}

func TestCheckMarriages_CrossFaction(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: 0, MaxAge: 70, Faction: "Iron"},
		{ID: "b", Name: "B", BirthYear: 0, MaxAge: 70, Faction: "Wood"},
	}
	rng := newTestRNG(42)
	events := CheckMarriages(figures, "settle", "Iron", 25, rng)
	if len(events) != 0 {
		t.Errorf("expected no cross-faction marriages, got %d", len(events))
	}
}

func TestCheckMarriages_AgeGate(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "a", Name: "Young", BirthYear: 15, MaxAge: 70, Faction: "Iron"},
		{ID: "b", Name: "Old", BirthYear: 30, MaxAge: 70, Faction: "Iron"},
	}
	rng := newTestRNG(42)
	events := CheckMarriages(figures, "settle", "Iron", 35, rng)
	if len(events) != 0 {
		t.Errorf("expected no marriages outside 20-25 age, got %d", len(events))
	}
}

func TestCheckMarriages_Determinism(t *testing.T) {
	figures1 := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: 0, MaxAge: 70, Faction: "Iron"},
		{ID: "b", Name: "B", BirthYear: 0, MaxAge: 70, Faction: "Iron"},
	}
	figures2 := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: 0, MaxAge: 70, Faction: "Iron"},
		{ID: "b", Name: "B", BirthYear: 0, MaxAge: 70, Faction: "Iron"},
	}
	rng1 := newTestRNG(42)
	rng2 := newTestRNG(42)
	e1 := CheckMarriages(figures1, "s", "Iron", 25, rng1)
	e2 := CheckMarriages(figures2, "s", "Iron", 25, rng2)
	if len(e1) != len(e2) {
		t.Errorf("determinism failed: %d vs %d marriages", len(e1), len(e2))
	}
}

func TestCheckMarriages_SkipsDead(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: 0, MaxAge: 24, Faction: "Iron", DeathYear: 25},
		{ID: "b", Name: "B", BirthYear: 0, MaxAge: 70, Faction: "Iron"},
	}
	rng := newTestRNG(42)
	events := CheckMarriages(figures, "settle", "Iron", 25, rng)
	if len(events) != 0 {
		t.Errorf("expected no marriages involving dead figures, got %d", len(events))
	}
}

func TestCheckMarriages_SkipsMarried(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "a", Name: "A", BirthYear: 0, MaxAge: 70, Faction: "Iron", Relationships: Relationships{Spouse: []string{"b"}}},
		{ID: "b", Name: "B", BirthYear: 0, MaxAge: 70, Faction: "Iron", Relationships: Relationships{Spouse: []string{"a"}}},
	}
	rng := newTestRNG(42)
	events := CheckMarriages(figures, "settle", "Iron", 25, rng)
	if len(events) != 0 {
		t.Errorf("expected no marriages for already-married figures, got %d", len(events))
	}
}
