package settlement

import (
	randv2 "math/rand/v2"
	"testing"
)

func TestGenerateNameIsDeterministic(t *testing.T) {
	rng1 := randv2.New(randv2.NewPCG(42, 42))
	rng2 := randv2.New(randv2.NewPCG(42, 42))

	name1 := GenerateName(rng1)
	name2 := GenerateName(rng2)

	if name1 != name2 {
		t.Errorf("same seed should produce same name, got %q and %q", name1, name2)
	}
}

func TestGenerateNameProducesNonEmpty(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(1, 1))
	name := GenerateName(rng)
	if name == "" {
		t.Error("name should not be empty")
	}
}

func TestEnsureUniqueNameNoCollision(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(0, 0))
	used := map[string]bool{}
	name := EnsureUniqueName(rng, used)
	if name == "" {
		t.Error("name should not be empty")
	}
}

func TestEnsureUniqueNameHandlesCollision(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(99, 99))
	name := GenerateName(rng)
	rng2 := randv2.New(randv2.NewPCG(99, 99))
	GenerateName(rng2) // advance to same state

	used := map[string]bool{name: true}
	result := EnsureUniqueName(rng2, used)
	if result == name {
		t.Errorf("expected different name due to collision, got same name %q", result)
	}
}
