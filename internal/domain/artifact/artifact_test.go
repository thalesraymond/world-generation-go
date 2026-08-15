package artifact

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestOwnerValidate(t *testing.T) {
	cases := []struct {
		kind    string
		wantErr bool
	}{
		{"figure", false},
		{"settlement", false},
		{"expedition", false},
		{"lost", false},
		{"unknown", false},
		{"invalid", true},
		{"", true},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			o := Owner{Kind: tc.kind, ID: "test-1"}
			err := o.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestOwnerJSONRoundTrip(t *testing.T) {
	o := Owner{Kind: "figure", ID: "Deepcrest-3"}
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var got Owner
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(got, o) {
		t.Fatalf("round-trip mismatch = %+v, want %+v", got, o)
	}
}

func TestCombatPowerInterface(t *testing.T) {
	var p Power = CombatPower{Base: 4}
	if p.Type() != "combat" {
		t.Fatalf("Type() = %q, want %q", p.Type(), "combat")
	}
	if p.BaseMagnitude() != 4 {
		t.Fatalf("BaseMagnitude() = %d, want 4", p.BaseMagnitude())
	}
	if got := p.EffectiveMagnitude(10); got != 8 {
		t.Fatalf("EffectiveMagnitude(10) = %d, want 8", got)
	}
	if got := p.EffectiveMagnitude(40); got != 20 {
		t.Fatalf("EffectiveMagnitude(40) = %d, want 20 (cap)", got)
	}
	if got := p.EffectiveMagnitude(100); got != 20 {
		t.Fatalf("EffectiveMagnitude(100) = %d, want 20 (cap)", got)
	}
}

func TestInfluencePowerInterface(t *testing.T) {
	var p Power = InfluencePower{Base: 3}
	if p.Type() != "influence" {
		t.Fatalf("Type() = %q, want %q", p.Type(), "influence")
	}
	if p.BaseMagnitude() != 3 {
		t.Fatalf("BaseMagnitude() = %d, want 3", p.BaseMagnitude())
	}
	if got := p.EffectiveMagnitude(20); got != 9 {
		t.Fatalf("EffectiveMagnitude(20) = %d, want 9", got)
	}
}

func TestNarrativePowerInterface(t *testing.T) {
	var p Power = NarrativePower{Effect: "inspires faith"}
	if p.Type() != "narrative" {
		t.Fatalf("Type() = %q, want %q", p.Type(), "narrative")
	}
	if p.BaseMagnitude() != 0 {
		t.Fatalf("BaseMagnitude() = %d, want 0", p.BaseMagnitude())
	}
	if got := p.EffectiveMagnitude(50); got != 0 {
		t.Fatalf("EffectiveMagnitude(50) = %d, want 0", got)
	}
}

func TestArtifactJSONRoundTripWithPowers(t *testing.T) {
	a := Artifact{
		ID:                 "artifact-settlement-0",
		Name:               "Crown of Deepcrest",
		Type:               "crown",
		SignificanceSource: "historical",
		Description:        "A gleaming crown.",
		Status:             "significant",
		SignificanceScore:  5,
		IsSignificant:      true,
		PivotalEventID:     "event-42-0",
		SignificanceYear:   42,
		Provenance: []ProvenanceEntry{
			{Year: 12, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-12-0", EventType: "Conquest"},
			{Year: 42, Owner: Owner{Kind: "figure", ID: "Deepcrest-3"}, EventID: "event-42-0", EventType: "War"},
		},
		AssociatedEventIDs: []string{"event-12-0", "event-42-0"},
		Powers: []Power{
			InfluencePower{Base: 4},
			InfluencePower{Base: 2},
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

	if len(got.Powers) != 2 {
		t.Fatalf("got %d powers, want 2", len(got.Powers))
	}

	if got.Powers[0].Type() != "influence" {
		t.Fatalf("Powers[0].Type() = %q, want %q", got.Powers[0].Type(), "influence")
	}
	if got.Powers[0].BaseMagnitude() != 4 {
		t.Fatalf("Powers[0].BaseMagnitude() = %d, want 4", got.Powers[0].BaseMagnitude())
	}

	if got.Powers[1].BaseMagnitude() != 2 {
		t.Fatalf("Powers[1].BaseMagnitude() = %d, want 2", got.Powers[1].BaseMagnitude())
	}

	if got.ID != a.ID || got.Name != a.Name || got.Type != a.Type {
		t.Fatalf("artifact fields mismatch after round trip")
	}
	if got.SignificanceScore != a.SignificanceScore || got.IsSignificant != a.IsSignificant {
		t.Fatalf("significance fields mismatch after round trip")
	}
	if got.PivotalEventID != a.PivotalEventID || got.SignificanceYear != a.SignificanceYear {
		t.Fatalf("pivotal fields mismatch after round trip")
	}
	if len(got.Provenance) != len(a.Provenance) {
		t.Fatalf("provenance length mismatch")
	}
	if !reflect.DeepEqual(got.Provenance, a.Provenance) {
		t.Fatalf("provenance mismatch")
	}
}

func TestArtifactJSONRoundTripMixedPowers(t *testing.T) {
	a := Artifact{
		ID:                 "artifact-ruin-0",
		Name:               "Sword of the Fallen",
		Type:               "weapon",
		SignificanceSource: "intrinsic",
		Status:             "lost",
		SignificanceScore:  3,
		IsSignificant:      true,
		Provenance:         []ProvenanceEntry{},
		Powers: []Power{
			CombatPower{Base: 2},
			NarrativePower{Effect: "survives calamity"},
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

	if len(got.Powers) != 2 {
		t.Fatalf("got %d powers, want 2", len(got.Powers))
	}

	cp, ok := got.Powers[0].(CombatPower)
	if !ok {
		t.Fatalf("Powers[0] type = %T, want CombatPower", got.Powers[0])
	}
	if cp.Base != 2 {
		t.Fatalf("CombatPower.Base = %d, want 2", cp.Base)
	}

	np, ok := got.Powers[1].(NarrativePower)
	if !ok {
		t.Fatalf("Powers[1] type = %T, want NarrativePower", got.Powers[1])
	}
	if np.Effect != "survives calamity" {
		t.Fatalf("NarrativePower.Effect = %q, want %q", np.Effect, "survives calamity")
	}
}

func TestArtifactJSONRoundTripNoPowers(t *testing.T) {
	a := Artifact{
		ID:                 "artifact-settlement-1",
		Name:               "Iron Shield",
		Type:               "armor",
		SignificanceSource: "historical",
		Status:             "held",
		SignificanceScore:  1,
		IsSignificant:      false,
		Provenance:         []ProvenanceEntry{{Year: 50, Owner: Owner{Kind: "settlement", ID: "Ironforge"}, EventID: "event-50-0", EventType: "Creation"}},
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got Artifact
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.Powers != nil {
		t.Fatalf("expected nil Powers, got %v", got.Powers)
	}
	if got.ID != a.ID || got.Name != a.Name || got.Status != a.Status {
		t.Fatalf("artifact mismatch after round trip")
	}
}

func TestArtifactBackwardCompatNoPowers(t *testing.T) {
	oldJSON := `{
		"id": "artifact-settlement-0",
		"name": "Old Relic",
		"type": "relic",
		"significanceSource": "intrinsic",
		"status": "lost",
		"significanceScore": 3,
		"isSignificant": true,
		"provenance": []
	}`

	var got Artifact
	if err := json.Unmarshal([]byte(oldJSON), &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.ID != "artifact-settlement-0" {
		t.Fatalf("ID = %q, want %q", got.ID, "artifact-settlement-0")
	}
	if got.Powers != nil {
		t.Fatalf("expected nil Powers, got %v", got.Powers)
	}
}

func TestArtifactUnmarshalUnknownPowerType(t *testing.T) {
	badJSON := `{
		"id": "x",
		"name": "x",
		"type": "weapon",
		"significanceSource": "historical",
		"status": "held",
		"significanceScore": 0,
		"isSignificant": false,
		"provenance": [],
		"powers": [{"type": "psionic", "base": 5}]
	}`

	var got Artifact
	if err := json.Unmarshal([]byte(badJSON), &got); err == nil {
		t.Fatal("expected error for unknown power type")
	}
}
