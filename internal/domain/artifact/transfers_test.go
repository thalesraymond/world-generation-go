package artifact

import (
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// figureArtifact builds a historical artifact held by a figure from year 1.
func figureArtifact(id, figureID string) Artifact {
	return Artifact{
		ID:                 id,
		Name:               "Crown of " + figureID,
		Type:               "crown",
		SignificanceSource: "historical",
		Status:             "held",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: figureID}, EventID: "event-1-0", EventType: "Discovery"},
		},
	}
}

// settlementArtifact builds a historical artifact held by a settlement from
// year 1.
func settlementArtifact(id, settlementID string) Artifact {
	return Artifact{
		ID:                 id,
		Name:               "Tome of " + settlementID,
		Type:               "tome",
		SignificanceSource: "historical",
		Status:             "held",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "settlement", ID: settlementID}, EventID: "event-1-0", EventType: "Creation"},
		},
	}
}

func TestTransfersDeathToHeir(t *testing.T) {
	artifacts := []Artifact{figureArtifact("artifact-1", "Deepcrest-1")}
	events := []simulation.Event{
		{Year: 50, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
	}
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 50},
		// Deepcrest-2 died before the event year: ineligible.
		{ID: "Deepcrest-2", Settlement: "Deepcrest", BirthYear: 20, DeathYear: 40, Parents: []string{"Deepcrest-1"}},
		// Deepcrest-3 died after the event year: eligible (DeathYear > 50).
		{ID: "Deepcrest-3", Settlement: "Deepcrest", BirthYear: 18, DeathYear: 60, Parents: []string{"Deepcrest-1"}},
	}}

	if err := PostProcess(artifacts, events, SignificanceContext{}, ctx); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("death event ArtifactID = %q, want artifact-1", events[0].ArtifactID)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 50, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-50-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", artifacts[0].Provenance, want)
	}
	if owner := CurrentOwner(artifacts[0]); owner.Kind != "figure" || owner.ID != "Deepcrest-3" {
		t.Errorf("current owner = %+v, want (figure, Deepcrest-3)", owner)
	}
	if wantIDs := []string{"event-50-0"}; !reflect.DeepEqual(artifacts[0].AssociatedEventIDs, wantIDs) {
		t.Errorf("associatedEventIDs = %v, want %v", artifacts[0].AssociatedEventIDs, wantIDs)
	}
}

func TestTransfersDeathWithoutHeirGoesToTreasury(t *testing.T) {
	artifacts := []Artifact{figureArtifact("artifact-1", "Deepcrest-1")}
	events := []simulation.Event{
		{Year: 40, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies childless"},
	}
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 40},
		{ID: "Deepcrest-2", Settlement: "Ironforge", BirthYear: 20, Parents: []string{"Other-1"}},
	}}

	if err := PostProcess(artifacts, events, SignificanceContext{}, ctx); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// No child of Deepcrest-1 exists: the artifact passes to the deceased's
	// settlement treasury, not to an unrelated figure's settlement.
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 40, Owner: Owner{Kind: "settlement", ID: "Deepcrest"}, EventID: "event-40-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", artifacts[0].Provenance, want)
	}
	if owner := CurrentOwner(artifacts[0]); owner.Kind != "settlement" || owner.ID != "Deepcrest" {
		t.Errorf("current owner = %+v, want (settlement, Deepcrest)", owner)
	}
}

func TestTransfersDeathOfUnknownFigureIsLost(t *testing.T) {
	artifacts := []Artifact{figureArtifact("artifact-1", "Ghost")}
	events := []simulation.Event{
		{Year: 30, Category: "Death", FigureID: "Ghost", Description: "unknown owner dies"},
	}
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-2", Settlement: "Deepcrest", BirthYear: 20, Parents: []string{"Deepcrest-1"}},
	}}

	if err := PostProcess(artifacts, events, SignificanceContext{}, ctx); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Ghost"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 30, Owner: Owner{Kind: "lost"}, EventID: "event-30-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", artifacts[0].Provenance, want)
	}
}

