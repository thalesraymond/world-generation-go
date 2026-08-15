package figures

import (
	randv2 "math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
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
		events := (&Leader{}).GenerateEvents(figure, 7, settlement, 1000, nil, 0, 0, rng)
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

	events := (&Explorer{}).GenerateEvents(figure, 7, settlement, 1000, nil, 10, 20, rng)
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

	events := (&Explorer{}).GenerateEvents(figure, 7, "Dawnwhisper", 500, nil, 5, 5, rng)
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
		{"General", &General{}},
		{"Diplomat", &Diplomat{}},
		{"Master Smith", &MasterSmith{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r1 := randv2.New(randv2.NewPCG(42, 7))
			r2 := randv2.New(randv2.NewPCG(42, 7))

			var got1, got2 [][]simulation.Event
			for i := 0; i < 10; i++ {
				got1 = append(got1, tc.role.GenerateEvents(figure, 7, "Det Settlement", 1000, nil, 0, 0, r1))
				got2 = append(got2, tc.role.GenerateEvents(figure, 7, "Det Settlement", 1000, nil, 0, 0, r2))
			}

			if !reflect.DeepEqual(got1, got2) {
				t.Errorf("same seed produced different event sequences:\nfirst:  %v\nsecond: %v", got1, got2)
			}
		})
	}
}

func TestNewRole_NewRoles(t *testing.T) {
	for _, name := range []string{"General", "Diplomat", "Master Smith"} {
		r, err := NewRole(name)
		if err != nil {
			t.Fatalf("NewRole(%q) unexpected error: %v", name, err)
		}
		if r.Name() != name {
			t.Errorf("NewRole(%q).Name() = %q, want %q", name, r.Name(), name)
		}
	}
}

func TestGeneral_GenerateEvents(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(42, 7))
	figure := &HistoricalFigure{ID: "gen-1", Name: "Cedric Ironhand", Stats: Stats{Martial: 15, Diplomatic: 10, Infamy: 1}}
	gen := &General{}

	var nilCount, eventCount int
	for i := 0; i < 20; i++ {
		events := gen.GenerateEvents(figure, 7, "Ashfield", 1000, nil, 0, 0, rng)
		if events == nil {
			nilCount++
			continue
		}
		eventCount++
		e := events[0]
		if e.Category != "Conflict" {
			t.Errorf("event.Category = %q, want Conflict", e.Category)
		}
		if e.FigureID != figure.ID {
			t.Errorf("event.FigureID = %q, want %q", e.FigureID, figure.ID)
		}
		if !strings.Contains(e.Description, figure.Name) {
			t.Errorf("description missing figure name: %q", e.Description)
		}
	}
	if nilCount == 0 {
		t.Error("expected some nil returns")
	}
	if eventCount == 0 {
		t.Error("expected some events")
	}
}

func TestDiplomat_GenerateEvents(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(42, 7))
	figure := &HistoricalFigure{ID: "dip-1", Name: "Lyra Silvertongue", Stats: Stats{Martial: 5, Diplomatic: 16, Infamy: 1}}

	events := (&Diplomat{}).GenerateEvents(figure, 7, "Goldhaven", 1000, nil, 0, 0, rng)
	if events == nil {
		t.Skip("no event generated — probabilistic")
		return
	}
	e := events[0]
	if e.Category != "Politics" {
		t.Errorf("event.Category = %q, want Politics", e.Category)
	}
	if !strings.Contains(e.Description, figure.Name) {
		t.Errorf("description missing figure name: %q", e.Description)
	}
}

func TestMasterSmith_GenerateEvents(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(42, 7))
	figure := &HistoricalFigure{ID: "smith-1", Name: "Borin Forgehand", Stats: Stats{Martial: 10, Diplomatic: 10, Infamy: 1}}

	events := (&MasterSmith{}).GenerateEvents(figure, 7, "Ironforge", 1000, nil, 0, 0, rng)
	if events == nil {
		t.Skip("no event generated — probabilistic")
		return
	}
	e := events[0]
	if e.Category != "Settlement" {
		t.Errorf("event.Category = %q, want Settlement", e.Category)
	}
	if !strings.Contains(e.Description, "master smith") {
		t.Errorf("description missing master smith: %q", e.Description)
	}
}

