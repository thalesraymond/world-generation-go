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
// Variables are substituted from context; missing variables emit "$name".
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

	alt := rule.Alternatives[rng.IntN(len(rule.Alternatives))]

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
