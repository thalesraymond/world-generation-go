package figures

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

type Diplomat struct{}

func (d *Diplomat) Name() string { return "Diplomat" }

func (d *Diplomat) GenerateEvents(figure *HistoricalFigure, year int, settlementName string, settlementPop float64, graph *pointcrawl.Graph, settlementX, settlementY int, rng *randv2.Rand) []simulation.Event {
	if rng.IntN(4) != 0 {
		return nil
	}

	factionNames := []string{"Freehold", "Ironbound", "Sylvani", "Ashen"}
	otherFaction := factionNames[rng.IntN(len(factionNames))]

	success := figure.Stats.InfluenceOutcome("Politics", rng)
	var desc string
	var repDelta int
	if success {
		desc = fmt.Sprintf("%s negotiated a trade agreement with the %s, strengthening %s", figure.Name, otherFaction, settlementName)
		repDelta = 2
	} else {
		desc = fmt.Sprintf("%s failed to secure an alliance with the %s, returning to %s empty-handed", figure.Name, otherFaction, settlementName)
		repDelta = -1
	}

	figure.AddReputation(ReputationEntry{Year: year, Event: "Diplomacy", Delta: repDelta, Description: desc})

	return []simulation.Event{{
		Category: "Politics", Description: desc, FigureID: figure.ID,
		SettlementName: settlementName, TargetSettlement: otherFaction,
	}}
}

func (d *Diplomat) CanTransitionTo(other Role) bool { return other.Name() == "Leader" }
