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
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
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

			statePath := filepath.Join(cfg.Output, "world_state.json")
			stateData, err := os.ReadFile(statePath)
			if err != nil {
				return fmt.Errorf("read world state: %w", err)
			}

			state, err := world.FromJSON(stateData)
			if err != nil {
				return fmt.Errorf("parse world state: %w", err)
			}

			var events []simulation.Event
			timelinePath := filepath.Join(cfg.Output, "timeline.json")
			if timelineData, err := os.ReadFile(timelinePath); err == nil {
				if err := json.Unmarshal(timelineData, &events); err != nil {
					return fmt.Errorf("parse timeline: %w", err)
				}
			}

			exporter := adapter.ObsidianExporter{}
			if err := exporter.Export(state, events, cfg.Output); err != nil {
				return fmt.Errorf("export world: %w", err)
			}

			factionSet := make(map[string]struct{})
			for _, s := range state.Settlements {
				factionSet[s.Faction] = struct{}{}
			}

			nodeCount := 0
			if state.PointcrawlGraph != nil {
				nodeCount = state.PointcrawlGraph.NodeCount()
			}

			cmd.Printf("Export complete: %d settlements, %d factions, %d pointcrawl nodes, %d artifacts, %d timeline events.\n",
				len(state.Settlements),
				len(factionSet),
				nodeCount,
				len(state.Artifacts),
				len(events),
			)
			return nil
		},
	}

	cmd.Flags().String("format", "obsidian", "Export format")

	viper.SetDefault("format", "obsidian")
	bindCommandFlag(cmd, "format")

	return cmd
}
