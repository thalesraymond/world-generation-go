package figures

// HistoricalFigure represents an individual actor in world history.
type HistoricalFigure struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	BirthYear     int            `json:"birthYear"`
	DeathYear     int            `json:"deathYear"`
	MaxAge        int            `json:"maxAge"`
	Role          string         `json:"role"`
	Faction       string         `json:"faction"`
	Relationships Relationships  `json:"relationships"`
}

// Relationships stores the immediate social ties of a historical figure.
type Relationships struct {
	Parents  []string `json:"parents"`
	Children []string `json:"children"`
	Spouse   []string `json:"spouse"`
}

// IsAlive reports whether the figure is currently alive.
func (f HistoricalFigure) IsAlive() bool {
	return f.DeathYear == 0
}

// Age returns the figure's age in the given year.
func (f HistoricalFigure) Age(currentYear int) int {
	return currentYear - f.BirthYear
}

// SetDeath records the year the figure died.
func (f *HistoricalFigure) SetDeath(year int) {
	f.DeathYear = year
}
