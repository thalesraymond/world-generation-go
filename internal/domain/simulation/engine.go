package simulation

import randv2 "math/rand/v2"

// Simulation manages entities and advances time chronologically.
type Simulation struct {
	startYear int
	endYear   int
	entities  []Entity
	rng       *randv2.Rand
}

// New creates a new Simulation engine for the given year range.
func New(startYear, endYear int, rng *randv2.Rand) *Simulation {
	return &Simulation{
		startYear: startYear,
		endYear:   endYear,
		rng:       rng,
	}
}

// AddEntity registers an entity to be ticked during the simulation.
func (s *Simulation) AddEntity(e Entity) {
	s.entities = append(s.entities, e)
}

// Run executes the simulation from startYear to endYear sequentially.
// It closes eventChan when simulation completes and all ticks are processed.
func (s *Simulation) Run(eventChan chan<- Event) {
	defer close(eventChan)

	for year := s.startYear; year <= s.endYear; year++ {
		for _, entity := range s.entities {
			entity.Tick(year, eventChan, s.rng)
		}
	}
}
