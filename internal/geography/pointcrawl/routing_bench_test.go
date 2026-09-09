package pointcrawl_test

import (
	"testing"
	"math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
	geo_pointcrawl "github.com/thalesraymond/world-generation-go/internal/geography/pointcrawl"
)

func BenchmarkConnectNodes(b *testing.B) {
	// Setup graph
	graph := pointcrawl.NewGraph()
	for i := 0; i < 500; i++ {
		graph.AddNode(&pointcrawl.Node{
			ID: i,
			X: rand.IntN(200),
			Y: rand.IntN(200),
		})
	}

	// Setup terrain
	tmap := &terrain.Map{
		Width:  200,
		Height: 200,
		Tiles:  make([]terrain.Tile, 200*200),
	}
	for i := range tmap.Tiles {
		tmap.Tiles[i] = terrain.Tile{
			Biome:     terrain.BiomeGrassland,
			Elevation: rand.Float64(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		geo_pointcrawl.ConnectNodes(graph, tmap, 30.0)
		graph.Edges = make([]pointcrawl.Edge, 0) // Clear edges for next run
	}
}