func TestTransfersConquestSpoilsGoToAggressor(t *testing.T) {
	artifacts := []Artifact{
		settlementArtifact("artifact-1", "Ironforge"),
		settlementArtifact("artifact-2", "Ironforge"),
	}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "Ironforge falls"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// The first terminated artifact in artifact order becomes the carrier.
	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("conquest event ArtifactID = %q, want artifact-1 (first terminated)", events[0].ArtifactID)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		{Year: 5, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-5-0", EventType: "Conquest"},
	}
	for i, art := range artifacts {
		if !reflect.DeepEqual(art.Provenance, want) {
			t.Errorf("artifacts[%d] provenance = %+v, want %+v", i, art.Provenance, want)
		}
		if wantIDs := []string{"event-5-0"}; !reflect.DeepEqual(art.AssociatedEventIDs, wantIDs) {
			t.Errorf("artifacts[%d] associatedEventIDs = %v, want %v", i, art.AssociatedEventIDs, wantIDs)
		}
		if owner := CurrentOwner(art); owner.Kind != "settlement" || owner.ID != "Blackgate" {
			t.Errorf("artifacts[%d] current owner = %+v, want (settlement, Blackgate)", i, owner)
		}
	}
	// Only the ArtifactID carrier earns significance weight (spec 4.1); the
	// second terminated artifact records provenance and association only.
	if artifacts[0].SignificanceScore != 3 {
		t.Errorf("carrier significance score = %d, want 3 (Conquest weight)", artifacts[0].SignificanceScore)
	}
	if artifacts[1].SignificanceScore != 0 {
		t.Errorf("non-carrier significance score = %d, want 0", artifacts[1].SignificanceScore)
	}
}

func TestTransfersRaidSpoilsGoToAggressor(t *testing.T) {
	artifacts := []Artifact{settlementArtifact("artifact-1", "Ironforge")}
	events := []simulation.Event{
		{Year: 7, Category: "Raid", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "raided"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("raid event ArtifactID = %q, want artifact-1", events[0].ArtifactID)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		{Year: 7, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-7-0", EventType: "Raid"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", artifacts[0].Provenance, want)
	}
	if owner := CurrentOwner(artifacts[0]); owner.Kind != "settlement" || owner.ID != "Blackgate" {
		t.Errorf("current owner = %+v, want (settlement, Blackgate)", owner)
	}
	if artifacts[0].SignificanceScore != 2 {
		t.Errorf("significance score = %d, want 2 (Raid weight)", artifacts[0].SignificanceScore)
	}
}

func TestTransfersDeathTransfersEveryOwnedArtifact(t *testing.T) {
	artifacts := []Artifact{
		figureArtifact("artifact-1", "Deepcrest-1"),
		figureArtifact("artifact-2", "Deepcrest-1"),
	}
	events := []simulation.Event{
		{Year: 25, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
	}
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 25},
		{ID: "Deepcrest-2", Settlement: "Deepcrest", BirthYear: 10, Parents: []string{"Deepcrest-1"}},
	}}

	if err := PostProcess(artifacts, events, SignificanceContext{}, ctx); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("death event ArtifactID = %q, want artifact-1 (first terminated)", events[0].ArtifactID)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 25, Owner: Owner{Kind: "figure", ID: "Deepcrest-2"}, EventID: "event-25-0", EventType: "Death"},
	}
	for i, art := range artifacts {
		if !reflect.DeepEqual(art.Provenance, want) {
			t.Errorf("artifacts[%d] provenance = %+v, want %+v", i, art.Provenance, want)
		}
		if owner := CurrentOwner(art); owner.Kind != "figure" || owner.ID != "Deepcrest-2" {
			t.Errorf("artifacts[%d] current owner = %+v, want (figure, Deepcrest-2)", i, owner)
		}
	}
}

