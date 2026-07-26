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
	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	domnarrative "github.com/thalesraymond/world-generation-go/internal/domain/narrative"
	dompointcrawl "github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/state"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	infranarrative "github.com/thalesraymond/world-generation-go/internal/infra/narrative"
	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

type settlementEntity struct {
	settlement      *world.Settlement
	figureRNG       *randv2.Rand
	pointcrawlGraph *dompointcrawl.Graph
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
	newborn := figures.CheckBirths(s.settlement.Figures, s.settlement.Population, year, s.figureRNG)
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

	// 5. Generate role events for figures with roles
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
		for _, e := range roleEvents {
			e.Year = year
			eventChan <- e
		}
	}

	// 6. Core settlement events (keep the original random events for backward compat)
	switch rng.IntN(5) {
	case 0:
		eventChan <- domsim.Event{Year: year, Category: "Conflict", Description: fmt.Sprintf("%s faces raiders from the borderlands", s.settlement.Name), SettlementName: s.settlement.Name}
	case 1:
		eventChan <- domsim.Event{Year: year, Category: "Disaster", Description: fmt.Sprintf("%s is struck by a terrible calamity", s.settlement.Name), SettlementName: s.settlement.Name}
	case 2:
		eventChan <- domsim.Event{Year: year, Category: "Politics", Description: fmt.Sprintf("%s holds a tense council of nobles", s.settlement.Name), SettlementName: s.settlement.Name}
	case 3:
		eventChan <- domsim.Event{Year: year, Category: "Discovery", Description: fmt.Sprintf("%s uncovers ancient secrets nearby", s.settlement.Name), SettlementName: s.settlement.Name}
	case 4:
		eventChan <- domsim.Event{Year: year, Category: "Settlement", Description: fmt.Sprintf("%s prospers under wise leadership", s.settlement.Name), SettlementName: s.settlement.Name}
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

			stateJSON, err := json.Marshal(worldState)
			if err != nil {
				return fmt.Errorf("marshal world state: %w", err)
			}

			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}

			statePath := filepath.Join(outputDir, "world_state.json")
			if err := os.WriteFile(statePath, stateJSON, 0644); err != nil {
				return fmt.Errorf("write world state: %w", err)
			}
			cmd.Printf("World state saved to %s\n", statePath)

			cmd.Printf("Starting timeline simulation for %d years with event density %q.\n", cfg.Years, cfg.Events)

			engine := state.NewEngine(uint64(cfg.Seed))
			timelineRNG := engine.GetPRNG("timeline")

			sim := domsim.New(1, cfg.Years, timelineRNG)
			for i := range worldState.Settlements {
				s := &worldState.Settlements[i]
				figureRNG := engine.GetPRNG("figures:" + s.Name)
				sim.AddEntity(&settlementEntity{
					settlement:      s,
					figureRNG:       figureRNG,
					pointcrawlGraph: worldState.PointcrawlGraph,
				})
			}

			eventChan := make(chan domsim.Event, 100)
			var events []domsim.Event
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for event := range eventChan {
					events = append(events, event)
					formatted := domsim.FormatEvent(event)
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), formatted)
				}
			}()
			sim.Run(eventChan)
			wg.Wait()

			timelineJSON, err := json.Marshal(events)
			if err != nil {
				return fmt.Errorf("marshal timeline: %w", err)
			}

			timelinePath := filepath.Join(outputDir, "timeline.json")
			if err := os.WriteFile(timelinePath, timelineJSON, 0644); err != nil {
				return fmt.Errorf("write timeline: %w", err)
			}
			cmd.Printf("Timeline saved to %s\n", timelinePath)

			narrativeEngine, err := domnarrative.NewEngineFromString(infranarrative.DefaultGrammar)
			if err != nil {
				return fmt.Errorf("create narrative engine: %w", err)
			}
			narrativeRNG := engine.GetPRNG("narrative")

			// Build figure lookup for narrative variable injection.
			figureLookup := make(map[string]figures.HistoricalFigure)
			for i := range worldState.Settlements {
				for j := range worldState.Settlements[i].Figures {
					f := &worldState.Settlements[i].Figures[j]
					figureLookup[f.ID] = *f
				}
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "--- Chronicle ---")
			for _, event := range events {
				var extra map[string]string
				if event.FigureID != "" {
					if fig, ok := figureLookup[event.FigureID]; ok {
						extra = map[string]string{
							"FigureName":     fig.Name,
							"FigureRole":     fig.Role,
							"SettlementName": event.SettlementName,
						}
					}
				}
				text, err := narrativeEngine.Narrate(event, extra, narrativeRNG)
				if err != nil {
					text = event.Description
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), text)
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
