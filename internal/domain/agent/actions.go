package agent

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// Expand action tuning.
const (
	ExpandMinPopulation = 50.0
	ExpandCost          = 200.0
	ExpandScoreAligned  = 3.0
)

// ExpandAction founds a new settlement on an unclaimed suitable site.
type ExpandAction struct{}

func (ExpandAction) Name() string { return "Expand" }

func (a ExpandAction) Preconditions(self *world.Settlement, all []world.Settlement, env AgentEnv) bool {
	if self.Population <= ExpandMinPopulation {
		return false
	}
	if self.Wealth <= ExpandCost {
		return false
	}
	return env != nil
}

func (ExpandAction) Score(self *world.Settlement) float64 {
	if hasGoal(self, "expand") {
		return ExpandScoreAligned
	}
	return 1.0
}

func (ExpandAction) Execute(self *world.Settlement, all *[]world.Settlement, env AgentEnv, rng *randv2.Rand) simulation.Event {
	event := simulation.Event{Year: -1, Category: "Expansion", SettlementName: self.Name}

	if env == nil {
		event.Description = fmt.Sprintf("%s expansion failed: no suitable targets", self.Name)
		return event
	}

	x, y, ok := env.FindExpansionTarget(self, rng)
	if !ok {
		event.Description = fmt.Sprintf("%s expansion failed: no suitable targets", self.Name)
		return event
	}

	childName := env.GenerateName(rng)
	child := world.Settlement{
		Name:             childName,
		Type:             "Outpost",
		X:                x,
		Y:                y,
		Faction:          self.Faction,
		Population:       self.Population * 0.2,
		MilitaryStrength: self.MilitaryStrength * 0.2,
		Wealth:           self.Wealth * 0.3,
		Figures:          figures.GenerateFounders(rng, childName, self.Faction, 0),
		Goals:            RandomGoals(rng),
	}
	child.Relations = world.InitRelations(child, *all)
	// Apply cross-faction friction for the new settlement.
	world.ApplySettlementCrossFactionFriction(&child, *all, rng)

	self.Wealth -= ExpandCost

	*all = append(*all, child)
	self.Relations[child.Name] = world.RelationShiftSameFactionBaseline
	world.ShiftRelations(findByName(all, child.Name), self.Name, world.RelationShiftSameFactionBaseline)

	event.Description = fmt.Sprintf("%s founded %s", self.Name, child.Name)
	return event
}

// Raid action tuning.
const (
	RaidMilitaryRatio  = 0.8
	RaidMaxRelations   = -0.5
	RaidSuccessChance  = 0.7
	RaidWealthTransfer = 50.0
	RaidScoreAligned   = 1.5
	RaidOutcomeSuccess = "success"
	RaidOutcomeFailure = "failure"
)

// RaidAction steals wealth from a hostile neighbor within range.
type RaidAction struct{}

func (RaidAction) Name() string { return "Raid" }

func (RaidAction) Preconditions(self *world.Settlement, all []world.Settlement, env AgentEnv) bool {
	if env == nil {
		return false
	}
	return raidTarget(self, all, env.MaxActionRange()) != ""
}

func (RaidAction) Score(self *world.Settlement) float64 {
	return 1.0
}

func (RaidAction) Execute(self *world.Settlement, all *[]world.Settlement, env AgentEnv, rng *randv2.Rand) simulation.Event {
	event := simulation.Event{Year: -1, Category: "Raid", SettlementName: self.Name}

	target := raidTarget(self, *all, env.MaxActionRange())
	if target == "" {
		event.Description = fmt.Sprintf("%s found no worthwhile raid targets", self.Name)
		return event
	}

	event.TargetSettlement = target
	targetSettlement := findByName(all, target)

	if rng.Float64() < RaidSuccessChance {
		if targetSettlement != nil {
			targetSettlement.Wealth -= RaidWealthTransfer
			world.ShiftRelations(targetSettlement, self.Name, world.RelationShiftRaidSuccessTarget)
		}
		self.Wealth += RaidWealthTransfer
		world.ShiftRelations(self, target, world.RelationShiftRaidSuccessSelf)

		event.Description = fmt.Sprintf("%s raided %s and seized %.0f wealth", self.Name, target, RaidWealthTransfer)
	} else {
		world.ShiftRelations(self, target, world.RelationShiftRaidFailureSelf)
		event.Description = fmt.Sprintf("%s raided %s but was driven off", self.Name, target)
	}

	return event
}

