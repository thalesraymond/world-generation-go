package artifact

import (
	"encoding/json"
	randv2 "math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// artifactsRNG returns the artifacts RNG lane for a seed, mirroring the
// engine's PCG construction used by the CLI wiring.
func artifactsRNG(seed uint64) *randv2.Rand {
	return randv2.New(randv2.NewPCG(seed, 1))
}

// Draw sequences below are verified against the fixed lane walk (type draw,
// gate draw, then a name draw on birth) for PCG(seed, 1):
//   - seed 1: ev1 weapon PASS "Warhammer"; ev2 armor fail; ev3 jewelry PASS
//   - seed 2: ev1 relic PASS "Idol"; ev2 relic PASS "Idol"; ev3 relic fail
//   - seed 3: ev1 armor fail; fallback name draw "Aegis"

func TestEmergencePassDiscoveryBirthsArtifact(t *testing.T) {
	events := []simulation.Event{
		{Year: 12, Category: "Discovery", FigureID: "Deepcrest-3", SettlementName: "Deepcrest", Description: "unearths a hoard"},
	}
	figures := []FigureContext{
		{ID: "Deepcrest-3", Settlement: "Deepcrest"},
	}

	got, err := EmergencePass(nil, events, figures, SignificanceContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}
	a := got[0]
	if a.ID != "artifact-Deepcrest-0" {
		t.Errorf("ID = %q, want artifact-Deepcrest-0", a.ID)
	}
	if a.Name != "Warhammer of Deepcrest" {
		t.Errorf("Name = %q, want Warhammer of Deepcrest", a.Name)
	}
	if a.Type != "weapon" {
		t.Errorf("Type = %q, want weapon", a.Type)
	}
	if a.SignificanceSource != "historical" {
		t.Errorf("SignificanceSource = %q, want historical", a.SignificanceSource)
	}
	if a.Status != "held" {
		t.Errorf("Status = %q, want held (created at birth, first transfer applied)", a.Status)
	}
	if a.IsSignificant || a.SignificanceScore != 0 {
		t.Errorf("Significance = (%v, %d), want (false, 0)", a.IsSignificant, a.SignificanceScore)
	}
	wantProv := []ProvenanceEntry{
		{Year: 12, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-12-0", EventType: "Discovery"},
	}
	if !reflect.DeepEqual(a.Provenance, wantProv) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, wantProv)
	}
	if wantIDs := []string{"event-12-0"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("AssociatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}
	if events[0].ArtifactID != "artifact-Deepcrest-0" {
		t.Errorf("birth event ArtifactID = %q, want artifact-Deepcrest-0", events[0].ArtifactID)
	}
	if owner := CurrentOwner(a); owner.Kind != "figure" || owner.ID != "Deepcrest-3" {
		t.Errorf("current owner = %+v, want (figure, Deepcrest-3)", owner)
	}
}

func TestEmergencePassConquestSpoils(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "Blackgate conquered Ironforge"},
	}

	got, err := EmergencePass(nil, events, nil, SignificanceContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}
	a := got[0]
	if a.ID != "artifact-Blackgate-0" {
		t.Errorf("ID = %q, want artifact-Blackgate-0", a.ID)
	}
	wantProv := []ProvenanceEntry{
		{Year: 5, Owner: Owner{Kind: "settlement", ID: "Blackgate"}, EventID: "event-5-0", EventType: "Conquest"},
	}
	if !reflect.DeepEqual(a.Provenance, wantProv) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, wantProv)
	}
	if owner := CurrentOwner(a); owner.Kind != "settlement" || owner.ID != "Blackgate" {
		t.Errorf("current owner = %+v, want (settlement, Blackgate)", owner)
	}
	if events[0].ArtifactID != "artifact-Blackgate-0" {
		t.Errorf("birth event ArtifactID = %q, want artifact-Blackgate-0", events[0].ArtifactID)
	}
}

func TestEmergencePassRarityGateFailNoBirth(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "Deepcrest-1", SettlementName: "Deepcrest", Description: "found"},
	}
	figures := []FigureContext{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", Reputation: []ReputationDelta{{Year: 5, Delta: 1, Event: "Discovery"}}},
	}

	got, err := EmergencePass(nil, events, figures, SignificanceContext{}, artifactsRNG(3))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("artifact count = %d, want 0 (gate failed, reputation below threshold)", len(got))
	}
	if events[0].ArtifactID != "" {
		t.Errorf("birth event ArtifactID = %q, want empty", events[0].ArtifactID)
	}
}

