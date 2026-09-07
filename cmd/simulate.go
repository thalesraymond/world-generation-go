package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appconfig "github.com/thalesraymond/world-generation-go/config"
	adapter "github.com/thalesraymond/world-generation-go/internal/adapter/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/state"
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

			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}

			cmd.Printf("Starting timeline simulation for %d years with event density %q.\n", cfg.Years, cfg.Events)

			events, worldState, err := ucsim.RunSimulation(cmd.Context(), ucsim.OrchestratorConfig{
				Seed:   uint64(cfg.Seed),
				Width:  width,
				Height: height,
				Years:  cfg.Years,
			})
			if err != nil {
				return fmt.Errorf("run simulation: %w", err)
			}

			cmd.Printf("World generated: %d x %d, %d settlements.\n", worldState.Width, worldState.Height, len(worldState.Settlements))

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

			narrativeRNG := state.NewEngine(uint64(cfg.Seed)).GetPRNG("narrative")
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
