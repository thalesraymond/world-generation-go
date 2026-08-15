package artifact

import (
	"encoding/json"
	"reflect"
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
