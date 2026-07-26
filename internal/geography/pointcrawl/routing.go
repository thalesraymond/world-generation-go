package pointcrawl

import (
	"math"
	"sort"

	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
)

// FrictionTable maps each biome to its base travel friction in watches.
var FrictionTable = map[terrain.BiomeType]int{
	terrain.BiomeWater:     10,
	terrain.BiomeDesert:    3,
	terrain.BiomeTundra:    3,
	terrain.BiomeForest:    3,
	terrain.BiomeGrassland: 1,
}

// HighElevationFrictionBonus is added to friction for tiles above this elevation.
const HighElevationFrictionBonus = 3

// HighElevationThreshold is the elevation above which the bonus applies.
const HighElevationThreshold = 0.8

// DefaultMaxConnectionDistance is the default cutoff distance for connecting nodes.
const DefaultMaxConnectionDistance = 30.0

// SampleFriction returns the friction for a single tile, including elevation bonus.
func SampleFriction(x, y int, terrainMap *terrain.Map) int {
	if terrainMap == nil {
		return 0
	}

	tile, ok := terrainMap.TileAt(x, y)
	if !ok {
		return 0
	}

	friction, ok := FrictionTable[tile.Biome]
	if !ok {
		friction = 1
	}

	if tile.Elevation > HighElevationThreshold {
		friction += HighElevationFrictionBonus
	}

	return friction
}

// CalculateEdgeCost computes the travel cost in watches between two nodes.
func CalculateEdgeCost(from, to *Node, terrainMap *terrain.Map) int {
	if from == nil || to == nil || terrainMap == nil {
		return 0
	}

	dx := float64(to.X - from.X)
	dy := float64(to.Y - from.Y)
	distance := math.Hypot(dx, dy)

	if distance == 0 {
		return 1
	}

	samples := 0
	totalFriction := 0

	steps := int(math.Ceil(distance))
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(float64(from.X) + dx*t))
		y := int(math.Round(float64(from.Y) + dy*t))
		totalFriction += SampleFriction(x, y, terrainMap)
		samples++
	}

	if samples == 0 {
		return 1
	}

	avgFriction := float64(totalFriction) / float64(samples)
	cost := int(math.Ceil(distance * avgFriction))
	if cost < 1 {
		cost = 1
	}

	return cost
}

// ConnectNodes adds bidirectional edges between pairs of nodes within maxDistance.
func ConnectNodes(graph *Graph, terrainMap *terrain.Map, maxDistance float64) {
	if graph == nil || terrainMap == nil {
		return
	}

	nodes := make([]*Node, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes = append(nodes, node)
	}

	// Sort by ID so iteration order is stable and output is deterministic.
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			a := nodes[i]
			b := nodes[j]

			dx := float64(b.X - a.X)
			dy := float64(b.Y - a.Y)
			d := math.Hypot(dx, dy)

			if d < maxDistance {
				cost := CalculateEdgeCost(a, b, terrainMap)
				graph.AddEdge(a.ID, b.ID, cost)
				graph.AddEdge(b.ID, a.ID, cost)
			}
		}
	}
}