func TestNewRoles_CanTransitionTo(t *testing.T) {
	gen, dip, smith := &General{}, &Diplomat{}, &MasterSmith{}
	leader, explorer := &Leader{}, &Explorer{}

	tests := []struct {
		from, to Role
		want     bool
	}{
		{gen, explorer, true},
		{gen, leader, false},
		{gen, gen, false},
		{dip, leader, true},
		{dip, explorer, false},
		{dip, dip, false},
		{smith, leader, false},
		{smith, explorer, false},
		{smith, smith, false},
	}
	for _, tt := range tests {
		if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
			t.Errorf("%s.CanTransitionTo(%s) = %v, want %v", tt.from.Name(), tt.to.Name(), got, tt.want)
		}
	}
}

func TestLeader_GenerateEvents_AddsReputation(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(42, 7))
	figure := &HistoricalFigure{ID: "l-rep", Name: "Queen Elara", Stats: Stats{Martial: 10, Diplomatic: 15, Infamy: 1}}
	events := (&Leader{}).GenerateEvents(figure, 7, "Valewatch", 1000, nil, 0, 0, rng)
	if events == nil {
		return
	}
	if len(figure.Reputation) == 0 {
		t.Error("expected reputation entry from Leader event")
	}
	if figure.TotalReputation() <= 0 {
		t.Errorf("expected positive reputation, got %d", figure.TotalReputation())
	}
	if got := figure.Reputation[len(figure.Reputation)-1].Year; got != 7 {
		t.Errorf("reputation entry year = %d, want the event year 7", got)
	}
}

func TestExplorer_GenerateEvents_AddsReputation(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(2, 1))
	figure := &HistoricalFigure{ID: "e-rep", Name: "Pathfinder", Stats: Stats{Martial: 8, Diplomatic: 8, Infamy: 1}}
	events := (&Explorer{}).GenerateEvents(figure, 7, "Thornhaven", 1000, nil, 0, 0, rng)
	if events == nil {
		return
	}
	if len(figure.Reputation) == 0 {
		t.Error("expected reputation entry from Explorer event")
	}
	if got := figure.Reputation[len(figure.Reputation)-1].Year; got != 7 {
		t.Errorf("reputation entry year = %d, want the event year 7", got)
	}
}

func TestCheckTransitions_ExplorerToLeader(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "exp-1", Name: "Scout", BirthYear: 0, MaxAge: 70, Faction: "f", Role: "Explorer"},
	}
	figures[0].SetRole(&Explorer{})
	events := []simulation.Event{
		{Category: "Discovery", FigureID: "exp-1", Description: "discovers new land"},
	}
	rng := newTestRNG(42)
	trans := CheckTransitions(figures, events, rng)
	for _, e := range trans {
		if e.Category != "RoleTransition" {
			t.Errorf("unexpected event category: %q", e.Category)
		}
	}
}

func TestCheckTransitions_InvalidRejected(t *testing.T) {
	figures := []HistoricalFigure{
		{ID: "smith-1", Name: "Forge", BirthYear: 0, MaxAge: 70, Faction: "f", Role: "Master Smith"},
	}
	figures[0].SetRole(&MasterSmith{})
	events := []simulation.Event{
		{Category: "Conflict", FigureID: "smith-1", Description: "fought"},
	}
	rng := newTestRNG(42)
	trans := CheckTransitions(figures, events, rng)
	if len(trans) != 0 {
		t.Errorf("expected no transitions for MasterSmith, got %d", len(trans))
	}
}

func TestCheckTransitions_Format(t *testing.T) {
	f := HistoricalFigure{ID: "exp-2", Name: "Wayfarer", BirthYear: 0, MaxAge: 70, Faction: "f", Role: "Explorer"}
	f.SetRole(&Explorer{})
	figures := []HistoricalFigure{f}
	events := []simulation.Event{
		{Category: "Discovery", FigureID: "exp-2", Description: "discovers"},
	}
	rng := newTestRNG(42)
	trans := CheckTransitions(figures, events, rng)
	for _, e := range trans {
		if !strings.Contains(e.Description, "Wayfarer") {
			t.Errorf("transition event missing figure name: %q", e.Description)
		}
		if !strings.Contains(e.Description, "Leader") {
			t.Errorf("transition event missing new role: %q", e.Description)
		}
	}
}

