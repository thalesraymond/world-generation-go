package simulation

import (
	"bytes"
	"context"
	"strings"
	"testing"

	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	infranarrative "github.com/thalesraymond/world-generation-go/internal/infra/narrative"
	ucsim "github.com/thalesraymond/world-generation-go/internal/usecase/simulation"
)

func newFactoryTestState() *world.State {
	return &world.State{
		Settlements: []world.Settlement{
			{
				Name: "Deepcrest",
				Figures: []figures.HistoricalFigure{
					{ID: "Deepcrest-0", Name: "Aldric", Role: "General"},
				},
			},
		},
	}
}

func factoryTestEvents() []domsim.Event {
	return []domsim.Event{
		{Year: 1, Category: "Conflict", FigureID: "Deepcrest-0", SettlementName: "Deepcrest", TargetSettlement: "Northhold", Description: "Deepcrest raided Northhold."},
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
		{Year: 1, Category: "Economy", SettlementName: "Ashfield", Description: "Ashfield prospers."},
		{Year: 2, Category: "Politics", SettlementName: "Deepcrest", TargetSettlement: "Ashfield", Description: "Deepcrest negotiated with Ashfield."},
	}
}

func TestNewChronicleForWorld(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(1, 0))
	state := newFactoryTestState()

	if _, err := NewChronicleForWorld(nil, state, "normal"); err == nil {
		t.Error("expected error for nil RNG")
	}
	if _, err := NewChronicleForWorld(rng, nil, "normal"); err == nil {
		t.Error("expected error for nil state")
	}
	if _, err := NewChronicleForWorld(rng, state, "loud"); err == nil {
		t.Error("expected error for invalid preset")
	} else if !strings.Contains(err.Error(), "invalid event preset") {
		t.Errorf("invalid preset error = %v, want it to mention the invalid preset", err)
	}

	c, err := NewChronicleForWorld(rng, state, "normal")
	if err != nil {
		t.Fatalf("NewChronicleForWorld valid deps error = %v", err)
	}
	if c == nil {
		t.Fatal("NewChronicleForWorld valid deps returned nil chronicle")
	}
}

func TestNewChronicleForWorld_ResolverWiring(t *testing.T) {
	run := func() string {
		c, err := NewChronicleForWorld(randv2.New(randv2.NewPCG(7, 0)), newFactoryTestState(), "normal")
		if err != nil {
			t.Fatalf("NewChronicleForWorld error = %v", err)
		}
		var buf bytes.Buffer
		if err := c.Stream(context.Background(), factoryTestEvents(), "normal", &buf); err != nil {
			t.Fatalf("Stream error = %v", err)
		}
		return buf.String()
	}

	out := run()
	if !strings.Contains(out, "Aldric") {
		t.Errorf("output must contain the resolved figure name, got %q", out)
	}
	if !strings.Contains(out, "Northhold") {
		t.Errorf("output must contain the target settlement, got %q", out)
	}
}

func TestNewChronicleForWorld_EquivalentToManualWiring(t *testing.T) {
	state := newFactoryTestState()
	events := factoryTestEvents()

	viaFactory, err := NewChronicleForWorld(randv2.New(randv2.NewPCG(42, 0)), state, "normal")
	if err != nil {
		t.Fatalf("NewChronicleForWorld error = %v", err)
	}
	var factoryBuf bytes.Buffer
	if err := viaFactory.Stream(context.Background(), events, "normal", &factoryBuf); err != nil {
		t.Fatalf("factory Stream error = %v", err)
	}

	manual, err := ucsim.NewChronicle(
		randv2.New(randv2.NewPCG(42, 0)),
		infranarrative.DefaultGrammarProvider{},
		ucsim.NewWorldFigureResolver(state),
	)
	if err != nil {
		t.Fatalf("manual NewChronicle error = %v", err)
	}
	var manualBuf bytes.Buffer
	if err := manual.Stream(context.Background(), events, "normal", &manualBuf); err != nil {
		t.Fatalf("manual Stream error = %v", err)
	}

	if factoryBuf.String() != manualBuf.String() {
		t.Errorf("factory wiring differs from manual wiring:\n--- factory ---\n%s\n--- manual ---\n%s", factoryBuf.String(), manualBuf.String())
	}
}

func TestNewChronicleForWorld_Determinism(t *testing.T) {
	events := append(factoryTestEvents(), domsim.Event{
		Year: 2, Category: "Death", FigureID: "Deepcrest-0", SettlementName: "Deepcrest", Description: "Aldric dies.",
	})

	run := func(preset string) string {
		c, err := NewChronicleForWorld(randv2.New(randv2.NewPCG(42, 0)), newFactoryTestState(), preset)
		if err != nil {
			t.Fatalf("NewChronicleForWorld(%s) error = %v", preset, err)
		}
		var buf bytes.Buffer
		if err := c.Stream(context.Background(), events, preset, &buf); err != nil {
			t.Fatalf("Stream(%s) error = %v", preset, err)
		}
		return buf.String()
	}

	for _, preset := range []string{"quiet", "normal", "verbose"} {
		first := run(preset)
		second := run(preset)
		if first != second {
			t.Errorf("preset %s: same seed produced different output:\n--- first ---\n%s\n--- second ---\n%s", preset, first, second)
		}
	}
}
