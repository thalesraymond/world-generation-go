package artifact

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

func TestIntrinsicPower(t *testing.T) {
	tests := []struct {
		typ      string
		wantType string
		wantBase int
		wantOK   bool
	}{
		{typ: "weapon", wantType: "combat", wantBase: 2, wantOK: true},
		{typ: "armor", wantType: "combat", wantBase: 2, wantOK: true},
		{typ: "crown", wantType: "influence", wantBase: 5, wantOK: true},
		{typ: "jewelry", wantType: "influence", wantBase: 1, wantOK: true},
		{typ: "relic", wantType: "narrative", wantOK: true},
		{typ: "tome", wantType: "narrative", wantOK: true},
		{typ: "lantern", wantOK: false},
		{typ: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			p, ok := IntrinsicPower(tt.typ)
			if ok != tt.wantOK {
				t.Fatalf("IntrinsicPower(%q) ok = %v, want %v", tt.typ, ok, tt.wantOK)
			}
			if !tt.wantOK {
				if p != nil {
					t.Fatalf("IntrinsicPower(%q) = %v, want nil", tt.typ, p)
				}
				return
			}
			if p.Type() != tt.wantType {
				t.Errorf("IntrinsicPower(%q) Type() = %q, want %q", tt.typ, p.Type(), tt.wantType)
			}
			if p.BaseMagnitude() != tt.wantBase {
				t.Errorf("IntrinsicPower(%q) BaseMagnitude() = %d, want %d", tt.typ, p.BaseMagnitude(), tt.wantBase)
			}
			if got := p.EffectiveMagnitude(0); got != tt.wantBase {
				t.Errorf("IntrinsicPower(%q) EffectiveMagnitude(0) = %d, want %d (effective equals base at creation)", tt.typ, got, tt.wantBase)
			}
			if source := powerSourceForTest(p); source != "intrinsic" {
				t.Errorf("IntrinsicPower(%q) Source = %q, want intrinsic", tt.typ, source)
			}
		})
	}
}

func TestIntrinsicPowerNarrativeEffects(t *testing.T) {
	tests := []struct {
		typ        string
		wantEffect string
		wantNoMag  bool
	}{
		{typ: "relic", wantEffect: "inspires faith in followers", wantNoMag: true},
		{typ: "tome", wantEffect: "reveals hidden knowledge", wantNoMag: true},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			p, ok := IntrinsicPower(tt.typ)
			if !ok {
				t.Fatalf("IntrinsicPower(%q) ok = false, want true", tt.typ)
			}
			n, isNarrative := p.(NarrativePower)
			if !isNarrative {
				t.Fatalf("IntrinsicPower(%q) = %T, want NarrativePower", tt.typ, p)
			}
			if n.Effect != tt.wantEffect {
				t.Errorf("IntrinsicPower(%q) Effect = %q, want %q", tt.typ, n.Effect, tt.wantEffect)
			}
			if p.BaseMagnitude() != 0 || p.EffectiveMagnitude(42) != 0 {
				t.Errorf("IntrinsicPower(%q) magnitudes = (%d, %d), want (0, 0)", tt.typ, p.BaseMagnitude(), p.EffectiveMagnitude(42))
			}
		})
	}
}

func TestArtifactPowerJSONRoundTrip(t *testing.T) {
	artifacts := []Artifact{
		{Powers: []Power{CombatPower{Base: 2, Source: "intrinsic"}}},
		{Powers: []Power{InfluencePower{Base: 5, Source: "intrinsic"}}},
		{Powers: []Power{NarrativePower{Effect: "reveals hidden knowledge", Source: "intrinsic"}}},
		{Powers: []Power{
			CombatPower{Base: 2, Source: "intrinsic"},
			InfluencePower{Base: 5, Source: "intrinsic"},
			NarrativePower{Effect: "inspires faith in followers", Source: "intrinsic"},
		}},
	}

	for i, want := range artifacts {
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatalf("case %d: marshal: %v", i, err)
		}
		var got Artifact
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("case %d: unmarshal: %v", i, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("case %d: round-trip mismatch:\n got %+v\nwant %+v", i, got, want)
		}
	}
}

func TestPowerFromJSONUnknownType(t *testing.T) {
	data := []byte(`{"powers":[{"type":"unknown","base":2}]}`)
	var a Artifact
	if err := json.Unmarshal(data, &a); err == nil {
		t.Fatal("expected error unmarshaling unknown power type, got nil")
	}
}

// poweredArtifact builds an artifact carrying one of each power kind so the
// AppliedPowers status tests exercise the full slice.
func poweredArtifact(status string) Artifact {
	return Artifact{
		ID:     "artifact-1",
		Name:   "Crown of Deepcrest",
		Type:   "crown",
		Status: status,
		Powers: []Power{
			CombatPower{Base: 2, Source: "intrinsic"},
			InfluencePower{Base: 5, Source: "intrinsic"},
			NarrativePower{Effect: "inspires faith in followers", Source: "intrinsic"},
		},
	}
}

func TestAppliedPowersDormantWhileLost(t *testing.T) {
	a := poweredArtifact("lost")
	if len(a.Powers) == 0 {
		t.Fatal("fixture must carry powers: dormancy means they exist but do not apply")
	}
	if got := AppliedPowers(a); got != nil {
		t.Errorf("AppliedPowers(lost) = %v, want nil (powers dormant, spec 7.7)", got)
	}
}

