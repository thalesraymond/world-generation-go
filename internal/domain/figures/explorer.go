package figures

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// Explorer represents a figure who ventures beyond settlements.
type Explorer struct{}

// Name returns the role identifier.
func (e *Explorer) Name() string { return "Explorer" }

// GenerateEvents produces a single Discovery event, possibly tied to the pointcrawl graph.
// Events are generated roughly 20% of the time.
func (e *Explorer) GenerateEvents(figure *HistoricalFigure, settlementName string, settlementPop float64, graph *pointcrawl.Graph, settlementX, settlementY int, rng *randv2.Rand) []simulation.Event {
	if rng.IntN(5) != 0 {
		return nil
	}

	var desc string
	if graph != nil {
		nodes := graph.GetUndiscoveredNear(settlementX, settlementY, 100.0)
		if len(nodes) > 0 {
			node := nodes[rng.IntN(len(nodes))]
			desc = fmt.Sprintf("%s discovers %s near %s", figure.Name, node.Name, settlementName)
		} else {
			desc = fmt.Sprintf("%s ventures beyond %s and finds nothing of note", figure.Name, settlementName)
		}
	} else {
		actions := []string{
			"scouts the wilderness around",
			"maps uncharted lands beyond",
			"explores ancient ruins near",
		}
		desc = fmt.Sprintf("%s %s %s", figure.Name, actions[rng.IntN(len(actions))], settlementName)
	}

	figure.AddReputation(ReputationEntry{Year: 0, Event: "Discovery", Delta: 1, Description: desc})

	return []simulation.Event{
		{
			Category:       "Discovery",
			Description:    desc,
			FigureID:       figure.ID,
			SettlementName: settlementName,
		},
	}
}

// CanTransitionTo reports which roles an Explorer may become.
func (e *Explorer) CanTransitionTo(other Role) bool {
	return other.Name() == "Leader"
}
