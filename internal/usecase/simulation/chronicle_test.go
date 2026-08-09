package simulation

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
	infranarrative "github.com/thalesraymond/world-generation-go/internal/infra/narrative"
)

// staticGrammar is a GrammarProvider over a fixed grammar source.
type staticGrammar struct{ src string }

func (g staticGrammar) Grammar() string { return g.src }

// testFigureResolver maps figure IDs to refs inline.
type testFigureResolver map[string]FigureRef

func (m testFigureResolver) Resolve(id string) (FigureRef, bool) {
	ref, ok := m[id]
	return ref, ok
}

func newTestChronicle(t *testing.T, grammar GrammarProvider, figures FigureResolver, seed uint64) *Chronicle {
	t.Helper()
	c, err := NewChronicle(randv2.New(randv2.NewPCG(seed, 0)), grammar, figures)
	if err != nil {
		t.Fatalf("NewChronicle error = %v", err)
	}
	return c
}

func TestParsePreset(t *testing.T) {
	for _, want := range []Preset{PresetQuiet, PresetNormal, PresetVerbose} {
		got, err := ParsePreset(string(want))
		if err != nil {
			t.Errorf("ParsePreset(%q) error = %v", want, err)
		}
		if got != want {
			t.Errorf("ParsePreset(%q) = %q, want %q", want, got, want)
		}
	}

	for _, bad := range []string{"", "loud", "NORMAL", " normal"} {
		_, err := ParsePreset(bad)
		if err == nil {
			t.Errorf("ParsePreset(%q) = nil error, want actionable error", bad)
			continue
		}
		if !strings.Contains(err.Error(), bad) {
			t.Errorf("ParsePreset(%q) error = %q, want it to name the invalid preset", bad, err)
		}
		if !strings.Contains(err.Error(), "quiet, normal, or verbose") {
			t.Errorf("ParsePreset(%q) error = %q, want the accepted values listed", bad, err)
		}
	}
}

func TestNewChronicleValidation(t *testing.T) {
	rng := randv2.New(randv2.NewPCG(1, 0))
	grammar := staticGrammar{src: infranarrative.DefaultGrammar}
	figures := testFigureResolver{}

	if _, err := NewChronicle(nil, grammar, figures); err == nil {
		t.Error("expected error for nil RNG")
	}
	if _, err := NewChronicle(rng, nil, figures); err == nil {
		t.Error("expected error for nil grammar provider")
	}
	if _, err := NewChronicle(rng, grammar, nil); err == nil {
		t.Error("expected error for nil figure resolver")
	}

	bad := staticGrammar{src: "this is not a valid grammar {"}
	if _, err := NewChronicle(rng, bad, figures); err == nil {
		t.Error("expected error for unparseable grammar")
	}

	if _, err := NewChronicle(rng, grammar, figures); err != nil {
		t.Errorf("NewChronicle valid deps error = %v", err)
	}
}

func TestChronicleStream_InvalidPreset(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 1)
	var buf bytes.Buffer
	err := c.Stream(context.Background(), []domsim.Event{}, "nope", &buf)
	if err == nil {
		t.Fatal("expected error for invalid preset")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "quiet, normal, or verbose") {
		t.Errorf("error = %q, want actionable invalid-preset message", err)
	}
	if buf.Len() != 0 {
		t.Errorf("invalid preset wrote %q", buf.String())
	}
}

func TestChronicleStream_EmptyEvents(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 1)
	var buf bytes.Buffer
	if err := c.Stream(context.Background(), nil, "normal", &buf); err != nil {
		t.Fatalf("Stream empty events error = %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("empty events produced %q", buf.String())
	}
}

