package demographics

import (
	"fmt"
	randv2 "math/rand/v2"

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
	RNG           *randv2.Rand
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

	nextPopulation := make([]float64, len(state.PopulationDensity))
	nextFaction := make([]string, len(state.FactionInfluence))

	for i := 0; i < config.Iterations; i++ {
		diffusePopulationInPlace(state, rate, nextPopulation)
		spreadFactionInfluenceInPlace(state, nextPopulation, config.MinPopulation, nextFaction)

		copy(state.PopulationDensity, nextPopulation)
		copy(state.FactionInfluence, nextFaction)
	}

	return nil
}

func diffusePopulationInPlace(state *world.State, rate float64, next []float64) {
	for idx, population := range state.PopulationDensity {
		next[idx] = population * (1 - rate)
	}

	type weightedNeighbor struct {
		idx    int
		weight float64
	}
	targets := make([]weightedNeighbor, 0, 8)

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

			targets = targets[:0]
			totalWeight := 0.0

			for dy := -1; dy <= 1; dy++ {
				ny := y + dy
				if ny < 0 || ny >= state.Height {
					continue
				}
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx := x + dx
					if nx < 0 || nx >= state.Width {
						continue
					}
					nIdx := ny*state.Width + nx

					if state.PopulationDensity[nIdx] >= population {
						continue
					}

					weight := state.Suitability[nIdx]
					if weight <= 0 {
						continue
					}

					targets = append(targets, weightedNeighbor{idx: nIdx, weight: weight})
					totalWeight += weight
				}
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
}

func spreadFactionInfluenceInPlace(state *world.State, nextPopulation []float64, minPopulation float64, next []string) {
	type factionScore struct {
		name  string
		score float64
	}
	var scores [8]factionScore

	for y := 0; y < state.Height; y++ {
		for x := 0; x < state.Width; x++ {
			idx, _ := state.Index(x, y)
			if nextPopulation[idx] < minPopulation {
				next[idx] = ""
				continue
			}

			scoreCount := 0

			for dy := -1; dy <= 1; dy++ {
				ny := y + dy
				if ny < 0 || ny >= state.Height {
					continue
				}
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx := x + dx
					if nx < 0 || nx >= state.Width {
						continue
					}
					nIdx := ny*state.Width + nx

					faction := state.FactionInfluence[nIdx]
					if faction == "" {
						continue
					}

					pop := state.PopulationDensity[nIdx]
					found := false
					for i := 0; i < scoreCount; i++ {
						if scores[i].name == faction {
							scores[i].score += pop
							found = true
							break
						}
					}
					if !found {
						scores[scoreCount] = factionScore{name: faction, score: pop}
						scoreCount++
					}
				}
			}

			bestFaction := state.FactionInfluence[idx]
			bestScore := 0.0
			for i := 0; i < scoreCount; i++ {
				if scores[i].score > bestScore {
					bestScore = scores[i].score
					bestFaction = scores[i].name
				}
			}

			next[idx] = bestFaction
		}
	}
}
