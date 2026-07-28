// Package agent implements settlement-level agent decision making: the six
// core actions (Expand, Raid, Conquer, Fortify, Ally, Prosper), goal-based
// scoring, and deterministic weighted-random selection.
package agent

import (
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// AgentEnv exposes the world context an action needs for preconditions and
// execution. Implemented by the simulation adapter (see cmd/simulate.go).
type AgentEnv interface {
	// Suitability returns the terrain suitability score at (x, y), or 0
	// when the coordinates are unknown.
	Suitability(x, y int) float64
	// FindExpansionTarget picks an unclaimed site for a new settlement and
	// returns its coordinates. ok is false when no site is available.
	FindExpansionTarget(self *world.Settlement, rng *randv2.Rand) (x, y int, ok bool)
	// GenerateName produces a unique settlement name.
	GenerateName(rng *randv2.Rand) string
	// MaxActionRange is the maximum Euclidean distance (in tiles) for
	// Raid and Conquer targets.
	MaxActionRange() float64
}

// Action is a settlement-level decision with preconditions, goal-aligned
// scoring, and an execution step that mutates state and emits an event.
type Action interface {
	Name() string
	Preconditions(self *world.Settlement, all []world.Settlement, env AgentEnv) bool
	Score(self *world.Settlement) float64
	Execute(self *world.Settlement, all *[]world.Settlement, env AgentEnv, rng *randv2.Rand) simulation.Event
}

// AllActions returns every action a settlement may consider, in a stable
// evaluation order.
func AllActions() []Action {
	return []Action{
		ExpandAction{},
		RaidAction{},
		ConquerAction{},
		FortifyAction{},
		AllyAction{},
		ProsperAction{},
	}
}

// findByName returns the settlement pointer inside all whose name matches,
// or nil when absent.
func findByName(all *[]world.Settlement, name string) *world.Settlement {
	for i := range *all {
		if (*all)[i].Name == name {
			return &(*all)[i]
		}
	}
	return nil
}

// hasGoal reports whether goal is in the settlement's goal list.
func hasGoal(self *world.Settlement, goal string) bool {
	for _, g := range self.Goals {
		if g == goal {
			return true
		}
	}
	return false
}

// withinRange reports whether other is inside the maximum action range.
func withinRange(self, other *world.Settlement, maxRange float64) bool {
	dx := float64(other.X - self.X)
	dy := float64(other.Y - self.Y)
	return dx*dx+dy*dy <= maxRange*maxRange
}
