package figures

import (
	"fmt"
	randv2 "math/rand/v2"
)

// HistoricalFigure represents an individual actor in world history.
type HistoricalFigure struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	BirthYear         int             `json:"birthYear"`
	DeathYear         int             `json:"deathYear"`
	MaxAge            int             `json:"maxAge"`
	RoleRole          Role            `json:"-"`
	Role              string          `json:"role"`
	Faction           string          `json:"faction"`
	Stats             Stats           `json:"stats"`
	Relationships     Relationships   `json:"relationships"`
	Reputation        []ReputationEntry  `json:"reputation"`
	ParentID          string          `json:"parentID,omitempty"`
	TransitionHistory []TransitionEntry  `json:"transitionHistory,omitempty"`
}

// Stats represents the character attributes of a historical figure.
type Stats struct {
	Martial    int `json:"martial"`
	Diplomatic int `json:"diplomatic"`
	Infamy     int `json:"infamy"`
}

// Normalize clamps all stats to the valid range [1, 20].
func (s Stats) Normalize() Stats {
	s.Martial = clamp(s.Martial)
	s.Diplomatic = clamp(s.Diplomatic)
	s.Infamy = clamp(s.Infamy)
	return s
}

func clamp(v int) int {
	if v < 1 {
		return 1
	}
	if v > 20 {
		return 20
	}
	return v
}

// Copy returns a value copy of the Stats.
func (s Stats) Copy() Stats { return s }

// InfluenceOutcome returns true when the figure succeeds in the given category.
func (s Stats) InfluenceOutcome(category string, rng *randv2.Rand) bool {
	switch category {
	case "Conflict":
		return rng.IntN(20) < s.Martial
	case "Politics":
		return rng.IntN(20) < s.Diplomatic
	default:
		return rng.IntN(100) < 50
	}
}

// GenerateStats creates random stats for a figure, optionally biased by role.
func GenerateStats(rng *randv2.Rand, role string) Stats {
	s := Stats{
		Martial:    1 + rng.IntN(18),
		Diplomatic: 1 + rng.IntN(18),
		Infamy:     1 + rng.IntN(18),
	}
	switch role {
	case "General":
		s.Martial = clamp(s.Martial + 2)
	case "Diplomat":
		s.Diplomatic = clamp(s.Diplomatic + 2)
	}
	return s
}

// ReputationEntry records a change in reputation for a figure.
type ReputationEntry struct {
	Year        int    `json:"year"`
	Event       string `json:"event"`
	Delta       int    `json:"delta"`
	Description string `json:"description"`
}

// AddReputation appends a reputation entry and increases infamy for negative deltas.
func (f *HistoricalFigure) AddReputation(entry ReputationEntry) {
	f.Reputation = append(f.Reputation, entry)
	if entry.Delta < 0 {
		s := f.Stats
		s.Infamy = clamp(s.Infamy + (-entry.Delta))
		f.Stats = s
	}
}

// TotalReputation returns the sum of all reputation deltas.
func (f HistoricalFigure) TotalReputation() int {
	total := 0
	for _, e := range f.Reputation {
		total += e.Delta
	}
	return total
}

// RecentReputation returns reputation entries within the lookback window.
func (f HistoricalFigure) RecentReputation(year, lookback int) []ReputationEntry {
	var recent []ReputationEntry
	for _, e := range f.Reputation {
		if e.Year >= year-lookback {
			recent = append(recent, e)
		}
	}
	return recent
}

// TransitionEntry records a role change in a figure's life.
type TransitionEntry struct {
	Year     int    `json:"year"`
	FromRole string `json:"fromRole"`
	ToRole   string `json:"toRole"`
	Reason   string `json:"reason"`
}

// SetRole assigns a role implementation to the figure and updates the string Role.
func (f *HistoricalFigure) SetRole(role Role) {
	f.RoleRole = role
	if role != nil {
		f.Role = role.Name()
	} else {
		f.Role = ""
	}
}

// GetRole returns the figure's role implementation, lazily initializing from the string Role.
func (f *HistoricalFigure) GetRole() Role {
	if f.RoleRole != nil {
		return f.RoleRole
	}
	if f.Role != "" {
		r, _ := NewRole(f.Role)
		if r != nil {
			f.RoleRole = r
			return r
		}
	}
	return nil
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

// String returns a human-readable summary of the figure.
func (f HistoricalFigure) String() string {
	return fmt.Sprintf("%s (%s, age %d, M:%d D:%d I:%d)", f.Name, f.Role, f.Age(0), f.Stats.Martial, f.Stats.Diplomatic, f.Stats.Infamy)
}

// Relationships stores the immediate social ties of a historical figure.
type Relationships struct {
	Parents  []string `json:"parents"`
	Children []string `json:"children"`
	Spouse   []string `json:"spouse"`
}

