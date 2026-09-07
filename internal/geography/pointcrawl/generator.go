package pointcrawl

import (
	"fmt"
	"math"
	randv2 "math/rand/v2"
	"sort"

	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// GeneratorConfig controls how POIs are extracted, sampled, and culled.
type GeneratorConfig struct {
	MinDistance        float64
	SampleStep         int
	MaxWildernessNodes int
	RNG                *randv2.Rand
}

// DefaultGeneratorConfig returns baseline pointcrawl generation rules.
func DefaultGeneratorConfig() GeneratorConfig {
	return GeneratorConfig{
		MinDistance:        5.0,
		SampleStep:         8,
		MaxWildernessNodes: 20,
	}
}

// Generate extracts and samples POIs into a pointcrawl Graph.
func Generate(state *world.State, terrainMap *terrain.Map, config GeneratorConfig) (*Graph, error) {
	if state == nil {
		return nil, fmt.Errorf("state is required")
	}

	if state.Width <= 0 || state.Height <= 0 {
		return nil, fmt.Errorf("state must have positive dimensions")
	}

	if terrainMap == nil {
		return nil, fmt.Errorf("terrain map is required")
	}

	if terrainMap.Width != state.Width || terrainMap.Height != state.Height {
		return nil, fmt.Errorf("terrain map dimensions %dx%d do not match state dimensions %dx%d",
			terrainMap.Width, terrainMap.Height, state.Width, state.Height)
	}

	nodes := make([]Node, 0)

	// 2.1 Settlement POIs are always Known.
	for i := range state.Settlements {
		s := &state.Settlements[i]
		nodes = append(nodes, Node{
			ID:         len(nodes),
			X:          s.X,
			Y:          s.Y,
			Visibility: Known,
			Name:       s.Name,
			Kind:       "settlement",
		})
	}

	terrainNodes := sampleTerrainNodes(terrainMap, config)
	for i := range terrainNodes {
		terrainNodes[i].ID = len(nodes)
		nodes = append(nodes, terrainNodes[i])
	}

	// 2.2 Cull overlapping nodes, preferring higher visibility.
	culled := cullNodes(nodes, config.MinDistance)

	graph := NewGraph()
	for i := range culled {
		graph.AddNode(&culled[i])
	}

	return graph, nil
}

func sampleTerrainNodes(terrainMap *terrain.Map, config GeneratorConfig) []Node {
	if config.SampleStep <= 0 {
		return nil
	}

	var rng *randv2.Rand
	if config.RNG == nil {
		rng = randv2.New(randv2.NewPCG(0, 0))
	} else {
		rng = config.RNG
	}

	candidates := make([]Node, 0)

	for y := 0; y < terrainMap.Height; y += config.SampleStep {
		for x := 0; x < terrainMap.Width; x += config.SampleStep {
			tile, ok := terrainMap.TileAt(x, y)
			if !ok {
				continue
			}

			switch {
			case tile.Elevation > 0.85 && tile.Biome != terrain.BiomeWater:
				candidates = append(candidates, Node{
					X:          x,
					Y:          y,
					Visibility: Unknown,
					Name:       fmt.Sprintf("Landmark-%d-%d", x, y),
					Kind:       "landmark",
				})
			case tile.Biome == terrain.BiomeForest:
				candidates = append(candidates, Node{
					X:          x,
					Y:          y,
					Visibility: Unknown,
					Name:       fmt.Sprintf("Wilderness-%d-%d", x, y),
					Kind:       "wilderness",
				})
			}
		}
	}

	// Randomly place hidden ruins in remote areas.
	maxRuins := min(5, config.MaxWildernessNodes/4+1)
	for i := 0; i < maxRuins; i++ {
		rx := rng.IntN(terrainMap.Width)
		ry := rng.IntN(terrainMap.Height)
		tile, ok := terrainMap.TileAt(rx, ry)
		if !ok || tile.Biome == terrain.BiomeWater {
			continue
		}

		candidates = append(candidates, Node{
			X:          rx,
			Y:          ry,
			Visibility: Hidden,
			Name:       fmt.Sprintf("Ruin-%d-%d", rx, ry),
			Kind:       "ruin",
		})
	}

	// Shuffle and limit wilderness nodes to keep the graph readable.
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	limit := config.MaxWildernessNodes
	if limit <= 0 {
		limit = len(candidates)
	}

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	return candidates
}

func cullNodes(nodes []Node, minDistance float64) []Node {
	if minDistance <= 0 {
		return nodes
	}

	// Sort by visibility priority so higher visibility nodes are kept.
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Visibility < nodes[j].Visibility
	})

	kept := make([]Node, 0, len(nodes))
	for _, candidate := range nodes {
		tooClose := false
		for _, existing := range kept {
			if distance(candidate.X, candidate.Y, existing.X, existing.Y) < minDistance {
				tooClose = true
				break
			}
		}
		if !tooClose {
			kept = append(kept, candidate)
		}
	}

	return kept
}

func distance(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Hypot(dx, dy)
}
