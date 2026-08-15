package artifact

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// Destruction draws consume the artifacts lane in the fixed stream-order
// walk (see destruction.go). The fixtures below use verified PCG(seed, 1)
// lane values:
//
//   - seed 1: draw1=99, draw2=10, draw3=86, draw4=91 — all FAIL
//   - seed 2: draw1=5,  draw2=5            — PASS
//   - seed 13: draw1=1 PASS, draw2=76 FAIL, draw3=22 FAIL
//   - seed 55 (born artifact): type=1 armor, gate=9 (birth), name=1
//     "Cuirass", destruction draw4=4 — PASS
//   - seed 20 (born artifact): type=4 armor, gate=3 (birth), name=2
//     "Shield", destruction draw4=69 — FAIL

func TestDestructionConquestDrawPass(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Tome of Ironforge",
		Type:               "tome",
		SignificanceSource: "historical",
		Status:             "held",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		},
		Powers: []Power{NarrativePower{Effect: "reveals hidden knowledge", Source: "intrinsic"}},
	}}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "Ironforge falls"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(2)); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.Status != "destroyed" {
		t.Errorf("Status = %q, want destroyed", a.Status)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		// Destruction provenance: the owner at destruction, the natural
		// event's ID and category. No transfer entry is recorded.
		{Year: 5, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-5-0", EventType: "Conquest"},
	}
	if !reflect.DeepEqual(a.Provenance, want) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, want)
	}
	if len(a.Powers) != 0 {
		t.Errorf("Powers = %v, want cleared (spec 7.7: powers vanish on destruction)", a.Powers)
	}
	// The natural event IS the lifecycle event (spec 6.1): it carries the
	// ArtifactID and is associated, and no synthetic event is minted.
	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("destruction event ArtifactID = %q, want artifact-1", events[0].ArtifactID)
	}
	if wantIDs := []string{"event-5-0"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("AssociatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}
	if len(events) != 1 || events[0].Category != "Conquest" {
		t.Errorf("events mutated: %+v, want the single natural Conquest event unchanged", events)
	}
	if got := DestructionYear(a); got != 5 {
		t.Errorf("DestructionYear = %d, want 5", got)
	}
}

func TestDestructionConquestDrawFailTransfers(t *testing.T) {
	artifacts := []Artifact{settlementArtifact("artifact-1", "Ironforge")}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "Ironforge falls"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(1)); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.Status != "held" {
		t.Errorf("Status = %q, want held (draw failed, normal transfer)", a.Status)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		{Year: 5, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-5-0", EventType: "Conquest"},
	}
	if !reflect.DeepEqual(a.Provenance, want) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, want)
	}
	if got := DestructionYear(a); got != 0 {
		t.Errorf("DestructionYear = %d, want 0 (not destroyed)", got)
	}
}

func TestDestructionDeathDrawPass(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Crown of Deepcrest",
		Type:               "crown",
		SignificanceSource: "historical",
		Status:             "held",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		},
		Powers: []Power{InfluencePower{Base: 5, Source: "intrinsic"}},
	}}
	events := []simulation.Event{
		{Year: 30, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies in war"},
	}
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 30},
		{ID: "Deepcrest-2", Settlement: "Deepcrest", BirthYear: 5, Parents: []string{"Deepcrest-1"}},
	}}

	if err := PostProcess(artifacts, events, SignificanceContext{}, ctx, artifactsRNG(2)); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.Status != "destroyed" {
		t.Errorf("Status = %q, want destroyed", a.Status)
	}
	// The war-death proxy: every Death of the owner figure triggers the
	// draw; the heir inherits nothing on a pass.
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 30, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-30-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(a.Provenance, want) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, want)
	}
	if len(a.Powers) != 0 {
		t.Errorf("Powers = %v, want cleared (spec 7.7)", a.Powers)
	}
	if events[0].ArtifactID != "artifact-1" {
		t.Errorf("death event ArtifactID = %q, want artifact-1", events[0].ArtifactID)
	}
	if wantIDs := []string{"event-30-0"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("AssociatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}
}

func TestDestructionDeathDrawFailInherits(t *testing.T) {
	artifacts := []Artifact{figureArtifact("artifact-1", "Deepcrest-1")}
	events := []simulation.Event{
		{Year: 30, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies"},
	}
	ctx := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 30},
		{ID: "Deepcrest-2", Settlement: "Deepcrest", BirthYear: 5, Parents: []string{"Deepcrest-1"}},
	}}

	if err := PostProcess(artifacts, events, SignificanceContext{}, ctx, artifactsRNG(1)); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.Status != "held" {
		t.Errorf("Status = %q, want held (draw failed, normal inheritance)", a.Status)
	}
	if owner := CurrentOwner(a); owner.Kind != "figure" || owner.ID != "Deepcrest-2" {
		t.Errorf("current owner = %+v, want (figure, Deepcrest-2)", owner)
	}
}

