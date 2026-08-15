package artifact

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// Draw sequences below are verified against the fixed lane walk for
// PCG(seed, 1) (the artifactsRNG harness used across the artifact tests).
// Interleaved sequences (the numbers after earlier draws on the same lane)
// were pinned by simulating the exact consumption order:
//   - seed 1 (after 2 fake-discovery IntN(2) draws): gates 86, 91 — both fail
//   - seed 7: gates 61, 80 on a fresh lane — both fail
//   - seed 8 (after 2 fake-discovery draws 0, 0): gate 58 (fail), gate 31
//     (pass, figure draw 1)

// plantedRelic builds a planted relic in the GeneratePlantedRelics shape:
// intrinsic significance, lost, no provenance.
func plantedRelic(id, name string) Artifact {
	return Artifact{
		ID:                 id,
		Name:               name,
		Type:               "weapon",
		SignificanceSource: "intrinsic",
		Status:             "lost",
		SignificanceScore:  3,
		IsSignificant:      true,
		SignificanceYear:   0,
	}
}

func TestFakeDiscoveryAssignsFiguresToPlantedRelics(t *testing.T) {
	artifacts := []Artifact{
		plantedRelic("artifact-ruin-0", "Relic of Old Keep"),
		plantedRelic("artifact-ruin-1", "Relic of Sunken Halls"),
	}
	events := []simulation.Event{
		{Year: 5, Category: "Birth", Description: "born"},
	}
	figures := []FigureContext{
		{ID: "Deepcrest-1", Settlement: "Deepcrest"},
		{ID: "Ironforge-1", Settlement: "Ironforge"},
	}
	agent := newFakeDiscoveryAgent(figures, artifactsRNG(1))

	got := fakeDiscovery(artifacts, events, agent)

	// Seed 1: IntN(2) draws 0 then 1, so the first relic goes to
	// Deepcrest-1 and the second to Ironforge-1. The minted events are
	// prepended in artifact order before the original stream.
	want := []simulation.Event{
		{Year: 0, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "artifact-ruin-0"},
		{Year: 0, Category: "Discovery", FigureID: "Ironforge-1", ArtifactID: "artifact-ruin-1"},
		{Year: 5, Category: "Birth", Description: "born"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("events = %+v, want %+v", got, want)
	}
	// A successful draw marks the relic held immediately (the walk then
	// records the year-0 ownership from the minted event).
	if artifacts[0].Status != "held" || artifacts[1].Status != "held" {
		t.Errorf("discovered relics must be held, got %+v", artifacts)
	}
}

func TestFakeDiscoverySkipsNonPlantedArtifacts(t *testing.T) {
	artifacts := []Artifact{
		{ID: "artifact-1", Name: "Crown", SignificanceSource: "historical", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		}},
		{ID: "artifact-ruin-0", Name: "Relic", SignificanceSource: "intrinsic", Status: "held"},
	}
	events := []simulation.Event{
		{Year: 5, Category: "Birth", Description: "born"},
	}
	agent := newFakeDiscoveryAgent([]FigureContext{{ID: "Deepcrest-1", Settlement: "Deepcrest"}}, artifactsRNG(1))

	got := fakeDiscovery(artifacts, events, agent)

	// Historical artifacts and already-held relics are never fake-discovered.
	if !reflect.DeepEqual(got, events) {
		t.Errorf("events = %+v, want unchanged %+v", got, events)
	}
}

func TestFakeDiscoveryNoFiguresLeavesRelicLost(t *testing.T) {
	artifacts := []Artifact{plantedRelic("artifact-ruin-0", "Relic of Old Keep")}
	events := []simulation.Event{
		{Year: 5, Category: "Birth", Description: "born"},
	}

	got := fakeDiscovery(artifacts, events, newFakeDiscoveryAgent(nil, artifactsRNG(1)))

	if !reflect.DeepEqual(got, events) {
		t.Errorf("events = %+v, want unchanged %+v (no figures, nothing minted)", got, events)
	}
	if artifacts[0].Status != "lost" {
		t.Errorf("status = %q, want lost (relic stays lost when the draw yields no figure)", artifacts[0].Status)
	}
}

