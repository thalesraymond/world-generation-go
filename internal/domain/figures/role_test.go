package figures

import (
	randv2 "math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

func TestLeader_Name(t *testing.T) {
	l := &Leader{}
	if got, want := l.Name(), "Leader"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestExplorer_Name(t *testing.T) {
	e := &Explorer{}
	if got, want := e.Name(), "Explorer"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestNewRole_ReturnsKnownRoles(t *testing.T) {
	leader, err := NewRole("Leader")
	if err != nil {
		t.Fatalf("NewRole(\"Leader\") unexpected error: %v", err)
	}
	if _, ok := leader.(*Leader); !ok {
		t.Errorf("NewRole(\"Leader\") returned %T, want *Leader", leader)
	}

	explorer, err := NewRole("Explorer")
	if err != nil {
		t.Fatalf("NewRole(\"Explorer\") unexpected error: %v", err)
	}
	if _, ok := explorer.(*Explorer); !ok {
		t.Errorf("NewRole(\"Explorer\") returned %T, want *Explorer", explorer)
	}
}

func TestNewRole_UnknownReturnsError(t *testing.T) {
	role, err := NewRole("Artisan")
	if err == nil {
		t.Fatalf("NewRole(\"Artisan\") expected error, got role %T", role)
	}
	if role != nil {
		t.Errorf("NewRole(\"Artisan\") returned non-nil role %T", role)
	}
}

func TestLeader_GenerateEvents(t *testing.T) {
	// Statistical test: over 20 calls some should fire and some should not.
	rng := randv2.New(randv2.NewPCG(9, 9))
	figure := &HistoricalFigure{ID: "fig-1", Name: "Aelar Thorne"}
	settlement := "Rivensprawl"

	allowed := map[string]struct{}{
		"Politics":   {},
		"Settlement": {},
		"Conflict":   {},
	}

	var nilCount, eventCount int
	for i := 0; i < 20; i++ {
		events := (&Leader{}).GenerateEvents(figure, settlement, 1000, nil, 0, 0, rng)
		if events == nil {
			nilCount++
			continue
		}
		if len(events) != 1 {
			t.Fatalf("iteration %d returned %d events, want 0 or 1", i, len(events))
		}
		eventCount++

		event := events[0]
		if _, ok := allowed[event.Category]; !ok {
			t.Errorf("event.Category = %q, want one of Politics/Settlement/Conflict", event.Category)
		}
		if event.FigureID != figure.ID {
			t.Errorf("event.FigureID = %q, want %q", event.FigureID, figure.ID)
		}
		if event.SettlementName != settlement {
			t.Errorf("event.SettlementName = %q, want %q", event.SettlementName, settlement)
		}
		if !strings.Contains(event.Description, figure.Name) {
			t.Errorf("event.Description = %q, missing figure name %q", event.Description, figure.Name)
		}
		if !strings.Contains(event.Description, settlement) {
			t.Errorf("event.Description = %q, missing settlement name %q", event.Description, settlement)
		}
	}

	if nilCount == 0 {
		t.Error("expected some GenerateEvents calls to return nil")
	}
	if eventCount == 0 {
		t.Error("expected some GenerateEvents calls to return an event")
	}
}

func TestExplorer_GenerateEvents(t *testing.T) {
	// Seed chosen so the first call triggers an event (IntN(5) == 0).
	rng := randv2.New(randv2.NewPCG(2, 1))
	figure := &HistoricalFigure{ID: "fig-2", Name: "Brisa Mosswood"}
	settlement := "Thornhaven"

	events := (&Explorer{}).GenerateEvents(figure, settlement, 1000, nil, 10, 20, rng)
	if len(events) != 1 {
		t.Fatalf("GenerateEvents returned %d events, want 1", len(events))
	}

	event := events[0]
	if event.Category != "Discovery" {
		t.Errorf("event.Category = %q, want Discovery", event.Category)
	}
	if event.FigureID != figure.ID {
		t.Errorf("event.FigureID = %q, want %q", event.FigureID, figure.ID)
	}
	if event.SettlementName != settlement {
		t.Errorf("event.SettlementName = %q, want %q", event.SettlementName, settlement)
	}
	if !strings.Contains(event.Description, figure.Name) {
		t.Errorf("event.Description = %q, missing figure name %q", event.Description, figure.Name)
	}
	if !strings.Contains(event.Description, settlement) {
		t.Errorf("event.Description = %q, missing settlement name %q", event.Description, settlement)
	}
}

func TestExplorer_GenerateEvents_NilGraph(t *testing.T) {
	// Seed chosen so the first call triggers an event with the nil-graph fallback.
	rng := randv2.New(randv2.NewPCG(2, 1))
	figure := &HistoricalFigure{ID: "fig-3", Name: "Caius Ashford"}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GenerateEvents panicked with nil graph: %v", r)
		}
	}()

	events := (&Explorer{}).GenerateEvents(figure, "Dawnwhisper", 500, nil, 5, 5, rng)
	if len(events) != 1 {
		t.Fatalf("nil graph fallback returned %d events, want 1", len(events))
	}
	if events[0].Category != "Discovery" {
		t.Errorf("event.Category = %q, want Discovery", events[0].Category)
	}
}

func TestCanTransitionTo(t *testing.T) {
	leader := &Leader{}
	explorer := &Explorer{}

	tests := []struct {
		from Role
		to   Role
		want bool
	}{
		{leader, explorer, true},
		{leader, leader, false},
		{explorer, leader, true},
		{explorer, explorer, false},
	}

	for _, tt := range tests {
		if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
			t.Errorf("%s.CanTransitionTo(%s) = %v, want %v", tt.from.Name(), tt.to.Name(), got, tt.want)
		}
	}
}

func TestRole_GenerateEvents_Determinism(t *testing.T) {
	figure := &HistoricalFigure{ID: "fig-det", Name: "Eldrin Fairwind"}

	cases := []struct {
		name string
		role Role
	}{
		{"Leader", &Leader{}},
		{"Explorer", &Explorer{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r1 := randv2.New(randv2.NewPCG(42, 7))
			r2 := randv2.New(randv2.NewPCG(42, 7))

			var got1, got2 [][]simulation.Event
			for i := 0; i < 10; i++ {
				got1 = append(got1, tc.role.GenerateEvents(figure, "Det Settlement", 1000, nil, 0, 0, r1))
				got2 = append(got2, tc.role.GenerateEvents(figure, "Det Settlement", 1000, nil, 0, 0, r2))
			}

			if !reflect.DeepEqual(got1, got2) {
				t.Errorf("same seed produced different event sequences:\nfirst:  %v\nsecond: %v", got1, got2)
			}
		})
	}
}