func TestEmergencePassFallbackBackdatesProvenance(t *testing.T) {
	events := []simulation.Event{
		{Year: 15, Category: "Discovery", FigureID: "D-1", SettlementName: "Deepcrest", Description: "found"},
	}
	figures := []FigureContext{
		{
			ID:         "D-1",
			Settlement: "Deepcrest",
			Reputation: []ReputationDelta{
				{Year: 8, Delta: 4, Event: "Raid"},
				{Year: 9, Delta: 3, Event: "Discovery"},
				{Year: 10, Delta: 4, Event: "Raid"},
			},
		},
	}

	got, err := EmergencePass(nil, events, figures, SignificanceContext{}, artifactsRNG(3))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1 (fallback birth)", len(got))
	}
	a := got[0]
	if a.ID != "artifact-Deepcrest-0" {
		t.Errorf("ID = %q, want artifact-Deepcrest-0", a.ID)
	}
	if a.Name != "Aegis of Deepcrest" {
		t.Errorf("Name = %q, want Aegis of Deepcrest", a.Name)
	}
	if a.Type != "armor" {
		t.Errorf("Type = %q, want armor", a.Type)
	}
	if a.SignificanceSource != "historical" || a.Status != "held" {
		t.Errorf("source/status = (%q, %q), want (historical, held)", a.SignificanceSource, a.Status)
	}
	// Reputation crosses 10 at year 10 (4+3+4): provenance is backdated to
	// the crossing entry, not to the qualifying event at year 15.
	wantProv := []ProvenanceEntry{
		{Year: 10, Owner: Owner{Kind: "figure", ID: "D-1"}, EventID: "", EventType: "Raid"},
	}
	if !reflect.DeepEqual(a.Provenance, wantProv) {
		t.Errorf("Provenance = %+v, want %+v", a.Provenance, wantProv)
	}
	if len(a.AssociatedEventIDs) != 0 {
		t.Errorf("AssociatedEventIDs = %v, want none (birth is not a stream event)", a.AssociatedEventIDs)
	}
	if events[0].ArtifactID != "" {
		t.Errorf("qualifying event ArtifactID = %q, want empty", events[0].ArtifactID)
	}
}

func TestEmergencePassFallbackRequiresThreshold(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "D-1", SettlementName: "Deepcrest", Description: "found"},
	}
	figures := []FigureContext{
		{ID: "D-1", Settlement: "Deepcrest", Reputation: []ReputationDelta{{Year: 5, Delta: 9, Event: "Raid"}}},
	}

	got, err := EmergencePass(nil, events, figures, SignificanceContext{}, artifactsRNG(3))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("artifact count = %d, want 0 (reputation 9 below threshold 10)", len(got))
	}
}

func TestEmergencePassIndexAndNameUniquePerOrigin(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "took Ironforge"},
		{Year: 7, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Haven", Description: "took Haven"},
	}

	got, err := EmergencePass(nil, events, nil, SignificanceContext{}, artifactsRNG(2))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("artifact count = %d, want 2", len(got))
	}
	if got[0].ID != "artifact-Blackgate-0" || got[1].ID != "artifact-Blackgate-1" {
		t.Errorf("IDs = (%q, %q), want (artifact-Blackgate-0, artifact-Blackgate-1)", got[0].ID, got[1].ID)
	}
	// Same type and same drawn word twice: the second name is disambiguated
	// so exported note filenames stay unique.
	if got[0].Name != "Idol of Blackgate" || got[1].Name != "Idol of Blackgate II" {
		t.Errorf("Names = (%q, %q), want (Idol of Blackgate, Idol of Blackgate II)", got[0].Name, got[1].Name)
	}
}

func TestEmergencePassSkipsEventsAlreadyInvolvingArtifact(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-ruin-0", Name: "Relic of Old Keep", Status: "lost"},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", ArtifactID: "artifact-ruin-0", Description: "spoils"},
	}

	got, err := EmergencePass(artifacts, events, nil, SignificanceContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1 (no emergence for events that already carry an artifact)", len(got))
	}
	if events[0].ArtifactID != "artifact-ruin-0" {
		t.Errorf("event ArtifactID = %q, want artifact-ruin-0 unchanged", events[0].ArtifactID)
	}
}

func TestEmergencePassSkipsNonQualifyingEvents(t *testing.T) {
	events := []simulation.Event{
		{Year: 1, Category: "Conquest", SettlementName: "Blackgate", Description: "sought conquest in vain"},
		{Year: 2, Category: "Raid", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "raided"},
		{Year: 3, Category: "War", SettlementName: "Blackgate", Description: "war"},
		{Year: 4, Category: "Discovery", Description: "found by no one"},
		{Year: 5, Category: "Economy", SettlementName: "Blackgate", Description: "trade"},
	}

	got, err := EmergencePass(nil, events, nil, SignificanceContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("artifact count = %d, want 0", len(got))
	}
}

