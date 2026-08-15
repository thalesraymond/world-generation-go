package figures

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// Role defines the behavior of a historical figure's position in society.
type Role interface {
	Name() string
	GenerateEvents(figure *HistoricalFigure, year int, settlementName string, settlementPop float64, graph *pointcrawl.Graph, settlementX, settlementY int, rng *randv2.Rand) []simulation.Event
	CanTransitionTo(other Role) bool
}

// NewRole returns the Role implementation registered for the given name.
func NewRole(name string) (Role, error) {
	switch name {
	case "Leader":
		return &Leader{}, nil
	case "Explorer":
		return &Explorer{}, nil
	case "General":
		return &General{}, nil
	case "Diplomat":
		return &Diplomat{}, nil
	case "Master Smith":
		return &MasterSmith{}, nil
	default:
		return nil, fmt.Errorf("unknown role: %q", name)
	}
}
