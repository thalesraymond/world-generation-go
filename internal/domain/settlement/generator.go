package settlement

import (
	"fmt"
	"math"
	randv2 "math/rand/v2"
	"sort"

	"github.com/thalesraymond/world-generation-go/internal/domain/agent"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// Agent state initialization defaults for newly generated settlements.
const (
	// MilitaryPopulationRatio derives initial military strength from population.
	MilitaryPopulationRatio = 0.1
	// InitialWealth is the default starting wealth for a new settlement.
	InitialWealth = 100.0
)

type Config struct {
	MinSuitability float64
	MinPopulation  float64
	MinDistance    float64
	MaxPopulation  float64
	MaxSettlements int
	MergeDistance  float64
	RNG            *randv2.Rand
}

func DefaultConfig() Config {
	return Config{
		MinSuitability: 0.65,
		MinPopulation:  0.35,
		MinDistance:    3,
		MaxPopulation:  100000,
		MaxSettlements: 0,
		MergeDistance:  0,
		RNG:            randv2.New(randv2.NewPCG(0, 0)),
	}
}

type candidate struct {
	index int
	x     int
	y     int
	score float64
}

func Generate(state *world.State, config Config) error {
	if state == nil {
		return fmt.Errorf("state is required")
	}

	if err := state.Validate(); err != nil {
		return err
	}

	candidates := findCandidates(state, config)
	selected := filterByDistance(candidates, config.MinDistance, config.MaxSettlements)

	usedNames := make(map[string]bool)
	settlements := make([]world.Settlement, 0, len(selected))
	for _, c := range selected {
		faction := state.FactionInfluence[c.index]
		if faction == "" {
			faction = "independent"
		}

		population := math.Round(state.PopulationDensity[c.index] * config.MaxPopulation)
		name := EnsureUniqueName(config.RNG, usedNames)
		usedNames[name] = true

		settlements = append(settlements, world.Settlement{
			Name:             name,
			Type:             Classify(population),
			X:                c.x,
			Y:                c.y,
			Faction:          faction,
			Population:       population,
			MilitaryStrength: population * MilitaryPopulationRatio,
			Wealth:           InitialWealth,
			Goals:            agent.RandomGoals(config.RNG),
		})
	}

	mergeDistance := config.MergeDistance
	if mergeDistance <= 0 {
		mergeDistance = config.MinDistance
	}
	settlements = ResolveProximityConflicts(settlements, mergeDistance)

	// Relations are initialized after the full settlement list is known so
	// every settlement sees every other settlement, including merged ones.
	for i := range settlements {
		settlements[i].Relations = world.InitRelations(settlements[i], settlements)
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
