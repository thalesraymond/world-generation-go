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

	if err := PostProcess(nil, events, SignificanceContext{}, TransferContext{}); err != nil {
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
		err := PostProcess(arts, evs, SignificanceContext{}, TransferContext{})
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

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
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
		{Year: 4, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "taken"},
		{Year: 5, Category: "Raid", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "raid on the fallen city"},
		{Year: 6, Category: "Raid", SettlementName: "Haven", TargetSettlement: "Blackgate", Description: "raid on the new owner"},
		{Year: 7, Category: "Conquest", SettlementName: "Haven", TargetSettlement: "Other", Description: "elsewhere"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// The conquest transfers the artifact to Blackgate, so the year-5 raid
	// on Ironforge no longer terminates the owner; the year-6 raid on the
	// artifact's new owner (Blackgate) does.
	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("conquest event ArtifactID = %q, want artifact-1", events[0].ArtifactID)
	}
	if events[1].ArtifactID != "" {
		t.Errorf("raid on former owner's city ArtifactID = %q, want empty", events[1].ArtifactID)
	}
	if events[2].ArtifactID != "artifact-1" {
		t.Errorf("raid on new owner ArtifactID = %q, want artifact-1", events[2].ArtifactID)
	}
	if events[3].ArtifactID != "" {
		t.Errorf("conquest of other settlement ArtifactID = %q, want empty", events[3].ArtifactID)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Transfer"},
		{Year: 4, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-4-0", EventType: "Conquest"},
		{Year: 6, Owner: Owner{Kind: "settlement", ID: "Haven"}, EventID: "event-6-0", EventType: "Raid"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", artifacts[0].Provenance, want)
	}
}

func TestPostProcessDoesNotAttachLostArtifacts(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Relic", Status: "lost"},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Death", FigureID: "Deepcrest-1", Description: "dies"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
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

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
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

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
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
	if owner := CurrentOwner(a); owner.Kind != "figure" || owner.ID != "Deepcrest-3" {
		t.Errorf("current owner = %+v, want (figure, Deepcrest-3)", owner)
	}
}

func TestPostProcessDiscoveryWithoutFigureRecordsNoEntry(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", Status: "lost"},
	}
	events := []simulation.Event{
		{Year: 12, Category: "Discovery", ArtifactID: "artifact-1", Description: "found by no one"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
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

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// The event carries an ArtifactID (attached by the lifecycle transfer
	// rules) and terminates the artifact's owner: the transfer is recorded
	// here per spec 6.3. The zero transfer context cannot resolve an heir or
	// treasury, so the destination falls back to lost.
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 5, Owner: Owner{Kind: "lost"}, EventID: "event-5-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", artifacts[0].Provenance, want)
	}
	if wantIDs := []string{"event-5-0"}; !reflect.DeepEqual(artifacts[0].AssociatedEventIDs, wantIDs) {
		t.Errorf("associatedEventIDs = %v, want %v", artifacts[0].AssociatedEventIDs, wantIDs)
	}
}

func TestPostProcessUnknownArtifactIDErrors(t *testing.T) {
	events := []simulation.Event{
		{Year: 1, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "ghost", Description: "found"},
	}

	if err := PostProcess(nil, events, SignificanceContext{}, TransferContext{}); err == nil {
		t.Fatal("PostProcess with unknown ArtifactID: expected error, got nil")
	}
}

// heldArtifact builds a historical artifact held by a figure from year 1,
// the fixture shape used across the significance tests.
func heldArtifact() Artifact {
	return Artifact{
		ID:                 "artifact-1",
		Name:               "Crown",
		SignificanceSource: "historical",
		Status:             "held",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		},
	}
}

