package demographics

import (
	"fmt"

	"github.com/thalesraymond/world-generation-go/internal/domain/spatial"
	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// SimulatorConfig controls diffusion and simulation steps.
type SimulatorConfig struct {
	Iterations    int
	DiffusionRate float64
	MinPopulation float64
	FactionNames  []string
}

// DefaultConfig returns a deterministic baseline simulation config.
func DefaultConfig() SimulatorConfig {
	return SimulatorConfig{
		Iterations:    8,
		DiffusionRate: 0.3,
		MinPopulation: 0.05,
		FactionNames:  []string{"auric", "verdant", "cinder"},
	}
}

// PreGenerateSuitability computes and stores suitability in the world state.
func PreGenerateSuitability(state *world.State, terrainMap terrain.Map) error {
	if state == nil {
		return fmt.Errorf("state is required")
	}

	if state.Width != terrainMap.Width || state.Height != terrainMap.Height {
		return fmt.Errorf("state and terrain dimensions differ: state=%dx%d terrain=%dx%d", state.Width, state.Height, terrainMap.Width, terrainMap.Height)
	}

	return state.SetSuitability(spatial.CalculateSuitabilityMap(terrainMap))
}

// SeedPopulationFromSuitability creates deterministic starting populations.
func SeedPopulationFromSuitability(state *world.State, config SimulatorConfig) error {
	if state == nil {
		return fmt.Errorf("state is required")
	}

	if err := state.Validate(); err != nil {
		return err
	}

	factions := config.FactionNames
	if len(factions) == 0 {
		factions = []string{"independent"}
	}

	for y := 0; y < state.Height; y++ {
		for x := 0; x < state.Width; x++ {
			idx, _ := state.Index(x, y)
			suitability := state.Suitability[idx]
			state.PopulationDensity[idx] = suitability * suitability

			if state.PopulationDensity[idx] < config.MinPopulation {
				state.FactionInfluence[idx] = ""
				continue
			}

			factionIdx := (x + y) % len(factions)
			state.FactionInfluence[idx] = factions[factionIdx]
		}
	}

	return nil
}

// Simulate runs the configured number of automata iterations.
func Simulate(state *world.State, config SimulatorConfig) error {
	if state == nil {
		return fmt.Errorf("state is required")
	}

	if err := state.Validate(); err != nil {
		return err
	}

	if config.Iterations <= 0 {
		return nil
	}

	rate := config.DiffusionRate
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}

	for i := 0; i < config.Iterations; i++ {
		nextPopulation := diffusePopulation(state, rate)
		nextFaction := spreadFactionInfluence(state, nextPopulation, config.MinPopulation)

		copy(state.PopulationDensity, nextPopulation)
		copy(state.FactionInfluence, nextFaction)
	}

	return nil
}

func diffusePopulation(state *world.State, rate float64) []float64 {
	next := make([]float64, len(state.PopulationDensity))
	for idx, population := range state.PopulationDensity {
		next[idx] = population * (1 - rate)
	}

	for y := 0; y < state.Height; y++ {
		for x := 0; x < state.Width; x++ {
			idx, _ := state.Index(x, y)
			population := state.PopulationDensity[idx]
			if population <= 0 {
				continue
			}

			transfer := population * rate
			if transfer == 0 {
				continue
			}

			type weightedNeighbor struct {
				idx    int
				weight float64
			}

			var targets []weightedNeighbor
			totalWeight := 0.0
			for _, n := range neighbors(state, x, y) {
				if state.PopulationDensity[n.idx] >= population {
					continue
				}

				weight := state.Suitability[n.idx]
				if weight <= 0 {
					continue
				}

				targets = append(targets, weightedNeighbor{idx: n.idx, weight: weight})
				totalWeight += weight
			}

			if totalWeight == 0 {
				next[idx] += transfer
				continue
			}

			for _, target := range targets {
				next[target.idx] += transfer * (target.weight / totalWeight)
			}
		}
	}

	return next
}

func spreadFactionInfluence(state *world.State, nextPopulation []float64, minPopulation float64) []string {
	next := make([]string, len(state.FactionInfluence))

	for y := 0; y < state.Height; y++ {
		for x := 0; x < state.Width; x++ {
			idx, _ := state.Index(x, y)
			if nextPopulation[idx] < minPopulation {
				next[idx] = ""
				continue
			}

			scores := map[string]float64{}
			for _, neighbor := range neighbors(state, x, y) {
				faction := state.FactionInfluence[neighbor.idx]
				if faction == "" {
					continue
				}

				scores[faction] += state.PopulationDensity[neighbor.idx]
			}

			bestFaction := state.FactionInfluence[idx]
			bestScore := 0.0
			for faction, score := range scores {
				if score > bestScore {
					bestScore = score
					bestFaction = faction
				}
			}

			next[idx] = bestFaction
		}
	}

	return next
}

type neighborCell struct {
	idx int
}

func neighbors(state *world.State, x, y int) []neighborCell {
	cells := make([]neighborCell, 0, 8)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			idx, ok := state.Index(x+dx, y+dy)
			if !ok {
				continue
			}

			cells = append(cells, neighborCell{idx: idx})
		}
	}

	return cells
}