// raidTarget returns the name of the most hostile in-range settlement the
// raider can militarily challenge, or "" when none qualifies. Iteration in
// slice order keeps the choice deterministic.
func raidTarget(self *world.Settlement, all []world.Settlement, maxRange float64) string {
	for i := range all {
		other := &all[i]
		if other.Name == self.Name {
			continue
		}
		if self.Relations[other.Name] >= RaidMaxRelations {
			continue
		}
		if !withinRange(self, other, maxRange) {
			continue
		}
		if self.MilitaryStrength <= other.MilitaryStrength*RaidMilitaryRatio {
			continue
		}
		return other.Name
	}
	return ""
}

// Conquer action tuning.
const (
	ConquerMilitaryRatio = 1.5
	ConquerMaxRelations  = -0.7
	ConquerWarCostRatio  = 0.2
	ConquerScoreAligned  = 2.0
)

// ConquerAction militarily absorbs a very hostile, weaker neighbor.
type ConquerAction struct{}

func (ConquerAction) Name() string { return "Conquer" }

func (ConquerAction) Preconditions(self *world.Settlement, all []world.Settlement, env AgentEnv) bool {
	if env == nil {
		return false
	}
	return conquerTarget(self, all, env.MaxActionRange()) != ""
}

func (ConquerAction) Score(self *world.Settlement) float64 {
	if hasGoal(self, "expand") {
		return ConquerScoreAligned
	}
	return 1.0
}

func (ConquerAction) Execute(self *world.Settlement, all *[]world.Settlement, env AgentEnv, rng *randv2.Rand) simulation.Event {
	event := simulation.Event{Year: -1, Category: "Conquest", SettlementName: self.Name}

	target := conquerTarget(self, *all, env.MaxActionRange())
	if target == "" {
		event.Description = fmt.Sprintf("%s found no conquerable targets", self.Name)
		return event
	}

	event.TargetSettlement = target
	targetSettlement := findByName(all, target)

	if targetSettlement != nil {
		targetSettlement.Faction = self.Faction
		world.ShiftRelations(targetSettlement, self.Name, world.RelationShiftConquer)
	}

	self.MilitaryStrength *= 1.0 - ConquerWarCostRatio
	world.ShiftRelations(self, target, world.RelationShiftConquer)

	event.Description = fmt.Sprintf("%s conquered %s", self.Name, target)
	return event
}

// conquerTarget returns the name of the first settlement satisfying the
// conquest preconditions, or "" when none qualifies.
func conquerTarget(self *world.Settlement, all []world.Settlement, maxRange float64) string {
	for i := range all {
		other := &all[i]
		if other.Name == self.Name {
			continue
		}
		if self.Relations[other.Name] >= ConquerMaxRelations {
			continue
		}
		if !withinRange(self, other, maxRange) {
			continue
		}
		if self.MilitaryStrength <= other.MilitaryStrength*ConquerMilitaryRatio {
			continue
		}
		return other.Name
	}
	return ""
}

// Fortify action tuning.
const (
	FortifyMinWealth    = 100.0
	FortifyConversion   = 100.0
	FortifyScoreAligned = 3.0
	FortifyScoreGrow    = 2.0
)

// FortifyAction converts wealth into military strength.
type FortifyAction struct{}

func (FortifyAction) Name() string { return "Fortify" }

func (FortifyAction) Preconditions(self *world.Settlement, all []world.Settlement, env AgentEnv) bool {
	return self.Wealth > FortifyMinWealth
}

func (FortifyAction) Score(self *world.Settlement) float64 {
	if hasGoal(self, "defend") {
		return FortifyScoreAligned
	}
	if hasGoal(self, "grow") {
		return FortifyScoreGrow
	}
	return 1.0
}

