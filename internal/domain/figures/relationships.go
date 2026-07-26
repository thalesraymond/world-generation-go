package figures

import (
	"fmt"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// AddParentChild adds child ID to parent's Children and parent ID to child's Parents.
func AddParentChild(parent, child *HistoricalFigure) {
	parent.Relationships.Children = append(parent.Relationships.Children, child.ID)
	child.Relationships.Parents = append(child.Relationships.Parents, parent.ID)
}

// AddSpouse adds each figure's ID to the other's Spouse slice.
func AddSpouse(f1, f2 *HistoricalFigure) {
	f1.Relationships.Spouse = append(f1.Relationships.Spouse, f2.ID)
	f2.Relationships.Spouse = append(f2.Relationships.Spouse, f1.ID)
}

// FormMarriage attempts to marry two figures, returning an event if successful.
func FormMarriage(a, b *HistoricalFigure, year int) (simulation.Event, bool) {
	if !a.IsAlive() || !b.IsAlive() {
		return simulation.Event{}, false
	}
	AddSpouse(a, b)
	return simulation.Event{
		Year:           year,
		Category:       "Marriage",
		Description:    fmt.Sprintf("%s marries %s", a.Name, b.Name),
		FigureID:       a.ID,
		RelatedFigures: []string{b.ID},
	}, true
}

// GetHeir returns the eldest living child of a deceased figure, or nil.
func GetHeir(figures []HistoricalFigure, deceasedID string) *HistoricalFigure {
	var eldest *HistoricalFigure
	eldestBirthYear := -1
	for i := range figures {
		if !figures[i].IsAlive() {
			continue
		}
		for _, parentID := range figures[i].Relationships.Parents {
			if parentID == deceasedID {
				if eldest == nil || figures[i].BirthYear < eldestBirthYear {
					eldest = &figures[i]
					eldestBirthYear = figures[i].BirthYear
				}
				break
			}
		}
	}
	return eldest
}