func TestEmergencePassBornArtifactJoinsProvenanceWalk(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "Deepcrest-3", SettlementName: "Deepcrest", Description: "found"},
		{Year: 9, Category: "Death", FigureID: "Deepcrest-3", Description: "owner dies"},
		{Year: 10, Category: "Death", FigureID: "Deepcrest-2", Description: "unrelated death"},
	}
	figures := []FigureContext{
		{ID: "Deepcrest-3", Settlement: "Deepcrest"},
	}

	got, err := EmergencePass(nil, events, figures, SignificanceContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}
	// The born artifact's owner dies at year 9: the death event is attached
	// and associated like any owned artifact.
	if events[1].ArtifactID != "artifact-Deepcrest-0" {
		t.Errorf("owner death ArtifactID = %q, want artifact-Deepcrest-0", events[1].ArtifactID)
	}
	if events[2].ArtifactID != "" {
		t.Errorf("unrelated death ArtifactID = %q, want empty", events[2].ArtifactID)
	}
	want := []string{"event-5-0", "event-9-0"}
	if !reflect.DeepEqual(got[0].AssociatedEventIDs, want) {
		t.Errorf("AssociatedEventIDs = %v, want %v", got[0].AssociatedEventIDs, want)
	}
}

func TestEmergencePassFallbackOncePerFigure(t *testing.T) {
	events := []simulation.Event{
		{Year: 10, Category: "Discovery", FigureID: "D-1", SettlementName: "Deepcrest", Description: "found"},
		{Year: 11, Category: "Discovery", FigureID: "D-1", SettlementName: "Deepcrest", Description: "found again"},
	}
	figures := []FigureContext{
		{ID: "D-1", Settlement: "Deepcrest", Reputation: []ReputationDelta{{Year: 3, Delta: 12, Event: "Raid"}}},
	}

	got, err := EmergencePass(nil, events, figures, SignificanceContext{}, artifactsRNG(4))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1 (fallback births at most one artifact per figure)", len(got))
	}
	if got[0].ID != "artifact-Deepcrest-0" {
		t.Errorf("ID = %q, want artifact-Deepcrest-0", got[0].ID)
	}
}

func TestEmergencePassUnknownFigureSkipsFallback(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "ghost", SettlementName: "Deepcrest", Description: "found"},
	}

	got, err := EmergencePass(nil, events, nil, SignificanceContext{}, artifactsRNG(3))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("artifact count = %d, want 0 (no figure context, no fallback)", len(got))
	}
}

func TestEmergencePassSkipsBirthWithoutOrigin(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "ghost", Description: "found"},
	}

	got, err := EmergencePass(nil, events, nil, SignificanceContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("artifact count = %d, want 0 (no origin to derive artifact-{origin}-{index})", len(got))
	}
}

func TestEmergencePassPropagatesPostProcessErrors(t *testing.T) {
	events := []simulation.Event{
		{Year: 1, Category: "Discovery", FigureID: "D-1", SettlementName: "Deepcrest", ArtifactID: "ghost", Description: "found"},
	}

	if _, err := EmergencePass(nil, events, nil, SignificanceContext{}, artifactsRNG(1)); err == nil {
		t.Fatal("EmergencePass with unknown ArtifactID: expected error, got nil")
	}
}

func TestEmergencePassDiscoveryWithoutSettlementUsesFigureOrigin(t *testing.T) {
	events := []simulation.Event{
		{Year: 12, Category: "Discovery", FigureID: "D-1", Description: "found"},
	}
	figures := []FigureContext{
		{ID: "D-1", Settlement: "Deepcrest"},
	}

	got, err := EmergencePass(nil, events, figures, SignificanceContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}
	if got[0].ID != "artifact-Deepcrest-0" {
		t.Errorf("ID = %q, want artifact-Deepcrest-0 (origin from figure context)", got[0].ID)
	}
	if got[0].Name != "Warhammer of Deepcrest" {
		t.Errorf("Name = %q, want Warhammer of Deepcrest", got[0].Name)
	}
}

func TestEmergencePassFallbackReputationEventFallback(t *testing.T) {
	events := []simulation.Event{
		{Year: 15, Category: "Discovery", FigureID: "D-1", SettlementName: "Deepcrest", Description: "found"},
	}
	figures := []FigureContext{
		{ID: "D-1", Settlement: "Deepcrest", Reputation: []ReputationDelta{{Year: 4, Delta: 12}}},
	}

	got, err := EmergencePass(nil, events, figures, SignificanceContext{}, artifactsRNG(3))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}
	wantProv := []ProvenanceEntry{
		{Year: 4, Owner: Owner{Kind: "figure", ID: "D-1"}, EventID: "", EventType: "Reputation"},
	}
	if !reflect.DeepEqual(got[0].Provenance, wantProv) {
		t.Errorf("Provenance = %+v, want %+v", got[0].Provenance, wantProv)
	}
}