func TestChronicleStream_Normal_AggregatesEconomyExpansion(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
	events := []domsim.Event{
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
		{Year: 1, Category: "Economy", SettlementName: "Ashfield", Description: "Ashfield prospers."},
		{Year: 1, Category: "Expansion", SettlementName: "Westwood", Description: "Westwood founded Newhold."},
		{Year: 2, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
	}

	var buf bytes.Buffer
	if err := c.Stream(context.Background(), events, "normal", &buf); err != nil {
		t.Fatalf("Stream error = %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	want := []string{
		"In 1, 2 settlements tended their wealth: Ashfield, Deepcrest.",
		"In 1, 1 settlement expanded their borders: Westwood.",
		"In 2, 1 settlement tended their wealth: Deepcrest.",
	}
	if len(lines) != len(want) {
		t.Fatalf("normal output lines = %d, want %d: %v", len(lines), len(want), lines)
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line %d = %q, want %q", i, lines[i], w)
		}
	}
}

func TestChronicleStream_Normal_KeepsEventfulCategoriesPerEvent(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
	events := []domsim.Event{
		{Year: 3, Category: "Raid", SettlementName: "Deepcrest", TargetSettlement: "Northhold", Description: "Deepcrest raided Northhold and seized 50 wealth"},
		{Year: 3, Category: "Conquest", SettlementName: "Deepcrest", TargetSettlement: "Northhold", Description: "Deepcrest conquered Northhold"},
		{Year: 3, Category: "Diplomacy", SettlementName: "Deepcrest", TargetSettlement: "Ashfield", Description: "Deepcrest signed an alliance with Ashfield"},
	}

	var buf bytes.Buffer
	if err := c.Stream(context.Background(), events, "normal", &buf); err != nil {
		t.Fatalf("Stream error = %v", err)
	}

	// Every eventful category is narrated individually (one line each), with
	// no $Variable leaks. Agent framing alternatives may or may not echo the
	// raw description — that is the grammar's choice, not a requirement.
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 per-event lines, got %d: %v", len(lines), lines)
	}
	for i, l := range lines {
		if strings.Contains(l, "$") {
			t.Errorf("line %d = %q contains a raw variable leak", i, l)
		}
	}
}

func TestChronicleStream_Quiet_OnlyHighSignal(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{
		"Deepcrest-0": {Name: "Aldric", Role: "Leader"},
	}, 7)
	events := []domsim.Event{
		{Year: 1, Category: "Birth", FigureID: "Deepcrest-0", SettlementName: "Deepcrest", Description: "Aldric is born in Deepcrest."},
		{Year: 1, Category: "Death", FigureID: "Deepcrest-0", SettlementName: "Deepcrest", Description: "Aldric dies."},
		{Year: 1, Category: "Raid", SettlementName: "Deepcrest", TargetSettlement: "Northhold", Description: "Deepcrest raided Northhold"},
		{Year: 1, Category: "Conflict", SettlementName: "Deepcrest", TargetSettlement: "Northhold", Description: "Deepcrest clashed with Northhold"},
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
		{Year: 1, Category: "Economy", SettlementName: "Ashfield", Description: "Ashfield prospers."},
	}

	var buf bytes.Buffer
	if err := c.Stream(context.Background(), events, "quiet", &buf); err != nil {
		t.Fatalf("Stream error = %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "born in") {
		t.Errorf("quiet must suppress Birth narration, got %q", out)
	}
	if strings.Contains(out, "raided Northhold") {
		t.Errorf("quiet must suppress Raid narration, got %q", out)
	}
	if !strings.Contains(out, "In 1, 2 settlements tended their wealth: Ashfield, Deepcrest.") {
		t.Errorf("quiet must keep aggregated Economy, got %q", out)
	}
	if !strings.Contains(out, "Aldric") {
		t.Errorf("quiet must keep Death narration (figure line), got %q", out)
	}
	if !strings.Contains(out, "Deepcrest clashed with Northhold") {
		t.Errorf("quiet must keep Conflict narration, got %q", out)
	}
}

