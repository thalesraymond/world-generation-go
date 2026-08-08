package agent

import (
	randv2 "math/rand/v2"
	"slices"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// testEnv is a deterministic AgentEnv stub.
type testEnv struct {
	suitability   float64
	targetX       int
	targetY       int
	targetOK      bool
	name          string
	maxRange      float64
	targetQueries int
}

func (e *testEnv) Suitability(x, y int) float64 { return e.suitability }
func (e *testEnv) FindExpansionTarget(self *world.Settlement, rng *randv2.Rand) (int, int, bool) {
	e.targetQueries++
	return e.targetX, e.targetY, e.targetOK
}
func (e *testEnv) GenerateName(rng *randv2.Rand) string { return e.name }
func (e *testEnv) MaxActionRange() float64              { return e.maxRange }

func newTestRNG() *randv2.Rand {
	return randv2.New(randv2.NewPCG(99, 13))
}

func approxEqual(a, b float64) bool {
	const eps = 1e-9
	d := a - b
	return d < eps && d > -eps
}

func baseSettlement() world.Settlement {
	return world.Settlement{
		Name:             "Alpha",
		X:                0,
		Y:                0,
		Faction:          "auric",
		Population:       100,
		MilitaryStrength: 100,
		Wealth:           300,
		Relations:        map[string]float64{},
		Goals:            []string{"grow"},
	}
}

// ── AllActions ─────────────────────────────────────────

func TestAllActionsReturnsSixActions(t *testing.T) {
	actions := AllActions()
	if len(actions) != 6 {
		t.Fatalf("AllActions() = %d actions, want 6", len(actions))
	}

	names := make([]string, 0, 6)
	for _, a := range actions {
		names = append(names, a.Name())
	}
	for _, want := range []string{"Expand", "Raid", "Conquer", "Fortify", "Ally", "Prosper"} {
		if !slices.Contains(names, want) {
			t.Fatalf("AllActions() missing %q (got %v)", want, names)
		}
	}
}

// ── ExpandAction ───────────────────────────────────────

func TestExpandPreconditions(t *testing.T) {
	env := &testEnv{targetOK: true, maxRange: 10}

	lowPop := baseSettlement()
	lowPop.Population = ExpandMinPopulation
	if (ExpandAction{}).Preconditions(&lowPop, nil, env) {
		t.Fatal("Expand precondition passed with insufficient population")
	}

	poor := baseSettlement()
	poor.Wealth = ExpandCost
	if (ExpandAction{}).Preconditions(&poor, nil, env) {
		t.Fatal("Expand precondition passed with insufficient wealth")
	}

	ok := baseSettlement()
	if !(ExpandAction{}).Preconditions(&ok, nil, env) {
		t.Fatal("Expand precondition failed for wealthy populous settlement")
	}

	if (ExpandAction{}).Preconditions(&ok, nil, nil) {
		t.Fatal("Expand precondition passed with nil env")
	}
}

func TestExpandScore(t *testing.T) {
	expander := baseSettlement()
	expander.Goals = []string{"expand"}
	if got := (ExpandAction{}).Score(&expander); got != ExpandScoreAligned {
		t.Fatalf("Expand score with expand goal = %v, want %v", got, ExpandScoreAligned)
	}

	neutral := baseSettlement()
	if got := (ExpandAction{}).Score(&neutral); got != 1.0 {
		t.Fatalf("Expand score without goal = %v, want 1.0", got)
	}
}

func TestExpandExecuteCreatesSettlement(t *testing.T) {
	env := &testEnv{targetX: 5, targetY: 5, targetOK: true, name: "Newhold", maxRange: 10}
	self := baseSettlement()
	all := []world.Settlement{self}

	event := (ExpandAction{}).Execute(&all[0], &all, env, newTestRNG())

	if event.Category != "Expansion" {
		t.Fatalf("event category = %q, want Expansion", event.Category)
	}
	if event.Description != "Alpha founded Newhold" {
		t.Fatalf("event description = %q", event.Description)
	}
	if len(all) != 2 {
		t.Fatalf("settlements = %d, want 2", len(all))
	}

	child := all[1]
	if child.Name != "Newhold" || child.X != 5 || child.Y != 5 {
		t.Fatalf("child = %+v", child)
	}
	if child.Faction != "auric" {
		t.Fatalf("child faction = %q, want auric", child.Faction)
	}
	if child.Population != self.Population*0.2 {
		t.Fatalf("child population = %v, want %v", child.Population, self.Population*0.2)
	}
	if child.MilitaryStrength != self.MilitaryStrength*0.2 {
		t.Fatalf("child military = %v, want %v", child.MilitaryStrength, self.MilitaryStrength*0.2)
	}
	if child.Wealth != self.Wealth*0.3 {
		t.Fatalf("child wealth = %v, want %v", child.Wealth, self.Wealth*0.3)
	}
	if len(child.Goals) < 2 || len(child.Goals) > 3 {
		t.Fatalf("child goals = %v, want 2-3 entries", child.Goals)
	}
	// Child has same-faction baseline with parent, then a founder shift is applied.
	if child.Relations["Alpha"] != 2*world.RelationShiftSameFactionBaseline {
		t.Fatalf("child relation to parent = %v, want %v", child.Relations["Alpha"], 2*world.RelationShiftSameFactionBaseline)
	}
	if all[0].Wealth != self.Wealth-ExpandCost {
		t.Fatalf("parent wealth = %v, want %v", all[0].Wealth, self.Wealth-ExpandCost)
	}
	if all[0].Relations["Newhold"] != world.RelationShiftSameFactionBaseline {
		t.Fatalf("parent relation to child = %v", all[0].Relations["Newhold"])
	}
}

func TestExpandExecuteNoTargets(t *testing.T) {
	env := &testEnv{targetOK: false, maxRange: 10}
	self := baseSettlement()
	all := []world.Settlement{self}

	event := (ExpandAction{}).Execute(&all[0], &all, env, newTestRNG())

	if event.Description != "Alpha found no site to settle" {
		t.Fatalf("event description = %q", event.Description)
	}
	if len(all) != 1 {
		t.Fatalf("settlements = %d, want 1 (no expansion)", len(all))
	}
	if all[0].Wealth != self.Wealth {
		t.Fatalf("wealth = %v, want %v (unchanged)", all[0].Wealth, self.Wealth)
	}
}

// ── RaidAction ─────────────────────────────────────────

func raidSetup() ([]world.Settlement, *testEnv) {
	raider := baseSettlement()
	target := world.Settlement{
		Name: "Beta", X: 5, Y: 0, Faction: "sylvani",
		Population: 80, MilitaryStrength: 50, Wealth: 200,
		Relations: map[string]float64{"Alpha": -0.6},
	}
	raider.Relations["Beta"] = -0.6
	all := []world.Settlement{raider, target}
	return all, &testEnv{maxRange: 10}
}

func TestRaidPreconditions(t *testing.T) {
	all, env := raidSetup()
	if !(RaidAction{}).Preconditions(&all[0], all, env) {
		t.Fatal("Raid precondition failed for valid hostile target")
	}

	// Insufficient military.
	all, env = raidSetup()
	all[0].MilitaryStrength = 30 // <= 50 * 0.8
	if (RaidAction{}).Preconditions(&all[0], all, env) {
		t.Fatal("Raid precondition passed with insufficient military")
	}

	// Relations not hostile enough.
	all, env = raidSetup()
	all[0].Relations["Beta"] = -0.4
	if (RaidAction{}).Preconditions(&all[0], all, env) {
		t.Fatal("Raid precondition passed with relations >= -0.5")
	}

	// Out of range.
	all, env = raidSetup()
	env.maxRange = 2
	if (RaidAction{}).Preconditions(&all[0], all, env) {
		t.Fatal("Raid precondition passed for out-of-range target")
	}
}

func TestRaidScore(t *testing.T) {
	self := baseSettlement()
	if got := (RaidAction{}).Score(&self); got != 1.0 {
		t.Fatalf("Raid score = %v, want 1.0", got)
	}
}

func TestRaidExecuteSuccess(t *testing.T) {
	all, env := raidSetup()
	rng := randv2.New(randv2.NewPCG(1, 1))

	event := (RaidAction{}).Execute(&all[0], &all, env, rng)

	if event.Category != "Raid" || event.TargetSettlement != "Beta" {
		t.Fatalf("event = %+v", event)
	}

	// Verify both outcomes across many seeds produce coherent state.
	successSeen, failureSeen := false, false
	for seed := uint64(0); seed < 50; seed++ {
		all, env := raidSetup()
		rng := randv2.New(randv2.NewPCG(seed, seed+1))
		before := all[0].Wealth
		event := (RaidAction{}).Execute(&all[0], &all, env, rng)

		if all[0].Wealth == before+RaidWealthTransfer {
			successSeen = true
			if event.Description != "Alpha raided Beta and seized 50 wealth" {
				t.Fatalf("success description = %q", event.Description)
			}
			if !approxEqual(all[1].Relations["Alpha"], -0.9) {
				t.Fatalf("target relations = %v, want -0.9", all[1].Relations["Alpha"])
			}
			if all[1].Wealth != 200-RaidWealthTransfer {
				t.Fatalf("target wealth = %v, want %v", all[1].Wealth, 200-RaidWealthTransfer)
			}
			if all[0].Relations["Beta"] != -1.0 {
				t.Fatalf("raider relations = %v, want -1.0 (clamped)", all[0].Relations["Beta"])
			}
		} else {
			failureSeen = true
			if event.Description != "Alpha raided Beta but was driven off" {
				t.Fatalf("failure description = %q", event.Description)
			}
			if !approxEqual(all[0].Relations["Beta"], -0.8) {
				t.Fatalf("raider relations after failure = %v, want -0.8", all[0].Relations["Beta"])
			}
			if all[1].Wealth != 200 {
				t.Fatalf("target wealth = %v, want 200 (unchanged)", all[1].Wealth)
			}
		}
	}
	if !successSeen || !failureSeen {
		t.Fatalf("raid outcomes not exercised: success=%v failure=%v", successSeen, failureSeen)
	}
}

func TestRaidExecuteNoTarget(t *testing.T) {
	self := baseSettlement()
	all := []world.Settlement{self}
	env := &testEnv{maxRange: 10}

	event := (RaidAction{}).Execute(&all[0], &all, env, newTestRNG())
	if event.Description != "Alpha sought war in vain" {
		t.Fatalf("event description = %q", event.Description)
	}
	if event.TargetSettlement != "" {
		t.Fatalf("target = %q, want empty", event.TargetSettlement)
	}
}

// ── ConquerAction ──────────────────────────────────────

func conquerSetup() ([]world.Settlement, *testEnv) {
	attacker := baseSettlement()
	target := world.Settlement{
		Name: "Beta", X: 5, Y: 0, Faction: "sylvani",
		Population: 80, MilitaryStrength: 40, Wealth: 200,
		Relations: map[string]float64{"Alpha": -0.8},
	}
	attacker.Relations["Beta"] = -0.8
	return []world.Settlement{attacker, target}, &testEnv{maxRange: 10}
}

func TestConquerPreconditions(t *testing.T) {
	all, env := conquerSetup()
	if !(ConquerAction{}).Preconditions(&all[0], all, env) {
		t.Fatal("Conquer precondition failed for valid target")
	}

	all, env = conquerSetup()
	all[0].MilitaryStrength = 50 // <= 40 * 1.5
	if (ConquerAction{}).Preconditions(&all[0], all, env) {
		t.Fatal("Conquer precondition passed with insufficient military")
	}

	all, env = conquerSetup()
	all[0].Relations["Beta"] = -0.6
	if (ConquerAction{}).Preconditions(&all[0], all, env) {
		t.Fatal("Conquer precondition passed with relations >= -0.7")
	}

	all, env = conquerSetup()
	env.maxRange = 2
	if (ConquerAction{}).Preconditions(&all[0], all, env) {
		t.Fatal("Conquer precondition passed for out-of-range target")
	}
}

func TestConquerScore(t *testing.T) {
	expansionist := baseSettlement()
	expansionist.Goals = []string{"expand"}
	if got := (ConquerAction{}).Score(&expansionist); got != ConquerScoreAligned {
		t.Fatalf("Conquer score with expand goal = %v, want %v", got, ConquerScoreAligned)
	}

	neutral := baseSettlement()
	if got := (ConquerAction{}).Score(&neutral); got != 1.0 {
		t.Fatalf("Conquer score without goal = %v, want 1.0", got)
	}
}

func TestConquerExecuteAbsorbsSettlement(t *testing.T) {
	all, env := conquerSetup()

	event := (ConquerAction{}).Execute(&all[0], &all, env, newTestRNG())

	if event.Category != "Conquest" || event.TargetSettlement != "Beta" {
		t.Fatalf("event = %+v", event)
	}
	if event.Description != "Alpha conquered Beta" {
		t.Fatalf("description = %q", event.Description)
	}
	if all[1].Faction != "auric" {
		t.Fatalf("target faction = %q, want auric", all[1].Faction)
	}
	if all[0].MilitaryStrength != 80.0 {
		t.Fatalf("attacker military = %v, want 80 (20%% war cost)", all[0].MilitaryStrength)
	}
	if all[0].Relations["Beta"] != -1.0 {
		t.Fatalf("attacker relations = %v, want -1.0 (clamped)", all[0].Relations["Beta"])
	}
	if all[1].Relations["Alpha"] != -1.0 {
		t.Fatalf("target relations = %v, want -1.0 (clamped)", all[1].Relations["Alpha"])
	}
}

// ── FortifyAction ──────────────────────────────────────

func TestFortifyPreconditions(t *testing.T) {
	rich := baseSettlement()
	if !(FortifyAction{}).Preconditions(&rich, nil, nil) {
		t.Fatal("Fortify precondition failed with wealth > 100")
	}

	poor := baseSettlement()
	poor.Wealth = FortifyMinWealth
	if (FortifyAction{}).Preconditions(&poor, nil, nil) {
		t.Fatal("Fortify precondition passed with wealth <= 100")
	}
}

func TestFortifyScore(t *testing.T) {
	defender := baseSettlement()
	defender.Goals = []string{"defend"}
	if got := (FortifyAction{}).Score(&defender); got != FortifyScoreAligned {
		t.Fatalf("Fortify score with defend goal = %v, want %v", got, FortifyScoreAligned)
	}

	grower := baseSettlement()
	if got := (FortifyAction{}).Score(&grower); got != FortifyScoreGrow {
		t.Fatalf("Fortify score with grow goal = %v, want %v", got, FortifyScoreGrow)
	}
}

func TestFortifyExecuteConversion(t *testing.T) {
	self := baseSettlement()
	all := []world.Settlement{self}

	event := (FortifyAction{}).Execute(&all[0], &all, nil, newTestRNG())

	if all[0].Wealth != self.Wealth-FortifyConversion {
		t.Fatalf("wealth = %v, want %v", all[0].Wealth, self.Wealth-FortifyConversion)
	}
	if all[0].MilitaryStrength != self.MilitaryStrength+FortifyConversion {
		t.Fatalf("military = %v, want %v", all[0].MilitaryStrength, self.MilitaryStrength+FortifyConversion)
	}
	if event.Category != "Economy" {
		t.Fatalf("event category = %q, want Economy", event.Category)
	}
	if event.Description != "Alpha invests in fortifications" {
		t.Fatalf("description = %q", event.Description)
	}
}

// ── AllyAction ─────────────────────────────────────────

func allySetup() []world.Settlement {
	self := baseSettlement()
	partner := world.Settlement{
		Name: "Beta", X: 5, Y: 0, Faction: "auric",
		Relations: map[string]float64{"Alpha": 0.55},
	}
	self.Relations["Beta"] = 0.55
	return []world.Settlement{self, partner}
}

func TestAllyPreconditions(t *testing.T) {
	all := allySetup()
	if !(AllyAction{}).Preconditions(&all[0], all, nil) {
		t.Fatal("Ally precondition failed for friendly target")
	}

	// Hostile settlement.
	all = allySetup()
	all[0].Relations["Beta"] = 0.4
	if (AllyAction{}).Preconditions(&all[0], all, nil) {
		t.Fatal("Ally precondition passed with relations <= 0.5")
	}

	// Existing alliance flag.
	all = allySetup()
	all[1].Relations["Alpha"] = AllyFloor
	if (AllyAction{}).Preconditions(&all[0], all, nil) {
		t.Fatal("Ally precondition passed with existing alliance")
	}
}

func TestAllyScore(t *testing.T) {
	self := baseSettlement()
	if got := (AllyAction{}).Score(&self); got != 1.0 {
		t.Fatalf("Ally score = %v, want 1.0", got)
	}
}

func TestAllyExecuteSetsAllianceFlag(t *testing.T) {
	all := allySetup()

	event := (AllyAction{}).Execute(&all[0], &all, nil, newTestRNG())

	if event.Category != "Diplomacy" || event.TargetSettlement != "Beta" {
		t.Fatalf("event = %+v", event)
	}
	if event.Description != "Alpha forms alliance with Beta" {
		t.Fatalf("description = %q", event.Description)
	}
	if !approxEqual(all[0].Relations["Beta"], 0.95) {
		t.Fatalf("self relations = %v, want 0.95", all[0].Relations["Beta"])
	}
	if !approxEqual(all[1].Relations["Alpha"], 0.95) {
		t.Fatalf("target relations = %v, want 0.95", all[1].Relations["Alpha"])
	}
}

func TestAllyExecuteAllianceFloor(t *testing.T) {
	all := allySetup()
	all[0].Relations["Beta"] = 0.51
	all[1].Relations["Alpha"] = 0.51

	(AllyAction{}).Execute(&all[0], &all, nil, newTestRNG())

	if !approxEqual(all[0].Relations["Beta"], 0.91) {
		t.Fatalf("self relations = %v, want 0.91", all[0].Relations["Beta"])
	}
}

// ── ProsperAction ──────────────────────────────────────

func TestProsperPreconditionsAlwaysTrue(t *testing.T) {
	self := baseSettlement()
	if !(ProsperAction{}).Preconditions(&self, nil, nil) {
		t.Fatal("Prosper precondition must always pass")
	}
}

func TestProsperScore(t *testing.T) {
	grower := baseSettlement()
	if got := (ProsperAction{}).Score(&grower); got != ProsperScoreGrow {
		t.Fatalf("Prosper score with grow goal = %v, want %v", got, ProsperScoreGrow)
	}

	other := baseSettlement()
	other.Goals = []string{"defend"}
	if got := (ProsperAction{}).Score(&other); got != 1.0 {
		t.Fatalf("Prosper score without grow goal = %v, want 1.0", got)
	}
}

func TestProsperExecuteGrowthScaledBySuitability(t *testing.T) {
	self := baseSettlement()
	other := world.Settlement{Name: "Beta", Faction: "auric", Relations: map[string]float64{}}
	all := []world.Settlement{self, other}
	env := &testEnv{suitability: 0.8}

	event := (ProsperAction{}).Execute(&all[0], &all, env, newTestRNG())

	wantPop := self.Population + ProsperPopulationGrowth*0.8
	if all[0].Population != wantPop {
		t.Fatalf("population = %v, want %v", all[0].Population, wantPop)
	}
	wantWealth := self.Wealth + ProsperWealthGrowth*0.8
	if all[0].Wealth != wantWealth {
		t.Fatalf("wealth = %v, want %v", all[0].Wealth, wantWealth)
	}
	if all[0].Relations["Beta"] != world.RelationShiftProsper {
		t.Fatalf("relations = %v, want %v", all[0].Relations["Beta"], world.RelationShiftProsper)
	}
	if event.Category != "Economy" || event.Description != "Alpha prospers" {
		t.Fatalf("event = %+v", event)
	}
}

func TestProsperExecuteDoesNotShiftCrossFactionRelations(t *testing.T) {
	self := baseSettlement()
	other := world.Settlement{Name: "Beta", Faction: "sylvani", Relations: map[string]float64{}}
	all := []world.Settlement{self, other}
	env := &testEnv{suitability: 0.8}

	(ProsperAction{}).Execute(&all[0], &all, env, newTestRNG())

	// Prosper must NOT affect relations with different-faction settlements.
	if _, exists := all[0].Relations["Beta"]; exists {
		t.Fatalf("Prosper affected cross-faction relations: %v", all[0].Relations["Beta"])
	}
}

func TestProsperExecuteNilEnvZeroSuitability(t *testing.T) {
	self := baseSettlement()
	all := []world.Settlement{self}

	(ProsperAction{}).Execute(&all[0], &all, nil, newTestRNG())

	if all[0].Population != self.Population || all[0].Wealth != self.Wealth {
		t.Fatalf("state changed with nil env: %+v", all[0])
	}
}

// ── Decision loop ──────────────────────────────────────

func TestChooseActionFallbackToProsper(t *testing.T) {
	self := baseSettlement()
	self.Wealth = 0
	self.Population = 10
	self.Relations = map[string]float64{}
	all := []world.Settlement{self}
	env := &testEnv{maxRange: 10, targetOK: false}

	// Wealth 0: Fortify fails. Pop 10 & wealth 0: Expand fails. No hostile or
	// friendly relations: Raid/Conquer/Ally fail. Only Prosper remains.
	action := ChooseAction(&all[0], all, env, newTestRNG())
	if action.Name() != "Prosper" {
		t.Fatalf("ChooseAction = %q, want Prosper", action.Name())
	}
}

func TestChooseActionPreconditionFiltering(t *testing.T) {
	self := baseSettlement()
	self.Wealth = 50 // Fortify (needs >100) and Expand (needs >200) fail.
	all := []world.Settlement{self}
	env := &testEnv{maxRange: 10}

	action := ChooseAction(&all[0], all, env, newTestRNG())
	if action.Name() != "Prosper" {
		t.Fatalf("ChooseAction = %q, want Prosper (only passing action)", action.Name())
	}
}

func TestChooseActionDeterminism(t *testing.T) {
	build := func() ([]world.Settlement, *testEnv) {
		self := baseSettlement()
		self.Goals = []string{"grow", "defend", "expand"}
		return []world.Settlement{self}, &testEnv{maxRange: 10, targetOK: true, targetX: 3, targetY: 3, name: "Newhold"}
	}

	allA, envA := build()
	allB, envB := build()

	gotA := ChooseAction(&allA[0], allA, envA, newTestRNG())
	gotB := ChooseAction(&allB[0], allB, envB, newTestRNG())

	if gotA.Name() != gotB.Name() {
		t.Fatalf("non-deterministic choice: %q vs %q", gotA.Name(), gotB.Name())
	}
}

func TestScoreActionGoalAlignment(t *testing.T) {
	self := baseSettlement()
	self.Goals = []string{"expand"}

	if got := ScoreAction(ExpandAction{}, self.Goals, &self); got != ExpandScoreAligned {
		t.Fatalf("Expand score = %v, want %v", got, ExpandScoreAligned)
	}
	if got := ScoreAction(FortifyAction{}, self.Goals, &self); got != 1.0 {
		t.Fatalf("Fortify score = %v, want 1.0", got)
	}

	self.Goals = []string{"grow"}
	if got := ScoreAction(ProsperAction{}, self.Goals, &self); got != ProsperScoreGrow {
		t.Fatalf("Prosper score = %v, want %v", got, ProsperScoreGrow)
	}
}

func TestWeightedRandomDeterminism(t *testing.T) {
	candidates := []weightedAction{
		{action: ExpandAction{}, weight: 3.0},
		{action: FortifyAction{}, weight: 1.0},
		{action: ProsperAction{}, weight: 2.0},
	}

	first := weightedRandom(candidates, newTestRNG())
	for i := 0; i < 10; i++ {
		if got := weightedRandom(candidates, newTestRNG()); got.Name() != first.Name() {
			t.Fatalf("non-deterministic weightedRandom: %q vs %q", first.Name(), got.Name())
		}
	}
}

func TestWeightedRandomZeroWeights(t *testing.T) {
	candidates := []weightedAction{
		{action: ExpandAction{}, weight: 0},
		{action: ProsperAction{}, weight: 0},
	}

	got := weightedRandom(candidates, newTestRNG())
	if got.Name() != "Expand" {
		t.Fatalf("weightedRandom with zero total = %q, want first candidate", got.Name())
	}
}

func TestWeightedRandomDistribution(t *testing.T) {
	candidates := []weightedAction{
		{action: ExpandAction{}, weight: 9.0},
		{action: ProsperAction{}, weight: 1.0},
	}

	counts := map[string]int{}
	for seed := uint64(0); seed < 200; seed++ {
		rng := randv2.New(randv2.NewPCG(seed, seed+1000))
		counts[weightedRandom(candidates, rng).Name()]++
	}

	if counts["Expand"] <= counts["Prosper"] {
		t.Fatalf("weights ignored: %v", counts)
	}
}

// ── RandomGoals ────────────────────────────────────────

func TestRandomGoalsCountAndUniqueness(t *testing.T) {
	for seed := uint64(0); seed < 100; seed++ {
		rng := randv2.New(randv2.NewPCG(seed, seed+1))
		goals := RandomGoals(rng)

		if len(goals) < 2 || len(goals) > 3 {
			t.Fatalf("goals = %v, want 2-3 entries", goals)
		}

		seen := map[string]bool{}
		for _, g := range goals {
			if seen[g] {
				t.Fatalf("duplicate goal in %v", goals)
			}
			seen[g] = true
			if !slices.Contains(GoalPool, g) {
				t.Fatalf("unknown goal %q in %v", g, goals)
			}
		}
	}
}

func TestRandomGoalsDeterminism(t *testing.T) {
	a := RandomGoals(newTestRNG())
	b := RandomGoals(newTestRNG())

	if !slices.Equal(a, b) {
		t.Fatalf("non-deterministic goals: %v vs %v", a, b)
	}
}

// ── Relations edge cases via actions ───────────────────

func TestRelationsShiftsAccumulateAcrossRaids(t *testing.T) {
	all, env := raidSetup()

	for i := 0; i < 3; i++ {
		// Force success by searching for a succeeding seed.
		rng := randv2.New(randv2.NewPCG(uint64(i), 1))
		(RaidAction{}).Execute(&all[0], &all, env, rng)
	}

	if all[0].Relations["Beta"] != -1.0 {
		t.Fatalf("relations after repeated raids = %v, want -1.0 (clamped)", all[0].Relations["Beta"])
	}
}

func TestRelationsNeverExceedBounds(t *testing.T) {
	all := allySetup()
	all[0].Relations["Beta"] = 0.99
	all[1].Relations["Alpha"] = 0.99

	(AllyAction{}).Execute(&all[0], &all, nil, newTestRNG())

	if all[0].Relations["Beta"] > 1.0 || all[1].Relations["Alpha"] > 1.0 {
		t.Fatalf("relations exceeded +1.0: %v / %v", all[0].Relations["Beta"], all[1].Relations["Alpha"])
	}
}

func TestAsymmetricRelations(t *testing.T) {
	all, env := raidSetup()
	all[0].Relations["Beta"] = -0.55
	all[1].Relations["Alpha"] = 0.9 // target likes the raider

	(RaidAction{}).Execute(&all[0], &all, env, randv2.New(randv2.NewPCG(1, 1)))

	// Whatever the raid outcome, the two directions move independently.
	if all[0].Relations["Beta"] == all[1].Relations["Alpha"] {
		t.Fatalf("relations should be asymmetric, both = %v", all[0].Relations["Beta"])
	}
}
