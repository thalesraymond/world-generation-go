package cmd

import (
	"encoding/json"
	"fmt"
	randv2 "math/rand/v2"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appconfig "github.com/thalesraymond/world-generation-go/config"
	adapter "github.com/thalesraymond/world-generation-go/internal/adapter/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/agent"
	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	dompointcrawl "github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/settlement"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/state"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

const (
	// agentMaxActionRange is the maximum Euclidean distance (in tiles) for
	// Raid and Conquer targets.
	agentMaxActionRange = 20.0
	// agentExpandMaxRange is the search radius for expansion targets.
	agentExpandMaxRange = 20.0
	// agentExpandMinDistance is the minimum distance between a new
	// settlement and every existing settlement.
	agentExpandMinDistance = 3.0
	// expansionHeadroom is the spare settlement-slice capacity reserved
	// before simulation so expansion appends never reallocate mid-run.
	expansionHeadroom = 1024
)

// agentEnv adapts the live world state to the agent.AgentEnv interface so
// domain actions can query suitability, expansion sites, and names without
// importing adapter packages.
type agentEnv struct {
	worldState *world.State
	graph      *dompointcrawl.Graph
	all        *[]world.Settlement
	usedNames  map[string]bool
}

func (e *agentEnv) Suitability(x, y int) float64 {
	if e.worldState == nil {
		return 0
	}
	idx, ok := e.worldState.Index(x, y)
	if !ok {
		return 0
	}
	return e.worldState.Suitability[idx]
}

func (e *agentEnv) FindExpansionTarget(self *world.Settlement, rng *randv2.Rand) (int, int, bool) {
	if e.graph == nil || e.all == nil {
		return 0, 0, false
	}

	sites := make([]dompointcrawl.SettlementSite, 0, len(*e.all))
	for _, s := range *e.all {
		sites = append(sites, dompointcrawl.SettlementSite{
			Name:    s.Name,
			X:       s.X,
			Y:       s.Y,
			Faction: s.Faction,
		})
	}

	node := dompointcrawl.FindExpansionTarget(e.graph, self.X, self.Y, self.Faction, sites, agentExpandMaxRange, agentExpandMinDistance, rng)
	if node == nil {
		return 0, 0, false
	}
	return node.X, node.Y, true
}

func (e *agentEnv) GenerateName(rng *randv2.Rand) string {
	name := settlement.EnsureUniqueName(rng, e.usedNames)
	e.usedNames[name] = true
	return name
}

func (e *agentEnv) MaxActionRange() float64 {
	return agentMaxActionRange
}

type settlementEntity struct {
	settlement      *world.Settlement
	figureRNG       *randv2.Rand
	agentRNG        *randv2.Rand
	pointcrawlGraph *dompointcrawl.Graph
	allSettlements  *[]world.Settlement
	env             *agentEnv
}

func (s *settlementEntity) Tick(year int, eventChan chan<- domsim.Event, rng *randv2.Rand) {
	// 1. Age figures
	// Figures' Age is computed dynamically, no field to increment

	// 2. Check deaths
	deathEvents := figures.CheckDeaths(s.settlement.Figures, year, s.figureRNG)
	for _, e := range deathEvents {
		e.Year = year
		e.SettlementName = s.settlement.Name
		eventChan <- e
	}

	// 3. Check births
	newborn := figures.CheckBirths(s.settlement.Figures, s.settlement.Population, year, s.settlement.Name, s.figureRNG)
	if newborn != nil {
		s.settlement.Figures = append(s.settlement.Figures, *newborn)
		eventChan <- domsim.Event{
			Year:           year,
			Category:       "Birth",
			Description:    newborn.Name + " is born in " + s.settlement.Name,
			FigureID:       newborn.ID,
			SettlementName: s.settlement.Name,
		}
	}

	// 4. Check role vacancies
	roleEvents := figures.AssignRoles(s.settlement.Figures, s.pointcrawlGraph, s.settlement.X, s.settlement.Y, s.figureRNG)
	for _, e := range roleEvents {
		e.Year = year
		e.SettlementName = s.settlement.Name
		eventChan <- e
	}

	// 4.5 Check marriages
	marriageEvents := figures.CheckMarriages(s.settlement.Figures, s.settlement.Name, s.settlement.Faction, year, s.figureRNG)
	for _, e := range marriageEvents {
		e.Year = year
		e.SettlementName = s.settlement.Name
		eventChan <- e
	}

	// 5. Generate role events for figures with roles
	var generatedEvents []domsim.Event
	for i := range s.settlement.Figures {
		if !s.settlement.Figures[i].IsAlive() {
			continue
		}
		if s.settlement.Figures[i].Role == "" {
			continue
		}
		role, err := figures.NewRole(s.settlement.Figures[i].Role)
		if err != nil {
			continue
		}
		roleEvents := role.GenerateEvents(&s.settlement.Figures[i], s.settlement.Name, s.settlement.Population, s.pointcrawlGraph, s.settlement.X, s.settlement.Y, s.figureRNG)
		for j := range roleEvents {
			roleEvents[j].Year = year
			roleEvents[j].SettlementName = s.settlement.Name
		}
		generatedEvents = append(generatedEvents, roleEvents...)
	}

	// 5.5 Check role transitions driven by recent events
	transEvents := figures.CheckTransitions(s.settlement.Figures, generatedEvents, s.figureRNG)
	for _, e := range transEvents {
		e.Year = year
		e.SettlementName = s.settlement.Name
		eventChan <- e
	}

	for _, e := range generatedEvents {
		eventChan <- e
	}

	// 6. Agent decision loop: evaluate state, pick a goal-aligned action,
	// execute it, and emit the resulting event. Expand may append a new
	// settlement to allSettlements, affecting subsequent years.
	if s.allSettlements != nil && s.agentRNG != nil {
		action := agent.ChooseAction(s.settlement, *s.allSettlements, s.env, s.agentRNG)
		event := action.Execute(s.settlement, s.allSettlements, s.env, s.agentRNG)
		event.Year = year
		eventChan <- event
	}
}

func newSimulateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Run the world simulation phases",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := appconfig.FromViper()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			width := viper.GetInt("width")
			height := viper.GetInt("height")
			outputDir := viper.GetString("output")

			cmd.Printf("Generating world: %dx%d with seed %d ...\n", width, height, cfg.Seed)

			worldConfig := ucsim.WorldGenConfig{
				Seed:   cfg.Seed,
				Width:  width,
				Height: height,
				Years:  cfg.Years,
			}

			worldState, err := ucsim.GenerateWorld(worldConfig)
			if err != nil {
				return fmt.Errorf("generate world: %w", err)
			}

			cmd.Printf("World generated: %d x %d, %d settlements.\n", worldState.Width, worldState.Height, len(worldState.Settlements))

			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}

			cmd.Printf("Starting timeline simulation for %d years with event density %q.\n", cfg.Years, cfg.Events)

			engine := state.NewEngine(uint64(cfg.Seed))
			timelineRNG := engine.GetPRNG("timeline")

			env := &agentEnv{
				worldState: worldState,
				graph:      worldState.PointcrawlGraph,
				all:        &worldState.Settlements,
				usedNames:  make(map[string]bool),
			}
			for i := range worldState.Settlements {
				env.usedNames[worldState.Settlements[i].Name] = true
			}

			sim := domsim.New(1, cfg.Years, timelineRNG)
			entities := make([]*settlementEntity, 0, len(worldState.Settlements))
			for i := range worldState.Settlements {
				s := &worldState.Settlements[i]
				entities = append(entities, &settlementEntity{
					settlement:      s,
					figureRNG:       engine.GetPRNG("figures:" + s.Name),
					agentRNG:        engine.GetPRNG("agent:" + s.Name),
					pointcrawlGraph: worldState.PointcrawlGraph,
					allSettlements:  &worldState.Settlements,
					env:             env,
				})
			}
			for _, entity := range entities {
				sim.AddEntity(entity)
			}

			// Pre-size the settlements slice so expansion appends never
			// reallocate mid-simulation, then anchor entity pointers to the
			// final backing array.
			if extra := cap(worldState.Settlements) - len(worldState.Settlements); extra < expansionHeadroom {
				grown := make([]world.Settlement, len(worldState.Settlements), len(worldState.Settlements)+expansionHeadroom)
				copy(grown, worldState.Settlements)
				worldState.Settlements = grown
			}
			env.all = &worldState.Settlements
			for i, entity := range entities {
				entity.settlement = &worldState.Settlements[i]
				entity.allSettlements = &worldState.Settlements
			}

			eventChan := make(chan domsim.Event, 100)
			var events []domsim.Event
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for event := range eventChan {
					events = append(events, event)
				}
			}()
			sim.Run(eventChan)
			wg.Wait()

			stateJSON, err := json.Marshal(worldState)
			if err != nil {
				return fmt.Errorf("marshal world state: %w", err)
			}

			statePath := filepath.Join(outputDir, "world_state.json")
			if err := os.WriteFile(statePath, stateJSON, 0644); err != nil {
				return fmt.Errorf("write world state: %w", err)
			}
			cmd.Printf("World state saved to %s\n", statePath)

			timelineJSON, err := json.Marshal(events)
			if err != nil {
				return fmt.Errorf("marshal timeline: %w", err)
			}

			timelinePath := filepath.Join(outputDir, "timeline.json")
			if err := os.WriteFile(timelinePath, timelineJSON, 0644); err != nil {
				return fmt.Errorf("write timeline: %w", err)
			}
			cmd.Printf("Timeline saved to %s\n", timelinePath)

			narrativeRNG := engine.GetPRNG("narrative")
			chronicle, err := adapter.NewChronicleForWorld(narrativeRNG, worldState, cfg.Events)
			if err != nil {
				return fmt.Errorf("create chronicle: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "--- Chronicle ---")
			if err := chronicle.Stream(cmd.Context(), events, cfg.Events, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("render chronicle: %w", err)
			}

			cmd.Println("\nSimulation completed successfully.")
			return nil
		},
	}

	cmd.Flags().Int("years", 100, "Number of years to simulate")
	cmd.Flags().String("events", "normal", "Event density preset")
	cmd.Flags().Int("width", 64, "World map width")
	cmd.Flags().Int("height", 64, "World map height")

	viper.SetDefault("years", 100)
	viper.SetDefault("events", "normal")
	viper.SetDefault("width", 64)
	viper.SetDefault("height", 64)
	bindCommandFlag(cmd, "years")
	bindCommandFlag(cmd, "events")
	bindCommandFlag(cmd, "width")
	bindCommandFlag(cmd, "height")

	return cmd
}
