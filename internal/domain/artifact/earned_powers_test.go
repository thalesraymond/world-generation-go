package artifact

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// pivotedArtifacts returns n historical artifacts held by figures, each with
// its own War event that crosses the significance threshold (spec 4.2: a
// single War is itself pivotal). War events carry the artifact's ID, so the
// post-process walk associates them and the significance evaluation grants
// one earned combat power per artifact in artifact order.
func pivotedArtifacts(n int) ([]Artifact, []simulation.Event) {
	artifacts := make([]Artifact, 0, n)
	events := make([]simulation.Event, 0, n)
	for i := 0; i < n; i++ {
		id := "artifact-" + string(rune('a'+i))
		artifacts = append(artifacts, Artifact{
			ID:                 id,
			Name:               id,
			SignificanceSource: "historical",
			Status:             "held",
			Provenance: []ProvenanceEntry{
				{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
			},
		})
		events = append(events, simulation.Event{Year: 2, Category: "War", ArtifactID: id, Description: "war"})
	}
	return artifacts, events
}

func TestEarnedNarrativeEffectDefault(t *testing.T) {
	if got := earnedNarrativeEffect("Settlement"); got != "shaped by its history, bearer gains renown" {
		t.Errorf("earnedNarrativeEffect(Settlement) = %q, want the default effect", got)
	}
}

func TestEarnedPowerGrantedAtPivotalCrossing(t *testing.T) {
	cases := []struct {
		name     string
		category string
		count    int // events needed to reach the threshold
		wantKind string
		wantBase bool
		wantEff  string
	}{
		{name: "war", category: "War", count: 1, wantKind: "combat", wantBase: true},
		{name: "conquest", category: "Conquest", count: 1, wantKind: "combat", wantBase: true},
		{name: "diplomacy", category: "Diplomacy", count: 2, wantKind: "influence", wantBase: true},
		{name: "politics", category: "Politics", count: 2, wantKind: "influence", wantBase: true},
		{name: "raid", category: "Raid", count: 2, wantKind: "narrative", wantEff: "survived a raid, bearer gains the raider's boldness"},
		{name: "expansion", category: "Expansion", count: 3, wantKind: "narrative", wantEff: "witnessed expansion, bearer gains a pioneer's resolve"},
		{name: "disaster", category: "Disaster", count: 3, wantKind: "narrative", wantEff: "survives calamity, bearer gains resilience"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			artifacts := []Artifact{heldArtifact()}
			var events []simulation.Event
			for i := 0; i < tc.count; i++ {
				events = append(events, simulation.Event{Year: 2 + i, Category: tc.category, ArtifactID: "artifact-1", Description: "event"})
			}

			if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(1)); err != nil {
				t.Fatalf("postProcess: %v", err)
			}

			a := artifacts[0]
			if len(a.Powers) != 1 {
				t.Fatalf("Powers = %v, want exactly one earned power", a.Powers)
			}
			p := a.Powers[0]
			if got := p.Type(); got != tc.wantKind {
				t.Errorf("Type() = %q, want %q", got, tc.wantKind)
			}
			switch v := p.(type) {
			case CombatPower:
				if v.Source != "earned" {
					t.Errorf("CombatPower.Source = %q, want earned", v.Source)
				}
				if tc.wantBase && (v.Base < earnedPowerMinBase || v.Base >= earnedPowerMinBase+earnedPowerBaseSpan) {
					t.Errorf("CombatPower.Base = %d, want in [%d,%d)", v.Base, earnedPowerMinBase, earnedPowerMinBase+earnedPowerBaseSpan)
				}
			case InfluencePower:
				if v.Source != "earned" {
					t.Errorf("InfluencePower.Source = %q, want earned", v.Source)
				}
				if tc.wantBase && (v.Base < earnedPowerMinBase || v.Base >= earnedPowerMinBase+earnedPowerBaseSpan) {
					t.Errorf("InfluencePower.Base = %d, want in [%d,%d)", v.Base, earnedPowerMinBase, earnedPowerMinBase+earnedPowerBaseSpan)
				}
			case NarrativePower:
				if v.Source != "earned" {
					t.Errorf("NarrativePower.Source = %q, want earned", v.Source)
				}
				if v.Effect != tc.wantEff {
					t.Errorf("NarrativePower.Effect = %q, want %q", v.Effect, tc.wantEff)
				}
			default:
				t.Fatalf("power = %T, want the %s concrete type", p, tc.wantKind)
			}

			lastID := events[len(events)-1].ID
			if a.PivotalEventID != lastID {
				t.Errorf("PivotalEventID = %q, want %q (crossing at the last event)", a.PivotalEventID, lastID)
			}
		})
	}
}

