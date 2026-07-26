package simulation

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	randv2 "math/rand/v2"

	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

type testEntity struct {
	name string
}

func (t testEntity) Tick(year int, eventChan chan<- domsim.Event, rng *randv2.Rand) {
	eventChan <- domsim.Event{
		Year:        year,
		Category:    "Test",
		Description: fmt.Sprintf("%s acted in year %d", t.name, year),
	}
}

func TestRunSimulation(t *testing.T) {
	var buf bytes.Buffer
	entities := []domsim.Entity{
		testEntity{name: "FactionA"},
	}

	err := RunSimulation(1, 2, entities, &buf, randv2.New(randv2.NewPCG(1, 0)))
	if err != nil {
		t.Fatalf("RunSimulation error = %v", err)
	}

	output := buf.String()
	expectedLines := []string{
		"[1] (Test) FactionA acted in year 1",
		"[2] (Test) FactionA acted in year 2",
	}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestRunSimulationInvalidRange(t *testing.T) {
	var buf bytes.Buffer
	err := RunSimulation(2, 1, nil, &buf, randv2.New(randv2.NewPCG(1, 0)))
	if err == nil {
		t.Fatalf("expected error for startYear > endYear, got nil")
	}
}
