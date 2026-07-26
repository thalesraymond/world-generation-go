package simulation

import (
	"fmt"
	"sync"
	"testing"

	randv2 "math/rand/v2"
)

type mockEntity struct {
	name   string
	events []Event
}

func (m *mockEntity) Tick(year int, eventChan chan<- Event, rng *randv2.Rand) {
	e := Event{
		Year:        year,
		Category:    "Test",
		Description: m.name + " ticked",
	}
	m.events = append(m.events, e)
	eventChan <- e
}

type rngMockEntity struct {
	name string
}

func (m *rngMockEntity) Tick(year int, eventChan chan<- Event, rng *randv2.Rand) {
	val := rng.IntN(100)
	e := Event{
		Year:        year,
		Category:    "RNGTest",
		Description: fmt.Sprintf("%s rolled %d", m.name, val),
	}
	eventChan <- e
}

func TestFormatEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{
			name: "with category",
			event: Event{
				Year:        100,
				Category:    "War",
				Description: "The siege began.",
			},
			expected: "[100] (War) The siege began.",
		},
		{
			name: "without category",
			event: Event{
				Year:        101,
				Description: "A peaceful year.",
			},
			expected: "[101] A peaceful year.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatEvent(tt.event)
			if got != tt.expected {
				t.Errorf("FormatEvent() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSimulationRunAndDeterminism(t *testing.T) {
	runSimulation := func() []Event {
		sim := New(1, 3, randv2.New(randv2.NewPCG(42, 0)))
		e1 := &mockEntity{name: "Alpha"}
		e2 := &mockEntity{name: "Beta"}
		sim.AddEntity(e1)
		sim.AddEntity(e2)

		eventChan := make(chan Event, 10)
		var collected []Event
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			for ev := range eventChan {
				collected = append(collected, ev)
			}
		}()

		sim.Run(eventChan)
		wg.Wait()
		return collected
	}

	res1 := runSimulation()
	res2 := runSimulation()

	if len(res1) != 6 {
		t.Fatalf("expected 6 events (2 entities * 3 years), got %d", len(res1))
	}

	if len(res1) != len(res2) {
		t.Fatalf("run lengths differ: %d vs %d", len(res1), len(res2))
	}

	for i := range res1 {
		if res1[i] != res2[i] {
			t.Errorf("determinism violation at index %d: %v vs %v", i, res1[i], res2[i])
		}
	}
}

func TestSimulationDeterminismWithRNGEntities(t *testing.T) {
	rngEntity := func(name string) Entity {
		return &rngMockEntity{name: name}
	}

	runSimulation := func() []Event {
		sim := New(1, 3, randv2.New(randv2.NewPCG(42, 0)))
		e1 := rngEntity("Alpha")
		e2 := rngEntity("Beta")
		sim.AddEntity(e1)
		sim.AddEntity(e2)

		eventChan := make(chan Event, 10)
		var collected []Event
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			for ev := range eventChan {
				collected = append(collected, ev)
			}
		}()

		sim.Run(eventChan)
		wg.Wait()
		return collected
	}

	res1 := runSimulation()
	res2 := runSimulation()

	if len(res1) != len(res2) {
		t.Fatalf("run lengths differ: %d vs %d", len(res1), len(res2))
	}

	for i := range res1 {
		if res1[i] != res2[i] {
			t.Errorf("determinism violation at index %d: %v vs %v", i, res1[i], res2[i])
		}
	}
}
