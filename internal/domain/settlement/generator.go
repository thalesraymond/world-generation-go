package settlement

import (
	"fmt"
	"math"
	randv2 "math/rand/v2"
	"sort"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// Config controls settlement candidate filtering and spacing.
type Config struct {
	MinSuitability float64
	MinPopulation  float64
	MinDistance    float64
	MaxPopulation  float64
	MaxSettlements int
	RNG            *randv2.Rand
}

// DefaultConfig returns baseline settlement generation rules.
func DefaultConfig() Config {
	return Config{
		MinSuitability: 0.65,
		MinPopulation:  0.35,
		MinDistance:    3,
		MaxPopulation:  100000,
		MaxSettlements: 0,
	}
}

type candidate struct {
	index int
	x     int
	y     int
	score float64
}

// Generate identifies candidates and creates settlement objects.
func Generate(state *world.State, config Config) error {
	if state == nil {
		return fmt.Errorf("state is required")
	}

	if err := state.Validate(); err != nil {
		return err
	}

	candidates := findCandidates(state, config)
	selected := filterByDistance(candidates, config.MinDistance, config.MaxSettlements)

	settlements := make([]world.Settlement, 0, len(selected))
	for idx, c := range selected {
		faction := state.FactionInfluence[c.index]
		if faction == "" {
			faction = "independent"
		}

		settlements = append(settlements, world.Settlement{
			Name:       fmt.Sprintf("Settlement-%03d", idx+1),
			X:          c.x,
			Y:          c.y,
			Faction:    faction,
			Population: math.Round(state.PopulationDensity[c.index] * config.MaxPopulation),
		})
	}

	state.Settlements = settlements
	return nil
}

func findCandidates(state *world.State, config Config) []candidate {
	candidates := make([]candidate, 0)
	for y := 0; y < state.Height; y++ {
		for x := 0; x < state.Width; x++ {
			idx, _ := state.Index(x, y)
			if state.Suitability[idx] < config.MinSuitability {
				continue
			}

			if state.PopulationDensity[idx] < config.MinPopulation {
				continue
			}

			candidates = append(candidates, candidate{
				index: idx,
				x:     x,
				y:     y,
				score: state.Suitability[idx] * state.PopulationDensity[idx],
			})
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates
}

func filterByDistance(candidates []candidate, minDistance float64, maxSettlements int) []candidate {
	selected := make([]candidate, 0, len(candidates))
	for _, c := range candidates {
		if maxSettlements > 0 && len(selected) >= maxSettlements {
			break
		}

		tooClose := false
		for _, existing := range selected {
			if distance(c.x, c.y, existing.x, existing.y) < minDistance {
				tooClose = true
				break
			}
		}

		if tooClose {
			continue
		}

		selected = append(selected, c)
	}

	return selected
}

func distance(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Hypot(dx, dy)
}
