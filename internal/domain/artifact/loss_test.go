package artifact

import (
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

func TestLossOnAbandonedSettlement(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Tome of Ironforge", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		}},
	}
	sig := SignificanceContext{SettlementClass: map[string]string{"Ironforge": "Abandoned"}}

	applyLoss(artifacts, 300, sig)

	a := artifacts[0]
	if a.Status != "lost" {
		t.Errorf("status = %q, want lost", a.Status)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		// The loss is a post-hoc observation, not a stream event: no EventID,
		// synthetic ArtifactLoss type (spec 6.1), dated at the horizon.
		{Year: 300, Owner: Owner{Kind: "lost"}, EventID: "", EventType: "ArtifactLoss"},
	}
	if !reflect.DeepEqual(a.Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", a.Provenance, want)
	}
	if year, ok := LostSinceYear(a); !ok || year != 300 {
		t.Errorf("LostSinceYear = (%d, %v), want (300, true)", year, ok)
	}
}

func TestLossSkipsNonAbandonedOwners(t *testing.T) {
	artifacts := []Artifact{
		// Figure owner: never lost by abandonment.
		{ID: "artifact-1", Name: "Crown", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "D-1"}, EventID: "event-1-0", EventType: "Discovery"},
		}},
		// Standing settlement (City): not lost.
		{ID: "artifact-2", Name: "Tome", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		}},
		// Settlement absent from the class map: class unknown, not lost.
		{ID: "artifact-3", Name: "Blade", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ghostfall"}, EventID: "event-1-0", EventType: "Creation"},
		}},
		// Already lost: never re-triggered.
		{ID: "artifact-4", Name: "Ring", Status: "lost", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "lost"}, EventID: "event-1-0", EventType: "Death"},
		}},
	}
	sig := SignificanceContext{SettlementClass: map[string]string{"Ironforge": "City"}}

	applyLoss(artifacts, 300, sig)

	for i, a := range artifacts {
		if len(a.Provenance) != 1 {
			t.Errorf("artifacts[%d] provenance = %+v, want unchanged single entry", i, a.Provenance)
		}
	}
	if artifacts[3].Status != "lost" {
		t.Errorf("artifacts[3] status = %q, want lost (unchanged)", artifacts[3].Status)
	}
}

// TestLossStatusSweepForMidStreamLosses pins the status propagation for the
// in-stream degenerate death path (spec 6.3: no heir and no settlement):
// recordTransfers records Owner{Kind: "lost"} during the walk and applyLoss
// must set Status "lost" so rediscovery and export see a consistent state.
func TestLossStatusSweepForMidStreamLosses(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", SignificanceSource: "historical", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "D-1"}, EventID: "event-1-0", EventType: "Discovery"},
		}},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Death", FigureID: "D-1", Description: "owner dies"},
	}

	got, _, err := EmergencePass(artifacts, events, nil, SignificanceContext{}, TransferContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	// The zero transfer context cannot resolve an heir or a treasury: the
	// death records the lost fallback and the sweep flips the status.
	a := got[0]
	if a.Status != "lost" {
		t.Errorf("status = %q, want lost (death with no heir and no settlement)", a.Status)
	}
	owner := CurrentOwner(a)
	if owner.Kind != "lost" {
		t.Errorf("current owner = %+v, want Owner{Kind: lost}", owner)
	}
	// Seed-1 rediscovery gate after the walk: 99 — fails, nothing minted.
	if year, ok := LostSinceYear(a); !ok || year != 5 {
		t.Errorf("LostSinceYear = (%d, %v), want (5, true)", year, ok)
	}
}

func TestLostSinceYear(t *testing.T) {
	cases := []struct {
		name string
		prov []ProvenanceEntry
		year int
		ok   bool
	}{
		{"no entries", nil, 0, false},
		{"held only", []ProvenanceEntry{{Year: 1, Owner: Owner{Kind: "figure", ID: "D-1"}}}, 0, false},
		{"lost entry", []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "D-1"}},
			{Year: 5, Owner: Owner{Kind: "lost"}},
		}, 5, true},
		{"multiple lost entries returns the last", []ProvenanceEntry{
			{Year: 5, Owner: Owner{Kind: "lost"}},
			{Year: 10, Owner: Owner{Kind: "figure", ID: "D-2"}},
			{Year: 12, Owner: Owner{Kind: "lost"}},
		}, 12, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			year, ok := LostSinceYear(Artifact{Provenance: tc.prov})
			if year != tc.year || ok != tc.ok {
				t.Errorf("LostSinceYear = (%d, %v), want (%d, %v)", year, ok, tc.year, tc.ok)
			}
		})
	}
}

// TestLossAndRediscoveryChronology exercises the full post-walk sequence for
// one artifact: abandonment loss at the horizon, then a passing rediscovery
// draw at the same year. The provenance chain must stay chronological (loss
// entry before rediscovery entry) and the status must end held.
func TestLossAndRediscoveryChronology(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Tome of Ironforge",
		SignificanceSource: "historical",
		Status:             "held",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		},
	}}
	events := []simulation.Event{
		{Year: 12, Category: "Birth", Description: "unrelated"},
	}
	sig := SignificanceContext{SettlementClass: map[string]string{"Ironforge": "Abandoned"}}
	figures := []FigureContext{{ID: "D-1", Settlement: "Deepcrest"}}

	got, _, err := EmergencePass(artifacts, events, figures, sig, TransferContext{}, artifactsRNG(2))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	a := got[0]
	if a.Status != "held" {
		t.Errorf("status = %q, want held (loss then immediate rediscovery)", a.Status)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		{Year: 12, Owner: Owner{Kind: "lost"}, EventID: "", EventType: "ArtifactLoss"},
		{Year: 12, Owner: Owner{Kind: "figure", ID: "D-1"}, EventID: "event-12-1", EventType: "Discovery"},
	}
	if !reflect.DeepEqual(a.Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", a.Provenance, want)
	}
	// The rediscovery entry closes the lost span: score freezes between the
	// entries (nothing accrues at the horizon anyway) and the status is held.
	if year, ok := LostSinceYear(a); !ok || year != 12 {
		t.Errorf("LostSinceYear = (%d, %v), want (12, true)", year, ok)
	}
}
