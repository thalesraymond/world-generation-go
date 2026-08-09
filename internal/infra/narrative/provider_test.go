package narrative

import (
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

// Compile-time check that DefaultGrammarProvider satisfies the usecase
// GrammarProvider interface without importing it (structural typing).
var _ simulation.GrammarProvider = DefaultGrammarProvider{}

func TestDefaultGrammarProvider_Grammar(t *testing.T) {
	p := DefaultGrammarProvider{}
	got := p.Grammar()
	if got == "" {
		t.Fatal("Grammar() returned empty source")
	}
	if got != DefaultGrammar {
		t.Error("Grammar() must return the DefaultGrammar source")
	}
}