func TestEarnedPowerNotGrantedOnAccrualCrossing(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Crown",
		SignificanceSource: "historical",
		Status:             "lost",
	}}
	events := []simulation.Event{
		{Year: 10, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "artifact-1", Description: "found"},
	}
	sig := SignificanceContext{
		FigureReputation: map[string]map[int]int{
			"Deepcrest-1": {10: 2, 12: 1, 14: 2},
		},
	}

	if err := PostProcess(artifacts, events, sig, TransferContext{}, artifactsRNG(1)); err != nil {
		t.Fatalf("postProcess: %v", err)
	}

	// The crossing happens via annual accrual in year 12: no pivotal event,
	// so no earned power (spec 7.4: pivotal events grant earned powers).
	a := artifacts[0]
	if a.PivotalEventID != "" {
		t.Errorf("PivotalEventID = %q, want empty (crossing was not an event)", a.PivotalEventID)
	}
	if len(a.Powers) != 0 {
		t.Errorf("Powers = %v, want none (accrual crossing grants nothing)", a.Powers)
	}
}

func TestEarnedPowerNotGrantedForIntrinsicArtifact(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-ruin-0",
		Name:               "Relic",
		Type:               "crown",
		SignificanceSource: "intrinsic",
		Status:             "lost",
		SignificanceScore:  3,
		IsSignificant:      true,
		SignificanceYear:   0,
		Powers:             []Power{InfluencePower{Base: 5, Source: "intrinsic"}},
	}}
	events := []simulation.Event{
		{Year: 2, Category: "Discovery", FigureID: "Deepcrest-1", ArtifactID: "artifact-ruin-0", Description: "found"},
		{Year: 4, Category: "War", ArtifactID: "artifact-ruin-0", Description: "war"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(1)); err != nil {
		t.Fatalf("postProcess: %v", err)
	}

	// Intrinsic artifacts carry the latch from creation with no pivotal
	// event (spec 4.3): events may raise the score but never grant a pivot
	// or an earned power.
	a := artifacts[0]
	if a.PivotalEventID != "" {
		t.Errorf("PivotalEventID = %q, want empty", a.PivotalEventID)
	}
	if len(a.Powers) != 1 {
		t.Fatalf("Powers = %v, want only the intrinsic power", a.Powers)
	}
	if a.Powers[0].Type() != "influence" {
		t.Errorf("Powers[0].Type() = %q, want influence", a.Powers[0].Type())
	}
}

func TestEarnedPowerStacksWithIntrinsic(t *testing.T) {
	artifacts := []Artifact{{
		ID:                 "artifact-1",
		Name:               "Crown of Deepcrest",
		Type:               "crown",
		SignificanceSource: "historical",
		Status:             "held",
		Powers:             []Power{InfluencePower{Base: 5, Source: "intrinsic"}},
		Provenance: []ProvenanceEntry{
			{Year: 1, Owner: Owner{Kind: "figure", ID: "Deepcrest-1"}, EventID: "event-1-0", EventType: "Discovery"},
		},
	}}
	events := []simulation.Event{
		{Year: 2, Category: "War", ArtifactID: "artifact-1", Description: "war"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(1)); err != nil {
		t.Fatalf("postProcess: %v", err)
	}

	// The crown carries its intrinsic influence power and earns a combat
	// power at the pivotal crossing: they stack (spec 7.1), earned appended
	// after intrinsic with a distinct source.
	a := artifacts[0]
	if len(a.Powers) != 2 {
		t.Fatalf("Powers = %v, want intrinsic + earned", a.Powers)
	}
	ip, ok := a.Powers[0].(InfluencePower)
	if !ok || ip.Source != "intrinsic" {
		t.Errorf("Powers[0] = %+v, want InfluencePower with source intrinsic", a.Powers[0])
	}
	cp, ok := a.Powers[1].(CombatPower)
	if !ok || cp.Source != "earned" {
		t.Errorf("Powers[1] = %+v, want CombatPower with source earned", a.Powers[1])
	}
}

func TestEarnedPowerGrantedOnce(t *testing.T) {
	artifacts := []Artifact{heldArtifact()}
	events := []simulation.Event{
		{Year: 2, Category: "War", ArtifactID: "artifact-1", Description: "war"},
		{Year: 5, Category: "War", ArtifactID: "artifact-1", Description: "war again"},
		{Year: 8, Category: "Conquest", ArtifactID: "artifact-1", Description: "conquest"},
	}

	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(1)); err != nil {
		t.Fatalf("postProcess: %v", err)
	}

	// The monotonic latch flips at the first crossing: later contributions
	// raise the score but never re-grant.
	a := artifacts[0]
	if a.PivotalEventID != "event-2-0" {
		t.Errorf("PivotalEventID = %q, want event-2-0", a.PivotalEventID)
	}
	if a.SignificanceScore != 9 {
		t.Errorf("SignificanceScore = %d, want 9", a.SignificanceScore)
	}
	if len(a.Powers) != 1 {
		t.Errorf("Powers = %v, want exactly one earned power", a.Powers)
	}
	if _, ok := a.Powers[0].(CombatPower); !ok {
		t.Errorf("Powers[0] = %T, want CombatPower from the first crossing", a.Powers[0])
	}
}