func TestAppliedPowersResumeOnRediscovery(t *testing.T) {
	a := poweredArtifact("lost")
	if AppliedPowers(a) != nil {
		t.Fatal("powers must be dormant while lost")
	}
	// Rediscovery transitions status back to held (issue #70); powers resume.
	a.Status = "held"
	if got := AppliedPowers(a); !reflect.DeepEqual(got, a.Powers) {
		t.Errorf("AppliedPowers(held after loss) = %v, want %v (powers resume, spec 7.7)", got, a.Powers)
	}
}

func TestAppliedPowersDestroyed(t *testing.T) {
	// Destruction clears powers (spec 7.7); the seam also never applies
	// powers to a terminal artifact even if one carries them.
	a := poweredArtifact("destroyed")
	if got := AppliedPowers(a); got != nil {
		t.Errorf("AppliedPowers(destroyed) = %v, want nil (powers vanish)", got)
	}
}

func TestAppliedPowersActiveStatuses(t *testing.T) {
	for _, status := range []string{"created", "held", "significant", "rediscovered"} {
		t.Run(status, func(t *testing.T) {
			a := poweredArtifact(status)
			if got := AppliedPowers(a); !reflect.DeepEqual(got, a.Powers) {
				t.Errorf("AppliedPowers(%s) = %v, want %v", status, got, a.Powers)
			}
		})
	}
}

// TestAppliedPowersFailClosedUnknownStatus pins the fail-closed contract:
// a status outside the documented vocabulary — empty (zero value) or
// unknown — applies no powers, so a typo or future status never leaks them.
func TestAppliedPowersFailClosedUnknownStatus(t *testing.T) {
	for _, status := range []string{"", "gremlin"} {
		t.Run(strconv.Quote(status), func(t *testing.T) {
			a := poweredArtifact(status)
			if got := AppliedPowers(a); got != nil {
				t.Errorf("AppliedPowers(%q) = %v, want nil (fail-closed)", status, got)
			}
		})
	}
}

// TestAppliedPowersUnaffectedByOwnershipChange pins the transfer half of
// spec 7.7: powers are intrinsic to the artifact, so an ownership change —
// recorded here as a new provenance entry, exactly what the transfer
// machinery produces (§6.3, transfers only mutate provenance) — never
// changes what applies. The seam gates on status alone.
func TestAppliedPowersUnaffectedByOwnershipChange(t *testing.T) {
	a := poweredArtifact("significant")
	before := AppliedPowers(a)
	a.Provenance = append(a.Provenance, ProvenanceEntry{
		Year:      12,
		Owner:     Owner{Kind: "settlement", ID: "Ironforge"},
		EventID:   "event-12-0",
		EventType: "Conquest",
	})
	if got := AppliedPowers(a); !reflect.DeepEqual(got, before) {
		t.Errorf("AppliedPowers after ownership change = %v, want %v (powers follow the artifact)", got, before)
	}
}

// TestEffectiveMagnitudeDeterministic runs the post-processing pass twice
// with identical seeds and inputs and requires byte-identical effective
// magnitudes (spec 7.6, issue #75 acceptance: same seed produces identical
// effective magnitudes). The pinned values catch regressions in the scaling
// formula itself, not just run-to-run drift.
func TestEffectiveMagnitudeDeterministic(t *testing.T) {
	run := func(seed uint64) []Artifact {
		artifacts, events := pivotedArtifacts(4)
		if err := PostProcess(artifacts, events, SignificanceContext{}, TransferContext{}, artifactsRNG(seed)); err != nil {
			t.Fatalf("PostProcess: %v", err)
		}
		return artifacts
	}

	first := run(1)
	second := run(1)
	for i := range first {
		a := first[i]
		if len(a.Powers) != 1 {
			t.Fatalf("artifact %s has %d powers, want 1", a.ID, len(a.Powers))
		}
		got := a.Powers[0].EffectiveMagnitude(a.SignificanceScore)
		if want := second[i].Powers[0].EffectiveMagnitude(second[i].SignificanceScore); got != want {
			t.Errorf("artifact %s effective magnitude differs across identical runs: %d vs %d", a.ID, got, want)
		}
		if got == 0 {
			t.Errorf("artifact %s effective magnitude = 0, want nonzero (score %d)", a.ID, a.SignificanceScore)
		}
	}

	// Pin expected magnitudes so a regression in the formula is caught.
	for i, want := range []int{3, 1, 3, 3} {
		a := first[i]
		if got := a.Powers[0].EffectiveMagnitude(a.SignificanceScore); got != want {
			t.Errorf("artifact %s effective magnitude = %d, want %d (score %d)", a.ID, got, want, a.SignificanceScore)
		}
	}

	// Magnitudes must derive from the lane, not hardcoded bases. Seeds 1 and
	// 2 are a fixed pair verified to diverge (seed 1: [3 1 3 3], seed 2:
	// [1 1 2 2] for PCG(seed, 1)), so the check is deterministic — no
	// probabilistic assertion that could flake in CI.
	other := run(2)
	for i := range first {
		got := first[i].Powers[0].EffectiveMagnitude(first[i].SignificanceScore)
		want := other[i].Powers[0].EffectiveMagnitude(other[i].SignificanceScore)
		if got != want {
			return // divergent, as verified
		}
	}
	t.Fatal("different seeds produced identical effective magnitudes: draws ignore the RNG lane")
}

// powerSourceForTest extracts the Source field from a power for assertions.
func powerSourceForTest(p Power) string {
	switch v := p.(type) {
	case CombatPower:
		return v.Source
	case InfluencePower:
		return v.Source
	case NarrativePower:
		return v.Source
	}
	return ""
}
