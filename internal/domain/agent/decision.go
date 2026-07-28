package agent

import (
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// weightedAction couples an action with its selection weight.
type weightedAction struct {
	action Action
	weight float64
}

// ChooseAction evaluates every known action against the settlement's state,
// filters by preconditions, scores by goal alignment, and selects one via
// deterministic weighted random. When nothing passes preconditions it falls
// back to Prosper, which always succeeds.
func ChooseAction(self *world.Settlement, all []world.Settlement, env AgentEnv, rng *randv2.Rand) Action {
	candidates := make([]weightedAction, 0, len(AllActions()))
	for _, action := range AllActions() {
		if !action.Preconditions(self, all, env) {
			continue
		}
		candidates = append(candidates, weightedAction{
			action: action,
			weight: ScoreAction(action, self.Goals, self),
		})
	}

	if len(candidates) == 0 {
		return ProsperAction{}
	}

	return weightedRandom(candidates, rng)
}

// ScoreAction returns the goal-alignment weight for an action: 3.0 when the
// action directly serves one of the settlement's goals, 2.0 for indirect
// support, 1.0 otherwise.
func ScoreAction(action Action, goals []string, self *world.Settlement) float64 {
	_ = goals
	return action.Score(self)
}

// weightedRandom picks one action proportionally to its weight using the
// provided RNG, making the selection deterministic for a given stream.
func weightedRandom(candidates []weightedAction, rng *randv2.Rand) Action {
	total := 0.0
	for _, c := range candidates {
		if c.weight > 0 {
			total += c.weight
		}
	}

	if total <= 0 {
		return candidates[0].action
	}

	draw := rng.Float64() * total
	for _, c := range candidates {
		if c.weight <= 0 {
			continue
		}
		draw -= c.weight
		if draw < 0 {
			return c.action
		}
	}

	return candidates[len(candidates)-1].action
}
