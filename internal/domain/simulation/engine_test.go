package simulation

import (
	"sync"
	"testing"
)

type mockEntity struct {
	name   string
	events []Event
}

func (m *mockEntity) Tick(year int, eventChan chan<- Event) {
	e := Event{
		Year:        year,
		Category:    "Test",
		Description: m.name + " ticked",
	}
	m.events = append(m.events, e)
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
		sim := New(1, 3)
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
