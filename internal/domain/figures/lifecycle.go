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
		maxAge := 70 + rng.IntN(21) // 70-90
		name := GenerateName(rng)
		founders[i] = HistoricalFigure{
			ID:            fmt.Sprintf("%s-%d", settlementName, i),
			Name:          name,
			BirthYear:     birthYear,
			MaxAge:        maxAge,
			Faction:       faction,
			Relationships: Relationships{},
		}
		// First founder is always Leader; others have no role.
		role := ""
		if i == 0 {
			role = "Leader"
			founders[i].SetRole(&Leader{})
		}
		founders[i].Stats = GenerateStats(rng, role)
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

			// Heir-first succession when a Leader dies.
			if figures[i].Role == "Leader" || (figures[i].RoleRole != nil && figures[i].RoleRole.Name() == "Leader") {
				heir := GetHeir(figures, figures[i].ID)
				if heir != nil {
					heir.SetRole(&Leader{})
					s := heir.Stats
					s.Martial = clamp(s.Martial + 1)
					s.Diplomatic = clamp(s.Diplomatic + 1)
					s.Infamy = clamp(s.Infamy + 1)
					heir.Stats = s
					heir.ParentID = figures[i].ID
					events = append(events, simulation.Event{
						Year:           currentYear,
						Category:       "Succession",
						Description:    fmt.Sprintf("%s inherits leadership from %s", heir.Name, figures[i].Name),
						FigureID:       heir.ID,
						SettlementName: figures[i].Faction,
					})
				}
			}
		}
	}
	return events
}

// CheckBirths evaluates whether a new figure is born this year.
func CheckBirths(figures []HistoricalFigure, population float64, currentYear int, settlementName string, rng *randv2.Rand) *HistoricalFigure {
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
		ID:            fmt.Sprintf("%s-%d", settlementName, idx),
		Name:          GenerateName(rng),
		BirthYear:     currentYear,
		MaxAge:        maxAge,
		Relationships: Relationships{},
		Stats:         GenerateStats(rng, ""),
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
			successor.SetRole(&Leader{})
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

// CheckMarriages evaluates eligible figures and forms marriages within the same faction.
func CheckMarriages(figures []HistoricalFigure, settlementName, faction string, year int, rng *randv2.Rand) []simulation.Event {
	var events []simulation.Event
	for i := range figures {
		if !figures[i].IsAlive() {
			continue
		}
		age := figures[i].Age(year)
		if age < 20 || age > 25 {
			continue
		}
		if len(figures[i].Relationships.Spouse) > 0 {
			continue
		}

		for j := range figures {
			if i == j {
				continue
			}
			if !figures[j].IsAlive() {
				continue
			}
			ageJ := figures[j].Age(year)
			if ageJ < 20 || ageJ > 25 {
				continue
			}
			if len(figures[j].Relationships.Spouse) > 0 {
				continue
			}
			if figures[i].Faction != figures[j].Faction {
				continue
			}

			if rng.IntN(3) == 0 {
				event, ok := FormMarriage(&figures[i], &figures[j], year)
				if ok {
					events = append(events, event)
				}
			}
		}
	}
	return events
}

// CheckTransitions evaluates role transitions driven by recent events.
func CheckTransitions(figures []HistoricalFigure, events []simulation.Event, rng *randv2.Rand) []simulation.Event {
	var transitionEvents []simulation.Event
	for i := range figures {
		if !figures[i].IsAlive() {
			continue
		}
		role := figures[i].GetRole()
		if role == nil {
			continue
		}

		if role.Name() == "Explorer" {
			for _, e := range events {
				if e.Category == "Discovery" && e.FigureID == figures[i].ID && rng.IntN(3) == 0 {
					leaderRole, _ := NewRole("Leader")
					if role.CanTransitionTo(leaderRole) {
						from := role.Name()
						figures[i].SetRole(leaderRole)
						figures[i].TransitionHistory = append(figures[i].TransitionHistory, TransitionEntry{Year: 0, FromRole: from, ToRole: "Leader", Reason: "founded settlement"})
						transitionEvents = append(transitionEvents, simulation.Event{
							Category: "RoleTransition", Description: fmt.Sprintf("%s becomes Leader after founding a settlement", figures[i].Name),
							FigureID: figures[i].ID, SettlementName: figures[i].Faction,
						})
					}
					break
				}
			}
		}

		if role.Name() == "Leader" && rng.IntN(50) == 0 {
			explorerRole, _ := NewRole("Explorer")
			if role.CanTransitionTo(explorerRole) {
				from := role.Name()
				figures[i].SetRole(explorerRole)
				figures[i].TransitionHistory = append(figures[i].TransitionHistory, TransitionEntry{Year: 0, FromRole: from, ToRole: "Explorer", Reason: "exiled"})
				transitionEvents = append(transitionEvents, simulation.Event{
					Category: "RoleTransition", Description: fmt.Sprintf("%s becomes Explorer after exile", figures[i].Name),
					FigureID: figures[i].ID, SettlementName: figures[i].Faction,
				})
			}
		}

		if role.Name() == "General" {
			for _, e := range events {
				if e.Category == "Conflict" && e.FigureID == figures[i].ID && !figures[i].Stats.InfluenceOutcome("Conflict", rng) && rng.IntN(3) == 0 {
					explorerRole, _ := NewRole("Explorer")
					if role.CanTransitionTo(explorerRole) {
						from := role.Name()
						figures[i].SetRole(explorerRole)
						figures[i].TransitionHistory = append(figures[i].TransitionHistory, TransitionEntry{Year: 0, FromRole: from, ToRole: "Explorer", Reason: "defeat"})
						transitionEvents = append(transitionEvents, simulation.Event{
							Category: "RoleTransition", Description: fmt.Sprintf("%s becomes Explorer after defeat", figures[i].Name),
							FigureID: figures[i].ID, SettlementName: figures[i].Faction,
						})
					}
					break
				}
			}
		}
	}
	return transitionEvents
}
