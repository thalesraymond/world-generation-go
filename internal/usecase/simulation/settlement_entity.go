package simulation

import (
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/agent"
	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	dompointcrawl "github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// ExpansionHeadroom is the spare settlement-slice capacity reserved before
// simulation so expansion appends never reallocate mid-run.
const ExpansionHeadroom = 1024

// SettlementEntity simulates one settlement per year: figure lifecycle,
// roles, marriages, and the agent decision loop.
type SettlementEntity struct {
	settlement      *world.Settlement
	figureRNG       *randv2.Rand
	agentRNG        *randv2.Rand
	pointcrawlGraph *dompointcrawl.Graph
	allSettlements  *[]world.Settlement
	env             agent.AgentEnv
}

// NewSettlementEntity creates an entity bound to a settlement. env is the
// AgentEnv adapter wiring the entity to the live world state.
func NewSettlementEntity(settlement *world.Settlement, figureRNG, agentRNG *randv2.Rand, pointcrawlGraph *dompointcrawl.Graph, allSettlements *[]world.Settlement, env agent.AgentEnv) *SettlementEntity {
	return &SettlementEntity{
		settlement:      settlement,
		figureRNG:       figureRNG,
		agentRNG:        agentRNG,
		pointcrawlGraph: pointcrawlGraph,
		allSettlements:  allSettlements,
		env:             env,
	}
}

// Tick advances the settlement one year and streams resulting events.
func (s *SettlementEntity) Tick(year int, eventChan chan<- domsim.Event, rng *randv2.Rand) {
	// 1. Age figures
	// Figures' Age is computed dynamically, no field to increment

	// 2. Check deaths
	deathEvents := figures.CheckDeaths(s.settlement.Figures, year, s.figureRNG)
	for _, e := range deathEvents {
		e.Year = year
		e.SettlementName = s.settlement.Name
		eventChan <- e
	}

	// 3. Check births
	newborn := figures.CheckBirths(s.settlement.Figures, s.settlement.Population, year, s.settlement.Name, s.figureRNG)
	if newborn != nil {
		s.settlement.Figures = append(s.settlement.Figures, *newborn)
		eventChan <- domsim.Event{
			Year:           year,
			Category:       "Birth",
			Description:    newborn.Name + " is born in " + s.settlement.Name,
			FigureID:       newborn.ID,
			SettlementName: s.settlement.Name,
		}
	}

	// 4. Check role vacancies
	roleEvents := figures.AssignRoles(s.settlement.Figures, s.pointcrawlGraph, s.settlement.X, s.settlement.Y, s.figureRNG)
	for _, e := range roleEvents {
		e.Year = year
		e.SettlementName = s.settlement.Name
		eventChan <- e
	}

	// 4.5 Check marriages
	marriageEvents := figures.CheckMarriages(s.settlement.Figures, s.settlement.Name, s.settlement.Faction, year, s.figureRNG)
	for _, e := range marriageEvents {
		e.Year = year
		e.SettlementName = s.settlement.Name
		eventChan <- e
	}

	// 5. Generate role events for figures with roles
	var generatedEvents []domsim.Event
	for i := range s.settlement.Figures {
		if !s.settlement.Figures[i].IsAlive() {
			continue
		}
		if s.settlement.Figures[i].Role == "" {
			continue
		}
		role, err := figures.NewRole(s.settlement.Figures[i].Role)
		if err != nil {
			continue
		}
		roleEvents := role.GenerateEvents(&s.settlement.Figures[i], s.settlement.Name, s.settlement.Population, s.pointcrawlGraph, s.settlement.X, s.settlement.Y, s.figureRNG)
		for j := range roleEvents {
			roleEvents[j].Year = year
			roleEvents[j].SettlementName = s.settlement.Name
		}
		generatedEvents = append(generatedEvents, roleEvents...)
	}

	// 5.5 Check role transitions driven by recent events
	transEvents := figures.CheckTransitions(s.settlement.Figures, generatedEvents, s.figureRNG)
	for _, e := range transEvents {
		e.Year = year
		e.SettlementName = s.settlement.Name
		eventChan <- e
	}

	for _, e := range generatedEvents {
		eventChan <- e
	}

	// 6. Agent decision loop: evaluate state, pick a goal-aligned action,
	// execute it, and emit the resulting event. Expand may append a new
	// settlement to allSettlements, affecting subsequent years.
	if s.allSettlements != nil && s.agentRNG != nil {
		action := agent.ChooseAction(s.settlement, *s.allSettlements, s.env, s.agentRNG)
		event := action.Execute(s.settlement, s.allSettlements, s.env, s.agentRNG)
		event.Year = year
		eventChan <- event
	}
}
