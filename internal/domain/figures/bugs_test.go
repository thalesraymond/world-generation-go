package figures

import (
	randv2 "math/rand/v2"
	"sync"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// settlementEntity implements simulation.Entity for a settlement's figures.
type settlementEntity struct {
	figures    []HistoricalFigure
	population float64
	faction    string
	settlement string
}

func (s *settlementEntity) Tick(year int, eventChan chan<- simulation.Event, rng *randv2.Rand) {
	deathEvents := CheckDeaths(s.figures, year, rng)
	for _, e := range deathEvents {
		eventChan <- e
	}

	child := CheckBirths(s.figures, s.population, year, rng)
	if child != nil {
		s.figures = append(s.figures, *child)
	}

	roleEvents := AssignRoles(s.figures, nil, 0, 0, rng)
	for _, e := range roleEvents {
		eventChan <- e
	}
}

func TestSimulation_BugsFixed(t *testing.T) {
	const seed = uint64(12345)
	const settlementName = "Bugville"
	const faction = "BugFaction"
	const foundingYear = 0 // Previously triggered negative founder birth years
	const simEndYear = 90
	const population = 5000.0

	rng := randv2.New(randv2.NewPCG(seed, seed+1))
	simRng := randv2.New(randv2.NewPCG(seed, seed+1))

	founders := GenerateFounders(rng, settlementName, faction, foundingYear)
	if len(founders) == 0 {
		t.Fatal("expected founders, got none")
	}

	ent := &settlementEntity{
		figures:    make([]HistoricalFigure, len(founders)),
		population: population,
		faction:    faction,
		settlement: settlementName,
	}
	copy(ent.figures, founders)

	sim := simulation.New(foundingYear, simEndYear, simRng)
	sim.AddEntity(ent)

	eventChan := make(chan simulation.Event, 1000)
	var collectedEvents []simulation.Event
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for ev := range eventChan {
			collectedEvents = append(collectedEvents, ev)
		}
	}()

	sim.Run(eventChan)
	wg.Wait()

	t.Logf("Simulation: years %d–%d, %d events collected, %d figures at end",
		foundingYear, simEndYear, len(collectedEvents), len(ent.figures))

	// --- Bug 1: Founders must not have negative or zero birth years ---
	for _, f := range founders {
		if f.BirthYear < 1 {
			t.Errorf("BUG 1: founder %q (id=%s) has non-positive birth year %d", f.Name, f.ID, f.BirthYear)
		}
	}
	// All figures must be born within the simulation range.
	for _, f := range ent.figures {
		if f.BirthYear < foundingYear {
			t.Errorf("BUG 1: figure %q (id=%s) born before simulation start (%d)", f.Name, f.ID, f.BirthYear)
		}
	}

	// --- Bug 2: Dead-but-alive — death events must correspond to figures with DeathYear set ---
	deathFigureIDs := make(map[string]int)
	for _, ev := range collectedEvents {
		if ev.Category == "Death" && ev.FigureID != "" {
			deathFigureIDs[ev.FigureID] = ev.Year
		}
	}
	t.Logf("Death events emitted: %d", len(deathFigureIDs))

	for _, f := range ent.figures {
		dy, found := deathFigureIDs[f.ID]
		if found {
			if f.DeathYear == 0 {
				t.Errorf("BUG 2: figure %q (id=%s) had a death event at year %d but DeathYear is 0 (still alive)", f.Name, f.ID, dy)
			}
			if f.DeathYear != dy {
				t.Errorf("BUG 2: figure %q (id=%s) death event year %d != DeathYear %d", f.Name, f.ID, dy, f.DeathYear)
			}
		}
	}

	// --- Bug 3: Newborns born after the first adult founders exist should have parents ---
	founderIDs := make(map[string]bool)
	for _, f := range founders {
		founderIDs[f.ID] = true
	}
	adultStartYear := foundingYear
	for _, f := range founders {
		if f.BirthYear+adultAge > adultStartYear {
			adultStartYear = f.BirthYear + adultAge
		}
	}
	newbornCount := 0
	newbornWithParents := 0
	for _, f := range ent.figures {
		if !founderIDs[f.ID] {
			newbornCount++
			if f.BirthYear >= adultStartYear && len(f.Relationships.Parents) > 0 {
				newbornWithParents++
			}
		}
	}
	t.Logf("Newborns: %d total, %d with parents", newbornCount, newbornWithParents)
	if newbornCount > 0 && newbornWithParents == 0 {
		t.Errorf("BUG 3: no newborns born after year %d have parents set", adultStartYear)
	}
}
