package simulation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

func TestRunSimulationHappyPath(t *testing.T) {
	events, worldState, err := ucsim.RunSimulation(context.Background(), ucsim.OrchestratorConfig{
		Seed: 42, Width: 48, Height: 48, Years: 25,
	})
	if err != nil {
		t.Fatalf("RunSimulation() error = %v", err)
	}

	if len(events) == 0 {
		t.Fatal("RunSimulation() returned no events")
	}
	if worldState == nil {
		t.Fatal("RunSimulation() returned nil world state")
	}
	if len(worldState.Settlements) == 0 {
		t.Fatal("RunSimulation() world has no settlements")
	}
	if len(worldState.Artifacts) == 0 {
		t.Fatal("RunSimulation() world has no artifacts")
	}
}

func TestRunSimulationDeterministic(t *testing.T) {
	config := ucsim.OrchestratorConfig{Seed: 42, Width: 48, Height: 48, Years: 25}

	firstEvents, firstState, err := ucsim.RunSimulation(context.Background(), config)
	if err != nil {
		t.Fatalf("RunSimulation() first run error = %v", err)
	}

	secondEvents, secondState, err := ucsim.RunSimulation(context.Background(), config)
	if err != nil {
		t.Fatalf("RunSimulation() second run error = %v", err)
	}

	firstStateJSON, err := json.Marshal(firstState)
	if err != nil {
		t.Fatalf("json.Marshal(first state) error = %v", err)
	}
	secondStateJSON, err := json.Marshal(secondState)
	if err != nil {
		t.Fatalf("json.Marshal(second state) error = %v", err)
	}
	if !bytes.Equal(firstStateJSON, secondStateJSON) {
		t.Fatal("RunSimulation() produced differing world state for same config")
	}

	firstEventsJSON, err := json.Marshal(firstEvents)
	if err != nil {
		t.Fatalf("json.Marshal(first events) error = %v", err)
	}
	secondEventsJSON, err := json.Marshal(secondEvents)
	if err != nil {
		t.Fatalf("json.Marshal(second events) error = %v", err)
	}
	if !bytes.Equal(firstEventsJSON, secondEventsJSON) {
		t.Fatal("RunSimulation() produced differing events for same config")
	}
}

func TestRunSimulationPlantedRelicLifecycle(t *testing.T) {
	events, worldState, err := ucsim.RunSimulation(context.Background(), ucsim.OrchestratorConfig{
		Seed: 42, Width: 48, Height: 48, Years: 25,
	})
	if err != nil {
		t.Fatalf("RunSimulation() error = %v", err)
	}

	// Every planted relic is fake-discovered at genesis year 0 by a drawn
	// figure: provenance begins with a year-0 Discovery and the status is
	// never the initial lost.
	discovered := 0
	for _, a := range worldState.Artifacts {
		if a.SignificanceSource != "intrinsic" {
			continue
		}
		discovered++
		if len(a.Provenance) == 0 {
			t.Errorf("planted relic %s has no provenance", a.ID)
			continue
		}
		first := a.Provenance[0]
		if first.Year != 0 || first.Owner.Kind != "figure" || first.EventType != "Discovery" {
			t.Errorf("planted relic %s first provenance = %+v, want year-0 figure Discovery", a.ID, first)
		}
	}
	if discovered == 0 {
		t.Fatal("world has no planted relics to exercise the lifecycle")
	}

	// The synthetic discovery events reach the returned stream (the caller
	// must consume the extended stream, not the pre-pass one).
	sawGenesisDiscovery := false
	for _, e := range events {
		if e.Year == 0 && e.Category == "Discovery" && e.ArtifactID != "" {
			sawGenesisDiscovery = true
			break
		}
	}
	if !sawGenesisDiscovery {
		t.Error("no year-0 synthetic Discovery event reached the event stream")
	}

	// Same seed reproduces the identical lifecycle: the deterministic run
	// below pins that the synthetic events and statuses are seeded, not
	// incidental.
	events2, state2, err := ucsim.RunSimulation(context.Background(), ucsim.OrchestratorConfig{
		Seed: 42, Width: 48, Height: 48, Years: 25,
	})
	if err != nil {
		t.Fatalf("RunSimulation() second run error = %v", err)
	}
	firstJSON, _ := json.Marshal(worldState.Artifacts)
	secondJSON, _ := json.Marshal(state2.Artifacts)
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Error("artifact lifecycle state differs across identical seeded runs")
	}
	ev1, _ := json.Marshal(events)
	ev2, _ := json.Marshal(events2)
	if !bytes.Equal(ev1, ev2) {
		t.Error("event streams differ across identical seeded runs")
	}
}

func TestRunSimulationInvalidDimensions(t *testing.T) {
	events, worldState, err := ucsim.RunSimulation(context.Background(), ucsim.OrchestratorConfig{
		Seed: 1, Width: 0, Height: 32, Years: 10,
	})
	if err == nil {
		t.Fatal("RunSimulation() expected error for zero width")
	}
	if events != nil || worldState != nil {
		t.Fatal("RunSimulation() expected nil results on error")
	}
	if !strings.Contains(err.Error(), "generate world: invalid dimensions") {
		t.Errorf("RunSimulation() error = %v, want wrapped generate world error", err)
	}
}

func TestRunSimulationCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	events, worldState, err := ucsim.RunSimulation(ctx, ucsim.OrchestratorConfig{
		Seed: 3, Width: 32, Height: 32, Years: 10,
	})
	if err != nil {
		t.Fatalf("RunSimulation() with cancelled context error = %v", err)
	}
	if worldState == nil {
		t.Fatal("RunSimulation() with cancelled context returned nil world state")
	}
	if len(events) != 0 {
		t.Logf("collected %d events despite cancelled context", len(events))
	}
}
