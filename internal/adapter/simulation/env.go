package simulation

import (
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/agent"
	dompointcrawl "github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/settlement"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

const (
	// agentMaxActionRange is the maximum Euclidean distance (in tiles) for
	// Raid and Conquer targets.
	agentMaxActionRange = 20.0
	// agentExpandMaxRange is the search radius for expansion targets.
	agentExpandMaxRange = 20.0
	// agentExpandMinDistance is the minimum distance between a new
	// settlement and every existing settlement.
	agentExpandMinDistance = 3.0
)

// NewAgentEnv adapts the live world state to the agent.AgentEnv interface so
// domain actions can query suitability, expansion sites, and names without
// importing adapter packages.
func NewAgentEnv(worldState *world.State, graph *dompointcrawl.Graph, settlements *[]world.Settlement, usedNames map[string]bool) agent.AgentEnv {
	return &agentEnv{
		worldState: worldState,
		graph:      graph,
		all:        settlements,
		usedNames:  usedNames,
	}
}

// agentEnv is the concrete AgentEnv adapter backed by the live world state.
type agentEnv struct {
	worldState *world.State
	graph      *dompointcrawl.Graph
	all        *[]world.Settlement
	usedNames  map[string]bool
}

func (e *agentEnv) Suitability(x, y int) float64 {
	if e.worldState == nil {
		return 0
	}
	idx, ok := e.worldState.Index(x, y)
	if !ok {
		return 0
	}
	return e.worldState.Suitability[idx]
}

func (e *agentEnv) FindExpansionTarget(self *world.Settlement, rng *randv2.Rand) (int, int, bool) {
	if e.graph == nil || e.all == nil {
		return 0, 0, false
	}

	sites := make([]dompointcrawl.SettlementSite, 0, len(*e.all))
	for _, s := range *e.all {
		sites = append(sites, dompointcrawl.SettlementSite{
			Name:    s.Name,
			X:       s.X,
			Y:       s.Y,
			Faction: s.Faction,
		})
	}

	node := dompointcrawl.FindExpansionTarget(e.graph, self.X, self.Y, self.Faction, sites, agentExpandMaxRange, agentExpandMinDistance, rng)
	if node == nil {
		return 0, 0, false
	}
	return node.X, node.Y, true
}

func (e *agentEnv) GenerateName(rng *randv2.Rand) string {
	name := settlement.EnsureUniqueName(rng, e.usedNames)
	e.usedNames[name] = true
	return name
}

func (e *agentEnv) MaxActionRange() float64 {
	return agentMaxActionRange
}
