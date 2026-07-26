package simulation

import (
	"fmt"
	"io"
	randv2 "math/rand/v2"
	"sync"

	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// RunSimulation runs the simulation from startYear to endYear, streaming formatted events to out.
func RunSimulation(startYear, endYear int, entities []domsim.Entity, out io.Writer, timelineRNG *randv2.Rand) error {
	if startYear > endYear {
		return fmt.Errorf("start year %d cannot be greater than end year %d", startYear, endYear)
	}

	sim := domsim.New(startYear, endYear, timelineRNG)
	for _, e := range entities {
		sim.AddEntity(e)
	}

	eventChan := make(chan domsim.Event, 100)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for event := range eventChan {
			formatted := domsim.FormatEvent(event)
			_, _ = fmt.Fprintln(out, formatted)
		}
	}()

	sim.Run(eventChan)
	wg.Wait()

	return nil
}
