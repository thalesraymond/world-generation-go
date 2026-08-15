package simulation

import (
	randv2 "math/rand/v2"
	"testing"

	dompointcrawl "github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func TestNewAgentEnvReturnsAgentEnv(t *testing.T) {
	env := NewAgentEnv(nil, nil, nil, nil)
	if env == nil {
		t.Fatal("NewAgentEnv returned nil")
	}
}

func TestAgentEnvSuitability(t *testing.T) {
	ws := world.NewState(4, 4)
	ws.Suitability[5] = 0.7

	env := NewAgentEnv(ws, nil, nil, nil)

	if got := env.Suitability(1, 1); got != 0.7 {
		t.Errorf("Suitability(1, 1) = %v, want 0.7", got)
	}
	if got := env.Suitability(3, 3); got != 0 {
		t.Errorf("Suitability(3, 3) = %v, want 0", got)
	}
	if got := env.Suitability(9, 0); got != 0 {
		t.Errorf("Suitability out of bounds = %v, want 0", got)
	}
}

func TestAgentEnvSuitabilityNilState(t *testing.T) {
	env := NewAgentEnv(nil, nil, nil, nil)
	if got := env.Suitability(0, 0); got != 0 {
		t.Errorf("Suitability with nil state = %v, want 0", got)
	}
}

func TestAgentEnvFindExpansionTargetNilGraph(t *testing.T) {
	self := &world.Settlement{Name: "Haven", X: 1, Y: 1, Faction: "testers"}
	settlements := []world.Settlement{*self}

	env := NewAgentEnv(nil, nil, &settlements, nil)
	if x, y, ok := env.FindExpansionTarget(self, randv2.New(randv2.NewPCG(1, 0))); ok {
		t.Errorf("FindExpansionTarget with nil graph = (%d, %d, true), want ok=false", x, y)
	}
}

func TestAgentEnvFindExpansionTarget(t *testing.T) {
	graph := dompointcrawl.NewGraph()
	graph.AddNode(&dompointcrawl.Node{ID: 1, X: 5, Y: 1, Visibility: dompointcrawl.Unknown, Name: "Ruin", Kind: "ruin"})

	self := &world.Settlement{Name: "Haven", X: 1, Y: 1, Faction: "testers"}
	settlements := []world.Settlement{*self}

	env := NewAgentEnv(nil, graph, &settlements, nil)
	x, y, ok := env.FindExpansionTarget(self, randv2.New(randv2.NewPCG(2, 0)))
	if !ok {
		t.Fatal("FindExpansionTarget did not find the ruin node")
	}
	if x != 5 || y != 1 {
		t.Errorf("FindExpansionTarget = (%d, %d), want (5, 1)", x, y)
	}
}

func TestAgentEnvGenerateNameUnique(t *testing.T) {
	env := NewAgentEnv(nil, nil, nil, map[string]bool{"Haven": true})
	rng := randv2.New(randv2.NewPCG(3, 0))

	first := env.GenerateName(rng)
	if first == "" {
		t.Fatal("GenerateName returned an empty name")
	}
	if first == "Haven" {
		t.Errorf("GenerateName returned used name %q", first)
	}

	second := env.GenerateName(rng)
	if second == first {
		t.Errorf("GenerateName returned duplicate %q", second)
	}
}

func TestAgentEnvMaxActionRange(t *testing.T) {
	env := NewAgentEnv(nil, nil, nil, nil)
	if got := env.MaxActionRange(); got != agentMaxActionRange {
		t.Errorf("MaxActionRange() = %v, want %v", got, agentMaxActionRange)
	}
}