func TestDestructionRaidNeverDraws(t *testing.T) {
	// Raid is plunder, not destruction (spec 6.6), and consumes no lane
	// values. With seed 13 draw1=1 PASS and draw2=76 FAIL, so the artifact
	// is destroyed at the year-8 Conquest only because the year-5 Raid did
	// not take a draw: a consuming Raid would have pushed the Conquest draw
	// to position 2 (FAIL) and the artifact would merely have transferred.
	artifacts := []Artifact{settlementArtifact("artifact-1", "Ironforge")}
	events := []simulation.Event{
		{Year: 5, Category: "Raid", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "raided"},
		{Year: 8, Category: "Conquest", SettlementName: "Haven", TargetSettlement: "Blackgate", Description: "conquered"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(13)); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.Status != "destroyed" {
		t.Errorf("Status = %q, want destroyed at the year-8 Conquest (Raid consumed no draw)", a.Status)
	}
	want := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		{Year: 5, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-5-0", EventType: "Raid"},
		{Year: 8, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-8-0", EventType: "Conquest"},
	}
	if !reflect.DeepEqual(a.Provenance, want) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, want)
	}
}

func TestDestructionTerminalExitsFurtherLifecycleProcessing(t *testing.T) {
	// Seed 13: draw1=1 PASS (A destroyed at year 5), draw2=76 FAIL (B
	// transferred to Blackgate), draw3=22 FAIL (B transferred to Haven).
	artifacts := []Artifact{
		settlementArtifact("artifact-1", "Ironforge"),
		settlementArtifact("artifact-2", "Ironforge"),
	}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "Ironforge falls"},
		{Year: 8, Category: "Conquest", SettlementName: "Haven", TargetSettlement: "Blackgate", Description: "Blackgate falls"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(13)); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.Status != "destroyed" {
		t.Fatalf("artifacts[0] Status = %q, want destroyed", a.Status)
	}
	wantA := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		{Year: 5, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-5-0", EventType: "Conquest"},
	}
	if !reflect.DeepEqual(a.Provenance, wantA) {
		t.Errorf("destroyed artifact provenance = %+v, want %+v (no further entries)", a.Provenance, wantA)
	}
	// The destroyed artifact exits all further lifecycle processing: the
	// year-8 conquest of its former city neither transfers nor associates
	// it, and the living artifact is the carrier instead.
	if wantIDs := []string{"event-5-0"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("destroyed artifact AssociatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}
	if events[1].ArtifactID != "artifact-2" {
		t.Errorf("year-8 event ArtifactID = %q, want artifact-2 (destroyed artifact must not be the carrier)", events[1].ArtifactID)
	}

	b := artifacts[1]
	if b.Status != "held" {
		t.Errorf("artifacts[1] Status = %q, want held", b.Status)
	}
	wantB := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		{Year: 5, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-5-0", EventType: "Conquest"},
		{Year: 8, Owner: Owner{Kind: "settlement", ID: "Haven"}, EventID: "event-8-0", EventType: "Conquest"},
	}
	if !reflect.DeepEqual(b.Provenance, wantB) {
		t.Errorf("living artifact provenance = %+v, want %+v", b.Provenance, wantB)
	}
	if wantIDs := []string{"event-5-0", "event-8-0"}; !reflect.DeepEqual(b.AssociatedEventIDs, wantIDs) {
		t.Errorf("living artifact AssociatedEventIDs = %v, want %v", b.AssociatedEventIDs, wantIDs)
	}
}

func TestDestructionNilRNGDisablesDraws(t *testing.T) {
	artifacts := []Artifact{settlementArtifact("artifact-1", "Ironforge")}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "Ironforge falls"},
	}

	// The zero-value contract: a nil lane disables destruction draws, so
	// the walk records transfers only (determinism is preserved when the
	// pipeline always supplies the lane).
	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, nil); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}
	if artifacts[0].Status != "held" {
		t.Errorf("Status = %q, want held (nil lane must not destroy)", artifacts[0].Status)
	}
	if owner := CurrentOwner(artifacts[0]); owner.Kind != "settlement" || owner.ID != "Blackgate" {
		t.Errorf("current owner = %+v, want (settlement, Blackgate)", owner)
	}
}

