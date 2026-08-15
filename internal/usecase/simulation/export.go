package simulation

import (
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// WorldExporter writes the world state and timeline events to a target
// directory in a presentation format. It is declared in the usecase layer
// and implemented by infra (internal/infra/exporter) with composition
// wiring in the adapter layer.
type WorldExporter interface {
	Export(state *world.State, events []domsim.Event, targetDir string) error
}