func TestTransfersNonOwnerEventsRecordNothing(t *testing.T) {
	artifacts := []Artifact{
		figureArtifact("artifact-1", "Deepcrest-1"),
		settlementArtifact("artifact-2", "Ironforge"),
	}
	events := []simulation.Event{
		{Year: 5, Category: "Death", FigureID: "Deepcrest-2", Description: "unrelated death"},
		{Year: 6, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Haven", Description: "elsewhere"},
		{Year: 7, Category: "Raid", SettlementName: "Blackgate", Description: "unresolvable raid target"},
		{Year: 8, Category: "Birth", Description: "unrelated"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	for _, e := range events {
		if e.ArtifactID != "" {
			t.Errorf("event %q ArtifactID = %q, want empty", e.ID, e.ArtifactID)
		}
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("figure artifact provenance = %+v, want unchanged %+v", artifacts[0].Provenance, want)
	}
	wantS := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
	}
	if !reflect.DeepEqual(artifacts[1].Provenance, wantS) {
		t.Errorf("settlement artifact provenance = %+v, want unchanged %+v", artifacts[1].Provenance, wantS)
	}
	if len(artifacts[0].AssociatedEventIDs) != 0 || len(artifacts[1].AssociatedEventIDs) != 0 {
		t.Errorf("non-owner events must not be associated: %v, %v", artifacts[0].AssociatedEventIDs, artifacts[1].AssociatedEventIDs)
	}
}

func TestTransfersUnresolvableSpoilsRecordNothing(t *testing.T) {
	artifacts := []Artifact{settlementArtifact("artifact-1", "Ironforge")}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", TargetSettlement: "Ironforge", Description: "conquered by no one"},
		{Year: 6, Category: "Raid", TargetSettlement: "Ironforge", Description: "raided by no one"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// The match rule needs only the target, but without the aggressor
	// (SettlementName) a spoils transfer has no destination: the events are
	// treated as terminating nothing.
	for _, e := range events {
		if e.ArtifactID != "" {
			t.Errorf("event %q ArtifactID = %q, want empty", e.ID, e.ArtifactID)
		}
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("provenance = %+v, want unchanged %+v", artifacts[0].Provenance, want)
	}
	if len(artifacts[0].AssociatedEventIDs) != 0 {
		t.Errorf("associatedEventIDs = %v, want none", artifacts[0].AssociatedEventIDs)
	}
}

func TestTransfersDeterministic(t *testing.T) {
	artifacts := []Artifact{
		figureArtifact("artifact-1", "Deepcrest-1"),
		settlementArtifact("artifact-2", "Ironforge"),
	}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "took Ironforge"},
		{Year: 8, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
		{Year: 9, Category: "Raid", SettlementName: "Haven", TargetSettlement: "Blackgate", Description: "raided"},
	}
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 8},
		{ID: "Deepcrest-2", Settlement: "Deepcrest", BirthYear: 10, Parents: []string{"Deepcrest-1"}},
	}}

	run := func() ([]Artifact, []simulation.Event, error) {
		arts := make([]Artifact, len(artifacts))
		copy(arts, artifacts)
		evs := make([]simulation.Event, len(events))
		copy(evs, events)
		err := PostProcess(arts, evs, SignificanceContext{}, ctx)
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
		t.Errorf("transfer provenance differs across identical runs:\nfirst=%+v\nsecond=%+v", firstArts, secondArts)
	}
	if !reflect.DeepEqual(firstEvents, secondEvents) {
		t.Errorf("event IDs differ across identical runs:\nfirst=%+v\nsecond=%+v", firstEvents, secondEvents)
	}
}

func TestTransfersHeirTieBreakUsesSliceOrder(t *testing.T) {
	artifacts := []Artifact{figureArtifact("artifact-1", "Deepcrest-1")}
	events := []simulation.Event{
		{Year: 30, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
	}
	// Twins: identical birth years, so the earlier index in the figures slice
	// (world-state order) must inherit.
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 30},
		{ID: "Deepcrest-2", Settlement: "Deepcrest", BirthYear: 10, Parents: []string{"Deepcrest-1"}},
		{ID: "Deepcrest-3", Settlement: "Deepcrest", BirthYear: 10, Parents: []string{"Deepcrest-1"}},
	}}

	if err := PostProcess(artifacts, events, SignificanceContext{}, ctx); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	if owner := CurrentOwner(artifacts[0]); owner.Kind != "figure" || owner.ID != "Deepcrest-2" {
		t.Errorf("current owner = %+v, want (figure, Deepcrest-2) (earlier slice index)", owner)
	}
}

func TestTransfersPowersFollowArtifact(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Blade of Deepcrest",
		Type:               "weapon",
		SignificanceSource: "historical",
		Status:             "held",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		},
		Powers: []Power{CombatPower{Base: 2}},
	}}
	events := []simulation.Event{
		{Year: 12, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
	}
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 12},
		{ID: "Deepcrest-2", Settlement: "Deepcrest", BirthYear: 5, Parents: []string{"Deepcrest-1"}},
	}}

	if err := PostProcess(artifacts, events, SignificanceContext{}, ctx); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// Powers are intrinsic to the artifact, not the owner (spec 7.7): a
	// transfer mutates provenance only.
	if owner := CurrentOwner(artifacts[0]); owner.Kind != "figure" || owner.ID != "Deepcrest-2" {
		t.Fatalf("current owner = %+v, want (figure, Deepcrest-2)", owner)
	}
	if len(artifacts[0].Powers) != 1 {
		t.Fatalf("power count = %d, want 1", len(artifacts[0].Powers))
	}
	p := artifacts[0].Powers[0]
	if p.Type() != "combat" || p.BaseMagnitude() != 2 {
		t.Errorf("power = (%s, %d), want (combat, 2)", p.Type(), p.BaseMagnitude())
	}
	if got := p.EffectiveMagnitude(artifacts[0].SignificanceScore); got != 2 {
		t.Errorf("effective magnitude = %d, want 2", got)
	}
}

