package cmd

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appconfig "github.com/thalesraymond/world-generation-go/config"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/state"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

type settlementEntity struct {
	settlement world.Settlement
}

func (s settlementEntity) Tick(year int, eventChan chan<- domsim.Event, rng *randv2.Rand) {
	if rng.IntN(10) < 3 {
		eventChan <- domsim.Event{
			Year:        year,
			Category:    "Settlement",
			Description: fmt.Sprintf("%s prospers", s.settlement.Name),
		}
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

			engine := state.NewEngine(uint64(cfg.Seed))
			timelineRNG := engine.GetPRNG("timeline")

			entities := make([]domsim.Entity, 0, len(worldState.Settlements))
			for _, settlement := range worldState.Settlements {
				entities = append(entities, settlementEntity{settlement: settlement})
			}

			if err := ucsim.RunSimulation(1, cfg.Years, entities, cmd.OutOrStdout(), timelineRNG); err != nil {
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
