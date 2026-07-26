package simulation

// Entity represents a world entity that can be ticked during simulation.
type Entity interface {
	Tick(year int, eventChan chan<- Event)
}
