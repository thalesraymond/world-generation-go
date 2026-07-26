package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	appconfig "github.com/thalesraymond/world-generation-go/config"
	"gopkg.in/yaml.v3"
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

			cfg.Width, cfg.Height = appconfig.ResolveSize(cfg.Size)

			data, err := yaml.Marshal(&cfg)
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			if err := os.WriteFile("worldgen.yaml", data, 0644); err != nil {
				return fmt.Errorf("write worldgen.yaml: %w", err)
			}

			cmd.Println("Project initialized: worldgen.yaml")
			return nil
		},
	}

	cmd.Flags().String("name", "", "World name")
	cmd.Flags().String("size", "medium", "World size preset")

	bindCommandFlag(cmd, "name")
	bindCommandFlag(cmd, "size")

	return cmd
}