func TestFakeDiscoveryAgentDrawsFromArtifactsLane(t *testing.T) {
	figures := []FigureContext{
		{ID: "A-1", Settlement: "A"},
		{ID: "B-1", Settlement: "B"},
		{ID: "C-1", Settlement: "C"},
	}
	cases := []struct {
		seed uint64
		want string
	}{
		// IntN(3) for PCG(seed, 1): seed 1 draws 2, seed 2 draws 0.
		{1, "C-1"},
		{2, "A-1"},
	}
	for _, tc := range cases {
		t.Run(string(rune('0'+tc.seed)), func(t *testing.T) {
			agent := newFakeDiscoveryAgent(figures, artifactsRNG(tc.seed))
			if got := agent.Discover(); got != tc.want {
				t.Errorf("Discover() = %q, want %q", got, tc.want)
			}
		})
	}

	// No figures: "" without consuming the lane (a second draw on a fresh
	// lane still yields the same figure sequence start).
	agent := newFakeDiscoveryAgent(nil, artifactsRNG(1))
	if got := agent.Discover(); got != "" {
		t.Errorf("Discover() with no figures = %q, want empty", got)
	}
}

// lostArtifact builds an artifact whose last provenance entry is a lost
// transfer (a degenerate death, spec 6.3) — the historically-lost shape the
// rediscovery step targets.
func lostArtifact(id string) Artifact {
	return Artifact{
		ID:                 id,
		Name:               "Crown of " + id,
		Type:               "crown",
		SignificanceSource: "historical",
		Status:             "lost",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
			{Year: 10, Owner: Owner{Kind: "lost"}, EventID: "event-10-0", EventType: "Death"},
		},
	}
}

func TestRediscoveryMintsSyntheticDiscovery(t *testing.T) {
	artifacts := []Artifact{lostArtifact("artifact-1")}
	events := []simulation.Event{
		{Year: 10, Category: "Death", FigureID: "Deepcrest-1", ID: "event-10-0", Description: "owner dies"},
	}
	agent := newFakeDiscoveryAgent([]FigureContext{{ID: "Deepcrest-1", Settlement: "Deepcrest"}}, artifactsRNG(2))
	// Seed 2: the gate draw (5) passes and the figure draw (IntN(1)) picks
	// Deepcrest-1.
	got := rediscovery(artifacts, events, agent, artifactsRNG(2), 10)

	wantEvents := []simulation.Event{
		{Year: 10, Category: "Death", FigureID: "Deepcrest-1", ID: "event-10-0", Description: "owner dies"},
		// The minted event continues the walk's year-count scheme: one
		// year-10 event already exists, so this is event-10-1.
		{Year: 10, Category: "Discovery", FigureID: "Deepcrest-1", ID: "event-10-1", ArtifactID: "artifact-1"},
	}
	if !reflect.DeepEqual(got, wantEvents) {
		t.Errorf("events = %+v, want %+v", got, wantEvents)
	}

	a := artifacts[0]
	if a.Status != "held" {
		t.Errorf("status = %q, want held (lost -> held on rediscovery)", a.Status)
	}
	wantProv := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 10, Owner: Owner{Kind: "lost"}, EventID: "event-10-0", EventType: "Death"},
		{Year: 10, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-10-1", EventType: "Discovery"},
	}
	if !reflect.DeepEqual(a.Provenance, wantProv) {
		t.Errorf("provenance = %+v, want %+v", a.Provenance, wantProv)
	}
	// The rediscovery event is associated like any in-stream Discovery
	// carrying the artifact's ID would be.
	if wantIDs := []string{"event-10-1"}; !reflect.DeepEqual(a.AssociatedEventIDs, wantIDs) {
		t.Errorf("associatedEventIDs = %v, want %v", a.AssociatedEventIDs, wantIDs)
	}
	if owner := CurrentOwner(a); owner.Kind != "figure" || owner.ID != "Deepcrest-1" {
		t.Errorf("current owner = %+v, want (figure, Deepcrest-1)", owner)
	}
}

