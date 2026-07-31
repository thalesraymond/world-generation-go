package figures

import (
	"encoding/json"
	randv2 "math/rand/v2"
	"strings"
	"testing"
)

func TestFigureCreation(t *testing.T) {
	rel := Relationships{
		Parents:  []string{"p1", "p2"},
		Children: []string{"c1"},
		Spouse:   []string{"s1"},
	}
	f := HistoricalFigure{
		ID:            "settlement-0",
		Name:          "Aelar Thorne",
		BirthYear:     100,
		DeathYear:     0,
		MaxAge:        80,
		Role:          "ruler",
		Faction:       "northern-league",
		Relationships: rel,
	}

	if f.ID != "settlement-0" {
		t.Errorf("ID = %q, want %q", f.ID, "settlement-0")
	}
	if f.Name != "Aelar Thorne" {
		t.Errorf("Name = %q, want %q", f.Name, "Aelar Thorne")
	}
	if f.BirthYear != 100 {
		t.Errorf("BirthYear = %d, want 100", f.BirthYear)
	}
	if f.DeathYear != 0 {
		t.Errorf("DeathYear = %d, want 0", f.DeathYear)
	}
	if f.MaxAge != 80 {
		t.Errorf("MaxAge = %d, want 80", f.MaxAge)
	}
	if f.Role != "ruler" {
		t.Errorf("Role = %q, want %q", f.Role, "ruler")
	}
	if f.Faction != "northern-league" {
		t.Errorf("Faction = %q, want %q", f.Faction, "northern-league")
	}
	if len(f.Relationships.Parents) != 2 {
		t.Errorf("len(Parents) = %d, want 2", len(f.Relationships.Parents))
	}
	if len(f.Relationships.Children) != 1 {
		t.Errorf("len(Children) = %d, want 1", len(f.Relationships.Children))
	}
	if len(f.Relationships.Spouse) != 1 {
		t.Errorf("len(Spouse) = %d, want 1", len(f.Relationships.Spouse))
	}
}

func TestJSONRoundTrip(t *testing.T) {
	rel := Relationships{
		Parents:  []string{"p1"},
		Children: []string{"c1", "c2"},
		Spouse:   []string{"s1"},
	}
	original := HistoricalFigure{
		ID:            "settlement-1",
		Name:          "Brisa Mosswood",
		BirthYear:     50,
		DeathYear:     120,
		MaxAge:        70,
		Role:          "scholar",
		Faction:       "southern-reach",
		Relationships: rel,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HistoricalFigure
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.BirthYear != original.BirthYear {
		t.Errorf("BirthYear = %d, want %d", decoded.BirthYear, original.BirthYear)
	}
	if decoded.DeathYear != original.DeathYear {
		t.Errorf("DeathYear = %d, want %d", decoded.DeathYear, original.DeathYear)
	}
	if decoded.MaxAge != original.MaxAge {
		t.Errorf("MaxAge = %d, want %d", decoded.MaxAge, original.MaxAge)
	}
	if decoded.Role != original.Role {
		t.Errorf("Role = %q, want %q", decoded.Role, original.Role)
	}
	if decoded.Faction != original.Faction {
		t.Errorf("Faction = %q, want %q", decoded.Faction, original.Faction)
	}
	if len(decoded.Relationships.Parents) != len(original.Relationships.Parents) {
		t.Errorf("len(Parents) = %d, want %d", len(decoded.Relationships.Parents), len(original.Relationships.Parents))
	}
	if len(decoded.Relationships.Children) != len(original.Relationships.Children) {
		t.Errorf("len(Children) = %d, want %d", len(decoded.Relationships.Children), len(original.Relationships.Children))
	}
	if len(decoded.Relationships.Spouse) != len(original.Relationships.Spouse) {
		t.Errorf("len(Spouse) = %d, want %d", len(decoded.Relationships.Spouse), len(original.Relationships.Spouse))
	}
}

func TestIsAlive(t *testing.T) {
	alive := HistoricalFigure{BirthYear: 10, DeathYear: 0}
	if !alive.IsAlive() {
		t.Errorf("IsAlive() = false, want true for DeathYear=0")
	}

	dead := HistoricalFigure{BirthYear: 10, DeathYear: 80}
	if dead.IsAlive() {
		t.Errorf("IsAlive() = true, want false for DeathYear=80")
	}
}

func TestAge(t *testing.T) {
	f := HistoricalFigure{BirthYear: 100}
	if got, want := f.Age(150), 50; got != want {
		t.Errorf("Age(150) = %d, want %d", got, want)
	}
	if got, want := f.Age(100), 0; got != want {
		t.Errorf("Age(100) = %d, want %d", got, want)
	}
}

func TestSetDeath(t *testing.T) {
	f := HistoricalFigure{BirthYear: 10, DeathYear: 0}
	f.SetDeath(75)
	if f.DeathYear != 75 {
		t.Errorf("DeathYear = %d, want 75", f.DeathYear)
	}
	if f.IsAlive() {
		t.Errorf("IsAlive() = true, want false after SetDeath")
	}
}

func TestGenerateNameDeterministic(t *testing.T) {
	r1 := randv2.New(randv2.NewPCG(1, 2))
	r2 := randv2.New(randv2.NewPCG(1, 2))

	name1 := GenerateName(r1)
	name2 := GenerateName(r2)

	if name1 != name2 {
		t.Errorf("same seed produced different names: %q vs %q", name1, name2)
	}
}

func TestGenerateNameFormat(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(1, 2))
	name := GenerateName(rng)

	parts := strings.Split(name, " ")
	if len(parts) != 2 {
		t.Errorf("name %q has %d parts, want 2", name, len(parts))
	}
	if parts[0] == "" || parts[1] == "" {
		t.Errorf("name %q has empty first or surname", name)
	}
}

