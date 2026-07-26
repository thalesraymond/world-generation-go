package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appconfig "github.com/thalesraymond/world-generation-go/config"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

type dummyEntity struct {
	name string
}

func (d dummyEntity) Tick(year int, eventChan chan<- domsim.Event) {
	if year == 1 || year%10 == 0 || year == 100 {
		eventChan <- domsim.Event{
			Year:        year,
			Category:    "Chronicle",
			Description: fmt.Sprintf("%s notes events in the realm during year %d.", d.name, year),
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

			cmd.Printf("Starting simulation for %d years with event density %q.\n", cfg.Years, cfg.Events)

			entities := []domsim.Entity{
				dummyEntity{name: "Realm of Eldoria"},
				dummyEntity{name: "The Iron Syndicate"},
			}

			if err := ucsim.RunSimulation(1, cfg.Years, entities, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("run simulation: %w", err)
			}

			cmd.Println("Simulation completed successfully.")
			return nil
		},
	}

	cmd.Flags().Int("years", 100, "Number of years to simulate")
	cmd.Flags().String("events", "normal", "Event density preset")

	viper.SetDefault("years", 100)
	viper.SetDefault("events", "normal")
	bindCommandFlag(cmd, "years")
	bindCommandFlag(cmd, "events")

	return cmd
}
