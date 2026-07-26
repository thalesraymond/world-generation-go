package simulation

import (
	"fmt"

	"github.com/thalesraymond/world-generation-go/internal/domain/demographics"
	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	"github.com/thalesraymond/world-generation-go/internal/domain/settlement"
	"github.com/thalesraymond/world-generation-go/internal/domain/state"
	"github.com/thalesraymond/world-generation-go/internal/domain/terrain"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	"github.com/thalesraymond/world-generation-go/internal/geography/pointcrawl"
)

type WorldGenConfig struct {
	Seed   int64
	Width  int
	Height int
	Years  int
}

func GenerateWorld(config WorldGenConfig) (*world.State, error) {
	if config.Width <= 0 || config.Height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", config.Width, config.Height)
	}

	engine := state.NewEngine(uint64(config.Seed))

	terrainRNG := engine.GetPRNG("terrain")
	climateRNG := engine.GetPRNG("climate")
	demographicsRNG := engine.GetPRNG("demographics")
	settlementsRNG := engine.GetPRNG("settlements")
	pointcrawlRNG := engine.GetPRNG("pointcrawl")

	terrainConfig := terrain.DefaultGeneratorConfig(config.Width, config.Height)
	terrainConfig.TerrainRNG = terrainRNG
	terrainConfig.ClimateRNG = climateRNG
	terrainMap := terrain.GenerateMap(terrainConfig)

	worldState := world.NewState(config.Width, config.Height)

	if err := demographics.PreGenerateSuitability(worldState, terrainMap); err != nil {
		return nil, fmt.Errorf("pre-generate suitability: %w", err)
	}

	demoConfig := demographics.DefaultConfig()
	demoConfig.RNG = demographicsRNG
	if err := demographics.SeedPopulationFromSuitability(worldState, demoConfig); err != nil {
		return nil, fmt.Errorf("seed population: %w", err)
	}

	if err := demographics.Simulate(worldState, demoConfig); err != nil {
		return nil, fmt.Errorf("simulate demographics: %w", err)
	}

	settlementConfig := settlement.DefaultConfig()
	settlementConfig.RNG = settlementsRNG
	if err := settlement.Generate(worldState, settlementConfig); err != nil {
		return nil, fmt.Errorf("generate settlements: %w", err)
	}

	for i := range worldState.Settlements {
		s := &worldState.Settlements[i]
		figureRNG := engine.GetPRNG("figures:" + s.Name)
		s.Figures = figures.GenerateFounders(figureRNG, s.Name, s.Faction, 0)
	}

	pointcrawlConfig := pointcrawl.DefaultGeneratorConfig()
	pointcrawlConfig.RNG = pointcrawlRNG

	graph, err := pointcrawl.Generate(worldState, &terrainMap, pointcrawlConfig)
	if err != nil {
		return nil, fmt.Errorf("generate pointcrawl graph: %w", err)
	}

	pointcrawl.ConnectNodes(graph, &terrainMap, pointcrawl.DefaultMaxConnectionDistance)

	worldState.PointcrawlGraph = graph

	return worldState, nil
}
