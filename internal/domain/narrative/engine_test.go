package narrative

import (
	"errors"
	"strings"
	"testing"

	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

type zeroSource struct{}

func (zeroSource) Uint64() uint64 { return 0 }

func TestEngine_Resolve_Terminal(t *testing.T) {
	eng := NewEngine(&Grammar{
		Rules: map[string]Rule{
			"greet": {Name: "greet", Alternatives: []Alternative{
				{Terminal{Text: "hello"}},
			}},
		},
	})

	rng := randv2.New(randv2.NewPCG(1, 2))
	got, err := eng.Resolve("greet", nil, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestEngine_Resolve_NonTerminal(t *testing.T) {
	s := `
phrase ::= <greeting> " world"
greeting ::= "hello"
`
	eng, err := NewEngineFromString(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	rng := randv2.New(randv2.NewPCG(3, 4))
	got, err := eng.Resolve("phrase", nil, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", got)
	}
}

func TestEngine_Resolve_Variable(t *testing.T) {
	eng := NewEngine(&Grammar{
		Rules: map[string]Rule{
			"say": {Name: "say", Alternatives: []Alternative{
				{Terminal{Text: "hello, "}, Variable{Name: "name"}},
			}},
		},
	})

	rng := randv2.New(randv2.NewPCG(5, 6))
	got, err := eng.Resolve("say", map[string]string{"name": "Ada"}, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello, Ada" {
		t.Fatalf("expected %q, got %q", "hello, Ada", got)
	}
}

func TestEngine_Resolve_MissingVariable(t *testing.T) {
	eng := NewEngine(&Grammar{
		Rules: map[string]Rule{
			"say": {Name: "say", Alternatives: []Alternative{
				{Variable{Name: "missing"}},
			}},
		},
	})

	rng := randv2.New(randv2.NewPCG(7, 8))
	got, err := eng.Resolve("say", nil, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "$missing" {
		t.Fatalf("expected %q, got %q", "$missing", got)
	}
}

func TestEngine_Resolve_RandomAlternative(t *testing.T) {
	eng := NewEngine(&Grammar{
		Rules: map[string]Rule{
			"coin": {Name: "coin", Alternatives: []Alternative{
				{Terminal{Text: "heads"}},
				{Terminal{Text: "tails"}},
			}},
		},
	})

	var heads, tails int
	rng := randv2.New(randv2.NewPCG(9, 10))
	for i := 0; i < 100; i++ {
		got, err := eng.Resolve("coin", nil, rng)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		switch got {
		case "heads":
			heads++
		case "tails":
			tails++
		default:
			t.Fatalf("unexpected output: %q", got)
		}
	}
	if heads == 0 || tails == 0 {
		t.Fatalf("expected both alternatives to be selected, got heads=%d tails=%d", heads, tails)
	}
}

func TestEngine_Resolve_DepthLimit(t *testing.T) {
	s := `
a ::= "(" <a> ")" | "x"
`
	eng, err := NewEngineFromString(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	rng := randv2.New(zeroSource{})

	got, err := eng.Resolve("a", nil, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// maxDepth=10, so recursion stops at depth 11, producing 11 nested wrappers.
	want := strings.Repeat("(", 11) + "[...]" + strings.Repeat(")", 11)
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEngine_Resolve_RuleNotFound(t *testing.T) {
	eng := NewEngine(&Grammar{
		Rules: map[string]Rule{
			"start": {Name: "start", Alternatives: []Alternative{
				{NonTerminal{Name: "missing"}},
			}},
		},
	})

	rng := randv2.New(randv2.NewPCG(13, 14))
	_, err := eng.Resolve("start", nil, rng)
	if err == nil {
		t.Fatal("expected error for missing rule")
	}
	if !errors.Is(err, ErrRuleNotFound) {
		t.Fatalf("expected ErrRuleNotFound, got %v", err)
	}
}

func TestEngine_Resolve_Determinism(t *testing.T) {
	s := `
color ::= "red" | "green" | "blue"
`
	eng, err := NewEngineFromString(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	const seed1, seed2 = 7, 42
	rng := randv2.New(randv2.NewPCG(seed1, seed2))
	first, err := eng.Resolve("color", nil, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rng2 := randv2.New(randv2.NewPCG(seed1, seed2))
	second, err := eng.Resolve("color", nil, rng2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != second {
		t.Fatalf("same seed produced different outputs: %q vs %q", first, second)
	}
}

func TestEngine_SetMaxDepth(t *testing.T) {
	s := `
a ::= "(" <a> ")" | "x"
`
	eng, err := NewEngineFromString(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	eng.SetMaxDepth(2)

	rng := randv2.New(zeroSource{})

	got, err := eng.Resolve("a", nil, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// maxDepth=2, so recursion stops at depth 3.
	want := strings.Repeat("(", 3) + "[...]" + strings.Repeat(")", 3)
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNewEngineFromString_ParseError(t *testing.T) {
	_, err := NewEngineFromString("not a grammar")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestEngine_Narrate_ResolvesRule(t *testing.T) {
	s := `
battle ::= "In " $year ", " $faction " fought at " $location "."
`
	eng, err := NewEngineFromString(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	event := simulation.Event{
		Year:        450,
		Category:    "battle",
		Description: "A skirmish occurred.",
	}
	extra := map[string]string{
		"faction":  "Red Legion",
		"location": "Ashmoor",
	}
	rng := randv2.New(randv2.NewPCG(1, 2))

	got, err := eng.Narrate(event, extra, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "In 450, Red Legion fought at Ashmoor."
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEngine_Narrate_RuleNotFound(t *testing.T) {
	s := `
battle ::= "In " $year " a battle raged."
`
	eng, err := NewEngineFromString(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	event := simulation.Event{
		Year:        450,
		Category:    "plague",
		Description: "Disease swept the land.",
	}
	rng := randv2.New(randv2.NewPCG(1, 2))

	got, err := eng.Narrate(event, nil, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != event.Description {
		t.Fatalf("expected %q, got %q", event.Description, got)
	}
}

func TestEngine_Narrate_ExtraContextMerged(t *testing.T) {
	s := `
greet ::= "Salutations, " $title " " $name ". The year is " $year "."
`
	eng, err := NewEngineFromString(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	event := simulation.Event{
		Year:        1000,
		Category:    "greet",
		Description: "ignored",
	}
	extra := map[string]string{
		"title": "Lord",
		"name":  "Vader",
	}
	rng := randv2.New(randv2.NewPCG(1, 2))

	got, err := eng.Narrate(event, extra, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Salutations, Lord Vader. The year is 1000."
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEngine_Narrate_Determinism(t *testing.T) {
	s := `
omen ::= "The omens were " $mood " in " $year "."
`
	eng, err := NewEngineFromString(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	event := simulation.Event{
		Year:        7,
		Category:    "omen",
		Description: "signs appeared",
	}
	extra := map[string]string{"mood": "dark"}

	const seed1, seed2 = 11, 13
	rng1 := randv2.New(randv2.NewPCG(seed1, seed2))
	first, err := eng.Narrate(event, extra, rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rng2 := randv2.New(randv2.NewPCG(seed1, seed2))
	second, err := eng.Narrate(event, extra, rng2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != second {
		t.Fatalf("same seed produced different outputs: %q vs %q", first, second)
	}
}