func TestPostProcessSignificanceWeightTable(t *testing.T) {
	artifacts := []Artifact{heldArtifact()}
	events := []simulation.Event{
		{Year: 2, Category: "War", ArtifactID: "artifact-1", Description: "war"},
		{Year: 3, Category: "Diplomacy", ArtifactID: "artifact-1", Description: "diplomacy"},
		{Year: 4, Category: "Politics", ArtifactID: "artifact-1", Description: "politics"},
		{Year: 5, Category: "Raid", ArtifactID: "artifact-1", Description: "raid"},
		{Year: 6, Category: "Expansion", ArtifactID: "artifact-1", Description: "expansion"},
		{Year: 7, Category: "Disaster", ArtifactID: "artifact-1", Description: "disaster"},
		{Year: 8, Category: "Economy", ArtifactID: "artifact-1", Description: "economy"},
		{Year: 9, Category: "Conquest", ArtifactID: "artifact-1", Description: "conquest"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.SignificanceScore != 14 {
		t.Errorf("SignificanceScore = %d, want 14", a.SignificanceScore)
	}
	if !a.IsSignificant {
		t.Error("IsSignificant = false, want true")
	}
	// The War event (year 2) is the first crossing: the monotonic latch must
	// record it even though later events accumulate more score.
	if a.PivotalEventID != "event-2-0" {
		t.Errorf("PivotalEventID = %q, want event-2-0 (first crossing)", a.PivotalEventID)
	}
	if a.SignificanceYear != 2 {
		t.Errorf("SignificanceYear = %d, want 2", a.SignificanceYear)
	}
}

func TestPostProcessSignificanceSingleEventPerCategory(t *testing.T) {
	cases := []struct {
		category   string
		wantScore  int
		wantSignif bool
	}{
		{"War", 3, true},
		{"Conquest", 3, true},
		{"Diplomacy", 2, false},
		{"Politics", 2, false},
		{"Raid", 2, false},
		{"Expansion", 1, false},
		{"Disaster", 1, false},
		{"Economy", 0, false},
		{"Settlement", 0, false},
		{"Birth", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.category, func(t *testing.T) {
			artifacts := []Artifact{heldArtifact()}
			events := []simulation.Event{
				{Year: 5, Category: tc.category, ArtifactID: "artifact-1", Description: "event"},
			}

			if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
				t.Fatalf("PostProcess: %v", err)
			}

			a := artifacts[0]
			if a.SignificanceScore != tc.wantScore {
				t.Errorf("SignificanceScore = %d, want %d", a.SignificanceScore, tc.wantScore)
			}
			if a.IsSignificant != tc.wantSignif {
				t.Errorf("IsSignificant = %v, want %v", a.IsSignificant, tc.wantSignif)
			}
			if tc.wantSignif {
				if a.PivotalEventID != "event-5-0" {
					t.Errorf("PivotalEventID = %q, want event-5-0", a.PivotalEventID)
				}
				if a.SignificanceYear != 5 {
					t.Errorf("SignificanceYear = %d, want 5", a.SignificanceYear)
				}
			}
		})
	}
}

func TestPostProcessSignificanceLatchIsMonotonic(t *testing.T) {
	t.Run("first crossing event wins", func(t *testing.T) {
		artifacts := []Artifact{heldArtifact()}
		events := []simulation.Event{
			{Year: 2, Category: "War", ArtifactID: "artifact-1", Description: "war"},
			{Year: 5, Category: "War", ArtifactID: "artifact-1", Description: "war again"},
		}

		if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
			t.Fatalf("PostProcess: %v", err)
		}

		a := artifacts[0]
		if a.SignificanceScore != 6 {
			t.Errorf("SignificanceScore = %d, want 6", a.SignificanceScore)
		}
		if a.PivotalEventID != "event-2-0" {
			t.Errorf("PivotalEventID = %q, want event-2-0", a.PivotalEventID)
		}
		if a.SignificanceYear != 2 {
			t.Errorf("SignificanceYear = %d, want 2", a.SignificanceYear)
		}
	})

	t.Run("carried latch never reverts", func(t *testing.T) {
		artifacts := []Artifact{{
			ID:                 "artifact-1",
			Name:               "Crown",
			SignificanceSource: "historical",
			Status:             "held",
			SignificanceScore:  4,
			IsSignificant:      true,
			PivotalEventID:     "event-9-0",
			SignificanceYear:   9,
			Provenance: []ProvenanceEntry{
				{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
			},
		}}
		events := []simulation.Event{
			{Year: 2, Category: "War", ArtifactID: "artifact-1", Description: "war"},
		}

		if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
			t.Fatalf("PostProcess: %v", err)
		}

		a := artifacts[0]
		if a.SignificanceScore != 7 {
			t.Errorf("SignificanceScore = %d, want 7", a.SignificanceScore)
		}
		if !a.IsSignificant {
			t.Error("IsSignificant = false, want true (latch must never revert)")
		}
		if a.PivotalEventID != "event-9-0" {
			t.Errorf("PivotalEventID = %q, want carried event-9-0", a.PivotalEventID)
		}
		if a.SignificanceYear != 9 {
			t.Errorf("SignificanceYear = %d, want carried 9", a.SignificanceYear)
		}
	})
}

