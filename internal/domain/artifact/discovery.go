package artifact

import (
	"fmt"

	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// Lifecycle constants (spec 6.5, issue #70). Fixed values keep the pass
// deterministic for a given seed.
const (
	// rediscoveryChancePercent is the pass probability of the seeded
	// rediscovery draw for historically-lost artifacts (spec 6.5). The spec
	// leaves the chance unspecified; the fixed 50% keeps the draw
	// deterministic from the artifacts lane.
	rediscoveryChancePercent = 50
	// genesisYear is the year world genesis happens: planted relics are
	// buried pre-timeline (spec 5.1), so fake-discovery events are minted at
	// this year and the provenance walk dates their first ownership to
	// creation.
	genesisYear = 0
)

// DiscoveryAgent answers "who discovers an artifact".
//
// TODO: temporary fake-discovery until expeditions exist (spec 5.2 and 9.5).
// Real expeditions will own their own discovery logic; this seeded figure
// draw is a placeholder standing behind the seam so the lifecycle pipeline
// (#70 loss/rediscovery, #71 destruction, #74 earned powers) never touches
// the expedition implementation.
type DiscoveryAgent interface {
	// Discover returns the figure who discovers the artifact, or "" when no
	// figure is available — the artifact then stays lost.
	Discover() string
}

// fakeDiscoveryAgent implements DiscoveryAgent by drawing the discovering
// figure uniformly from the world's figures on the artifacts RNG lane. The
// draw consumes the lane exactly once per call; with no figures available it
// returns "" without consuming anything.
type fakeDiscoveryAgent struct {
	figures []FigureContext
	rng     *randv2.Rand
}

func newFakeDiscoveryAgent(figures []FigureContext, rng *randv2.Rand) *fakeDiscoveryAgent {
	return &fakeDiscoveryAgent{figures: figures, rng: rng}
}

func (a *fakeDiscoveryAgent) Discover() string {
	if len(a.figures) == 0 {
		return ""
	}
	return a.figures[a.rng.IntN(len(a.figures))].ID
}

// fakeDiscovery runs lifecycle step 1, before the provenance walk: every
// planted relic (intrinsic significance source, still lost) performs one
// seeded draw through the DiscoveryAgent that assigns the discovering
// figure. A hit mints a synthetic Discovery event at the genesis year
// carrying the relic's ArtifactID and marks the relic held (the walk then
// records the first ownership via the existing Discovery rule,
// recordProvenance); the minted events are PREPENDED to the stream in
// artifact order so the provenance walk assigns their IDs (event-0-{n}). A
// miss — the world has no figures — leaves the relic lost with nothing
// minted.
//
// Lane consumption: fake-discovery draws come first, before destruction
// (#71), rediscovery, earned-power (#74), and emergence draws.
func fakeDiscovery(artifacts []Artifact, events []simulation.Event, agent DiscoveryAgent) []simulation.Event {
	var minted []simulation.Event
	for i := range artifacts {
		a := &artifacts[i]
		if a.SignificanceSource != "intrinsic" || a.Status != "lost" {
			continue
		}
		figure := agent.Discover()
		if figure == "" {
			continue
		}
		a.Status = "held"
		minted = append(minted, simulation.Event{
			Year:       genesisYear,
			Category:   "Discovery",
			FigureID:   figure,
			ArtifactID: a.ID,
		})
	}
	return append(minted, events...)
}

// rediscovery runs lifecycle step 4, after the walk and loss detection:
// every artifact still lost — planted relics that failed fake-discovery,
// historically-lost artifacts, and settlement-abandonment losses — performs
// one pass/fail gate draw on the artifacts lane. A passing draw then draws
// the discovering figure through the same DiscoveryAgent seam and mints a
// synthetic Discovery event at the horizon year, APPENDED to the stream
// after the walk. The appended events miss the walk's ID pass, so their IDs
// continue the event-{year}-{index} scheme manually (index = count of events
// already at that year); the artifact records the rediscovery provenance
// entry, is associated with the event, and its status returns to held. A
// failing draw — or a pass with no figure available — leaves the artifact
// lost and mints nothing.
//
// Lane consumption: rediscovery draws come after fake-discovery (#70) and
// destruction (#71) draws, and before earned-power (#74) and emergence draws.
func rediscovery(artifacts []Artifact, events []simulation.Event, agent DiscoveryAgent, rng *randv2.Rand, horizon int) []simulation.Event {
	for i := range artifacts {
		a := &artifacts[i]
		if a.Status != "lost" {
			continue
		}
		if rng.IntN(100) >= rediscoveryChancePercent {
			continue
		}
		figure := agent.Discover()
		if figure == "" {
			continue
		}
		event := simulation.Event{
			Year:       horizon,
			Category:   "Discovery",
			FigureID:   figure,
			ID:         syntheticEventID(events, horizon),
			ArtifactID: a.ID,
		}
		events = append(events, event)
		a.Provenance = append(a.Provenance, ProvenanceEntry{
			Year:      horizon,
			Owner:     Owner{Kind: "figure", ID: figure},
			EventID:   event.ID,
			EventType: "Discovery",
		})
		a.AssociatedEventIDs = append(a.AssociatedEventIDs, event.ID)
		a.Status = "held"
	}
	return events
}

// syntheticEventID continues the walk's event-{year}-{index} scheme (spec
// 10.2) for events minted after the walk: the index is the count of events
// already recorded at that year, so appended IDs never collide with the
// walk-assigned ones.
func syntheticEventID(events []simulation.Event, year int) string {
	n := 0
	for i := range events {
		if events[i].Year == year {
			n++
		}
	}
	return fmt.Sprintf("event-%d-%d", year, n)
}
