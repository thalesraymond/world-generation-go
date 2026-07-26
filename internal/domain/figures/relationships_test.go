package figures

import (
	randv2 "math/rand/v2"
	"reflect"
	"testing"
)

func TestAddParentChild(t *testing.T) {
	parent := &HistoricalFigure{ID: "parent-1", Name: "Parent"}
	child := &HistoricalFigure{ID: "child-1", Name: "Child"}

	AddParentChild(parent, child)

	if len(parent.Relationships.Children) != 1 || parent.Relationships.Children[0] != "child-1" {
		t.Errorf("expected parent's child to be child-1, got %v", parent.Relationships.Children)
	}
	if len(child.Relationships.Parents) != 1 || child.Relationships.Parents[0] != "parent-1" {
		t.Errorf("expected child's parent to be parent-1, got %v", child.Relationships.Parents)
	}
}

func TestAddSpouse(t *testing.T) {
	a := &HistoricalFigure{ID: "a", Name: "A"}
	b := &HistoricalFigure{ID: "b", Name: "B"}

	AddSpouse(a, b)

	if len(a.Relationships.Spouse) != 1 || a.Relationships.Spouse[0] != "b" {
		t.Errorf("expected a's spouse to be b, got %v", a.Relationships.Spouse)
	}
	if len(b.Relationships.Spouse) != 1 || b.Relationships.Spouse[0] != "a" {
		t.Errorf("expected b's spouse to be a, got %v", b.Relationships.Spouse)
	}
}

func TestFormMarriage_Alive(t *testing.T) {
	a := &HistoricalFigure{ID: "a", Name: "A"}
	b := &HistoricalFigure{ID: "b", Name: "B"}
	year := 50

	event, ok := FormMarriage(a, b, year)
	if !ok {
		t.Fatalf("expected marriage to succeed")
	}
	if event.Year != year {
		t.Errorf("expected event year %d, got %d", year, event.Year)
	}
	if event.Category != "Marriage" {
		t.Errorf("expected category Marriage, got %q", event.Category)
	}
	if event.FigureID != "a" {
		t.Errorf("expected FigureID a, got %q", event.FigureID)
	}
	if !reflect.DeepEqual(event.RelatedFigures, []string{"b"}) {
		t.Errorf("expected related figures [b], got %v", event.RelatedFigures)
	}
}

func TestFormMarriage_DeadFigure(t *testing.T) {
	a := &HistoricalFigure{ID: "a", Name: "A", DeathYear: 1}
	b := &HistoricalFigure{ID: "b", Name: "B"}

	_, ok := FormMarriage(a, b, 50)
	if ok {
		t.Fatalf("expected marriage to fail when one figure is dead")
	}
}

func TestGetHeir_WithChildren(t *testing.T) {
	parent := HistoricalFigure{ID: "p", Name: "Parent"}
	eldest := HistoricalFigure{ID: "eldest", Name: "Eldest", BirthYear: 10}
	youngest := HistoricalFigure{ID: "youngest", Name: "Youngest", BirthYear: 20}

	AddParentChild(&parent, &eldest)
	AddParentChild(&parent, &youngest)

	figures := []HistoricalFigure{parent, eldest, youngest}
	heir := GetHeir(figures, parent.ID)
	if heir == nil {
		t.Fatalf("expected heir, got nil")
	}
	if heir.ID != eldest.ID {
		t.Errorf("expected eldest to be heir, got %q", heir.ID)
	}
}

func TestGetHeir_NoChildren(t *testing.T) {
	parent := HistoricalFigure{ID: "p", Name: "Parent"}
	figures := []HistoricalFigure{parent}
	heir := GetHeir(figures, parent.ID)
	if heir != nil {
		t.Fatalf("expected nil heir, got %q", heir.ID)
	}
}

func TestGetHeir_SkipsDeadChildren(t *testing.T) {
	parent := HistoricalFigure{ID: "p", Name: "Parent"}
	deadChild := HistoricalFigure{ID: "dead", Name: "Dead", BirthYear: 10, DeathYear: 20}
	livingChild := HistoricalFigure{ID: "living", Name: "Living", BirthYear: 20}

	AddParentChild(&parent, &deadChild)
	AddParentChild(&parent, &livingChild)

	figures := []HistoricalFigure{parent, deadChild, livingChild}
	heir := GetHeir(figures, parent.ID)
	if heir == nil {
		t.Fatalf("expected heir, got nil")
	}
	if heir.ID != "living" {
		t.Errorf("expected living child to be heir, got %q", heir.ID)
	}
}

func TestDeterminism_RelationshipOperations(t *testing.T) {
	seed := uint64(777)
	rng1 := randv2.New(randv2.NewPCG(seed, seed+1))
	rng2 := randv2.New(randv2.NewPCG(seed, seed+1))

	a1 := &HistoricalFigure{ID: "a", Name: GenerateName(rng1)}
	b1 := &HistoricalFigure{ID: "b", Name: GenerateName(rng1)}
	_, ok1 := FormMarriage(a1, b1, 10)

	a2 := &HistoricalFigure{ID: "a", Name: GenerateName(rng2)}
	b2 := &HistoricalFigure{ID: "b", Name: GenerateName(rng2)}
	_, ok2 := FormMarriage(a2, b2, 10)

	if ok1 != ok2 {
		t.Fatalf("marriage outcomes differ")
	}
	if !reflect.DeepEqual(a1.Relationships, a2.Relationships) {
		t.Errorf("spouse relationships differ: %v vs %v", a1.Relationships, a2.Relationships)
	}
}
