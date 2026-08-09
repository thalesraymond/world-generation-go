package simulation

import (
	"context"
	"errors"
	"fmt"
	"io"
	randv2 "math/rand/v2"
	"sort"
	"strings"

	"github.com/thalesraymond/world-generation-go/internal/domain/narrative"
	domsim "github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// GrammarProvider returns the CFG grammar source used to build the narrative
// engine. It is declared in the usecase layer and implemented by infra
// (internal/infra/narrative) so dependency direction stays inward.
type GrammarProvider interface {
	Grammar() string
}

// FigureRef is the resolved narrative identity of a figure.
type FigureRef struct {
	Name string
	Role string
}

// FigureResolver resolves a figure ID to its narrative identity. The default
// implementation is a map built from world state.
type FigureResolver interface {
	Resolve(id string) (FigureRef, bool)
}

// MapFigureResolver is a map-based FigureResolver keyed by figure ID.
type MapFigureResolver map[string]FigureRef

var _ FigureResolver = MapFigureResolver{}

// Resolve looks up the figure by ID.
func (m MapFigureResolver) Resolve(id string) (FigureRef, bool) {
	ref, ok := m[id]
	return ref, ok
}

// NewWorldFigureResolver builds a MapFigureResolver from world state
// settlements. IDs are unique after the figure-identity fix, so lookup is exact.
func NewWorldFigureResolver(state *world.State) MapFigureResolver {
	resolver := make(MapFigureResolver)
	for i := range state.Settlements {
		for j := range state.Settlements[i].Figures {
			f := state.Settlements[i].Figures[j]
			resolver[f.ID] = FigureRef{Name: f.Name, Role: f.Role}
		}
	}
	return resolver
}

// Preset controls the volume and shape of the chronicle stream.
type Preset string

const (
	// PresetQuiet emits a high-signal subset only: Death, RoleTransition,
	// Conflict, Conquest, plus aggregated Economy|Expansion.
	PresetQuiet Preset = "quiet"
	// PresetNormal narrates every event after cross-settlement per-year
	// aggregation of Economy|Expansion.
	PresetNormal Preset = "normal"
	// PresetVerbose narrates every event individually plus one raw
	// FormatEvent line per event, with no aggregation.
	PresetVerbose Preset = "verbose"
)

// ParsePreset validates a preset string and returns its typed value.
func ParsePreset(s string) (Preset, error) {
	switch Preset(s) {
	case PresetQuiet, PresetNormal, PresetVerbose:
		return Preset(s), nil
	}
	return "", fmt.Errorf("invalid event preset %q: want quiet, normal, or verbose", s)
}

// Chronicle renders a stream of simulation events into narrated prose.
// It is constructed with the narrative RNG, a GrammarProvider, and a
// FigureResolver, and renders post-hoc over a collected events slice.
type Chronicle struct {
	engine  *narrative.Engine
	rng     *randv2.Rand
	figures FigureResolver
}

// NewChronicle constructs a Chronicle from the narrative RNG, a grammar
// provider, and a figure resolver.
func NewChronicle(rng *randv2.Rand, grammar GrammarProvider, figures FigureResolver) (*Chronicle, error) {
	if rng == nil {
		return nil, fmt.Errorf("narrative RNG is nil")
	}
	if grammar == nil {
		return nil, fmt.Errorf("grammar provider is nil")
	}
	if figures == nil {
		return nil, fmt.Errorf("figure resolver is nil")
	}
	engine, err := narrative.NewEngineFromString(grammar.Grammar())
	if err != nil {
		return nil, fmt.Errorf("create narrative engine: %w", err)
	}
	return &Chronicle{engine: engine, rng: rng, figures: figures}, nil
}

// Stream validates the preset and renders the events to out.
//
// Verbose emits one raw FormatEvent line per event and narrates every event
// individually. Normal and quiet apply cross-settlement per-year aggregation
// to Economy|Expansion and suppress their per-event narration.
func (c *Chronicle) Stream(ctx context.Context, events []domsim.Event, preset string, out io.Writer) error {
	p, err := ParsePreset(preset)
	if err != nil {
		return err
	}

	if p == PresetVerbose {
		return c.streamVerbose(ctx, events, out)
	}
	return c.streamAggregated(ctx, events, p, out)
}

// streamVerbose emits one raw line and one narrated line per event.
func (c *Chronicle) streamVerbose(ctx context.Context, events []domsim.Event, out io.Writer) error {
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, domsim.FormatEvent(event)); err != nil {
			return fmt.Errorf("write verbose line: %w", err)
		}
		text, err := c.renderEvent(event)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, text); err != nil {
			return fmt.Errorf("write narrative line: %w", err)
		}
	}
	return nil
}

// streamAggregated renders normal/quiet output: aggregated Economy|Expansion
// summaries (one per year per category, at the position of the first event in
// the group) and per-event narration for everything else. In quiet, only the
// high-signal categories are narrated.
func (c *Chronicle) streamAggregated(ctx context.Context, events []domsim.Event, p Preset, out io.Writer) error {
	summaries := buildAggregates(events)
	emitted := make(map[yearCategoryKey]bool)

	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return err
		}

		if isAggregatedCategory(event.Category) {
			key := yearCategoryKey{Year: event.Year, Category: event.Category}
			if emitted[key] {
				continue
			}
			emitted[key] = true
			line, ok := summaries[key]
			if !ok {
				continue
			}
			if _, err := fmt.Fprintln(out, line); err != nil {
				return fmt.Errorf("write aggregate line: %w", err)
			}
			continue
		}

		if p == PresetQuiet && !isQuietCategory(event.Category) {
			continue
		}

		text, err := c.renderEvent(event)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, text); err != nil {
			return fmt.Errorf("write narrative line: %w", err)
		}
	}
	return nil
}

