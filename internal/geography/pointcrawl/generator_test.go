package pointcrawl

import (
	randv2 "math/rand/v2"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func makeTerrainMap(width, height int) *terrain.Map {
	tiles := make([]terrain.Tile, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			tiles[idx] = terrain.Tile{
				Elevation:   0.3,
				Temperature: 0.5,
				Humidity:    0.5,
				Biome:       terrain.BiomeGrassland,
			}
		}
	}
	return &terrain.Map{Width: width, Height: height, Tiles: tiles}
}

func TestGenerate_ConvertsSettlementsToKnownNodes(t *testing.T) {
	state := world.NewState(32, 32)
	state.Settlements = []world.Settlement{
		{Name: "Aldburg", X: 10, Y: 10},
	}

	terrainMap := makeTerrainMap(32, 32)
	config := DefaultGeneratorConfig()

	graph, err := Generate(state, terrainMap, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	found := false
	for _, node := range graph.Nodes {
		if node.Kind == "settlement" && node.Name == "Aldburg" {
			found = true
			if node.Visibility != Known {
				t.Errorf("settlement visibility = %v, want Known", node.Visibility)
			}
		}
	}
	if !found {
		t.Errorf("expected a settlement node named Aldburg")
	}
}

func TestGenerate_GeneratesWildernessNodes(t *testing.T) {
	state := world.NewState(32, 32)
	terrainMap := makeTerrainMap(32, 32)
	for y := 0; y < terrainMap.Height; y++ {
		for x := 0; x < terrainMap.Width; x++ {
			terrainMap.Tiles[y*terrainMap.Width+x].Biome = terrain.BiomeForest
		}
	}

	config := DefaultGeneratorConfig()
	graph, err := Generate(state, terrainMap, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if graph.NodeCount() == 0 {
		t.Fatalf("expected wilderness nodes, got none")
	}

	foundWilderness := false
	for _, node := range graph.Nodes {
		if node.Kind == "wilderness" {
			foundWilderness = true
			break
		}
	}
	if !foundWilderness {
		t.Errorf("expected at least one wilderness node")
	}
}

func TestGenerate_SpatialCullingMergesCloseNodes(t *testing.T) {
	state := world.NewState(32, 32)
	state.Settlements = []world.Settlement{
		{Name: "Near1", X: 10, Y: 10},
		{Name: "Near2", X: 11, Y: 11},
	}

	terrainMap := makeTerrainMap(32, 32)
	config := DefaultGeneratorConfig()
	config.MinDistance = 10.0

	graph, err := Generate(state, terrainMap, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	settlementCount := 0
	for _, node := range graph.Nodes {
		if node.Kind == "settlement" {
			settlementCount++
		}
	}

	if settlementCount != 1 {
		t.Errorf("expected 1 settlement after culling, got %d", settlementCount)
	}
}

func TestGenerate_Determinism(t *testing.T) {
	state := world.NewState(32, 32)
	state.Settlements = []world.Settlement{
		{Name: "Aldburg", X: 10, Y: 10},
	}

	terrainMap := makeTerrainMap(32, 32)
	for i := range terrainMap.Tiles {
		terrainMap.Tiles[i].Biome = terrain.BiomeForest
		terrainMap.Tiles[i].Elevation = 0.6
	}

	seed := uint64(12345)
	rng := randv2.New(randv2.NewPCG(seed, seed))
	config := GeneratorConfig{
		MinDistance:        5.0,
		SampleStep:         8,
		MaxWildernessNodes: 20,
		RNG:                rng,
	}

	graphA, err := Generate(state, terrainMap, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	rng = randv2.New(randv2.NewPCG(seed, seed))
	config.RNG = rng
	graphB, err := Generate(state, terrainMap, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if graphA.NodeCount() != graphB.NodeCount() {
		t.Fatalf("node count mismatch: %d vs %d", graphA.NodeCount(), graphB.NodeCount())
	}

	for id, nodeA := range graphA.Nodes {
		nodeB, ok := graphB.Nodes[id]
		if !ok {
			t.Fatalf("graphB missing node %d", id)
		}
		if *nodeA != *nodeB {
			t.Errorf("node %d differs: %+v vs %+v", id, *nodeA, *nodeB)
		}
	}
}

func TestGenerate_EmptyStateProducesGraphWithOnlyWilderness(t *testing.T) {
	state := world.NewState(32, 32)
	terrainMap := makeTerrainMap(32, 32)
	for i := range terrainMap.Tiles {
		terrainMap.Tiles[i].Biome = terrain.BiomeForest
	}

	config := DefaultGeneratorConfig()
	config.MaxWildernessNodes = 5

	graph, err := Generate(state, terrainMap, config)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	for _, node := range graph.Nodes {
		if node.Kind == "settlement" {
			t.Errorf("expected no settlement nodes, got %+v", node)
		}
	}

	if graph.NodeCount() == 0 {
		t.Errorf("expected wilderness nodes in empty state graph")
	}
}

func TestGenerate_RejectsMismatchedDimensions(t *testing.T) {
	state := world.NewState(32, 32)
	terrainMap := makeTerrainMap(16, 16)

	_, err := Generate(state, terrainMap, DefaultGeneratorConfig())
	if err == nil {
		t.Fatalf("expected error for mismatched dimensions")
	}
}
