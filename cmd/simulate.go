package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appconfig "github.com/thalesraymond/world-generation-go/config"
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

			cmd.Printf("Starting simulation for %d years with event density %q (status: queued).\n", cfg.Years, cfg.Events)
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
