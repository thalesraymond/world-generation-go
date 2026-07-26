package simulation

import randv2 "math/rand/v2"

// Entity represents a world entity that can be ticked during simulation.
type Entity interface {
	Tick(year int, eventChan chan<- Event, rng *randv2.Rand)
}
