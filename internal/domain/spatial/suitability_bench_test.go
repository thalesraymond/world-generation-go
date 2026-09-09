package spatial_test

import (
	"testing"
	"math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/spatial"
	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
)

func BenchmarkCalculateSuitabilityMap(b *testing.B) {
	tmap := terrain.Map{
		Width:  500,
		Height: 500,
		Tiles:  make([]terrain.Tile, 500*500),
	}

	for i := range tmap.Tiles {
		tmap.Tiles[i] = terrain.Tile{
			Biome:     terrain.BiomeGrassland,
			Elevation: rand.Float64(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spatial.CalculateSuitabilityMap(&tmap)
	}
}