// buildAggregates groups Economy|Expansion events by (year, category) and
// renders one deterministic, RNG-free summary line per group.
func buildAggregates(events []domsim.Event) map[yearCategoryKey]string {
	groupSet := make(map[yearCategoryKey]map[string]struct{})
	for _, e := range events {
		if !isAggregatedCategory(e.Category) {
			continue
		}
		key := yearCategoryKey{Year: e.Year, Category: e.Category}
		if groupSet[key] == nil {
			groupSet[key] = make(map[string]struct{})
		}
		if e.SettlementName != "" {
			groupSet[key][e.SettlementName] = struct{}{}
		}
	}

	summaries := make(map[yearCategoryKey]string, len(groupSet))
	for key, settlements := range groupSet {
		names := make([]string, 0, len(settlements))
		for name := range settlements {
			names = append(names, name)
		}
		sort.Strings(names)
		summaries[key] = aggregateSummary(key.Category, key.Year, names)
	}
	return summaries
}

type yearCategoryKey struct {
	Year     int
	Category string
}

// aggregateSummary renders a single summary line for a (year, category)
// group. It is composed without RNG draws and is stable under iteration.
func aggregateSummary(category string, year int, names []string) string {
	count := len(names)
	noun := "settlements"
	if count == 1 {
		noun = "settlement"
	}
	list := ""
	if count > 0 {
		list = ": " + strings.Join(names, ", ")
	}
	switch category {
	case "Economy":
		return fmt.Sprintf("In %d, %d %s tended their wealth%s.", year, count, noun, list)
	case "Expansion":
		return fmt.Sprintf("In %d, %d %s expanded their borders%s.", year, count, noun, list)
	}
	return ""
}

// renderEvent builds the narrative context for an event and dispatches it
// through the fallback chain: .figure -> base rule -> event.Description.
func (c *Chronicle) renderEvent(event domsim.Event) (string, error) {
	ctx := c.contextFor(event)

	if isFigureCategory(event.Category) && ctx["FigureName"] != "" {
		text, err := c.engine.Resolve(event.Category+".figure", ctx, c.rng)
		if err == nil {
			return text, nil
		}
		if !errors.Is(err, narrative.ErrRuleNotFound) && !errors.Is(err, narrative.ErrNoEligibleAlternative) {
			return "", fmt.Errorf("narrate %s.figure: %w", event.Category, err)
		}
	}

	text, err := c.engine.Resolve(event.Category, ctx, c.rng)
	if err != nil {
		if errors.Is(err, narrative.ErrRuleNotFound) || errors.Is(err, narrative.ErrNoEligibleAlternative) {
			return event.Description, nil
		}
		return "", fmt.Errorf("narrate %s: %w", event.Category, err)
	}
	return text, nil
}

// contextFor builds the complete narrative context for an event. Agent fields
// (TargetSettlement, Outcome, ActionType, Amount) are copied from the event for
// all categories. Figure fields are added when the event carries a resolvable
// figure ID. SettlementName is always set from the event.
func (c *Chronicle) contextFor(event domsim.Event) map[string]string {
	ctx := map[string]string{
		"year":             fmt.Sprint(event.Year),
		"category":         event.Category,
		"description":      event.Description,
		"SettlementName":   event.SettlementName,
		"TargetSettlement": event.TargetSettlement,
		"ActionType":       event.Category,
		"Outcome":          event.Description,
		"Amount":           extractAmount(event.Description),
	}
	if event.FigureID != "" {
		if ref, ok := c.figures.Resolve(event.FigureID); ok {
			ctx["FigureName"] = ref.Name
			ctx["FigureRole"] = ref.Role
		}
	}
	return ctx
}

// isAggregatedCategory reports whether the category participates in
// cross-settlement per-year aggregation.
func isAggregatedCategory(category string) bool {
	return category == "Economy" || category == "Expansion"
}

// isQuietCategory reports whether the category is narrated per-event in the
// quiet preset.
func isQuietCategory(category string) bool {
	switch category {
	case "Death", "RoleTransition", "Conflict", "Conquest":
		return true
	}
	return false
}

// isFigureCategory reports whether the category has a .figure rule variant.
func isFigureCategory(category string) bool {
	switch category {
	case "Conflict", "Politics", "Discovery":
		return true
	}
	return false
}

// extractAmount pulls the first integer-like token out of an event
// description (e.g. "50" from "... seized 50 wealth"). It returns an empty
// string when no amount is present.
func extractAmount(description string) string {
	fields := strings.Fields(description)
	for i, field := range fields {
		num := strings.Trim(field, ".,;:")
		if num == "" {
			continue
		}
		isNumber := true
		for _, r := range num {
			if r < '0' || r > '9' {
				isNumber = false
				break
			}
		}
		if isNumber && i+1 < len(fields) && strings.Trim(fields[i+1], ".,;:") == "wealth" {
			return num
		}
	}
	return ""
}
