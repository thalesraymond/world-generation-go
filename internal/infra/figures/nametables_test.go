package figures_test

import (
	randv2 "math/rand/v2"
	"strings"
	"testing"

	domainfigures "github.com/thalesraymond/world-generation-go/internal/domain/figures"
)

func TestInfraNameGenerationDeterministic(t *testing.T) {
	r1 := randv2.New(randv2.NewPCG(99, 88))
	r2 := randv2.New(randv2.NewPCG(99, 88))

	for i := 0; i < 20; i++ {
		n1 := domainfigures.GenerateName(r1)
		n2 := domainfigures.GenerateName(r2)
		if n1 != n2 {
			t.Errorf("iteration %d: names differ: %q vs %q", i, n1, n2)
		}
	}
}

func TestInfraNameFormat(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(99, 88))
	for i := 0; i < 50; i++ {
		name := domainfigures.GenerateName(rng)
		parts := strings.SplitN(name, " ", 2)
		if len(parts) != 2 {
			t.Errorf("name %q has %d parts, want 2", name, len(parts))
		}
		if parts[0] == "" || parts[1] == "" {
			t.Errorf("name %q has empty component", name)
		}
	}
}

func TestInfraNameVariety(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(99, 88))
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := domainfigures.GenerateName(rng)
		seen[name] = true
	}
	if len(seen) < 30 {
		t.Errorf("name variety too low: got %d unique names from 100 calls, want >= 30", len(seen))
	}
}