func (FortifyAction) Execute(self *world.Settlement, all *[]world.Settlement, env AgentEnv, rng *randv2.Rand) simulation.Event {
	self.Wealth -= FortifyConversion
	self.MilitaryStrength += FortifyConversion

	return simulation.Event{
		Year:           -1,
		Category:       "Economy",
		SettlementName: self.Name,
		Description:    fmt.Sprintf("%s invests in fortifications", self.Name),
	}
}

// Ally action tuning.
const (
	AllyMinRelations = 0.5
	AllyFloor        = 0.6
	AllyScoreAligned = 1.5
)

// AllyAction formalizes an alliance with a friendly settlement.
type AllyAction struct{}

func (AllyAction) Name() string { return "Ally" }

func (AllyAction) Preconditions(self *world.Settlement, all []world.Settlement, env AgentEnv) bool {
	return allyTarget(self, all) != ""
}

func (AllyAction) Score(self *world.Settlement) float64 {
	return 1.0
}

func (AllyAction) Execute(self *world.Settlement, all *[]world.Settlement, env AgentEnv, rng *randv2.Rand) simulation.Event {
	event := simulation.Event{Year: -1, Category: "Diplomacy", SettlementName: self.Name}

	target := allyTarget(self, *all)
	if target == "" {
		event.Description = fmt.Sprintf("%s found no willing allies", self.Name)
		return event
	}

	event.TargetSettlement = target

	world.ShiftRelations(self, target, world.RelationShiftAlly)
	if self.Relations[target] < AllyFloor {
		self.Relations[target] = AllyFloor
	}

	targetSettlement := findByName(all, target)
	if targetSettlement != nil {
		world.ShiftRelations(targetSettlement, self.Name, world.RelationShiftAlly)
		if targetSettlement.Relations[self.Name] < AllyFloor {
			targetSettlement.Relations[self.Name] = AllyFloor
		}
	}

	event.Description = fmt.Sprintf("%s forms alliance with %s", self.Name, target)
	return event
}

// allyTarget returns the name of the first friendly settlement without an
// existing alliance, or "" when none qualifies.
func allyTarget(self *world.Settlement, all []world.Settlement) string {
	for i := range all {
		other := &all[i]
		if other.Name == self.Name {
			continue
		}
		if self.Relations[other.Name] <= AllyMinRelations {
			continue
		}
		if other.Relations[self.Name] >= AllyFloor {
			continue // alliance flag already set
		}
		return other.Name
	}
	return ""
}

// Prosper action tuning.
const (
	ProsperPopulationGrowth = 2.0
	ProsperWealthGrowth     = 5.0
	ProsperScoreGrow        = 2.0
)

// ProsperAction grows population and wealth based on local suitability.
type ProsperAction struct{}

func (ProsperAction) Name() string { return "Prosper" }

func (ProsperAction) Preconditions(self *world.Settlement, all []world.Settlement, env AgentEnv) bool {
	return true
}

func (ProsperAction) Score(self *world.Settlement) float64 {
	if hasGoal(self, "grow") {
		return ProsperScoreGrow
	}
	return 1.0
}

func (ProsperAction) Execute(self *world.Settlement, all *[]world.Settlement, env AgentEnv, rng *randv2.Rand) simulation.Event {
	suitability := 0.0
	if env != nil {
		suitability = env.Suitability(self.X, self.Y)
	}

	self.Population += ProsperPopulationGrowth * suitability
	self.Wealth += ProsperWealthGrowth * suitability

	for i := range *all {
		other := &(*all)[i]
		if other.Name == self.Name {
			continue
		}
		if other.Faction == self.Faction {
			world.ShiftRelations(self, other.Name, world.RelationShiftProsper)
		}
	}

	return simulation.Event{
		Year:           -1,
		Category:       "Economy",
		SettlementName: self.Name,
		Description:    fmt.Sprintf("%s prospers", self.Name),
	}
}