func TestRelationshipsSerializationRoundTrip(t *testing.T) {
	original := Relationships{
		Parents:  []string{"p1", "p2"},
		Children: []string{"c1"},
		Spouse:   []string{"s1", "s2"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Relationships
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(decoded.Parents) != len(original.Parents) {
		t.Errorf("len(Parents) = %d, want %d", len(decoded.Parents), len(original.Parents))
	}
	if decoded.Parents[0] != original.Parents[0] || decoded.Parents[1] != original.Parents[1] {
		t.Errorf("Parents = %v, want %v", decoded.Parents, original.Parents)
	}
	if len(decoded.Children) != len(original.Children) {
		t.Errorf("len(Children) = %d, want %d", len(decoded.Children), len(original.Children))
	}
	if decoded.Children[0] != original.Children[0] {
		t.Errorf("Children = %v, want %v", decoded.Children, original.Children)
	}
	if len(decoded.Spouse) != len(original.Spouse) {
		t.Errorf("len(Spouse) = %d, want %d", len(decoded.Spouse), len(original.Spouse))
	}
	if decoded.Spouse[0] != original.Spouse[0] || decoded.Spouse[1] != original.Spouse[1] {
		t.Errorf("Spouse = %v, want %v", decoded.Spouse, original.Spouse)
	}
}

func TestGenerateStats_Determinism(t *testing.T) {
	rng1 := newTestRNG(42)
	rng2 := newTestRNG(42)
	s1 := GenerateStats(rng1, "General")
	s2 := GenerateStats(rng2, "General")
	if s1 != s2 {
		t.Errorf("same seed produced different stats: %+v vs %+v", s1, s2)
	}
}

func TestGenerateStats_RoleBias(t *testing.T) {
	genMartialZero, diplZero := 0, 0
	for i := 0; i < 100; i++ {
		g := GenerateStats(randv2.New(randv2.NewPCG(uint64(i*2), uint64(i*2+1))), "General")
		d := GenerateStats(randv2.New(randv2.NewPCG(uint64(i*2+1000), uint64(i*2+1001))), "Diplomat")
		if g.Martial < 3 {
			genMartialZero++
		}
		if d.Diplomatic < 3 {
			diplZero++
		}
	}
	if genMartialZero > 0 || diplZero > 0 {
		t.Errorf("stats with role bias should not be below 3")
	}
}

func TestStats_Normalize(t *testing.T) {
	s := Stats{Martial: 25, Diplomatic: -5, Infamy: 10}
	s = s.Normalize()
	if s.Martial != 20 || s.Diplomatic != 1 || s.Infamy != 10 {
		t.Errorf("normalized stats: %+v, want Martial=20 Diplomatic=1 Infamy=10", s)
	}
}

func TestStats_Copy(t *testing.T) {
	s := Stats{Martial: 15, Diplomatic: 10, Infamy: 5}
	c := s.Copy()
	c.Martial = 20
	if s.Martial == 20 {
		t.Errorf("copy should not affect original")
	}
	if c.Diplomatic != 10 {
		t.Errorf("copy Diplomatic = %d, want 10", c.Diplomatic)
	}
}

func TestStats_InfluenceOutcome(t *testing.T) {
	s := Stats{Martial: 20, Diplomatic: 0, Infamy: 0}
	rng := newTestRNG(1)
	conflictSuccesses := 0
	for i := 0; i < 100; i++ {
		if s.InfluenceOutcome("Conflict", rng) {
			conflictSuccesses++
		}
	}
	if conflictSuccesses != 100 {
		t.Errorf("Martial 20 should always succeed on Conflict, got %d/100", conflictSuccesses)
	}

	s2 := Stats{Martial: 1, Diplomatic: 0, Infamy: 0}
	rng2 := newTestRNG(1)
	if s2.InfluenceOutcome("Conflict", rng2) {
		t.Errorf("Martial 1 should have low chance of success")
	}

	s3 := Stats{Martial: 0, Diplomatic: 20, Infamy: 0}
	rng3 := newTestRNG(1)
	politicsSuccesses := 0
	for i := 0; i < 100; i++ {
		if s3.InfluenceOutcome("Politics", rng3) {
			politicsSuccesses++
		}
	}
	if politicsSuccesses != 100 {
		t.Errorf("Diplomatic 20 should always succeed on Politics, got %d/100", politicsSuccesses)
	}

	// Default category uses 50% chance.
	s4 := Stats{Martial: 0, Diplomatic: 0, Infamy: 0}
	rng4 := newTestRNG(1)
	s4.InfluenceOutcome("Other", rng4)
}

func TestReputation_AddAndTotal(t *testing.T) {
	f := &HistoricalFigure{Stats: Stats{Infamy: 1}}
	f.AddReputation(ReputationEntry{Year: 100, Event: "Battle", Delta: 5, Description: "Won a battle"})
	f.AddReputation(ReputationEntry{Year: 105, Event: "Raid", Delta: -3, Description: "Led a raid"})
	if f.TotalReputation() != 2 {
		t.Errorf("total reputation = %d, want 2", f.TotalReputation())
	}
	if len(f.Reputation) != 2 {
		t.Errorf("reputation entries = %d, want 2", len(f.Reputation))
	}
	if f.Stats.Infamy != 4 {
		t.Errorf("infamy = %d, want 4 (1 + 3)", f.Stats.Infamy)
	}
}

func TestReputation_RecentEntries(t *testing.T) {
	f := &HistoricalFigure{}
	f.AddReputation(ReputationEntry{Year: 100, Delta: 1, Description: "old"})
	f.AddReputation(ReputationEntry{Year: 195, Delta: 2, Description: "recent"})
	f.AddReputation(ReputationEntry{Year: 200, Delta: 3, Description: "current"})
	recent := f.RecentReputation(200, 10)
	if len(recent) != 2 {
		t.Errorf("recent entries = %d, want 2", len(recent))
	}
}

func TestReputation_Empty(t *testing.T) {
	f := &HistoricalFigure{}
	if f.TotalReputation() != 0 {
		t.Errorf("empty total = %d, want 0", f.TotalReputation())
	}
	if len(f.RecentReputation(100, 50)) != 0 {
		t.Errorf("empty recent = %d entries, want 0", len(f.RecentReputation(100, 50)))
	}
}

func TestSetRole_GetRole(t *testing.T) {
	f := &HistoricalFigure{}
	f.SetRole(&Leader{})
	r := f.GetRole()
	if r.Name() != "Leader" {
		t.Errorf("GetRole().Name() = %q, want Leader", r.Name())
	}
	if f.Role != "Leader" {
		t.Errorf("string Role = %q, want Leader", f.Role)
	}

	f.SetRole(nil)
	if f.GetRole() != nil {
		t.Error("expected nil role after SetRole(nil)")
	}
	if f.Role != "" {
		t.Errorf("string Role = %q, want empty", f.Role)
	}
}

func TestGetRole_LazyInit(t *testing.T) {
	f := &HistoricalFigure{Role: "Explorer"}
	r := f.GetRole()
	if r == nil {
		t.Fatal("GetRole returned nil for Explorer string")
	}
	if r.Name() != "Explorer" {
		t.Errorf("GetRole().Name() = %q, want Explorer", r.Name())
	}
}

func TestGetRole_UnknownRole(t *testing.T) {
	f := &HistoricalFigure{Role: "Bogus"}
	r := f.GetRole()
	if r != nil {
		t.Errorf("GetRole returned non-nil for unknown role: %T", r)
	}
}

func TestGenerateFounders_HasStats(t *testing.T) {
	rng := newTestRNG(42)
	founders := GenerateFounders(rng, "Test", "faction", 0)
	for _, f := range founders {
		if f.Stats.Martial < 1 || f.Stats.Martial > 20 {
			t.Errorf("founder %s has invalid Martial: %d", f.Name, f.Stats.Martial)
		}
		if f.Stats.Diplomatic < 1 || f.Stats.Diplomatic > 20 {
			t.Errorf("founder %s has invalid Diplomatic: %d", f.Name, f.Stats.Diplomatic)
		}
	}
}

func TestCheckBirths_HasStats(t *testing.T) {
	rng := newTestRNG(1)
	child := CheckBirths(nil, 20000, 100, rng)
	if child == nil {
		t.Fatal("no birth")
	}
	if child.Stats.Martial < 1 || child.Stats.Martial > 20 {
		t.Errorf("newborn has invalid stats: %+v", child.Stats)
	}
}

func TestStringMethod(t *testing.T) {
	f := HistoricalFigure{Name: "Aldric", Role: "Leader", BirthYear: 50, Stats: Stats{Martial: 15, Diplomatic: 10, Infamy: 3}}
	s := f.String()
	if !strings.Contains(s, "Aldric") || !strings.Contains(s, "Leader") || !strings.Contains(s, "M:15") {
		t.Errorf("String() = %q, missing expected content", s)
	}
}

func TestJSONRoundTrip_WithNewFields(t *testing.T) {
	original := HistoricalFigure{
		ID:    "fig-1", Name: "Test", BirthYear: 100, Role: "Leader",
		Faction: "f", Stats: Stats{Martial: 15, Diplomatic: 10, Infamy: 5},
		Reputation:        []ReputationEntry{{Year: 100, Event: "E", Delta: 1, Description: "test"}},
		ParentID:          "parent-1",
		TransitionHistory: []TransitionEntry{{Year: 105, FromRole: "Explorer", ToRole: "Leader", Reason: "promotion"}},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded HistoricalFigure
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Stats.Martial != 15 {
		t.Errorf("stats Martial = %d", decoded.Stats.Martial)
	}
	if len(decoded.Reputation) != 1 {
		t.Errorf("reputation len = %d", len(decoded.Reputation))
	}
	if decoded.ParentID != original.ParentID {
		t.Errorf("parentID = %q, want %q", decoded.ParentID, original.ParentID)
	}
	if len(decoded.TransitionHistory) != 1 {
		t.Errorf("transitionHistory len = %d", len(decoded.TransitionHistory))
	}
}