func TestTransfersZeroValueContextRecordsLost(t *testing.T) {
	artifacts := []Artifact{figureArtifact("artifact-1", "Deepcrest-1")}
	events := []simulation.Event{
		{Year: 20, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	// With no transfer context there is no heir and no treasury to resolve:
	// the transfer is still recorded, with the lost fallback.
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 20, Owner: Owner{Kind: "lost"}, EventID: "event-20-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(artifacts[0].Provenance, want) {
		t.Errorf("provenance = %+v, want %+v", artifacts[0].Provenance, want)
	}
}

// TestEmergencePassTransferChain exercises the full acceptance path for
// emergent artifacts: birth on a Conquest draw, then a Raid and a Conquest
// transferring the born artifact onward. Draw outcomes are the verified
// seed-1 lane (ev1 weapon PASS "Warhammer"); the later transfer events carry
// the born artifact's ID, so no further draws happen.
func TestEmergencePassTransferChain(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "spoils"},
		{Year: 8, Category: "Raid", SettlementName: "Haven", TargetSettlement: "Blackgate", Description: "raided"},
		{Year: 12, Category: "Conquest", SettlementName: "Rift", TargetSettlement: "Haven", Description: "conquered"},
	}

	run := func() ([]Artifact, []simulation.Event, error) {
		evs := make([]simulation.Event, len(events))
		copy(evs, events)
		arts, evs, err := EmergencePass(nil, evs, nil, SignificanceContext{}, TransferContext{}, artifactsRNG(1))
		return arts, evs, err
	}

	got, gotEvents, err := run()
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}

	a := got[0]
	if a.ID != "artifact-Blackgate-0" || a.Name != "Warhammer of Blackgate" {
		t.Errorf("born artifact = (%q, %q), want (artifact-Blackgate-0, Warhammer of Blackgate)", a.ID, a.Name)
	}
	wantProv := []ProvenanceEntry{
		{Year: 5, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-5-0", EventType: "Conquest"},
		{Year: 8, Owner: Owner{Kind: "settlement", ID: "Haven"}, EventID: "event-8-0", EventType: "Raid"},
		{Year: 12, Owner: Owner{Kind: "settlement", ID: "Rift"}, EventID: "event-12-0", EventType: "Conquest"},
	}
	if !reflect.DeepEqual(a.Provenance, wantProv) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, wantProv)
	}
	if wantIDs := []string{"event-5-0", "event-8-0", "event-12-0"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("AssociatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}
	if owner := CurrentOwner(a); owner.Kind != "settlement" || owner.ID != "Rift" {
		t.Errorf("current owner = %+v, want (settlement, Rift)", owner)
	}
	if gotEvents[1].ArtifactID != a.ID {
		t.Errorf("raid event ArtifactID = %q, want %q (transfer attach, no draw)", gotEvents[1].ArtifactID, a.ID)
	}
	if gotEvents[2].ArtifactID != a.ID {
		t.Errorf("conquest event ArtifactID = %q, want %q (transfer attach, no draw)", gotEvents[2].ArtifactID, a.ID)
	}

	// Same seed and inputs => identical transfer sequence.
	second, secondEvents, err := run()
	if err != nil {
		t.Fatalf("second EmergencePass: %v", err)
	}
	if !reflect.DeepEqual(got, second) {
		t.Errorf("transfer sequence differs across identical runs:\nfirst=%+v\nsecond=%+v", got, second)
	}
	if !reflect.DeepEqual(gotEvents, secondEvents) {
		t.Errorf("events differ across identical runs:\nfirst=%+v\nsecond=%+v", gotEvents, secondEvents)
	}
}

// TestEmergencePassDeathTransferChain covers the death path for born
// artifacts: a Discovery births an artifact to a figure, then the figure's
// death transfers it to the heir from the transfer context.
func TestEmergencePassDeathTransferChain(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "Deepcrest-3", SettlementName: "Deepcrest", Description: "found"},
		{Year: 9, Category: "Death", FigureID: "Deepcrest-3", Description: "owner dies"},
	}
	figures := []FigureContext{
		{ID: "Deepcrest-3", Settlement: "Deepcrest"},
	}
	transfers := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-3", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 9},
		{ID: "Deepcrest-4", Settlement: "Deepcrest", BirthYear: 4, Parents: []string{"Deepcrest-3"}},
	}}

	got, evs, err := EmergencePass(nil, events, figures, SignificanceContext{}, transfers, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}

	a := got[0]
	wantProv := []ProvenanceEntry{
		{Year: 5, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-5-0", EventType: "Discovery"},
		{Year: 9, Owner: Owner{Kind: "figure", ID: "Deepcrest-4"}, EventID: "event-9-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(a.Provenance, wantProv) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, wantProv)
	}
	if owner := CurrentOwner(a); owner.Kind != "figure" || owner.ID != "Deepcrest-4" {
		t.Errorf("current owner = %+v, want (figure, Deepcrest-4)", owner)
	}
	if evs[1].ArtifactID != a.ID {
		t.Errorf("death event ArtifactID = %q, want %q", events[1].ArtifactID, a.ID)
	}
}
