package settlement

import (
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func TestResolveProximityConflictsMergesClose(t *testing.T) {
	settlements := []world.Settlement{
		{Name: "A", X: 0, Y: 0, Population: 1000, Type: TypeVillage},
		{Name: "B", X: 1, Y: 1, Population: 500, Type: TypeAbandoned},
	}

	result := ResolveProximityConflicts(settlements, 3.0)
	if len(result) != 1 {
		t.Fatalf("expected 1 settlement after merge, got %d", len(result))
	}
	if result[0].Population != 1500 {
		t.Errorf("expected combined population 1500, got %v", result[0].Population)
	}
	if result[0].Type != TypeVillage {
		t.Errorf("expected reclassified type Village, got %s", result[0].Type)
	}
}

func TestResolveProximityConflictsKeepsDistant(t *testing.T) {
	settlements := []world.Settlement{
		{Name: "A", X: 0, Y: 0, Population: 1000},
		{Name: "B", X: 10, Y: 10, Population: 500},
	}

	result := ResolveProximityConflicts(settlements, 3.0)
	if len(result) != 2 {
		t.Fatalf("expected 2 settlements, got %d", len(result))
	}
}

func TestResolveProximityConflictsTieBreaking(t *testing.T) {
	settlements := []world.Settlement{
		{Name: "A", X: 0, Y: 0, Population: 1000, Type: TypeVillage},
		{Name: "B", X: 1, Y: 1, Population: 1000, Type: TypeVillage},
	}

	result := ResolveProximityConflicts(settlements, 3.0)
	if len(result) != 1 {
		t.Fatalf("expected 1 settlement after merge, got %d", len(result))
	}
	if result[0].Name != "A" {
		t.Errorf("expected lower-index settlement A to survive, got %s", result[0].Name)
	}
	if result[0].Population != 2000 {
		t.Errorf("expected combined population 2000, got %v", result[0].Population)
	}
	if result[0].Type != TypeVillage {
		t.Errorf("expected reclassified type Village, got %s", result[0].Type)
	}
}

func TestResolveProximityConflictsLargerSurvives(t *testing.T) {
	settlements := []world.Settlement{
		{Name: "Small", X: 0, Y: 0, Population: 100, Type: TypeAbandoned},
		{Name: "Big", X: 1, Y: 1, Population: 50000, Type: TypeMajorCity},
	}

	result := ResolveProximityConflicts(settlements, 3.0)
	if len(result) != 1 {
		t.Fatalf("expected 1 settlement after merge, got %d", len(result))
	}
	if result[0].Name != "Big" {
		t.Errorf("expected larger settlement Big to survive, got %s", result[0].Name)
	}
}
