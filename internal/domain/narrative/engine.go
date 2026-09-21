package narrative

import (
	"errors"
	"fmt"
	"os"
	"strings"

	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

const defaultMaxDepth = 10

var (
	// ErrRuleNotFound is returned when Resolve targets a rule absent from the grammar.
	ErrRuleNotFound = fmt.Errorf("rule not found in grammar")
	// ErrNoEligibleAlternative is returned when no alternative of a rule has all
	// of its direct variables present and non-empty in the context.
	ErrNoEligibleAlternative = fmt.Errorf("no eligible alternative")
)

// Engine expands context-free grammar rules into narrative text.
type Engine struct {
	grammar  *Grammar
	maxDepth int
}

// NewEngine creates an Engine from a parsed grammar.
func NewEngine(g *Grammar) *Engine {
	return &Engine{
		grammar:  g,
		maxDepth: defaultMaxDepth,
	}
}

// NewEngineFromFile reads a grammar file from disk and returns an Engine.
func NewEngineFromFile(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read grammar file: %w", err)
	}
	return NewEngineFromString(string(data))
}

// NewEngineFromString parses a grammar string and returns an Engine.
func NewEngineFromString(grammarStr string) (*Engine, error) {
	g, err := Parse(grammarStr)
	if err != nil {
		return nil, fmt.Errorf("parse grammar: %w", err)
	}
	return NewEngine(g), nil
}

// SetMaxDepth configures the recursion limit used during resolution.
func (e *Engine) SetMaxDepth(d int) {
	e.maxDepth = d
}

// Resolve expands the named rule into narrative text.
//
// An alternative is eligible when every direct variable it references is
// present and non-empty in context; the RNG draws uniformly among the
// eligible alternatives only. If no alternative is eligible, the error
// ErrNoEligibleAlternative (wrapped with the rule name) is returned and
// no backtracking is attempted — callers implement the fallback chain.
// If recursion exceeds maxDepth, the expansion falls back to "[...]".
func (e *Engine) Resolve(ruleName string, context map[string]string, rng *randv2.Rand) (string, error) {
	return e.resolve(ruleName, context, rng, 0)
}

func (e *Engine) Narrate(event simulation.Event, extraContext map[string]string, rng *randv2.Rand) (string, error) {
	return e.NarrateWithRule(event, extraContext, rng, event.Category)
}

// NarrateWithRule expands an event using a specific grammar rule instead of the
// event's category. It falls back to the event description if the rule is not
// found.
func (e *Engine) NarrateWithRule(event simulation.Event, extraContext map[string]string, rng *randv2.Rand, ruleName string) (string, error) {
	ctx := make(map[string]string, len(extraContext)+3)
	for k, v := range extraContext {
		ctx[k] = v
	}
	ctx["year"] = fmt.Sprint(event.Year)
	ctx["category"] = event.Category
	ctx["description"] = event.Description

	text, err := e.Resolve(ruleName, ctx, rng)
	if err != nil {
		if errors.Is(err, ErrRuleNotFound) {
			return event.Description, nil
		}
		return "", err
	}
	return text, nil
}

// AlternativeVariables returns the names of all direct variables referenced
// in an alternative. Variables nested inside referenced non-terminals are
// not included — each rule filters those when it expands.
func AlternativeVariables(alt Alternative) []string {
	var names []string
	for _, sym := range alt {
		if v, ok := sym.(Variable); ok {
			names = append(names, v.Name)
		}
	}
	return names
}

// isAlternativeEligible reports whether every direct variable referenced by
// the alternative is present and non-empty in the context.
func isAlternativeEligible(alt Alternative, context map[string]string) bool {
	for _, name := range AlternativeVariables(alt) {
		if v, ok := context[name]; !ok || v == "" {
			return false
		}
	}
	return true
}

func (e *Engine) resolve(ruleName string, context map[string]string, rng *randv2.Rand, depth int) (string, error) {
	rule, ok := e.grammar.Rules[ruleName]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrRuleNotFound, ruleName)
	}

	if depth > e.maxDepth {
		return "[...]", nil
	}

	if len(rule.Alternatives) == 0 {
		return "", fmt.Errorf("rule %q has no alternatives", ruleName)
	}

	var eligible []Alternative
	for _, alt := range rule.Alternatives {
		if isAlternativeEligible(alt, context) {
			eligible = append(eligible, alt)
		}
	}
	if len(eligible) == 0 {
		return "", fmt.Errorf("%w: %q", ErrNoEligibleAlternative, ruleName)
	}

	alt := eligible[rng.IntN(len(eligible))]

	var out strings.Builder
	for _, sym := range alt {
		switch s := sym.(type) {
		case Terminal:
			out.WriteString(s.Text)
		case NonTerminal:
			text, err := e.resolve(s.Name, context, rng, depth+1)
			if err != nil {
				return "", err
			}
			out.WriteString(text)
		case Variable:
			if v, ok := context[s.Name]; ok {
				out.WriteString(v)
			} else {
				out.WriteString("$" + s.Name)
			}
		}
	}

	return out.String(), nil
}