func TestEarnedPowerSkippedWithoutLane(t *testing.T) {
	artifacts := []Artifact{heldArtifact()}
	events := []simulation.Event{
		{Year: 2, Category: "War", ArtifactID: "artifact-1", Description: "war"},
	}

	// The nil-lane contract: a nil rng disables all lane draws, so
	// significance still crosses and records the pivotal event, but no
	// earned power is granted. The pipeline always runs through
	// EmergencePass, which threads the lane.
	if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, nil); err != nil {
		t.Fatalf("PostProcess: %v", err)
	}

	a := artifacts[0]
	if a.PivotalEventID != "event-2-0" {
		t.Errorf("PivotalEventID = %q, want event-2-0", a.PivotalEventID)
	}
	if len(a.Powers) != 0 {
		t.Errorf("Powers = %v, want none without the artifacts lane", a.Powers)
	}
}

func TestEarnedPowerMagnitudeDeterministic(t *testing.T) {
	run := func(seed uint64) []Artifact {
		artifacts, events := pivotedArtifacts(12)
		if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(seed)); err != nil {
			t.Fatalf("postProcess: %v", err)
		}
		return artifacts
	}

	first := run(1)
	second := run(1)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("identical seeds produced different powers:\nfirst=%+v\nsecond=%+v", first, second)
	}

	sumBases := func(artifacts []Artifact) int {
		total := 0
		for _, a := range artifacts {
			if len(a.Powers) != 1 {
				t.Fatalf("artifact %s has %d powers, want 1", a.ID, len(a.Powers))
			}
			total += a.Powers[0].BaseMagnitude()
		}
		return total
	}

	// 12 independent draws from a 3-value range: two fixed seeds producing
	// identical sums is improbable enough to pin as a regression check.
	if got, other := sumBases(first), sumBases(run(2)); got == other {
		t.Fatalf("different seeds produced identical magnitude sums %d", got)
	}
}

func TestEmergencePassEarnedPowersDeterministic(t *testing.T) {
	artifacts, events := pivotedArtifacts(6)
	events = append(events, simulation.Event{
		Year: 4, Category: "Discovery", FigureID: "Deepcrest-3", SettlementName: "Deepcrest", Description: "unearths a hoard",
	})
	figures := []FigureContext{
		{ID: "Deepcrest-3", Settlement: "Deepcrest"},
	}

	run := func(seed uint64) ([]Artifact, error) {
		arts := make([]Artifact, len(artifacts))
		copy(arts, artifacts)
		evs := make([]simulation.Event, len(events))
		copy(evs, events)
		got, _, err := EmergencePass(arts, evs, figures, SignificanceContext{}, TransferContext{}, artifactsRNG(seed))
		if err != nil {
			return nil, err
		}
		return got, nil
	}

	first, err := run(1)
	if err != nil {
		t.Fatalf("first EmergencePass: %v", err)
	}
	second, err := run(1)
	if err != nil {
		t.Fatalf("second EmergencePass: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("identical seeds produced different artifacts:\nfirst=%+v\nsecond=%+v", first, second)
	}
	if len(first) != 7 {
		t.Fatalf("artifact count = %d, want 7 (6 initial + 1 emergent)", len(first))
	}
	for i := 0; i < 6; i++ {
		cp, ok := first[i].Powers[0].(CombatPower)
		if !ok || cp.Source != "earned" {
			t.Errorf("initial artifact %d power = %+v, want earned CombatPower", i, first[i].Powers[0])
		}
	}
}

func TestEarnedPowerJSONRoundTripKeepsSource(t *testing.T) {
	a := Artifact{
		ID:                 "artifact-1",
		Name:               "Blade",
		Type:               "weapon",
		SignificanceSource: "historical",
		Status:             "significant",
		Powers: []Power{
			CombatPower{Base: 2, Source: "intrinsic"},
			CombatPower{Base: 3, Source: "earned"},
			InfluencePower{Base: 1, Source: "earned"},
			NarrativePower{Effect: "survives calamity, bearer gains resilience", Source: "earned"},
		},
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got Artifact
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	want := []struct {
		typ    string
		source string
		base   int
	}{
		{"combat", "intrinsic", 2},
		{"combat", "earned", 3},
		{"influence", "earned", 1},
		{"narrative", "earned", 0},
	}
	if len(got.Powers) != len(want) {
		t.Fatalf("got %d powers, want %d", len(got.Powers), len(want))
	}
	for i, w := range want {
		p := got.Powers[i]
		if p.Type() != w.typ || p.BaseMagnitude() != w.base {
			t.Errorf("power %d = (%s, base %d), want (%s, base %d)", i, p.Type(), p.BaseMagnitude(), w.typ, w.base)
		}
		switch v := p.(type) {
		case CombatPower:
			if v.Source != w.source {
				t.Errorf("power %d source = %q, want %q", i, v.Source, w.source)
			}
		case InfluencePower:
			if v.Source != w.source {
				t.Errorf("power %d source = %q, want %q", i, v.Source, w.source)
			}
		case NarrativePower:
			if v.Source != w.source {
				t.Errorf("power %d source = %q, want %q", i, v.Source, w.source)
			}
		}
	}
}
