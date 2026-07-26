package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appconfig "github.com/thalesraymond/world-generation-go/config"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
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

			cmd.Printf("Starting timeline simulation for %d years with event density %q.\n", cfg.Years, cfg.Events)

			entities := []domsim.Entity{}

			if err := ucsim.RunSimulation(1, cfg.Years, entities, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("run simulation: %w", err)
			}

			cmd.Println("Simulation completed successfully.")
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
