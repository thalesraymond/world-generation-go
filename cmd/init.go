package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	appconfig "github.com/thalesraymond/world-generation-go/config"
)

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a world generation project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := appconfig.FromViper()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			cmd.Printf("Initialization acknowledged for world %q with size %q.\n", cfg.Name, cfg.Size)
			return nil
		},
	}

	cmd.Flags().String("name", "", "World name")
	cmd.Flags().String("size", "medium", "World size preset")

	bindCommandFlag(cmd, "name")
	bindCommandFlag(cmd, "size")

	return cmd
}
