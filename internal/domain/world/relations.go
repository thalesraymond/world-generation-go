package world

import (
	randv2 "math/rand/v2"
)

// CrossFactionFrictionMax is the maximum negative friction magnitude applied
// to cross-faction settlement pairs so rivalries can emerge naturally.
const CrossFactionFrictionMax = 0.6

// Relation shift magnitudes per action type.
const (
	// RelationShiftSameFactionBaseline is the initial relations bonus for
	// settlements sharing a non-independent faction.
	RelationShiftSameFactionBaseline = 0.3

	// RelationShiftRaidSuccessSelf is applied to the raider's relations
	// toward the target after a successful raid.
	RelationShiftRaidSuccessSelf = -0.4
	// RelationShiftRaidSuccessTarget is applied to the target's relations
	// toward the raider after a successful raid.
	RelationShiftRaidSuccessTarget = -0.3
	// RelationShiftRaidFailureSelf is applied to the raider's relations
	// toward the target after a failed raid.
	RelationShiftRaidFailureSelf = -0.2

	// RelationShiftConquer is applied in both directions after a conquest.
	RelationShiftConquer = -0.8

	// RelationShiftAlly is applied in both directions after an alliance.
	RelationShiftAlly = 0.4

	// RelationShiftProsper is applied toward every other settlement
	// when a settlement prospers.
	RelationShiftProsper = 0.05
)

// Relation bounds.
const (
	RelationMin = -1.0
	RelationMax = 1.0
)

// InitRelations builds the baseline relations map for a settlement against
// all other known settlements. Settlements sharing the same non-independent
// faction start at +0.3; all others start at 0.0. Self is excluded.
func InitRelations(self Settlement, allSettlements []Settlement) map[string]float64 {
	relations := make(map[string]float64, len(allSettlements))
	for _, other := range allSettlements {
		if other.Name == self.Name {
			continue
		}

		baseline := 0.0
		if other.Faction == self.Faction && self.Faction != "independent" {
			baseline = RelationShiftSameFactionBaseline
		}
		relations[other.Name] = baseline
	}

	return relations
}

// ShiftRelations adjusts the settlement's relations toward the target by
// delta, clamping the result to [-1.0, +1.0]. Missing entries are treated
// as 0.0 before the shift.
func ShiftRelations(self *Settlement, target string, delta float64) {
	if self == nil || target == "" || target == self.Name {
		return
	}

	if self.Relations == nil {
		self.Relations = make(map[string]float64)
	}

	self.Relations[target] = clampRelation(self.Relations[target] + delta)
}

func clampRelation(value float64) float64 {
	if value < RelationMin {
		return RelationMin
	}
	if value > RelationMax {
		return RelationMax
	}
	return value
}

// ApplyCrossFactionFriction applies random negative friction to cross-faction
// settlement pairs so rivalries can emerge naturally. Independent factions
// are excluded. Must be called after InitRelations.
func ApplyCrossFactionFriction(settlements []Settlement, rng *randv2.Rand) {
	for i := range settlements {
		for j := i + 1; j < len(settlements); j++ {
			applyFrictionPair(&settlements[i], &settlements[j], rng)
		}
	}
}

// ApplySettlementCrossFactionFriction applies cross-faction friction from a
// single settlement toward all others. Used when a new settlement is founded
// mid-simulation via expansion.
func ApplySettlementCrossFactionFriction(self *Settlement, all []Settlement, rng *randv2.Rand) {
	for i := range all {
		other := &all[i]
		if other.Name == self.Name {
			continue
		}
		applyFrictionPair(self, other, rng)
	}
}

func applyFrictionPair(a, b *Settlement, rng *randv2.Rand) {
	if a.Faction == b.Faction {
		return
	}
	if a.Faction == "independent" || b.Faction == "independent" {
		return
	}
	friction := -rng.Float64() * CrossFactionFrictionMax
	a.Relations[b.Name] = friction
	b.Relations[a.Name] = friction
}