func TestRediscoveryFailedDrawLeavesLost(t *testing.T) {
	artifacts := []Artifact{lostArtifact("artifact-1")}
	events := []simulation.Event{
		{Year: 10, Category: "Death", FigureID: "Deepcrest-1", ID: "event-10-0", Description: "owner dies"},
	}
	agent := newFakeDiscoveryAgent([]FigureContext{{ID: "Deepcrest-1", Settlement: "Deepcrest"}}, artifactsRNG(1))
	// Seed 1: the gate draw (99) fails.
	got := rediscovery(artifacts, events, agent, artifactsRNG(1), 10)

	if len(got) != 1 {
		t.Errorf("event count = %d, want 1 (failed draw mints nothing)", len(got))
	}
	if artifacts[0].Status != "lost" {
		t.Errorf("status = %q, want lost", artifacts[0].Status)
	}
	if len(artifacts[0].Provenance) != 2 {
		t.Errorf("provenance = %+v, want unchanged (no rediscovery entry)", artifacts[0].Provenance)
	}
	if len(artifacts[0].AssociatedEventIDs) != 0 {
		t.Errorf("associatedEventIDs = %v, want none", artifacts[0].AssociatedEventIDs)
	}
}

func TestRediscoveryNoFigureLeavesLost(t *testing.T) {
	artifacts := []Artifact{lostArtifact("artifact-1")}
	events := []simulation.Event{
		{Year: 10, Category: "Death", FigureID: "Deepcrest-1", ID: "event-10-0", Description: "owner dies"},
	}
	agent := newFakeDiscoveryAgent(nil, artifactsRNG(2))
	// Seed 2: the gate passes but no figure exists to rediscover it.
	got := rediscovery(artifacts, events, agent, artifactsRNG(2), 10)

	if len(got) != 1 {
		t.Errorf("event count = %d, want 1 (no figure, nothing minted)", len(got))
	}
	if artifacts[0].Status != "lost" {
		t.Errorf("status = %q, want lost", artifacts[0].Status)
	}
}

func TestRediscoverySkipsHeldArtifactsWithoutConsumingDraws(t *testing.T) {
	held := plantedRelic("artifact-ruin-0", "Relic of Old Keep")
	held.Status = "held"
	run := func(artifacts []Artifact) []simulation.Event {
		evs := []simulation.Event{
			{Year: 10, Category: "Death", FigureID: "Deepcrest-1", ID: "event-10-0", Description: "owner dies"},
		}
		agent := newFakeDiscoveryAgent([]FigureContext{{ID: "Deepcrest-1", Settlement: "Deepcrest"}}, artifactsRNG(2))
		return rediscovery(artifacts, evs, agent, artifactsRNG(2), 10)
	}

	// The held artifact must not consume a gate draw: the minted event for
	// the lost artifact is identical with or without the held one in front.
	withHeld := run([]Artifact{held, lostArtifact("artifact-1")})
	withoutHeld := run([]Artifact{lostArtifact("artifact-1")})

	if !reflect.DeepEqual(withHeld, withoutHeld) {
		t.Errorf("held artifact shifted the lane:\nwith held=%+v\nwithout  =%+v", withHeld, withoutHeld)
	}
	if held.Status != "held" {
		t.Errorf("held artifact status = %q, want held (untouched)", held.Status)
	}
}

