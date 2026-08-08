package cmd

import (
	"fmt"
	randv2 "math/rand/v2"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	dompointcrawl "github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/state"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// newAgentTestSettlement builds a settlement with agent state for Tick tests.
func newAgentTestSettlement(name string, population float64) world.Settlement {
	return world.Settlement{
		Name:             name,
		Type:             "Town",
		X:                1,
		Y:                1,
		Faction:          "testers",
		Population:       population,
		MilitaryStrength: population * 0.1,
		Wealth:           100,
		Goals:            []string{"grow"},
		Relations:        map[string]float64{},
	}
}

func TestAgentRNGDerivation(t *testing.T) {
	engineA := state.NewEngine(42)
	engineB := state.NewEngine(42)

	agentA := engineA.GetPRNG("agent:Haven")
	agentB := engineB.GetPRNG("agent:Haven")

	// Same seed produces identical agent streams across runs.
	for i := 0; i < 8; i++ {
		if agentA.Uint64() != agentB.Uint64() {
			t.Fatalf("agent RNG diverged at draw %d for same seed", i)
		}
	}

	// Agent RNG differs from figure RNG for the same settlement.
	engineC := state.NewEngine(42)
	figureC := engineC.GetPRNG("figures:Haven")
	agentC := engineC.GetPRNG("agent:Haven")
	same := true
	for i := 0; i < 8; i++ {
		if figureC.Uint64() != agentC.Uint64() {
			same = false
			break
		}
	}
	if same {
		t.Fatal("agent RNG must differ from figure RNG for the same settlement")
	}
}

func TestSettlementTickEmitsAgentEvents(t *testing.T) {
	settlements := []world.Settlement{
		newAgentTestSettlement("Haven", 1000),
		newAgentTestSettlement("Blackgate", 800),
	}
	settlements[0].Relations = world.InitRelations(settlements[0], settlements)
	settlements[1].Relations = world.InitRelations(settlements[1], settlements)

	engine := state.NewEngine(7)
	ws := world.NewState(8, 8)
	for i := range ws.Suitability {
		ws.Suitability[i] = 0.8
	}
	env := &agentEnv{
		worldState: ws,
		graph:      nil,
		all:        &settlements,
		usedNames:  map[string]bool{"Haven": true, "Blackgate": true},
	}

	entity := &settlementEntity{
		settlement:      &settlements[0],
		figureRNG:       engine.GetPRNG("figures:Haven"),
		agentRNG:        engine.GetPRNG("agent:Haven"),
		pointcrawlGraph: dompointcrawl.NewGraph(),
		allSettlements:  &settlements,
		env:             env,
	}

	eventChan := make(chan domsim.Event, 256)
	years := 10
	go func() {
		for year := 1; year <= years; year++ {
			entity.Tick(year, eventChan, engine.GetPRNG("timeline"))
		}
		close(eventChan)
	}()

	agentEvents := 0
	agentCategories := 0
	initialWealth := 100.0
	for event := range eventChan {
		if event.Year < 1 || event.Year > years {
			t.Errorf("event year %d out of range", event.Year)
		}
		if isAgentCategory(event.Category) {
			agentCategories++
		}
		if event.SettlementName == "Haven" {
			agentEvents++
		}
	}

	if agentCategories != years {
		t.Errorf("expected one agent-category event per year (%d), got %d", years, agentCategories)
	}
	if agentEvents == 0 {
		t.Error("expected events for settlement Haven")
	}

	// Haven only had Fortify/Prosper available (no env-backed Expand target,
	// no hostile neighbors, ally requires reciprocal relations). Because
	// Fortify requires wealth > 100 and the settlement starts at 100, only
	// Prosper can run and its growth depends on suitability. Provide an env
	// with non-zero suitability so the state visibly changes.
	if settlements[0].Wealth == initialWealth && settlements[0].MilitaryStrength == 100 {
		t.Logf("final state: wealth=%v military=%v", settlements[0].Wealth, settlements[0].MilitaryStrength)
		t.Error("expected agent actions to mutate settlement state over 10 years")
	}
}

func TestSettlementTickExpandAppendsSettlement(t *testing.T) {
	// Build a world state with an undiscovered node adjacent to Haven so
	// Expand finds a target.
	ws := world.NewState(32, 32)
	for i := range ws.Suitability {
		ws.Suitability[i] = 0.8
	}

	graph := dompointcrawl.NewGraph()
	graph.AddNode(&dompointcrawl.Node{ID: 1, X: 5, Y: 1, Visibility: dompointcrawl.Unknown, Name: "Ruin", Kind: "ruin"})

	haven := newAgentTestSettlement("Haven", 1000)
	haven.X, haven.Y = 1, 1
	haven.Wealth = 1000 // enough for Expand (cost 200)
	haven.Goals = []string{"expand"}

	settlements := []world.Settlement{haven}

	env := &agentEnv{
		worldState: ws,
		graph:      graph,
		all:        &settlements,
		usedNames:  map[string]bool{"Haven": true},
	}

	engine := state.NewEngine(11)
	entity := &settlementEntity{
		settlement:      &settlements[0],
		figureRNG:       engine.GetPRNG("figures:Haven"),
		agentRNG:        engine.GetPRNG("agent:Haven"),
		pointcrawlGraph: graph,
		allSettlements:  &settlements,
		env:             env,
	}

	eventChan := make(chan domsim.Event, 256)
	go func() {
		for year := 1; year <= 20 && len(settlements) == 1; year++ {
			entity.Tick(year, eventChan, engine.GetPRNG("timeline"))
		}
		close(eventChan)
	}()

	expansionSeen := false
	for event := range eventChan {
		if event.Category == "Expansion" && strings.Contains(event.Description, "founded") {
			expansionSeen = true
		}
	}

	if !expansionSeen {
		t.Fatal("expected an Expansion event founding a new settlement")
	}
	if len(settlements) != 2 {
		t.Fatalf("expected expansion to append a settlement, got %d", len(settlements))
	}

	child := settlements[1]
	if child.Faction != "testers" {
		t.Errorf("child faction = %q, want parent faction %q", child.Faction, "testers")
	}
	if child.Name == "Haven" {
		t.Error("child settlement must have a unique name")
	}
	if child.Relations["Haven"] < 2*world.RelationShiftSameFactionBaseline-1e-9 || child.Relations["Haven"] > 2*world.RelationShiftSameFactionBaseline+1e-9 {
		t.Errorf("child relations toward parent = %v, want %v", child.Relations["Haven"], 2*world.RelationShiftSameFactionBaseline)
	}
	if settlements[0].Wealth != 800 {
		t.Errorf("parent wealth = %v, want 800 after expansion cost", settlements[0].Wealth)
	}
}

func TestExtractAmount(t *testing.T) {
	cases := []struct {
		description string
		want        string
	}{
		{"Haven raided Blackgate and seized 50 wealth", "50"},
		{"Haven raided Blackgate and seized 50 wealth.", "50"},
		{"Haven prospers", ""},
		{"no numbers here", ""},
		{"year 42 saw 100 wealth change hands", "100"}, // 100 followed by wealth
	}
	for _, tc := range cases {
		if got := extractAmount(tc.description); got != tc.want {
			t.Errorf("extractAmount(%q) = %q, want %q", tc.description, got, tc.want)
		}
	}
}

func TestIsAgentCategory(t *testing.T) {
	for _, cat := range []string{"Expansion", "Raid", "Conquest", "Diplomacy", "Economy"} {
		if !isAgentCategory(cat) {
			t.Errorf("isAgentCategory(%q) = false, want true", cat)
		}
	}
	for _, cat := range []string{"Birth", "Death", "Conflict", "", "expansion"} {
		if isAgentCategory(cat) {
			t.Errorf("isAgentCategory(%q) = true, want false", cat)
		}
	}
}

// TestAgentRNGAbsenceSkipsDecisionLoop guards the nil-safety of the agent
// decision loop for legacy entity construction.
func TestAgentRNGAbsenceSkipsDecisionLoop(t *testing.T) {
	settlements := []world.Settlement{newAgentTestSettlement("Haven", 100)}
	entity := &settlementEntity{
		settlement:      &settlements[0],
		figureRNG:       randv2.New(randv2.NewPCG(1, 2)),
		pointcrawlGraph: dompointcrawl.NewGraph(),
		allSettlements:  &settlements,
	}

	eventChan := make(chan domsim.Event, 16)
	go func() {
		entity.Tick(1, eventChan, randv2.New(randv2.NewPCG(3, 4)))
		close(eventChan)
	}()

	for event := range eventChan {
		if isAgentCategory(event.Category) {
			t.Errorf("legacy entity without agentRNG must not emit agent events, got %+v", event)
		}
	}
}

// TestSettlementTickNoYearZeroEvents verifies that role-generated events are
// stamped with the current year rather than left as year 0.
func TestSettlementTickNoYearZeroEvents(t *testing.T) {
	s := newAgentTestSettlement("Haven", 1000)
	s.Figures = []figures.HistoricalFigure{
		{ID: "Haven-0", Name: "Aldric", BirthYear: 0, MaxAge: 70, Role: "Leader"},
	}
	settlements := []world.Settlement{s}

	engine := state.NewEngine(1)
	env := &agentEnv{all: &settlements, usedNames: map[string]bool{"Haven": true}}
	entity := &settlementEntity{
		settlement:      &settlements[0],
		figureRNG:       engine.GetPRNG("figures:Haven"),
		agentRNG:        engine.GetPRNG("agent:Haven"),
		pointcrawlGraph: dompointcrawl.NewGraph(),
		allSettlements:  &settlements,
		env:             env,
	}

	eventChan := make(chan domsim.Event, 512)
	go func() {
		for year := 1; year <= 50; year++ {
			entity.Tick(year, eventChan, engine.GetPRNG("timeline"))
		}
		close(eventChan)
	}()

	for event := range eventChan {
		if event.Year == 0 {
			t.Fatalf("event has year 0: %+v", event)
		}
	}
}

// TestTickDeterministicPerSeed verifies that two identically-seeded entity
// setups produce byte-identical event streams.
func TestTickDeterministicPerSeed(t *testing.T) {
	run := func() string {
		settlements := []world.Settlement{
			newAgentTestSettlement("Haven", 1000),
			newAgentTestSettlement("Blackgate", 800),
		}
		settlements[0].Relations = world.InitRelations(settlements[0], settlements)
		settlements[1].Relations = world.InitRelations(settlements[1], settlements)

		engine := state.NewEngine(99)
		env := &agentEnv{all: &settlements, usedNames: map[string]bool{"Haven": true, "Blackgate": true}}

		var b strings.Builder
		eventChan := make(chan domsim.Event, 512)
		entities := []*settlementEntity{
			{
				settlement:      &settlements[0],
				figureRNG:       engine.GetPRNG("figures:Haven"),
				agentRNG:        engine.GetPRNG("agent:Haven"),
				pointcrawlGraph: dompointcrawl.NewGraph(),
				allSettlements:  &settlements,
				env:             env,
			},
			{
				settlement:      &settlements[1],
				figureRNG:       engine.GetPRNG("figures:Blackgate"),
				agentRNG:        engine.GetPRNG("agent:Blackgate"),
				pointcrawlGraph: dompointcrawl.NewGraph(),
				allSettlements:  &settlements,
				env:             env,
			},
		}

		go func() {
			for year := 1; year <= 15; year++ {
				for _, e := range entities {
					e.Tick(year, eventChan, engine.GetPRNG("timeline"))
				}
			}
			close(eventChan)
		}()

		for event := range eventChan {
			fmt.Fprintf(&b, "%s|%s|%s|%s\n", fmt.Sprint(event.Year), event.Category, event.SettlementName, event.Description)
		}
		return b.String()
	}

	first := run()
	second := run()
	if first != second {
		t.Fatalf("same seed produced different event streams:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}
