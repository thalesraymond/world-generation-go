package world

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
