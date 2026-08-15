package simulation

import (
	"fmt"

	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	"github.com/thalesraymond/world-generation-go/internal/infra/exporter"
	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

var _ ucsim.WorldExporter = ObsidianExporter{}

// ObsidianExporter writes the world to Obsidian-compatible markdown via the
// infra exporter. It is the composition wiring for the usecase WorldExporter
// interface and lives in the adapter layer so cmd never imports infra.
type ObsidianExporter struct{}

// Export writes all world note types (bases, factions, pointcrawl, figures,
// chronicle, artifacts) plus the artifacts index under targetDir.
func (ObsidianExporter) Export(state *world.State, events []domsim.Event, targetDir string) error {
	if err := exporter.Export(state, targetDir); err != nil {
		return fmt.Errorf("export world: %w", err)
	}
	if err := exporter.ExportPointcrawl(state, targetDir); err != nil {
		return fmt.Errorf("export pointcrawl: %w", err)
	}
	if err := exporter.ExportTimeline(state, events, targetDir); err != nil {
		return fmt.Errorf("export timeline: %w", err)
	}
	if err := exporter.ExportFigures(state, events, targetDir); err != nil {
		return fmt.Errorf("export figures: %w", err)
	}
	if err := exporter.ExportArtifacts(state, events, targetDir); err != nil {
		return fmt.Errorf("export artifacts: %w", err)
	}
	return nil
}
