package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appconfig "github.com/thalesraymond/world-generation-go/config"
)

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export generated world data",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := appconfig.FromViper()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			cmd.Printf("Export acknowledged for format %q to %q.\n", cfg.Format, cfg.Output)
			return nil
		},
	}

	cmd.Flags().String("format", "obsidian", "Export format")

	viper.SetDefault("format", "obsidian")
	bindCommandFlag(cmd, "format")

	return cmd
}
