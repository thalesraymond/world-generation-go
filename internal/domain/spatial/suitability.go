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
			tile, ok := terrainMap.TileAt(x, y)
			if !ok {
				continue
			}

			scores[idx] = EvaluateTileSuitability(tile, hasNearbyWater(terrainMap, x, y, 2), localElevationVariance(terrainMap, x, y))
		}
	}

	return scores
}

func hasNearbyWater(terrainMap terrain.Map, x, y, radius int) bool {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			neighbor, ok := terrainMap.TileAt(x+dx, y+dy)
			if !ok {
				continue
			}

			if neighbor.Biome == terrain.BiomeWater {
				return true
			}
		}
	}

	return false
}

func localElevationVariance(terrainMap terrain.Map, x, y int) float64 {
	minElevation := 1.0
	maxElevation := 0.0
	found := false

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			neighbor, ok := terrainMap.TileAt(x+dx, y+dy)
			if !ok {
				continue
			}

			if neighbor.Elevation < minElevation {
				minElevation = neighbor.Elevation
			}

			if neighbor.Elevation > maxElevation {
				maxElevation = neighbor.Elevation
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
