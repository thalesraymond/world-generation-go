package artifact

import (
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

func TestPostProcessAssignsEventIDs(t *testing.T) {
	events := []simulation.Event{
		{Year: 1, Category: "Birth", Description: "A"},
		{Year: 1, Category: "Death", Description: "B"},
		{Year: 3, Category: "Politics", Description: "C"},
		{Year: 1, Category: "Conquest", Description: "D"},
		{Year: 3, Category: "Economy", Description: "E"},
	}

	if err := PostProcess(nil, events); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	want := []string{"event-1-0", "event-1-1", "event-3-0", "event-1-2", "event-3-1"}
	for i, id := range want {
		if events[i].ID != id {
			t.Errorf("events[%d].ID = %q, want %q", i, events[i].ID, id)
		}
	}
}

func TestPostProcessDeterministic(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", Status: "lost"},
	}
	events := []simulation.Event{
		{Year: 1, Category: "Discovery", Description: "found", FigureID: "Deepcrest-1", ArtifactID: "artifact-1"},
		{Year: 2, Category: "Death", Description: "dies", FigureID: "Deepcrest-1", ArtifactID: "artifact-1"},
	}

	run := func() ([]Artifact, []simulation.Event, error) {
		arts := make([]Artifact, len(artifacts))
		copy(arts, artifacts)
		evs := make([]simulation.Event, len(events))
		copy(evs, events)
		err := PostProcess(arts, evs)
		return arts, evs, err
	}

	firstArts, firstEvents, err := run()
	if err != nil {
		t.Fatalf("first PostProcess: %v", err)
	}
	secondArts, secondEvents, err := run()
	if err != nil {
		t.Fatalf("second PostProcess: %v", err)
	}

	if !reflect.DeepEqual(firstArts, secondArts) {
		t.Errorf("provenance chains differ across identical runs:\nfirst=%+v\nsecond=%+v", firstArts, secondArts)
	}
	if !reflect.DeepEqual(firstEvents, secondEvents) {
		t.Errorf("event IDs differ across identical runs:\nfirst=%+v\nsecond=%+v", firstEvents, secondEvents)
	}
}

func TestPostProcessAttachesArtifactIDToOwnerDeath(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		}},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
		{Year: 5, Category: "Death", FigureID: "Deepcrest-2", Description: "unrelated dies"},
		{Year: 6, Category: "Birth", FigureID: "Deepcrest-3", Description: "birth"},
	}

	if err := PostProcess(artifacts, events); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("owner death event ArtifactID = %q, want artifact-1", events[0].ArtifactID)
	}
	if events[1].ArtifactID != "" {
		t.Errorf("unrelated death event ArtifactID = %q, want empty", events[1].ArtifactID)
	}
	if events[2].ArtifactID != "" {
		t.Errorf("birth event ArtifactID = %q, want empty", events[2].ArtifactID)
	}
}

func TestPostProcessAttachesArtifactIDToOwnerSettlementAttack(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Tome", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Transfer"},
		}},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "taken"},
		{Year: 5, Category: "Raid", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "raided"},
		{Year: 6, Category: "Conquest", SettlementName: "Haven", TargetSettlement: "Other", Description: "elsewhere"},
	}

	if err := PostProcess(artifacts, events); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("conquest event ArtifactID = %q, want artifact-1", events[0].ArtifactID)
	}
	if events[1].ArtifactID != "artifact-1" {
		t.Errorf("raid event ArtifactID = %q, want artifact-1", events[1].ArtifactID)
	}
	if events[2].ArtifactID != "" {
		t.Errorf("conquest of other settlement ArtifactID = %q, want empty", events[2].ArtifactID)
	}
}

func TestPostProcessDoesNotAttachLostArtifacts(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Relic", Status: "lost"},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Death", FigureID: "Deepcrest-1", Description: "dies"},
	}

	if err := PostProcess(artifacts, events); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	if events[0].ArtifactID != "" {
		t.Errorf("lost artifact implicated in death event: ArtifactID = %q, want empty", events[0].ArtifactID)
	}
}

func TestPostProcessTieBreakDeterministic(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-b", Name: "B", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		}},
		{ID: "artifact-a", Name: "A", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		}},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
	}

	if err := PostProcess(artifacts, events); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// Both artifacts match the same death event; the first in artifact order
	// wins so the attachment is independent of map iteration order.
	if events[0].ArtifactID != "artifact-b" {
		t.Errorf("death event ArtifactID = %q, want artifact-b (first in artifact order)", events[0].ArtifactID)
	}
}

func TestPostProcessDiscoveryRecordsProvenance(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", Status: "lost"},
	}
	events := []simulation.Event{
		{Year: 12, Category: "Discovery", FigureID: "Deepcrest-3", ArtifactID: "artifact-1", Description: "found"},
	}

	if err := PostProcess(artifacts, events); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	want := []ProvenanceEntry{
		{Year: 12, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-12-0", EventType: "Discovery"},
	}
	if !reflect.DeepEqual(a.Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", a.Provenance, want)
	}
	if wantIDs := []string{"event-12-0"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("associatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}

	// Current owner derives from the last provenance entry.
	if kind, id := currentOwner(a); kind != "figure" || id != "Deepcrest-3" {
		t.Errorf("current owner = (%q, %q), want (figure, Deepcrest-3)", kind, id)
	}
}

func TestPostProcessDiscoveryWithoutFigureRecordsNoEntry(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", Status: "lost"},
	}
	events := []simulation.Event{
		{Year: 12, Category: "Discovery", ArtifactID: "artifact-1", Description: "found by no one"},
	}

	if err := PostProcess(artifacts, events); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	if len(artifacts[0].Provenance) != 0 {
		t.Errorf("provenance = %+v, want no entries", artifacts[0].Provenance)
	}
	if wantIDs := []string{"event-12-0"}; !reflect.DeepEqual(artifacts[0].AssociatedEventIDs, wantIDs) {
		t.Errorf("associatedEventIDs = %v, want %v", artifacts[0].AssociatedEventIDs, wantIDs)
	}
}

func TestPostProcessAssociatesNonDiscoveryEvents(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		}},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Death", FigureID: "Deepcrest-1", ArtifactID: "artifact-1", Description: "owner dies"},
	}

	if err := PostProcess(artifacts, events); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// The event carries an ArtifactID (attached by the lifecycle transfer
	// rules); it is associated with the artifact but its transfer destination
	// is computed by those rules, so no provenance entry is appended here.
	if len(artifacts[0].Provenance) != 1 {
		t.Errorf("provenance = %+v, want the single discovery entry only", artifacts[0].Provenance)
	}
	if wantIDs := []string{"event-5-0"}; !reflect.DeepEqual(artifacts[0].AssociatedEventIDs, wantIDs) {
		t.Errorf("associatedEventIDs = %v, want %v", artifacts[0].AssociatedEventIDs, wantIDs)
	}
}

func TestPostProcessUnknownArtifactIDErrors(t *testing.T) {
	events := []simulation.Event{
		{Year: 1, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "ghost", Description: "found"},
	}

	if err := PostProcess(nil, events); err == nil {
		t.Fatal("PostProcess with unknown ArtifactID: expected error, got nil")
	}
}