func TestSyntheticEventIDContinuesYearScheme(t *testing.T) {
	events := []simulation.Event{
		{Year: 10, ID: "event-10-0"},
		{Year: 10, ID: "event-10-1"},
		{Year: 5, ID: "event-5-0"},
		{Year: 0, ID: "event-0-0"},
	}
	cases := []struct {
		year int
		want string
	}{
		{10, "event-10-2"},
		{5, "event-5-1"},
		{0, "event-0-1"},
		{99, "event-99-0"},
	}
	for _, tc := range cases {
		if got := syntheticEventID(events, tc.year); got != tc.want {
			t.Errorf("syntheticEventID(year %d) = %q, want %q", tc.year, got, tc.want)
		}
	}
}

// TestEmergencePassFakeDiscoveryBeforeEmergenceDraws pins the canonical lane
// order: the fake-discovery figure draws happen before the emergence draws.
// Seed 1 with two relics and two figures consumes IntN(2)=0,1 for discovery,
// then the emergence walk draws type=armor (1), gate=91 (fails — common gate
// is 25%), so no artifact is born. If emergence drew first, the type draw
// would be different (IntN(8) drawn from a fresh lane).
func TestEmergencePassFakeDiscoveryBeforeEmergenceDraws(t *testing.T) {
	artifacts := []Artifact{
		plantedRelic("artifact-ruin-0", "Relic of Old Keep"),
		plantedRelic("artifact-ruin-1", "Relic of Sunken Halls"),
	}
	events := []simulation.Event{
		{Year: 5, Category: "Discovery", FigureID: "D-1", SettlementName: "Deepcrest", Description: "unearths a hoard"},
	}
	figures := []FigureContext{
		{ID: "D-1", Settlement: "Deepcrest"},
		{ID: "D-2", Settlement: "Ironforge"},
	}

	got, gotEvents, err := EmergencePass(artifacts, events, figures, SignificanceContext{}, TransferContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	// Fake-discovery consumed the first two lane draws: relic-ruin-0 ->
	// D-1 (IntN(2)=0), relic-ruin-1 -> D-2 (IntN(2)=1). The prepended
	// events carry walk-assigned IDs.
	if len(gotEvents) != 3 {
		t.Fatalf("event count = %d, want 3 (two discoveries + the stream event)", len(gotEvents))
	}
	wantFirst := []simulation.Event{
		{Year: 0, Category: "Discovery", FigureID: "D-1", ID: "event-0-0", ArtifactID: "artifact-ruin-0"},
		{Year: 0, Category: "Discovery", FigureID: "D-2", ID: "event-0-1", ArtifactID: "artifact-ruin-1"},
	}
	for i := range wantFirst {
		if !reflect.DeepEqual(gotEvents[i], wantFirst[i]) {
			t.Errorf("events[%d] = %+v, want %+v", i, gotEvents[i], wantFirst[i])
		}
	}
	// The emergence walk then drew type=armor, gate=91: the common gate
	// (25%) fails, so no emergent artifact is born.
	if len(got) != 2 {
		t.Errorf("artifact count = %d, want 2 (no emergence birth)", len(got))
	}
	// The walk recorded the discoveries as the relics' first ownership.
	for i, wantOwner := range []string{"D-1", "D-2"} {
		prov := got[i].Provenance
		if len(prov) != 1 || prov[0].Year != 0 || prov[0].Owner.Kind != "figure" || prov[0].Owner.ID != wantOwner {
			t.Errorf("artifacts[%d] provenance = %+v, want single year-0 figure entry for %s", i, prov, wantOwner)
		}
		if got[i].Status != "held" {
			t.Errorf("artifacts[%d] status = %q, want held (discovered)", i, got[i].Status)
		}
	}
}

// TestEmergencePassLossRediscoveryDeterministic is the determinism gate for
// the #70 lifecycle: identical seed and inputs must produce byte-identical
// artifacts AND events (the loss and rediscovery sequences included), and a
// different seed must diverge.
func TestEmergencePassLossRediscoveryDeterministic(t *testing.T) {
	artifacts := []Artifact{
		plantedRelic("artifact-ruin-0", "Relic of Old Keep"),
		plantedRelic("artifact-ruin-1", "Relic of Sunken Halls"),
		// Historically lost: a degenerate death (no heir, no settlement)
		// already recorded in its provenance at year 5.
		{ID: "artifact-1", Name: "Crown of Deepcrest", Type: "crown", SignificanceSource: "historical", Status: "lost", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "D-1"}, EventID: "event-1-0", EventType: "Discovery"},
			{Year: 5, Owner: Owner{Kind: "lost"}, EventID: "event-5-0", EventType: "Death"},
		}},
		// Owned by a settlement whose FINAL class is Abandoned.
		{ID: "artifact-2", Name: "Tome of Ironforge", Type: "tome", SignificanceSource: "historical", Status: "held", Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		}},
	}
	events := []simulation.Event{
		{Year: 8, Category: "War", ArtifactID: "artifact-1", Description: "war while lost"},
		{Year: 12, Category: "Birth", Description: "unrelated"},
	}
	figures := []FigureContext{
		{ID: "D-1", Settlement: "Deepcrest"},
		{ID: "D-2", Settlement: "Ironforge"},
	}
	sig := SignificanceContext{
		FigureReputation: map[string]map[int]int{
			// D-1's year-6 +5 falls inside artifact-1's lost span [5, 12)
			// and must never accrue.
			"D-1": {1: 1, 2: 2, 6: 5},
		},
		SettlementClass: map[string]string{"Ironforge": "Abandoned"},
	}

	run := func() ([]Artifact, []simulation.Event, error) {
		arts := make([]Artifact, len(artifacts))
		copy(arts, artifacts)
		evs := make([]simulation.Event, len(events))
		copy(evs, events)
		return EmergencePass(arts, evs, figures, sig, TransferContext{}, artifactsRNG(8))
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
		t.Errorf("lifecycle artifacts differ across identical runs:\nfirst=%+v\nsecond=%+v", firstArts, secondArts)
	}
	if !reflect.DeepEqual(evJSON(firstEvents), evJSON(secondEvents)) {
		t.Errorf("lifecycle events differ across identical runs:\nfirst=%+v\nsecond=%+v", firstEvents, secondEvents)
	}

	// A different seed must change the rediscovery sequence.
	arts := make([]Artifact, len(artifacts))
	copy(arts, artifacts)
	evs := make([]simulation.Event, len(events))
	copy(evs, events)
	otherArts, _, err := EmergencePass(arts, evs, figures, sig, TransferContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass with other seed: %v", err)
	}
	if reflect.DeepEqual(artJSON(firstArts), artJSON(otherArts)) {
		t.Errorf("expected a different lifecycle stream for a different seed")
	}

	// Pin the seed-8 outcome so a regression in expected values is caught,
	// not just run-to-run divergence. Lane: 2 fake-discovery IntN(2) draws
	// (both 0 -> D-1), then rediscovery gates in artifact order: 58 (fail)
	// for artifact-1, 31 (pass, figure draw 1 -> D-2) for artifact-2.
	//
	// artifact-1: rediscovery fails, score frozen.
	a1 := firstArts[2]
	if a1.Status != "lost" {
		t.Errorf("artifact-1 status = %q, want lost (gate 58 fails)", a1.Status)
	}
	wantProv1 := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "figure", ID: "D-1"}, EventID: "event-1-0", EventType: "Discovery"},
		{Year: 5, Owner: Owner{Kind: "lost"}, EventID: "event-5-0", EventType: "Death"},
	}
	if !reflect.DeepEqual(a1.Provenance, wantProv1) {
		t.Errorf("artifact-1 provenance = %+v, want %+v", a1.Provenance, wantProv1)
	}
	// Accrual only for held years [1, 5): the year-6 +5 and the year-8 War
	// both fall inside the lost span [5, 12) and are skipped (spec 4.6).
	if a1.SignificanceScore != 3 {
		t.Errorf("artifact-1 SignificanceScore = %d, want 3 (frozen while lost)", a1.SignificanceScore)
	}
	if !a1.IsSignificant || a1.SignificanceYear != 2 {
		t.Errorf("artifact-1 significance = (%v, year %d), want (true, 2)", a1.IsSignificant, a1.SignificanceYear)
	}
	for _, e := range firstEvents {
		if e.Category == "Discovery" && e.ArtifactID == a1.ID {
			t.Errorf("lost artifact %s was minted an event: %+v", a1.ID, e)
		}
	}

	// artifact-2: abandonment loss at the horizon, then rediscovery by the
	// drawn figure at the same year; the entry chain stays chronological.
	a2 := firstArts[3]
	if a2.Status != "held" {
		t.Errorf("artifact-2 status = %q, want held (gate 31 passes)", a2.Status)
	}
	wantProv2 := []ProvenanceEntry{
		{Year: 1, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-1-0", EventType: "Creation"},
		{Year: 12, Owner: Owner{Kind: "lost"}, EventID: "", EventType: "ArtifactLoss"},
		{Year: 12, Owner: Owner{Kind: "figure", ID: "D-2"}, EventID: "event-12-1", EventType: "Discovery"},
	}
	if !reflect.DeepEqual(a2.Provenance, wantProv2) {
		t.Errorf("artifact-2 provenance = %+v, want %+v", a2.Provenance, wantProv2)
	}
	// The minted rediscovery event is appended after the stream events and
	// carries the artifact's ID, so the emergence second walk skips it.
	if len(firstEvents) != 5 {
		t.Fatalf("event count = %d, want 5 (two discoveries + two stream events + rediscovery)", len(firstEvents))
	}
	last := firstEvents[len(firstEvents)-1]
	if last.ID != "event-12-1" || last.Category != "Discovery" || last.FigureID != "D-2" || last.ArtifactID != a2.ID {
		t.Errorf("minted rediscovery event = %+v, want {12, Discovery, D-2, event-12-1, %s}", last, a2.ID)
	}
	// Seed-1 run: both rediscovery gates (86, 91) fail — artifact-2 stays
	// lost there, proving the sequences diverge per seed.
	if otherArts[3].Status != "lost" {
		t.Errorf("seed-1 artifact-2 status = %q, want lost (both seed-1 gates fail)", otherArts[3].Status)
	}
}

