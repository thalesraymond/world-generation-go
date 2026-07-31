package figures

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

type General struct{}

func (g *General) Name() string { return "General" }

func (g *General) GenerateEvents(figure *HistoricalFigure, settlementName string, settlementPop float64, graph *pointcrawl.Graph, settlementX, settlementY int, rng *randv2.Rand) []simulation.Event {
	if rng.IntN(4) != 0 {
		return nil
	}

	targetNames := []string{"Blackdale", "Thornfield", "Ashgate", "Ironpeak"}
	target := targetNames[rng.IntN(len(targetNames))]

	success := figure.Stats.InfluenceOutcome("Conflict", rng)
	var desc string
	var repDelta int
	if success {
		desc = fmt.Sprintf("%s led a successful raid on %s, bringing glory to %s", figure.Name, target, settlementName)
		repDelta = 2
	} else {
		desc = fmt.Sprintf("%s led a failed assault on %s and retreated to %s", figure.Name, target, settlementName)
		repDelta = -1
	}

	figure.AddReputation(ReputationEntry{Year: 0, Event: "Raid", Delta: repDelta, Description: desc})

	return []simulation.Event{{
		Category: "Conflict", Description: desc, FigureID: figure.ID,
		SettlementName: settlementName, TargetSettlement: target,
	}}
}

func (g *General) CanTransitionTo(other Role) bool { return other.Name() == "Explorer" }
