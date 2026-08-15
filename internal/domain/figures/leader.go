package figures

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// Leader represents a figure governing a settlement.
type Leader struct{}

// Name returns the role identifier.
func (l *Leader) Name() string { return "Leader" }

// GenerateEvents produces a single political, settlement, or conflict event.
// Events are generated roughly 25% of the time.
func (l *Leader) GenerateEvents(figure *HistoricalFigure, year int, settlementName string, settlementPop float64, graph *pointcrawl.Graph, settlementX, settlementY int, rng *randv2.Rand) []simulation.Event {
	if rng.IntN(4) != 0 {
		return nil
	}

	categories := []string{"Politics", "Settlement", "Conflict"}
	category := categories[rng.IntN(len(categories))]

	var desc string
	switch category {
	case "Politics":
		actions := []string{"holds a tense council", "negotiates a treaty", "quells a rebellion", "addresses the populace"}
		desc = fmt.Sprintf("%s %s in %s", figure.Name, actions[rng.IntN(len(actions))], settlementName)
	case "Settlement":
		actions := []string{"expands the borders of", "establishes new trade routes through", "declares a festival in"}
		desc = fmt.Sprintf("%s %s %s", figure.Name, actions[rng.IntN(len(actions))], settlementName)
	case "Conflict":
		actions := []string{"leads a skirmish near", "fortifies the defenses of", "rallies the militia of"}
		desc = fmt.Sprintf("%s %s %s", figure.Name, actions[rng.IntN(len(actions))], settlementName)
	}

	figure.AddReputation(ReputationEntry{Year: year, Event: category, Delta: 1, Description: desc})

	return []simulation.Event{
		{
			Category:       category,
			Description:    desc,
			FigureID:       figure.ID,
			SettlementName: settlementName,
		},
	}
}

// CanTransitionTo reports which roles a Leader may become.
func (l *Leader) CanTransitionTo(other Role) bool {
	return other.Name() == "Explorer"
}