// TestEmergencePassRediscoveryClosesLostSpan pins the significance freeze:
// an artifact lost mid-stream (degenerate death) accrues nothing while lost,
// and the score resumes after the synthetic rediscovery entry closes the
// span.
func TestEmergencePassRediscoveryClosesLostSpan(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Crown of Deepcrest",
		Type:               "crown",
		SignificanceSource: "historical",
		Status:             "held",
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "D-1"}, EventID: "event-1-0", EventType: "Discovery"},
		},
	}}
	events := []simulation.Event{
		{Year: 5, Category: "Death", FigureID: "D-1", Description: "owner dies childless"},
		{Year: 8, Category: "War", ArtifactID: "artifact-1", Description: "war while lost"},
		{Year: 12, Category: "Birth", Description: "unrelated"},
	}
	figures := []FigureContext{{ID: "D-1", Settlement: "Deepcrest"}}
	sig := SignificanceContext{
		FigureReputation: map[string]map[int]int{
			"D-1": {1: 1, 2: 2, 6: 5},
		},
	}

	got, _, err := EmergencePass(artifacts, events, figures, sig, TransferContext{}, artifactsRNG(1))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	a := got[0]
	// Seed 1: the death destruction draw (99) fails, so the degenerate death
	// records the lost fallback; the rediscovery gate (10) passes, so the
	// lost span [5, 12) is closed by the synthetic Discovery at the horizon.
	if a.Status != "held" {
		t.Errorf("status = %q, want held (rediscovered)", a.Status)
	}
	// Score = year-1 (+1) and year-2 (+2) accrual only. The +5 in year 6
	// falls inside the lost span and the year-8 War is skipped while lost
	// (spec 4.6): the score froze at 3 and resumes after the span closes.
	if a.SignificanceScore != 3 {
		t.Errorf("SignificanceScore = %d, want 3 (frozen while lost, no lost-span accrual)", a.SignificanceScore)
	}
	if !a.IsSignificant || a.SignificanceYear != 2 {
		t.Errorf("significance = (%v, year %d), want (true, 2) — crossed by accrual in year 2", a.IsSignificant, a.SignificanceYear)
	}
}