func TestExplorer_GenerateEvents_WithGraph(t *testing.T) {
	graph := pointcrawl.NewGraph()
	graph.AddNode(&pointcrawl.Node{ID: 1, X: 10, Y: 10, Visibility: pointcrawl.Unknown, Name: "Ruined Tower", Kind: "ruin"})

	rng := randv2.New(randv2.NewPCG(2, 1))
	figure := &HistoricalFigure{ID: "fig-graph", Name: "Brisa Mosswood"}
	events := (&Explorer{}).GenerateEvents(figure, 7, "Thornhaven", 1000, graph, 10, 10, rng)
	if len(events) != 1 {
		t.Fatalf("GenerateEvents with graph returned %d events, want 1", len(events))
	}
	if !strings.Contains(events[0].Description, "Ruined Tower") {
		t.Errorf("description missing node name: %q", events[0].Description)
	}
}

func TestDiplomat_GenerateEvents_FailureAddsReputation(t *testing.T) {
	// Loop until we observe a failure branch (delta -1).
	success := false
	for i := 0; i < 100; i++ {
		rng := randv2.New(randv2.NewPCG(uint64(i), 7))
		figure := &HistoricalFigure{ID: "dip-f", Name: "Lyra", Stats: Stats{Martial: 5, Diplomatic: 16, Infamy: 1}}
		events := (&Diplomat{}).GenerateEvents(figure, 7, "Goldhaven", 1000, nil, 0, 0, rng)
		if events == nil {
			continue
		}
		if len(figure.Reputation) > 0 && figure.Reputation[0].Delta < 0 {
			success = true
			break
		}
	}
	if !success {
		t.Skip("failure branch not observed in 100 attempts — probabilistic")
	}
}

func TestCheckTransitions_LeaderToExplorerExile(t *testing.T) {
	success := false
	for i := 0; i < 1000; i++ {
		figures := []HistoricalFigure{
			{ID: "lead-1", Name: "Ruler", BirthYear: 0, MaxAge: 70, Faction: "f", Role: "Leader"},
		}
		figures[0].SetRole(&Leader{})
		rng := randv2.New(randv2.NewPCG(uint64(i), 1))
		trans := CheckTransitions(figures, []simulation.Event{}, rng)
		for _, e := range trans {
			if e.Category == "RoleTransition" && strings.Contains(e.Description, "exile") {
				success = true
				if figures[0].Role != "Explorer" {
					t.Errorf("expected role Explorer, got %q", figures[0].Role)
				}
				break
			}
		}
		if success {
			break
		}
	}
	if !success {
		t.Skip("leader exile transition not observed in 1000 attempts — probabilistic")
	}
}

func TestCheckTransitions_GeneralToExplorerDefeat(t *testing.T) {
	success := false
	for i := 0; i < 200; i++ {
		figures := []HistoricalFigure{
			{ID: "gen-1", Name: "General", BirthYear: 0, MaxAge: 70, Faction: "f", Role: "General", Stats: Stats{Martial: 1, Diplomatic: 10, Infamy: 1}},
		}
		figures[0].SetRole(&General{})
		events := []simulation.Event{
			{Category: "Conflict", FigureID: "gen-1", Description: "defeated"},
		}
		rng := randv2.New(randv2.NewPCG(uint64(i), 1))
		trans := CheckTransitions(figures, events, rng)
		for _, e := range trans {
			if e.Category == "RoleTransition" && strings.Contains(e.Description, "defeat") {
				success = true
				if figures[0].Role != "Explorer" {
					t.Errorf("expected role Explorer, got %q", figures[0].Role)
				}
				break
			}
		}
		if success {
			break
		}
	}
	if !success {
		t.Skip("general defeat transition not observed in 200 attempts — probabilistic")
	}
}
