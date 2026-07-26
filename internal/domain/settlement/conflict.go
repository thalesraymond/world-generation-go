package settlement

import (
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func ResolveProximityConflicts(settlements []world.Settlement, mergeDistance float64) []world.Settlement {
	if len(settlements) <= 1 {
		return settlements
	}

	absorbed := make([]bool, len(settlements))

	for i := 0; i < len(settlements); i++ {
		if absorbed[i] {
			continue
		}
		for j := i + 1; j < len(settlements); j++ {
			if absorbed[j] {
				continue
			}
			d := distance(settlements[i].X, settlements[i].Y, settlements[j].X, settlements[j].Y)
			if d < mergeDistance {
				if settlements[i].Population >= settlements[j].Population {
					settlements[i].Population += settlements[j].Population
					settlements[i].Type = Classify(settlements[i].Population)
					absorbed[j] = true
				} else {
					settlements[j].Population += settlements[i].Population
					settlements[j].Type = Classify(settlements[j].Population)
					absorbed[i] = true
					break
				}
			}
		}
	}

	result := make([]world.Settlement, 0, len(settlements))
	for i, a := range absorbed {
		if !a {
			result = append(result, settlements[i])
		}
	}
	return result
}
