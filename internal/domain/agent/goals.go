package agent

import (
	randv2 "math/rand/v2"
	"sort"
)

// GoalPool enumerates every goal a settlement may pursue.
var GoalPool = []string{"grow", "defend", "expand"}

// RandomGoals selects 2–3 unique goals from the goal pool. The result is
// deterministic for a given RNG stream.
func RandomGoals(rng *randv2.Rand) []string {
	count := 2 + rng.IntN(2) // 2 or 3 goals

	pool := append([]string(nil), GoalPool...)
	selected := make([]string, 0, count)
	for len(selected) < count && len(pool) > 0 {
		idx := rng.IntN(len(pool))
		selected = append(selected, pool[idx])
		pool = append(pool[:idx], pool[idx+1:]...)
	}

	sort.Strings(selected)
	return selected
}
