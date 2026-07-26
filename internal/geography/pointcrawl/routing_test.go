package pointcrawl

import (
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
)

func TestFrictionTableValues(t *testing.T) {
	cases := []struct {
		biome    terrain.BiomeType
		expected int
	}{
		{terrain.BiomeWater, 10},
		{terrain.BiomeDesert, 3},
		{terrain.BiomeTundra, 3},
		{terrain.BiomeForest, 3},
		{terrain.BiomeGrassland, 1},
	}

	for _, tc := range cases {
		if got := FrictionTable[tc.biome]; got != tc.expected {
			t.Errorf("FrictionTable[%q] = %d, want %d", tc.biome, got, tc.expected)
		}
	}
}

func TestSampleFriction(t *testing.T) {
	m := &terrain.Map{
		Width:  3,
		Height: 3,
		Tiles: []terrain.Tile{
			{Elevation: 0.1, Biome: terrain.BiomeGrassland},
			{Elevation: 0.1, Biome: terrain.BiomeGrassland},
			{Elevation: 0.1, Biome: terrain.BiomeGrassland},
			{Elevation: 0.1, Biome: terrain.BiomeForest},
			{Elevation: 0.85, Biome: terrain.BiomeGrassland},
			{Elevation: 0.1, Biome: terrain.BiomeWater},
			{Elevation: 0.1, Biome: terrain.BiomeDesert},
			{Elevation: 0.1, Biome: terrain.BiomeTundra},
			{Elevation: 0.9, Biome: terrain.BiomeForest},
		},
	}

	tests := []struct {
		x, y int
		want int
	}{
		{0, 0, 1},
		{1, 1, 4},
		{2, 1, 10},
		{0, 2, 3},
		{1, 2, 3},
		{2, 2, 6},
	}

	for _, tc := range tests {
		got := SampleFriction(tc.x, tc.y, m)
		if got != tc.want {
			t.Errorf("SampleFriction(%d,%d) = %d, want %d", tc.x, tc.y, got, tc.want)
		}
	}
}

func TestCalculateEdgeCost_MinimumForAdjacentGrassland(t *testing.T) {
	m := &terrain.Map{
		Width:  3,
		Height: 1,
		Tiles: []terrain.Tile{
			{Elevation: 0.1, Biome: terrain.BiomeGrassland},
			{Elevation: 0.1, Biome: terrain.BiomeGrassland},
			{Elevation: 0.1, Biome: terrain.BiomeGrassland},
		},
	}

	a := &Node{X: 0, Y: 0}
	b := &Node{X: 1, Y: 0}

	cost := CalculateEdgeCost(a, b, m)
	if cost != 1 {
		t.Errorf("CalculateEdgeCost adjacent grassland = %d, want 1", cost)
	}
}

func TestCalculateEdgeCost_HigherForMountainTerrain(t *testing.T) {
	m := &terrain.Map{
		Width:  5,
		Height: 1,
		Tiles: []terrain.Tile{
			{Elevation: 0.1, Biome: terrain.BiomeGrassland},
			{Elevation: 0.9, Biome: terrain.BiomeGrassland},
			{Elevation: 0.9, Biome: terrain.BiomeGrassland},
			{Elevation: 0.9, Biome: terrain.BiomeGrassland},
			{Elevation: 0.1, Biome: terrain.BiomeGrassland},
		},
	}

	a := &Node{X: 0, Y: 0}
	b := &Node{X: 4, Y: 0}

	cost := CalculateEdgeCost(a, b, m)
	if cost <= 1 {
		t.Errorf("CalculateEdgeCost mountain path = %d, want > 1", cost)
	}
}

func TestConnectNodes_CreatesEdgesForNearbyNodes(t *testing.T) {
	m := &terrain.Map{
		Width:  10,
		Height: 10,
		Tiles:  make([]terrain.Tile, 100),
	}
	for i := range m.Tiles {
		m.Tiles[i] = terrain.Tile{Elevation: 0.1, Biome: terrain.BiomeGrassland}
	}

	g := NewGraph()
	g.AddNode(&Node{ID: 1, X: 0, Y: 0})
	g.AddNode(&Node{ID: 2, X: 5, Y: 0})
	g.AddNode(&Node{ID: 3, X: 9, Y: 0})

	ConnectNodes(g, m, 6.0)

	if g.EdgeCount() == 0 {
		t.Errorf("expected edges between nearby nodes, got none")
	}
}

func TestConnectNodes_DoesNotConnectDistantNodes(t *testing.T) {
	m := &terrain.Map{
		Width:  100,
		Height: 100,
		Tiles:  make([]terrain.Tile, 10000),
	}
	for i := range m.Tiles {
		m.Tiles[i] = terrain.Tile{Elevation: 0.1, Biome: terrain.BiomeGrassland}
	}

	g := NewGraph()
	g.AddNode(&Node{ID: 1, X: 0, Y: 0})
	g.AddNode(&Node{ID: 2, X: 50, Y: 0})

	ConnectNodes(g, m, 30.0)

	if g.EdgeCount() != 0 {
		t.Errorf("expected no edges, got %d", g.EdgeCount())
	}
}

func TestConnectNodes_Deterministic(t *testing.T) {
	m := &terrain.Map{
		Width:  10,
		Height: 10,
		Tiles:  make([]terrain.Tile, 100),
	}
	for i := range m.Tiles {
		m.Tiles[i] = terrain.Tile{Elevation: 0.1, Biome: terrain.BiomeGrassland}
	}

	makeGraph := func() *Graph {
		g := NewGraph()
		g.AddNode(&Node{ID: 1, X: 0, Y: 0})
		g.AddNode(&Node{ID: 2, X: 5, Y: 0})
		g.AddNode(&Node{ID: 3, X: 9, Y: 0})
		ConnectNodes(g, m, 6.0)
		return g
	}

	first := makeGraph()
	second := makeGraph()

	if first.EdgeCount() != second.EdgeCount() {
		t.Fatalf("edge count mismatch: %d vs %d", first.EdgeCount(), second.EdgeCount())
	}

	edgeSet := make(map[Edge]int)
	for _, e := range first.Edges {
		edgeSet[e]++
	}
	for _, e := range second.Edges {
		edgeSet[e]--
	}
	for e, count := range edgeSet {
		if count != 0 {
			t.Errorf("edge counts differ for %+v", e)
		}
	}
}

func TestDefaultGeneratorConfig(t *testing.T) {
	cfg := DefaultGeneratorConfig()
	if cfg.MinDistance <= 0 {
		t.Errorf("MinDistance should be positive, got %v", cfg.MinDistance)
	}
	if cfg.SampleStep <= 0 {
		t.Errorf("SampleStep should be positive, got %v", cfg.SampleStep)
	}
}
