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
