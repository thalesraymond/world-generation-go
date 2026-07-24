package terrain

const (
	DefaultWaterThreshold   = 0.45
	DefaultElevationCooling = 0.35
)

// BiomeType describes the terrain classification for a tile.
type BiomeType string

const (
	BiomeWater     BiomeType = "water"
	BiomeDesert    BiomeType = "desert"
	BiomeTundra    BiomeType = "tundra"
	BiomeForest    BiomeType = "forest"
	BiomeGrassland BiomeType = "grassland"
)

// Tile stores the generated environmental values for a map coordinate.
type Tile struct {
	Elevation   float64
	Temperature float64
	Humidity    float64
	Biome       BiomeType
}

// Map stores a 2D terrain grid in row-major order.
type Map struct {
	Width  int
	Height int
	Tiles  []Tile
}

// TileAt returns the tile at the requested coordinate when it is in bounds.
func (m Map) TileAt(x, y int) (Tile, bool) {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return Tile{}, false
	}

	return m.Tiles[y*m.Width+x], true
}
