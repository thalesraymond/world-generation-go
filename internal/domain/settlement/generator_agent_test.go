package settlement

import (
	randv2 "math/rand/v2"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// agentStateFixture builds a small deterministic world for agent state tests.
func agentStateFixture(t *testing.T, seed uint64) *world.State {
	t.Helper()

	state := world.NewState(3, 1)
	state.Suitability = []float64{0.9, 0.8, 0.7}
	state.PopulationDensity = []float64{0.9, 0.8, 0.7}
	state.FactionInfluence = []string{"auric", "auric", "verdant"}

	config := DefaultConfig()
	config.MinSuitability = 0.6
	config.MinPopulation = 0.6
	config.MinDistance = 1
	config.MaxPopulation = 1000
	config.RNG = randv2.New(randv2.NewPCG(seed, seed^0x9e3779b9))

	if err := Generate(state, config); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	return state
}

func TestGenerateInitializesAgentState(t *testing.T) {
	state := agentStateFixture(t, 42)

	if len(state.Settlements) != 3 {
		t.Fatalf("settlement count = %d, want 3", len(state.Settlements))
	}

	seen := make(map[string]bool, len(state.Settlements))
	for _, s := range state.Settlements {
		seen[s.Name] = true

		wantMilitary := s.Population * MilitaryPopulationRatio
		if s.MilitaryStrength != wantMilitary {
			t.Errorf("%s: MilitaryStrength = %v, want %v", s.Name, s.MilitaryStrength, wantMilitary)
		}

		if s.Wealth != InitialWealth {
			t.Errorf("%s: Wealth = %v, want %v", s.Name, s.Wealth, InitialWealth)
		}

		if len(s.Goals) < 2 || len(s.Goals) > 3 {
			t.Errorf("%s: goal count = %d, want 2-3 (%v)", s.Name, len(s.Goals), s.Goals)
		}

		if len(s.Relations) != len(state.Settlements)-1 {
			t.Errorf("%s: relations count = %d, want %d", s.Name, len(s.Relations), len(state.Settlements)-1)
		}
		if _, ok := s.Relations[s.Name]; ok {
			t.Errorf("%s: relations must exclude self", s.Name)
		}
	}

	// Same-faction pairs share the +0.3 baseline; cross-faction pairs are
	// negative after ApplyCrossFactionFriction.
	for _, s := range state.Settlements {
		for _, other := range state.Settlements {
			if other.Name == s.Name {
				continue
			}
			got := s.Relations[other.Name]
			if other.Faction == s.Faction && s.Faction != "independent" {
				want := world.RelationShiftSameFactionBaseline
				if got != want {
					t.Errorf("%s relations toward %s = %v, want %v", s.Name, other.Name, got, want)
				}
			} else {
				// Cross-faction (or independent) — must be ≤ 0 after friction.
				if got > 0 {
					t.Errorf("%s relations toward %s = %v, want ≤ 0", s.Name, other.Name, got)
				}
			}
		}
	}
}

func TestGenerateAgentStateDeterministic(t *testing.T) {
	first := agentStateFixture(t, 42)
	second := agentStateFixture(t, 42)

	if len(first.Settlements) != len(second.Settlements) {
		t.Fatalf("settlement counts differ: %d vs %d", len(first.Settlements), len(second.Settlements))
	}

	for i := range first.Settlements {
		a := first.Settlements[i]
		b := second.Settlements[i]

		if a.Name != b.Name {
			t.Fatalf("settlement %d name differs: %q vs %q", i, a.Name, b.Name)
		}
		if a.MilitaryStrength != b.MilitaryStrength {
			t.Errorf("%s: MilitaryStrength differs: %v vs %v", a.Name, a.MilitaryStrength, b.MilitaryStrength)
		}
		if a.Wealth != b.Wealth {
			t.Errorf("%s: Wealth differs: %v vs %v", a.Name, a.Wealth, b.Wealth)
		}
		if len(a.Goals) != len(b.Goals) {
			t.Fatalf("%s: goal count differs: %d vs %d", a.Name, len(a.Goals), len(b.Goals))
		}
		for j := range a.Goals {
			if a.Goals[j] != b.Goals[j] {
				t.Errorf("%s: goal %d differs: %q vs %q", a.Name, j, a.Goals[j], b.Goals[j])
			}
		}
		if len(a.Relations) != len(b.Relations) {
			t.Fatalf("%s: relations count differs: %d vs %d", a.Name, len(a.Relations), len(b.Relations))
		}
		for name, value := range a.Relations {
			if b.Relations[name] != value {
				t.Errorf("%s: relations[%q] differs: %v vs %v", a.Name, name, value, b.Relations[name])
			}
		}
	}
}