func TestPostProcessSignificanceFigureAnnualAccrual(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Crown",
		SignificanceSource: "historical",
		Status:             "lost",
	}}
	events := []simulation.Event{
		{Year: 10, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "artifact-1", Description: "found"},
		{Year: 16, Category: "Birth", Description: "unrelated"},
	}
	sig := SignificanceContext{
		FigureReputation: map[string]map[int]int{
			"Deepcrest-1": {10: 2, 11: -3, 12: 1, 14: 2},
		},
	}

	if err := PostProcess(artifacts, events, sig, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// Held years are [10, 16): year 10 accrues +2, year 11's -3 clamps to 0,
	// year 12 accrues +1, year 14 accrues +2 -> 5. The crossing happens via
	// accrual in year 12, so there is no pivotal event.
	a := artifacts[0]
	if a.SignificanceScore != 5 {
		t.Errorf("SignificanceScore = %d, want 5", a.SignificanceScore)
	}
	if !a.IsSignificant {
		t.Error("IsSignificant = false, want true")
	}
	if a.SignificanceYear != 12 {
		t.Errorf("SignificanceYear = %d, want 12 (accrual crossing year)", a.SignificanceYear)
	}
	if a.PivotalEventID != "" {
		t.Errorf("PivotalEventID = %q, want empty (crossing was not an event)", a.PivotalEventID)
	}
}

func TestPostProcessSignificanceSettlementLumpSum(t *testing.T) {
	cases := []struct {
		class      string
		wantScore  int
		wantSignif bool
	}{
		{"MajorCity", 3, true},
		{"City", 2, false},
		{"Village", 1, false},
		{"Abandoned", 0, false},
		{"", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			artifacts := []Artifact{{
				ID:                 "artifact-1",
				Name:               "Tome",
				SignificanceSource: "historical",
				Status:             "held",
				Provenance: []ProvenanceEntry{
					{Year: 5, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-5-0", EventType: "Transfer"},
				},
			}}
			sig := SignificanceContext{SettlementClass: map[string]string{"Ironforge": tc.class}}

			if err := PostProcess(artifacts, nil, sig, TransferContext{}); err != nil {
				t.Fatalf("PostProcess: %v", err)
			}

			a := artifacts[0]
			if a.SignificanceScore != tc.wantScore {
				t.Errorf("SignificanceScore = %d, want %d", a.SignificanceScore, tc.wantScore)
			}
			if a.IsSignificant != tc.wantSignif {
				t.Errorf("IsSignificant = %v, want %v", a.IsSignificant, tc.wantSignif)
			}
			if tc.wantSignif {
				// The lump sum is not an event: the acquisition-year
				// crossing must not mint a pivotal event.
				if a.PivotalEventID != "" {
					t.Errorf("PivotalEventID = %q, want empty", a.PivotalEventID)
				}
				if a.SignificanceYear != 5 {
					t.Errorf("SignificanceYear = %d, want 5", a.SignificanceYear)
				}
			}
		})
	}

	t.Run("lump once at acquisition, events can still cross", func(t *testing.T) {
		artifacts := []Artifact{{
			ID:                 "artifact-1",
			Name:               "Tome",
			SignificanceSource: "historical",
			Status:             "held",
			Provenance: []ProvenanceEntry{
				{Year: 5, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-5-0", EventType: "Transfer"},
			},
		}}
		events := []simulation.Event{
			{Year: 8, Category: "War", ArtifactID: "artifact-1", Description: "war"},
			{Year: 12, Category: "Birth", Description: "unrelated"},
		}
		sig := SignificanceContext{SettlementClass: map[string]string{"Ironforge": "City"}}

		if err := PostProcess(artifacts, events, sig, TransferContext{}); err != nil {
			t.Fatalf("PostProcess: %v", err)
		}

		// The City lump (2) applies once at year 5, never again for later
		// years; the War event at year 8 crosses the threshold.
		a := artifacts[0]
		if a.SignificanceScore != 5 {
			t.Errorf("SignificanceScore = %d, want 5", a.SignificanceScore)
		}
		if !a.IsSignificant {
			t.Error("IsSignificant = false, want true")
		}
		if a.PivotalEventID != "event-8-0" {
			t.Errorf("PivotalEventID = %q, want event-8-0", a.PivotalEventID)
		}
		if a.SignificanceYear != 8 {
			t.Errorf("SignificanceYear = %d, want 8", a.SignificanceYear)
		}
	})
}

func TestPostProcessSignificanceFreezesWhileLost(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Crown",
		SignificanceSource: "historical",
		Status:             "lost",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
			{Year: 5, Owner: Owner{Kind: "lost", ID: ""}, EventID: "event-5-0", EventType: "Owner death"},
		},
	}}
	events := []simulation.Event{
		{Year: 7, Category: "War", ArtifactID: "artifact-1", Description: "war while lost"},
		{Year: 10, Category: "Discovery", FigureID: "Deepcrest-2", ArtifactID: "artifact-1", Description: "refound"},
		{Year: 12, Category: "War", ArtifactID: "artifact-1", Description: "war after rediscovery"},
		{Year: 15, Category: "Birth", Description: "unrelated"},
	}
	sig := SignificanceContext{
		FigureReputation: map[string]map[int]int{
			// Deepcrest-1's +5 in year 6 falls inside the lost span [5, 10)
			// and must not accrue; +1 in year 2 accrues while held.
			"Deepcrest-1": {2: 1, 6: 5},
		},
	}

	if err := PostProcess(artifacts, events, sig, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.SignificanceScore != 4 {
		t.Errorf("SignificanceScore = %d, want 4 (year 2 accrual + year 12 war only)", a.SignificanceScore)
	}
	if !a.IsSignificant {
		t.Error("IsSignificant = false, want true")
	}
	if a.PivotalEventID != "event-12-0" {
		t.Errorf("PivotalEventID = %q, want event-12-0 (post-rediscovery war)", a.PivotalEventID)
	}
	if a.SignificanceYear != 12 {
		t.Errorf("SignificanceYear = %d, want 12", a.SignificanceYear)
	}
	if len(a.Provenance) != 3 {
		t.Errorf("provenance length = %d, want 3", len(a.Provenance))
	}
}

