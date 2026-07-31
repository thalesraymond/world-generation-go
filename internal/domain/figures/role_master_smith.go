package figures

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

type MasterSmith struct{}

func (ms *MasterSmith) Name() string { return "Master Smith" }

func (ms *MasterSmith) GenerateEvents(figure *HistoricalFigure, settlementName string, settlementPop float64, graph *pointcrawl.Graph, settlementX, settlementY int, rng *randv2.Rand) []simulation.Event {
	if rng.IntN(5) != 0 {
		return nil
	}

	items := []string{"a new plow", "reinforced gates", "a ceremonial sword", "new armor", "a magnificent crown", "fine tools"}
	item := items[rng.IntN(len(items))]

	desc := fmt.Sprintf("The master smith %s of %s forged %s", figure.Name, settlementName, item)
	figure.AddReputation(ReputationEntry{Year: 0, Event: "Craftsmanship", Delta: 1, Description: desc})

	return []simulation.Event{{
		Category: "Settlement", Description: desc, FigureID: figure.ID,
		SettlementName: settlementName,
	}}
}

func (ms *MasterSmith) CanTransitionTo(other Role) bool { return false }
