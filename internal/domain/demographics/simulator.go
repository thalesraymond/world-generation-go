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

	width := state.Width
	height := state.Height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx, _ := state.Index(x, y)
			population := state.PopulationDensity[idx]
			if population <= 0 {
				continue
			}

			transfer := population * rate
			if transfer == 0 {
				continue
			}

			// OPTIMIZATION: Use stack-allocated array instead of dynamic slice to avoid heap allocations in tight loop
			type weightedNeighbor struct {
				idx    int
				weight float64
			}

			var targets [8]weightedNeighbor
			targetsCount := 0

			totalWeight := 0.0

			// OPTIMIZATION: Inline neighbor calculation to avoid allocations and function call overhead
			var nIdxs [8]int
			nCount := 0

			if y > 0 {
				if x > 0 {
					nIdxs[nCount] = (y-1)*width + (x - 1)
					nCount++
				}
				nIdxs[nCount] = (y-1)*width + x
				nCount++
				if x < width-1 {
					nIdxs[nCount] = (y-1)*width + (x + 1)
					nCount++
				}
			}
			if x > 0 {
				nIdxs[nCount] = y*width + (x - 1)
				nCount++
			}
			if x < width-1 {
				nIdxs[nCount] = y*width + (x + 1)
				nCount++
			}
			if y < height-1 {
				if x > 0 {
					nIdxs[nCount] = (y+1)*width + (x - 1)
					nCount++
				}
				nIdxs[nCount] = (y+1)*width + x
				nCount++
				if x < width-1 {
					nIdxs[nCount] = (y+1)*width + (x + 1)
					nCount++
				}
			}

			for i := 0; i < nCount; i++ {
				nIdx := nIdxs[i]
				if state.PopulationDensity[nIdx] >= population {
					continue
				}

				weight := state.Suitability[nIdx]
				if weight <= 0 {
					continue
				}

				targets[targetsCount] = weightedNeighbor{idx: nIdx, weight: weight}
				targetsCount++
				totalWeight += weight
			}

			if totalWeight == 0 {
				next[idx] += transfer
				continue
			}

			for i := 0; i < targetsCount; i++ {
				target := targets[i]
				next[target.idx] += transfer * (target.weight / totalWeight)
			}
		}
	}

	return next
}

func spreadFactionInfluence(state *world.State, nextPopulation []float64, minPopulation float64) []string {
	next := make([]string, len(state.FactionInfluence))

	width := state.Width
	height := state.Height

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx, _ := state.Index(x, y)
			if nextPopulation[idx] < minPopulation {
				next[idx] = ""
				continue
			}

			// OPTIMIZATION: Use stack-allocated array and linear search instead of map[string]float64
			// to avoid map allocation and hashing overhead per cell. Linear search on max 8 elements is faster.
			type factionScore struct {
				faction string
				score   float64
			}
			var scores [8]factionScore
			scoresCount := 0

			// OPTIMIZATION: Inline neighbor calculation
			var nIdxs [8]int
			nCount := 0

			if y > 0 {
				if x > 0 {
					nIdxs[nCount] = (y-1)*width + (x - 1)
					nCount++
				}
				nIdxs[nCount] = (y-1)*width + x
				nCount++
				if x < width-1 {
					nIdxs[nCount] = (y-1)*width + (x + 1)
					nCount++
				}
			}
			if x > 0 {
				nIdxs[nCount] = y*width + (x - 1)
				nCount++
			}
			if x < width-1 {
				nIdxs[nCount] = y*width + (x + 1)
				nCount++
			}
			if y < height-1 {
				if x > 0 {
					nIdxs[nCount] = (y+1)*width + (x - 1)
					nCount++
				}
				nIdxs[nCount] = (y+1)*width + x
				nCount++
				if x < width-1 {
					nIdxs[nCount] = (y+1)*width + (x + 1)
					nCount++
				}
			}

			for i := 0; i < nCount; i++ {
				nIdx := nIdxs[i]
				faction := state.FactionInfluence[nIdx]
				if faction == "" {
					continue
				}

				found := false
				for j := 0; j < scoresCount; j++ {
					if scores[j].faction == faction {
						scores[j].score += state.PopulationDensity[nIdx]
						found = true
						break
					}
				}
				if !found {
					scores[scoresCount] = factionScore{faction: faction, score: state.PopulationDensity[nIdx]}
					scoresCount++
				}
			}

			bestFaction := state.FactionInfluence[idx]
			bestScore := 0.0
			for j := 0; j < scoresCount; j++ {
				if scores[j].score > bestScore {
					bestScore = scores[j].score
					bestFaction = scores[j].faction
				}
			}

			next[idx] = bestFaction
		}
	}

	return next
}