func TestDestructionBornArtifactSecondWalkPass(t *testing.T) {
	// Seed 55 second-walk lane: type=1 armor, gate=9 (birth), name=1
	// (Cuirass), destruction draw=4 -> PASS.
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "Deepcrest-3", SettlementName: "Deepcrest", Description: "found"},
		{Year: 9, Category: "Death", FigureID: "Deepcrest-3", Description: "owner dies in war"},
	}
	figures := []FigureContext{
		{ID: "Deepcrest-3", Settlement: "Deepcrest"},
	}
	transfers := TransferContext{Figures: []FigureLifecycle{
		{ID: "Deepcrest-3", Settlement: "Deepcrest", BirthYear: 1, DeathYear: 9},
		{ID: "Deepcrest-4", Settlement: "Deepcrest", BirthYear: 4, Parents: []string{"Deepcrest-3"}},
	}}

	got, evs, err := EmergencePass(nil, events, figures, SignificanceContext{}, transfers, artifactsRNG(55))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}

	a := got[0]
	if a.Name != "Cuirass of Deepcrest" || a.Type != "armor" {
		t.Errorf("born artifact = (%q, %q), want (Cuirass of Deepcrest, armor)", a.Name, a.Type)
	}
	if a.Status != "destroyed" {
		t.Errorf("Status = %q, want destroyed (second-walk draw passes)", a.Status)
	}
	wantProv := []ProvenanceEntry{
		{Year: 5, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-5-0", EventType: "Discovery"},
		{Year: 9, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-9-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(a.Provenance, wantProv) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, wantProv)
	}
	if len(a.Powers) != 0 {
		t.Errorf("Powers = %v, want cleared (spec 7.7)", a.Powers)
	}
	if wantIDs := []string{"event-5-0", "event-9-0"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("AssociatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}
	// The pass returns the extended stream: fake-discovery prepends minted
	// events into a fresh slice, so assertions must read the returned slice,
	// not the caller's original.
	if evs[1].ArtifactID != a.ID {
		t.Errorf("death event ArtifactID = %q, want %q", evs[1].ArtifactID, a.ID)
	}
	// No synthetic event is minted for the born artifact's destruction.
	if len(evs) != 2 {
		t.Errorf("event count = %d, want 2 (no synthetic events)", len(evs))
	}
	for _, e := range evs {
		if e.Category == "ArtifactDestruction" {
			t.Errorf("synthetic destruction event minted: %+v", e)
		}
	}
}

func TestDestructionBornArtifactSecondWalkFail(t *testing.T) {
	// Seed 20 second-walk lane: type=4 armor, gate=3 (birth), name=2
	// (Shield), destruction draw=69 -> FAIL (normal death transfer).
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

	got, _, err := EmergencePass(nil, events, figures, SignificanceContext{}, transfers, artifactsRNG(20))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}

	a := got[0]
	if a.Status != "held" {
		t.Errorf("Status = %q, want held (draw failed)", a.Status)
	}
	wantProv := []ProvenanceEntry{
		{Year: 5, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-5-0", EventType: "Discovery"},
		{Year: 9, Owner: Owner{Kind: "figure", ID: "Deepcrest-4"}, EventID: "event-9-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(a.Provenance, wantProv) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, wantProv)
	}
	if len(a.Powers) != 1 {
		t.Errorf("Powers = %v, want the intrinsic armor power kept", a.Powers)
	}
}

func TestDestructionDeterministic(t *testing.T) {
	artifacts := []Artifact{settlementArtifact("artifact-1", "Ironforge")}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "Ironforge falls"},
	}

	run := func(seed uint64) ([]Artifact, []simulation.Event, error) {
		arts := make([]Artifact, len(artifacts))
		copy(arts, artifacts)
		evs := make([]simulation.Event, len(events))
		copy(evs, events)
		err := PostProcess(arts, evs, SignificanceContext{}, TransferContext{}, artifactsRNG(seed))
		return arts, evs, err
	}

	firstArts, firstEvents, err := run(2)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	secondArts, secondEvents, err := run(2)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	artJSON := func(arts []Artifact) []byte {
		b, err := json.Marshal(arts)
		if err != nil {
			t.Fatalf("marshal artifacts: %v", err)
		}
		return b
	}
	if !reflect.DeepEqual(artJSON(firstArts), artJSON(secondArts)) {
		t.Errorf("identical seeds must produce identical destruction outcomes:\nfirst=%+v\nsecond=%+v", firstArts, secondArts)
	}
	if !reflect.DeepEqual(firstEvents, secondEvents) {
		t.Errorf("events differ across identical runs:\nfirst=%+v\nsecond=%+v", firstEvents, secondEvents)
	}

	// A different seed draws differently: seed 1 fails, seed 2 passes.
	otherArts, _, err := run(1)
	if err != nil {
		t.Fatalf("other seed: %v", err)
	}
	if otherArts[0].Status != "held" || firstArts[0].Status != "destroyed" {
		t.Errorf("different seeds must draw differently: seed1 = %q, seed2 = %q", otherArts[0].Status, firstArts[0].Status)
	}
}

func TestDestructionProbabilityIsLow(t *testing.T) {
	// Statistical sanity on many deterministic seeds: the pass probability
	// is 10%, so across 2000 independent single-draw runs the destruction
	// count must sit in a band far from both 0% and 100% (the assertion is
	// deterministic — the lane values are fixed per seed).
	const trials = 2000
	destroyed := 0
	for seed := uint64(1); seed <= trials; seed++ {
		arts := []Artifact{settlementArtifact("artifact-1", "Ironforge")}
		evs := []simulation.Event{
			{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "falls"},
		}
		if err := PostProcess(arts, evs, SignificanceContext{}, TransferContext{}, artifactsRNG(seed)); err != nil {
			t.Fatalf("PostProcess(seed %d): %v", seed, err)
		}
		if arts[0].Status == "destroyed" {
			destroyed++
		}
	}
	if destroyed < 100 || destroyed > 300 {
		t.Errorf("destructions in %d seeds = %d, want a low-probability band around 200 (10%% per draw)", trials, destroyed)
	}
	if destroyed >= trials/2 {
		t.Errorf("destructions in %d seeds = %d, must stay well below 50%%", trials, destroyed)
	}
}

func TestDestructionSignificanceStopsAtDestructionYear(t *testing.T) {
	// Seed 2: the death draw (lane position 1) passes. The artifact accrues
	// figure reputation only for held years before the destruction year, and
	// events after it contribute nothing.
	artifacts := []Artifact{figureArtifact("artifact-1", "Deepcrest-1")}
	events := []simulation.Event{
		{Year: 3, Category: "War", ArtifactID: "artifact-1", Description: "war"},
		{Year: 5, Category: "Death", FigureID: "Deepcrest-1", Description: "owner dies in war"},
		{Year: 7, Category: "War", ArtifactID: "artifact-1", Description: "war after the end"},
	}
	sig := SignificanceContext{FigureReputation: map[string]map[int]int{
		"Deepcrest-1": {1: 2, 4: 3, 6: 5},
	}}

	if err := PostProcess(artifacts, events, sig, TransferContext{}, artifactsRNG(2)); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.Status != "destroyed" {
		t.Fatalf("Status = %q, want destroyed", a.Status)
	}
	// 3 (War, year 3) + 2 + 3 (reputation for held years 1-4) = 8. The
	// year-6 reputation delta and the year-7 War contribute nothing; the
	// destruction event itself (Death, year 5) carries no weight.
	if a.SignificanceScore != 8 {
		t.Errorf("SignificanceScore = %d, want 8", a.SignificanceScore)
	}
	if !a.IsSignificant {
		t.Error("IsSignificant = false, want true (latch before destruction is preserved)")
	}
	if a.PivotalEventID != "event-3-0" || a.SignificanceYear != 3 {
		t.Errorf("pivotal = (%q, %d), want (event-3-0, 3)", a.PivotalEventID, a.SignificanceYear)
	}
	// The year-7 event must not associate: destroyed artifacts exit all
	// further lifecycle processing.
	if wantIDs := []string{"event-3-0", "event-5-0"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("AssociatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}
}

func TestDestructionSignificanceSettlementLumpSumOnce(t *testing.T) {
	// The destruction provenance entry records the end of tenure, not an
	// acquisition: a settlement owner's size-class lump sum is awarded only
	// at the acquisition year (year 1), never at the destruction year.
	artifacts := []Artifact{settlementArtifact("artifact-1", "Ironforge")}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "Ironforge falls"},
	}
	sig := SignificanceContext{SettlementClass: map[string]string{"Ironforge": "City"}}

	if err := PostProcess(artifacts, events, sig, TransferContext{}, artifactsRNG(2)); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.Status != "destroyed" {
		t.Fatalf("Status = %q, want destroyed", a.Status)
	}
	// 2 (City lump at acquisition) + 3 (Conquest weight of the destruction
	// event itself, spec 6.1) = 5. No lump at the destruction year; nothing
	// after it.
	if a.SignificanceScore != 5 {
		t.Errorf("SignificanceScore = %d, want 5", a.SignificanceScore)
	}
	if !a.IsSignificant {
		t.Error("IsSignificant = false, want true (Conquest weight crosses the threshold)")
	}
	if a.PivotalEventID != "event-5-0" || a.SignificanceYear != 5 {
		t.Errorf("pivotal = (%q, %d), want (event-5-0, 5)", a.PivotalEventID, a.SignificanceYear)
	}
}

func TestDestructionSignificanceEventTruncationGuard(t *testing.T) {
	// The stream walk never records event contributions after the
	// destruction year (PostProcess skips destroyed carriers), so this
	// exercises the defensive truncation in buildContributions directly:
	// the destruction event's own weight counts (spec 6.1), anything after
	// the destruction year must be dropped.
	a := Artifact{
		ID:     "artifact-1",
		Status: "destroyed",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
			{Year: 5, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-5-0", EventType: "Death"},
		},
	}
	events := []significanceEvent{
		{year: 3, weight: 3, eventID: "event-3-0"},
		{year: 5, weight: 3, eventID: "event-5-0"},
		{year: 8, weight: 3, eventID: "event-8-0"},
	}

	contribs := buildContributions(&a, events, 10, SignificanceContext{})
	for _, c := range contribs {
		if c.year > 5 {
			t.Errorf("contribution at year %d must be truncated (destruction year 5): %+v", c.year, c)
		}
	}
	if len(contribs) != 2 || contribs[0].year != 3 || contribs[1].year != 5 {
		t.Errorf("contributions = %+v, want exactly the year-3 and year-5 events", contribs)
	}
}
