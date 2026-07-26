package figures

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

const (
	maxFiguresPerSettlement = 15
	minFiguresPerSettlement = 10
	maxFounderAgeOffset     = 20
	adultAge                = 18
	eventRiskMinAge         = 30
)

func GenerateFounders(rng *randv2.Rand, settlementName, faction string, foundingYear int) []HistoricalFigure {
	count := 3 + rng.IntN(3) // 3, 4, or 5
	founders := make([]HistoricalFigure, count)
	for i := range count {
		birthYear := foundingYear - rng.IntN(maxFounderAgeOffset)
		if birthYear < 1 {
			birthYear = 1
		}
		maxAge := 70 + rng.IntN(21) // 70-90
		founders[i] = HistoricalFigure{
			ID:            fmt.Sprintf("%s-%d", settlementName, i),
			Name:          GenerateName(rng),
			BirthYear:     birthYear,
			MaxAge:        maxAge,
			Faction:       faction,
			Relationships: Relationships{},
		}
	}
	// First founder is always Leader
	founders[0].Role = "Leader"

	// Form spouse pairs among founders
	if len(founders) >= 2 {
		for i := 0; i < len(founders)-1; i += 2 {
			AddSpouse(&founders[i], &founders[i+1])
		}
	}
	return founders
}

// CheckDeaths processes death checks for all living figures in a settlement year.
func CheckDeaths(figures []HistoricalFigure, currentYear int, rng *randv2.Rand) []simulation.Event {
	var events []simulation.Event
	for i := range figures {
		if !figures[i].IsAlive() {
			continue
		}
		age := figures[i].Age(currentYear)
		died := false
		if age >= figures[i].MaxAge {
			died = true
		} else if age >= eventRiskMinAge && rng.IntN(100) < 2 {
			died = true
		}
		if died {
			figures[i].SetDeath(currentYear)
			who := figures[i].Name
			if figures[i].Role != "" {
				who += " (" + figures[i].Role + ")"
			}
			events = append(events, simulation.Event{
				Year:           currentYear,
				Category:       "Death",
				Description:    fmt.Sprintf("%s has died at age %d", who, age),
				FigureID:       figures[i].ID,
				SettlementName: figures[i].Faction,
			})
		}
	}
	return events
}

// CheckBirths evaluates whether a new figure is born this year.
func CheckBirths(figures []HistoricalFigure, population float64, currentYear int, rng *randv2.Rand) *HistoricalFigure {
	aliveCount := 0
	for _, f := range figures {
		if f.IsAlive() {
			aliveCount++
		}
	}
	if aliveCount >= maxFiguresPerSettlement {
		return nil
	}
	capFactor := 1.0 - float64(aliveCount)/float64(maxFiguresPerSettlement)
	birthProb := (population / 10000.0) * capFactor * 0.5
	if rng.Float64() >= birthProb {
		return nil
	}
	idx := len(figures)
	maxAge := 70 + rng.IntN(21)
	figure := &HistoricalFigure{
		ID:            fmt.Sprintf("born-%d", idx),
		Name:          GenerateName(rng),
		BirthYear:     currentYear,
		MaxAge:        maxAge,
		Relationships: Relationships{},
	}

	// Assign 1-2 parents from existing living adult figures
	var parents []int
	for i := range figures {
		if figures[i].IsAlive() && figures[i].Age(currentYear) >= adultAge {
			parents = append(parents, i)
		}
	}
	if len(parents) > 0 {
		numParents := 1 + rng.IntN(min(2, len(parents)))
		for j := 0; j < numParents && j < len(parents); j++ {
			k := j + rng.IntN(len(parents)-j)
			parents[j], parents[k] = parents[k], parents[j]
			AddParentChild(&figures[parents[j]], figure)
		}
	}
	return figure
}

// AssignRoles checks for role vacancies and assigns roles.
func AssignRoles(figures []HistoricalFigure, graph *pointcrawl.Graph, settlementX, settlementY int, rng *randv2.Rand) []simulation.Event {
	var events []simulation.Event
	hasLeader := false
	for _, f := range figures {
		if f.IsAlive() && f.Role == "Leader" {
			hasLeader = true
			break
		}
	}
	if !hasLeader {
		var successor *HistoricalFigure
		for i := range figures {
			if figures[i].IsAlive() && figures[i].Role == "" {
				successor = &figures[i]
				break
			}
		}
		if successor != nil {
			successor.Role = "Leader"
			events = append(events, simulation.Event{
				Year:        successor.BirthYear + adultAge,
				Category:    "Politics",
				Description: fmt.Sprintf("%s rises as the new leader", successor.Name),
				FigureID:    successor.ID,
			})
		}
	}
	return events
}
