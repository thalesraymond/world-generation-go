package simulation

// Simulation manages entities and advances time chronologically.
type Simulation struct {
	startYear int
	endYear   int
	entities  []Entity
}

// New creates a new Simulation engine for the given year range.
func New(startYear, endYear int) *Simulation {
	return &Simulation{
		startYear: startYear,
		endYear:   endYear,
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
			entity.Tick(year, eventChan)
		}
	}
}