// TestEmergencePassRediscoverySkipsMintedEventsInEmergenceWalk verifies that
// the appended rediscovery events carry ArtifactID, so the emergence second
// walk's "already" check skips them instead of birthing a new artifact.
func TestEmergencePassRediscoverySkipsMintedEventsInEmergenceWalk(t *testing.T) {
	artifacts := []Artifact{lostArtifact("artifact-1")}
	events := []simulation.Event{
		{Year: 10, Category: "Death", FigureID: "D-1", Description: "owner dies"},
	}
	figures := []FigureContext{{ID: "D-1", Settlement: "Deepcrest"}}

	got, gotEvents, err := EmergencePass(artifacts, events, figures, SignificanceContext{}, TransferContext{}, artifactsRNG(2))
	if err != nil {
		t.Fatalf("EmergencePass: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("artifact count = %d, want 1 (no emergence birth from the minted event)", len(got))
	}
	if len(gotEvents) != 2 {
		t.Fatalf("event count = %d, want 2 (stream event + minted rediscovery)", len(gotEvents))
	}
	if gotEvents[1].ID != "event-10-1" || gotEvents[1].Category != "Discovery" || gotEvents[1].FigureID != "D-1" || gotEvents[1].ArtifactID != "artifact-1" {
		t.Errorf("minted event = %+v, want a Discovery carrying artifact-1 as event-10-1", gotEvents[1])
	}
}

// TestRediscoveryUsesPassedLane verifies all lifecycle draws consume the
// artifacts lane passed by the caller: the pinned gate values (61, 80 for
// PCG(7, 1) on a fresh lane) only reproduce when every draw comes from that
// lane.
func TestRediscoveryUsesPassedLane(t *testing.T) {
	artifacts := []Artifact{
		lostArtifact("artifact-1"),
		lostArtifact("artifact-2"),
	}
	events := []simulation.Event{
		{Year: 10, Category: "Death", FigureID: "Deepcrest-1", ID: "event-10-0", Description: "owner dies"},
	}
	figures := []FigureContext{
		{ID: "Deepcrest-1", Settlement: "Deepcrest"},
		{ID: "Ironforge-1", Settlement: "Ironforge"},
	}
	// Seed 7 on a fresh lane: gates 61 and 80 — both fail, so nothing is
	// minted and both artifacts stay lost.
	got := rediscovery(artifacts, events, newFakeDiscoveryAgent(figures, artifactsRNG(7)), artifactsRNG(7), 10)

	if artifacts[0].Status != "lost" {
		t.Errorf("artifact-1 status = %q, want lost (gate 61 fails)", artifacts[0].Status)
	}
	if artifacts[1].Status != "lost" {
		t.Errorf("artifact-2 status = %q, want lost (gate 80 fails)", artifacts[1].Status)
	}
	if len(got) != 1 {
		t.Fatalf("event count = %d, want 1 (failed draws mint nothing)", len(got))
	}
	if len(artifacts[0].Provenance) != 2 || len(artifacts[1].Provenance) != 2 {
		t.Errorf("failed draws must not append provenance entries: %+v", artifacts)
	}
}
