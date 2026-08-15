package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appconfig "github.com/thalesraymond/world-generation-go/config"
	adapter "github.com/thalesraymond/world-generation-go/internal/adapter/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/artifact"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/state"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

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

			usedNames := make(map[string]bool)
			for i := range worldState.Settlements {
				usedNames[worldState.Settlements[i].Name] = true
			}
			env := adapter.NewAgentEnv(worldState, worldState.PointcrawlGraph, &worldState.Settlements, usedNames)

			// Pre-size the settlements slice so expansion appends never
			// reallocate mid-simulation; entities below anchor to the final
			// backing array.
			if extra := cap(worldState.Settlements) - len(worldState.Settlements); extra < ucsim.ExpansionHeadroom {
				grown := make([]world.Settlement, len(worldState.Settlements), len(worldState.Settlements)+ucsim.ExpansionHeadroom)
				copy(grown, worldState.Settlements)
				worldState.Settlements = grown
			}

			sim := domsim.New(1, cfg.Years, timelineRNG)
			entities := make([]*ucsim.SettlementEntity, 0, len(worldState.Settlements))
			for i := range worldState.Settlements {
				s := &worldState.Settlements[i]
				entities = append(entities, ucsim.NewSettlementEntity(
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
				for event := range eventChan {
					events = append(events, event)
				}
			}()
			sim.Run(eventChan)
			wg.Wait()

			if err := artifact.PostProcess(worldState.Artifacts, events); err != nil {
				return fmt.Errorf("post-process artifact state: %w", err)
			}

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
