package simulation

import (
	"context"
	"fmt"
	"sync"

	"github.com/thalesraymond/world-generation-go/internal/domain/artifact"
	"github.com/thalesraymond/world-generation-go/internal/domain/settlement"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/state"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// OrchestratorConfig holds the simulation parameters for RunSimulation.
type OrchestratorConfig struct {
	Seed   uint64
	Width  int
	Height int
	Years  int
}

// RunSimulation generates the world and runs the full timeline simulation,
// returning the raw event stream and the final world state. All RNG lanes
// (world generation, timeline, per-settlement figures and agents, artifacts)
// are derived from the master seed, so identical configs produce identical
// events and state. The call is synchronous; ctx is honored only as a
// cancellation escape hatch for the event collector.
func RunSimulation(ctx context.Context, config OrchestratorConfig) ([]domsim.Event, *world.State, error) {
	worldState, err := GenerateWorld(WorldGenConfig{
		Seed:   int64(config.Seed),
		Width:  config.Width,
		Height: config.Height,
		Years:  config.Years,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("generate world: %w", err)
	}

	engine := state.NewEngine(config.Seed)
	timelineRNG := engine.GetPRNG("timeline")

	usedNames := make(map[string]bool)
	for i := range worldState.Settlements {
		usedNames[worldState.Settlements[i].Name] = true
	}
	env := NewAgentEnv(worldState, worldState.PointcrawlGraph, &worldState.Settlements, usedNames)

	// Pre-size the settlements slice so expansion appends never reallocate
	// mid-simulation; entities below anchor to the final backing array.
	if extra := cap(worldState.Settlements) - len(worldState.Settlements); extra < ExpansionHeadroom {
		grown := make([]world.Settlement, len(worldState.Settlements), len(worldState.Settlements)+ExpansionHeadroom)
		copy(grown, worldState.Settlements)
		worldState.Settlements = grown
	}

	sim := domsim.New(1, config.Years, timelineRNG)
	entities := make([]*SettlementEntity, 0, len(worldState.Settlements))
	for i := range worldState.Settlements {
		s := &worldState.Settlements[i]
		entities = append(entities, NewSettlementEntity(
			s,
			engine.GetPRNG("figures:"+s.Name),
			engine.GetPRNG("agent:"+s.Name),
			worldState.PointcrawlGraph,
			&worldState.Settlements,
			env,
		))
	}
	for _, entity := range entities {
		sim.AddEntity(entity)
	}

	eventChan := make(chan domsim.Event, 100)
	var events []domsim.Event
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				// Cancellation stops recording, but the channel must still
				// be drained until sim.Run closes it so the synchronous
				// simulation never blocks on a full buffer.
				for range eventChan {
				}
				return
			case event, ok := <-eventChan:
				if !ok {
					return
				}
				events = append(events, event)
			}
		}
	}()
	sim.Run(eventChan)
	wg.Wait()

	artifactRNG := engine.GetPRNG("artifacts")
	worldState.Artifacts, err = artifact.EmergencePass(worldState.Artifacts, events, buildFigureContexts(worldState.Settlements), buildSignificanceContext(worldState), artifactRNG)
	if err != nil {
		return nil, nil, fmt.Errorf("post-process artifact state: %w", err)
	}

	return events, worldState, nil
}

// buildFigureContexts summarizes the figures the emergence fallback needs:
// home settlement (the artifact ID origin) and reputation history (the
// threshold-crossing year).
func buildFigureContexts(settlements []world.Settlement) []artifact.FigureContext {
	ctx := make([]artifact.FigureContext, 0, len(settlements)*4)
	for i := range settlements {
		for j := range settlements[i].Figures {
			f := &settlements[i].Figures[j]
			rep := make([]artifact.ReputationDelta, 0, len(f.Reputation))
			for _, e := range f.Reputation {
				rep = append(rep, artifact.ReputationDelta{Year: e.Year, Delta: e.Delta, Event: e.Event})
			}
			ctx = append(ctx, artifact.FigureContext{ID: f.ID, Settlement: settlements[i].Name, Reputation: rep})
		}
	}
	return ctx
}

// buildSignificanceContext derives the artifact significance inputs from the
// world state: per-year figure reputation deltas and settlement size classes
// (spec 4.4). The world state records no historical population, so the size
// class is classified from the settlement's recorded population at pass time;
// the lump sum is still awarded exactly once, at the acquisition year.
func buildSignificanceContext(state *world.State) artifact.SignificanceContext {
	sig := artifact.SignificanceContext{
		FigureReputation: make(map[string]map[int]int),
		SettlementClass:  make(map[string]string),
	}
	for i := range state.Settlements {
		s := &state.Settlements[i]
		sig.SettlementClass[s.Name] = settlement.Classify(s.Population)
		for j := range s.Figures {
			f := &s.Figures[j]
			byYear := make(map[int]int)
			for _, e := range f.Reputation {
				byYear[e.Year] += e.Delta
			}
			sig.FigureReputation[f.ID] = byYear
		}
	}
	return sig
}
