package spatial

import "github.com/thalesraymond/world-generation-go/internal/domain/terrain"

// EvaluateTileSuitability scores one tile in the range [0,1].
func EvaluateTileSuitability(tile terrain.Tile, nearWater bool, elevationVariance float64) float64 {
	if tile.Biome == terrain.BiomeWater {
		return 0
	}

	waterScore := 0.2
	if nearWater {
		waterScore = 1
	}

	flatness := clamp01(1 - elevationVariance*3)
	biomeScore := biomeLivability(tile.Biome)
	heightPenalty := clamp01(1 - max(0, tile.Elevation-0.85)*6)

	return clamp01((0.4*waterScore + 0.3*flatness + 0.3*biomeScore) * heightPenalty)
}

// CalculateSuitabilityMap precomputes per-tile suitability for simulation.
func CalculateSuitabilityMap(terrainMap terrain.Map) []float64 {
	cellCount := terrainMap.Width * terrainMap.Height
	if cellCount <= 0 {
		return nil
	}

	scores := make([]float64, cellCount)
	for y := 0; y < terrainMap.Height; y++ {
		for x := 0; x < terrainMap.Width; x++ {
			idx := y*terrainMap.Width + x
			tile := terrainMap.Tiles[idx]
			scores[idx] = EvaluateTileSuitability(tile, hasNearbyWater(terrainMap, x, y, 2), localElevationVariance(terrainMap, x, y))
		}
	}

	return scores
}

// hasNearbyWater determines if there is any water tile within the given radius.
// Optimized to avoid function call overhead and redundant boundary checks.
func hasNearbyWater(terrainMap terrain.Map, x, y, radius int) bool {
	minY := y - radius
	if minY < 0 {
		minY = 0
	}
	maxY := y + radius
	if maxY >= terrainMap.Height {
		maxY = terrainMap.Height - 1
	}

	minX := x - radius
	if minX < 0 {
		minX = 0
	}
	maxX := x + radius
	if maxX >= terrainMap.Width {
		maxX = terrainMap.Width - 1
	}

	for dy := minY; dy <= maxY; dy++ {
		rowOffset := dy * terrainMap.Width
		for dx := minX; dx <= maxX; dx++ {
			if terrainMap.Tiles[rowOffset+dx].Biome == terrain.BiomeWater {
				return true
			}
		}
	}

	return false
}

// localElevationVariance calculates the difference between max and min elevation in the 3x3 area.
// Optimized to avoid function call overhead and redundant boundary checks.
func localElevationVariance(terrainMap terrain.Map, x, y int) float64 {
	minY := y - 1
	if minY < 0 {
		minY = 0
	}
	maxY := y + 1
	if maxY >= terrainMap.Height {
		maxY = terrainMap.Height - 1
	}

	minX := x - 1
	if minX < 0 {
		minX = 0
	}
	maxX := x + 1
	if maxX >= terrainMap.Width {
		maxX = terrainMap.Width - 1
	}

	minElevation := 1.0
	maxElevation := 0.0
	found := false

	for dy := minY; dy <= maxY; dy++ {
		rowOffset := dy * terrainMap.Width
		for dx := minX; dx <= maxX; dx++ {
			e := terrainMap.Tiles[rowOffset+dx].Elevation
			if e < minElevation {
				minElevation = e
			}
			if e > maxElevation {
				maxElevation = e
			}
			found = true
		}
	}

	if !found {
		return 1
	}

	return maxElevation - minElevation
}

func biomeLivability(biome terrain.BiomeType) float64 {
	switch biome {
	case terrain.BiomeGrassland:
		return 1
	case terrain.BiomeForest:
		return 0.85
	case terrain.BiomeTundra:
		return 0.25
	case terrain.BiomeDesert:
		return 0.1
	default:
		return 0
	}
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}

	if value > 1 {
		return 1
	}

	return value
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}

	return b
}