func TestChronicleStream_Verbose_RawLinesAndNoAggregation(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
	events := []domsim.Event{
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
		{Year: 1, Category: "Economy", SettlementName: "Ashfield", Description: "Ashfield prospers."},
	}

	var buf bytes.Buffer
	if err := c.Stream(context.Background(), events, "verbose", &buf); err != nil {
		t.Fatalf("Stream error = %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("verbose: want 4 lines (raw+narrated per event), got %d: %v", len(lines), lines)
	}

	// Raw lines first, matching FormatEvent shape.
	if !strings.HasPrefix(lines[0], "[1] (Economy)") {
		t.Errorf("line 0 = %q, want a raw FormatEvent line", lines[0])
	}
	if !strings.HasPrefix(lines[2], "[1] (Economy)") {
		t.Errorf("line 2 = %q, want a raw FormatEvent line", lines[2])
	}

	// No aggregation: both events are narrated individually.
	if strings.Contains(buf.String(), "tended their wealth: Ashfield") {
		t.Errorf("verbose must not aggregate Economy, got %q", buf.String())
	}
}

func TestChronicleStream_FigureFallbackToBaseRule(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{
		"Deepcrest-0": {Name: "Aldric", Role: "General"},
	}, 7)
	events := []domsim.Event{
		// Figure event with a target: .figure rule fires and names the figure.
		{
			Year: 1, Category: "Conflict", FigureID: "Deepcrest-0",
			SettlementName: "Deepcrest", TargetSettlement: "Northhold",
			Description: "Deepcrest raided Northhold.",
		},
		// Figure event without a target: .figure alternatives are all
		// ineligible, so the base Conflict rule fires.
		{
			Year: 2, Category: "Conflict", FigureID: "Deepcrest-0",
			SettlementName: "Deepcrest", TargetSettlement: "",
			Description: "Deepcrest was attacked.",
		},
	}

	var buf bytes.Buffer
	if err := c.Stream(context.Background(), events, "normal", &buf); err != nil {
		t.Fatalf("Stream error = %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "Aldric") {
		t.Errorf("figure line = %q, want figure name (Conflict.figure)", lines[0])
	}
	if !strings.Contains(lines[0], "Northhold") {
		t.Errorf("figure line = %q, want target settlement", lines[0])
	}
	if strings.Contains(lines[1], "Aldric") {
		t.Errorf("base-fallback line = %q must not name the figure", lines[1])
	}
}

func TestChronicleStream_FallbackToDescription(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
	// Unknown category: no rule -> event.Description.
	events := []domsim.Event{
		{Year: 1, Category: "CustomCategory", Description: "Something unusual happened."},
	}

	var buf bytes.Buffer
	if err := c.Stream(context.Background(), events, "normal", &buf); err != nil {
		t.Fatalf("Stream error = %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "Something unusual happened." {
		t.Errorf("fallback = %q, want raw description", got)
	}
}

func TestChronicleStream_SettlementNameAbsentFallsBack(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{
		"Deepcrest-0": {Name: "Aldric", Role: ""},
	}, 7)
	// A Birth event whose settlement name is absent: every Birth alternative
	// requires $SettlementName, so the base rule is ineligible and the
	// chronicle falls back to the clean description.
	events := []domsim.Event{
		{Year: 1, Category: "Birth", FigureID: "Deepcrest-0", SettlementName: "", Description: "A child was born on the road."},
	}

	var buf bytes.Buffer
	if err := c.Stream(context.Background(), events, "normal", &buf); err != nil {
		t.Fatalf("Stream error = %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "A child was born on the road." {
		t.Errorf("settlement-less Birth = %q, want clean description fallback", got)
	}
}

func TestChronicleStream_ContextCompleteForAllCategories(t *testing.T) {
	// Probe grammar that echoes every agent context field. If the chronicle
	// omits any field, the variable renders as a literal leak instead. The
	// rule name must match the event category.
	grammar := staticGrammar{src: `CustomProbe ::= $ActionType "|" $TargetSettlement "|" $Outcome "|" $Amount`}
	c := newTestChronicle(t, grammar, testFigureResolver{}, 7)
	events := []domsim.Event{
		{
			Year: 5, Category: "CustomProbe",
			TargetSettlement: "Northhold",
			Description:      "Deepcrest raided Northhold and seized 50 wealth",
		},
	}

	var buf bytes.Buffer
	if err := c.Stream(context.Background(), events, "normal", &buf); err != nil {
		t.Fatalf("Stream error = %v", err)
	}
	want := "CustomProbe|Northhold|Deepcrest raided Northhold and seized 50 wealth|50"
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Errorf("context probe = %q, want %q", got, want)
	}
}

func TestChronicleStream_Determinism(t *testing.T) {
	figures := testFigureResolver{"Deepcrest-0": {Name: "Aldric", Role: "General"}}
	events := []domsim.Event{
		{Year: 1, Category: "Conflict", FigureID: "Deepcrest-0", SettlementName: "Deepcrest", TargetSettlement: "Northhold", Description: "Deepcrest raided Northhold."},
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
		{Year: 1, Category: "Economy", SettlementName: "Ashfield", Description: "Ashfield prospers."},
		{Year: 2, Category: "Politics", SettlementName: "Deepcrest", TargetSettlement: "Ashfield", Description: "Deepcrest negotiated with Ashfield."},
		{Year: 2, Category: "Death", FigureID: "Deepcrest-0", SettlementName: "Deepcrest", Description: "Aldric dies."},
	}

	run := func(preset string) string {
		c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, figures, 42)
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

func TestChronicleStream_CancelledContext(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var buf bytes.Buffer
	err := c.Stream(ctx, []domsim.Event{
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
	}, "normal", &buf)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestChronicleStream_VerboseCancelledContext(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var buf bytes.Buffer
	err := c.Stream(ctx, []domsim.Event{
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
	}, "verbose", &buf)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

// failingWriter fails on every write.
type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) { return 0, errors.New("write failed") }

// nthFailingWriter succeeds for the first n writes, then fails.
type nthFailingWriter struct {
	remaining int
}

func (w *nthFailingWriter) Write(p []byte) (int, error) {
	if w.remaining == 0 {
		return 0, errors.New("write failed")
	}
	w.remaining--
	return len(p), nil
}

func TestChronicleStream_WriteErrors(t *testing.T) {
	events := []domsim.Event{
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
		{Year: 1, Category: "Raid", SettlementName: "Deepcrest", TargetSettlement: "Northhold", Description: "Deepcrest raided Northhold"},
	}

	for _, preset := range []string{"quiet", "normal", "verbose"} {
		c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
		err := c.Stream(context.Background(), events, preset, failingWriter{})
		if err == nil {
			t.Errorf("preset %s: expected write error, got nil", preset)
			continue
		}
		if !strings.Contains(err.Error(), "write failed") {
			t.Errorf("preset %s: error = %q, want the underlying write error", preset, err)
		}
	}
}

// TestChronicleStream_VerboseNarrativeWriteError covers the write path for the
// narrated (second) line in verbose mode, where the raw line write succeeds
// and the narrative write fails.
func TestChronicleStream_VerboseNarrativeWriteError(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
	w := &nthFailingWriter{remaining: 1}
	err := c.Stream(context.Background(), []domsim.Event{
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
	}, "verbose", w)
	if err == nil {
		t.Fatal("expected write error on the narrative line")
	}
	if !strings.Contains(err.Error(), "write failed") {
		t.Errorf("error = %q, want the underlying write error", err)
	}
}

// TestChronicleStream_AggregatedNarrativeWriteError covers the per-event
// narrative write path in normal mode (the aggregate line write succeeds and
// the per-event write fails).
func TestChronicleStream_AggregatedNarrativeWriteError(t *testing.T) {
	c := newTestChronicle(t, staticGrammar{src: infranarrative.DefaultGrammar}, testFigureResolver{}, 7)
	w := &nthFailingWriter{remaining: 1}
	err := c.Stream(context.Background(), []domsim.Event{
		{Year: 1, Category: "Economy", SettlementName: "Deepcrest", Description: "Deepcrest prospers."},
		{Year: 1, Category: "Raid", SettlementName: "Deepcrest", TargetSettlement: "Northhold", Description: "Deepcrest raided Northhold"},
	}, "normal", w)
	if err == nil {
		t.Fatal("expected write error on the per-event narrative line")
	}
	if !strings.Contains(err.Error(), "write failed") {
		t.Errorf("error = %q, want the underlying write error", err)
	}
}

func TestMapFigureResolver(t *testing.T) {
	m := MapFigureResolver{"a-1": {Name: "Aldric", Role: "General"}}
	ref, ok := m.Resolve("a-1")
	if !ok || ref.Name != "Aldric" || ref.Role != "General" {
		t.Errorf("Resolve(a-1) = %+v, %v; want Aldric/General, true", ref, ok)
	}
	if _, ok := m.Resolve("missing"); ok {
		t.Error("Resolve(missing) = true, want false")
	}
}

func TestNewWorldFigureResolver(t *testing.T) {
	state := &world.State{
		Settlements: []world.Settlement{
			{
				Name: "Deepcrest",
				Figures: []figures.HistoricalFigure{
					{ID: "Deepcrest-0", Name: "Aldric", Role: "General"},
					{ID: "Deepcrest-1", Name: "Mira", Role: ""},
				},
			},
			{
				Name: "Ashfield",
				Figures: []figures.HistoricalFigure{
					{ID: "Ashfield-0", Name: "Brogan", Role: "Diplomat"},
				},
			},
		},
	}

	resolver := NewWorldFigureResolver(state)
	ref, ok := resolver.Resolve("Deepcrest-0")
	if !ok || ref.Name != "Aldric" || ref.Role != "General" {
		t.Errorf("Resolve(Deepcrest-0) = %+v, %v; want Aldric/General, true", ref, ok)
	}
	ref, ok = resolver.Resolve("Ashfield-0")
	if !ok || ref.Name != "Brogan" || ref.Role != "Diplomat" {
		t.Errorf("Resolve(Ashfield-0) = %+v, %v; want Brogan/Diplomat, true", ref, ok)
	}
	if _, ok := resolver.Resolve("missing"); ok {
		t.Error("Resolve(missing) = true, want false")
	}
}

func TestExtractAmount(t *testing.T) {
	cases := []struct {
		description string
		want        string
	}{
		{"Deepcrest raided Northhold and seized 50 wealth", "50"},
		{"Deepcrest raided Northhold and seized 50 wealth.", "50"},
		{"Deepcrest prospers", ""},
		{"no numbers here", ""},
		{"year 42 saw 100 wealth change hands", "100"},
		{"the . council weighed options before a 60 wealth levy", "60"},
		{"coins , ; scattered but none counted wealth here", ""},
	}
	for _, tc := range cases {
		if got := extractAmount(tc.description); got != tc.want {
			t.Errorf("extractAmount(%q) = %q, want %q", tc.description, got, tc.want)
		}
	}
}

func TestAggregateSummary(t *testing.T) {
	cases := []struct {
		category string
		year     int
		names    []string
		want     string
	}{
		{"Economy", 1, []string{"Ashfield", "Deepcrest"}, "In 1, 2 settlements tended their wealth: Ashfield, Deepcrest."},
		{"Economy", 1, []string{"Deepcrest"}, "In 1, 1 settlement tended their wealth: Deepcrest."},
		{"Expansion", 7, []string{"Westwood"}, "In 7, 1 settlement expanded their borders: Westwood."},
		{"Economy", 2, nil, "In 2, 0 settlements tended their wealth."},
		{"Unknown", 3, []string{"X"}, ""},
	}
	for _, tc := range cases {
		if got := aggregateSummary(tc.category, tc.year, tc.names); got != tc.want {
			t.Errorf("aggregateSummary(%s, %d, %v) = %q, want %q", tc.category, tc.year, tc.names, got, tc.want)
		}
	}
}