func TestPostProcessSignificanceIntrinsicBypass(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-ruin-0",
		Name:               "Relic",
		Type:               "crown",
		SignificanceSource: "intrinsic",
		Status:             "lost",
		SignificanceScore:  3,
		IsSignificant:      true,
		SignificanceYear:   0,
	}}
	events := []simulation.Event{
		{Year: 2, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "artifact-ruin-0", Description: "found"},
		{Year: 4, Category: "War", ArtifactID: "artifact-ruin-0", Description: "war"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	// The intrinsic latch and creation year are carried untouched; events may
	// still raise the score beyond the threshold (spec 4.3).
	if a.SignificanceScore != 6 {
		t.Errorf("SignificanceScore = %d, want 6", a.SignificanceScore)
	}
	if !a.IsSignificant {
		t.Error("IsSignificant = false, want true")
	}
	if a.PivotalEventID != "" {
		t.Errorf("PivotalEventID = %q, want empty", a.PivotalEventID)
	}
	if a.SignificanceYear != 0 {
		t.Errorf("SignificanceYear = %d, want 0 (creation year)", a.SignificanceYear)
	}
}

func TestPostProcessSignificanceDeterministic(t *testing.T) {
	artifacts := []Artifact{
		heldArtifact(),
		{
			ID:                 "artifact-2",
			Name:               "Tome",
			SignificanceSource: "historical",
			Status:             "lost",
		},
		{
			ID:                 "artifact-ruin-0",
			Name:               "Relic",
			SignificanceSource: "intrinsic",
			Status:             "lost",
			SignificanceScore:  3,
			IsSignificant:      true,
			SignificanceYear:   0,
		},
	}
	events := []simulation.Event{
		{Year: 1, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "artifact-2", Description: "found"},
		{Year: 2, Category: "War", ArtifactID: "artifact-1", Description: "war"},
		{Year: 3, Category: "Raid", ArtifactID: "artifact-2", Description: "raid"},
		{Year: 4, Category: "Politics", ArtifactID: "artifact-1", Description: "politics"},
		{Year: 5, Category: "Birth", Description: "unrelated"},
	}
	sig := SignificanceContext{
		FigureReputation: map[string]map[int]int{
			"Deepcrest-1": {1: 1, 2: -2, 3: 2},
		},
		SettlementClass: map[string]string{"Ironforge": "City"},
	}

	run := func() ([]Artifact, []simulation.Event, error) {
		arts := make([]Artifact, len(artifacts))
		copy(arts, artifacts)
		evs := make([]simulation.Event, len(events))
		copy(evs, events)
		err := PostProcess(arts, evs, sig, TransferContext{})
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
		t.Errorf("significance evaluation differs across identical runs:\nfirst=%+v\nsecond=%+v", firstArts, secondArts)
	}
	if !reflect.DeepEqual(firstEvents, secondEvents) {
		t.Errorf("event IDs differ across identical runs:\nfirst=%+v\nsecond=%+v", firstEvents, secondEvents)
	}

	// Pin the evaluated state so a regression in the expected values is
	// caught, not just a run-to-run difference.
	if got := firstArts[0].SignificanceScore; got != 8 {
		t.Errorf("artifact-1 significance score = %d, want 8", got)
	}
	if got := firstArts[0].PivotalEventID; got != "event-2-0" {
		t.Errorf("artifact-1 pivotal event = %q, want event-2-0", got)
	}
	if got := firstArts[1].SignificanceScore; got != 5 {
		t.Errorf("artifact-2 significance score = %d, want 5", got)
	}
	if got := firstArts[1].PivotalEventID; got != "event-3-0" {
		t.Errorf("artifact-2 pivotal event = %q, want event-3-0", got)
	}
}