func TestEmergencePassNilRNGReturnsError(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "D-1", SettlementName: "Deepcrest", Description: "found"},
	}
	if _, err := EmergencePass(nil, events, nil, SignificanceContext{}, nil); err == nil {
		t.Fatal("EmergencePass with nil RNG: expected error, got nil")
	}
}

func TestEmergencePassPreservesPostProcessBehavior(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", Status: "lost"},
	}
	events := []simulation.Event{
		{Year: 1, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "artifact-1", Description: "found"},
		{Year: 2, Category: "Birth", Description: "born"},
		{Year: 2, Category: "Economy", Description: "trade"},
	}

	got, err := EmergencePass(artifacts, events, nil, SignificanceContext{}, artifactsRNG(3))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}
	if !reflect.DeepEqual(got[0], artifacts[0]) {
		t.Errorf("existing artifact mutated: got %+v, want %+v", got[0], artifacts[0])
	}
	wantIDs := []string{"event-1-0", "event-2-0", "event-2-1"}
	for i, id := range wantIDs {
		if events[i].ID != id {
			t.Errorf("events[%d].ID = %q, want %q", i, events[i].ID, id)
		}
	}
}

func TestEmergencePassDeterministic(t *testing.T) {
	events := []simulation.Event{
		{Year: 1, Category: "Birth", Description: "A"},
		{Year: 2, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "took Ironforge"},
		{Year: 3, Category: "Discovery", FigureID: "Deepcrest-1", SettlementName: "Deepcrest", Description: "found"},
		{Year: 4, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Haven", Description: "took Haven"},
		{Year: 5, Category: "Discovery", FigureID: "D-1", SettlementName: "Haven", Description: "found"},
		{Year: 6, Category: "Economy", Description: "trade"},
	}
	figures := []FigureContext{
		{ID: "Deepcrest-1", Settlement: "Deepcrest", Reputation: []ReputationDelta{{Year: 2, Delta: 3, Event: "Discovery"}, {Year: 3, Delta: 4, Event: "Raid"}, {Year: 4, Delta: 5, Event: "War"}}},
		{ID: "D-1", Settlement: "Haven", Reputation: []ReputationDelta{{Year: 3, Delta: 2, Event: "Discovery"}}},
	}

	run := func() ([]Artifact, []simulation.Event, error) {
		evs := make([]simulation.Event, len(events))
		copy(evs, events)
		arts, err := EmergencePass(nil, evs, figures, SignificanceContext{}, artifactsRNG(42))
		return arts, evs, err
	}

	firstArts, firstEvents, err := run()
	if err != nil {
		t.Fatalf("first EmergencePass: %v", err)
	}
	secondArts, secondEvents, err := run()
	if err != nil {
		t.Fatalf("second EmergencePass: %v", err)
	}

	artJSON := func(arts []Artifact) []byte {
		b, err := json.Marshal(arts)
		if err != nil {
			t.Fatalf("marshal artifacts: %v", err)
		}
		return b
	}
	evJSON := func(evs []simulation.Event) []byte {
		b, err := json.Marshal(evs)
		if err != nil {
			t.Fatalf("marshal events: %v", err)
		}
		return b
	}

	if !reflect.DeepEqual(artJSON(firstArts), artJSON(secondArts)) {
		t.Errorf("emergent artifacts differ across identical runs:\nfirst=%+v\nsecond=%+v", firstArts, secondArts)
	}
	if !reflect.DeepEqual(evJSON(firstEvents), evJSON(secondEvents)) {
		t.Errorf("events differ across identical runs:\nfirst=%+v\nsecond=%+v", firstEvents, secondEvents)
	}

	// A different seed must change the birth stream: emergence is seeded.
	evs := make([]simulation.Event, len(events))
	copy(evs, events)
	otherArts, err := EmergencePass(nil, evs, figures, SignificanceContext{}, artifactsRNG(7))
	if err != nil {
		t.Fatalf("EmergencePass with other seed: %v", err)
	}
	if reflect.DeepEqual(artJSON(firstArts), artJSON(otherArts)) {
		t.Errorf("expected a different artifact stream for a different seed")
	}
	if len(otherArts) != len(firstArts) {
		t.Logf("note: seeds 42 and 7 minted different artifact counts (%d vs %d)", len(firstArts), len(otherArts))
	}
}

func TestEmergencePassNamesAreSettlementSuffixed(t *testing.T) {
	events := []simulation.Event{
		{Year: 5, Category: "Conquest", SettlementName: "Blackgate", TargetSettlement: "Ironforge", Description: "took Ironforge"},
	}
	got, err := EmergencePass(nil, events, nil, SignificanceContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(got))
	}
	if !strings.HasSuffix(got[0].Name, " of Blackgate") {
		t.Errorf("Name = %q, want suffix \" of Blackgate\"", got[0].Name)
	}
}
