package narrative

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

func grammarPath(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "grammars", name)
	if _, err := os.Stat(path); err != nil {
		// Fall back to project-root relative path when running from repo root.
		path = filepath.Join("grammars", name)
	}
	return path
}

func TestNewEngineFromFile_LoadsMythicalGrammar(t *testing.T) {
	eng, err := NewEngineFromFile(grammarPath(t, "mythical.bnf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		event    simulation.Event
		extra    map[string]string
		contains []string
	}{
		{
			name: "battle",
			event: simulation.Event{
				Year:        450,
				Category:    "battle",
				Description: "A battle occurred.",
			},
			extra: map[string]string{
				"attacker": "Red Legion",
				"defender": "Blue League",
				"location": "Ashmoor",
				"outcome":  "The Red Legion prevailed",
			},
			contains: []string{"Red Legion", "Blue League", "Ashmoor"},
		},
		{
			name: "founding",
			event: simulation.Event{
				Year:        500,
				Category:    "founding",
				Description: "A settlement was founded.",
			},
			extra: map[string]string{
				"settlement": "Rivensprawl",
				"faction":    "the Dominion of Sun",
				"location":   "the Thornwood Vale",
			},
			contains: []string{"Rivensprawl", "the Dominion of Sun"},
		},
		{
			name: "diplomacy",
			event: simulation.Event{
				Year:        550,
				Category:    "diplomacy",
				Description: "A diplomat was dispatched.",
			},
			extra: map[string]string{
				"faction_a": "Silver Compact",
				"faction_b": "Iron Circle",
				"result":    "peace",
			},
			contains: []string{"Silver Compact", "Iron Circle", "peace"},
		},
		{
			name: "disaster",
			event: simulation.Event{
				Year:        600,
				Category:    "disaster",
				Description: "A disaster struck.",
			},
			extra: map[string]string{
				"type":     "earthquake",
				"location": "Drakefall",
			},
			contains: []string{"earthquake", "Drakefall"},
		},
		{
			name: "discovery",
			event: simulation.Event{
				Year:        650,
				Category:    "discovery",
				Description: "A discovery was made.",
			},
			extra: map[string]string{
				"explorer": "Elara Moonwhisper",
				"thing":    "the Crown of Stars",
				"place":    "the Sunken Vale",
				"location": "the Mournful Reach",
			},
			contains: []string{"Elara Moonwhisper", "the Crown of Stars", "the Sunken Vale"},
		},
	}

	rng := randv2.New(randv2.NewPCG(1, 2))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eng.Narrate(tt.event, tt.extra, rng)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(got) == "" {
				t.Fatal("expected non-empty narrative")
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Fatalf("expected narrative to contain %q, got %q", want, got)
				}
			}
		})
	}
}

func TestNewEngineFromFile_LoadsSimpleGrammar(t *testing.T) {
	eng, err := NewEngineFromFile(grammarPath(t, "simple.bnf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event := simulation.Event{
		Year:        100,
		Category:    "variable_test",
		Description: "A simple event.",
	}
	rng := randv2.New(randv2.NewPCG(1, 2))

	got, err := eng.Narrate(event, map[string]string{"name": "broom", "action": "sweep", "place": "hall"}, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "broom") {
		t.Fatalf("expected narrative to contain %q, got %q", "broom", got)
	}
	if !strings.Contains(got, "hall") {
		t.Fatalf("expected narrative to contain %q, got %q", "hall", got)
	}
}

func TestNewEngineFromFile_Determinism(t *testing.T) {
	eng, err := NewEngineFromFile(grammarPath(t, "mythical.bnf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event := simulation.Event{
		Year:        300,
		Category:    "battle",
		Description: "An ancient battle.",
	}
	extra := map[string]string{
		"attacker": "Old Kingdom",
		"defender": "Wild Tribes",
		"location": "Bonefield",
		"outcome":  "The Old Kingdom routed the invaders",
	}

	rng1 := randv2.New(randv2.NewPCG(42, 0))
	first, err := eng.Narrate(event, extra, rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rng2 := randv2.New(randv2.NewPCG(42, 0))
	second, err := eng.Narrate(event, extra, rng2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != second {
		t.Fatalf("same seed produced different outputs: %q vs %q", first, second)
	}
}

func TestNewEngineFromFile_UnknownCategoryFallback(t *testing.T) {
	eng, err := NewEngineFromFile(grammarPath(t, "mythical.bnf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event := simulation.Event{
		Year:        999,
		Category:    "unknown_category",
		Description: "Something mysterious happened.",
	}
	rng := randv2.New(randv2.NewPCG(1, 2))

	got, err := eng.Narrate(event, nil, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != event.Description {
		t.Fatalf("expected %q, got %q", event.Description, got)
	}
}

func TestNewEngineFromFile_NotFound(t *testing.T) {
	_, err := NewEngineFromFile(grammarPath(t, "does-not-exist.bnf"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}
