package simulation

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	infranarrative "github.com/thalesraymond/world-generation-go/internal/infra/narrative"
	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

// NewChronicleForWorld constructs a usecase Chronicle wired to the default
// grammar provider and a figure resolver built from the given world state.
// The preset string is validated at construction so invalid presets fail
// early. Composition wiring between usecase interfaces and infra
// implementations lives in the adapter layer, keeping both cmd and usecase
// free of infra imports.
func NewChronicleForWorld(rng *randv2.Rand, state *world.State, preset string) (*ucsim.Chronicle, error) {
	if state == nil {
		return nil, fmt.Errorf("world state is nil")
	}
	if _, err := ucsim.ParsePreset(preset); err != nil {
		return nil, fmt.Errorf("parse event preset %q: %w", preset, err)
	}
	return ucsim.NewChronicle(rng, infranarrative.DefaultGrammarProvider{}, ucsim.NewWorldFigureResolver(state))
}
