package artifact

import (
	"math"
	"sort"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// significanceThreshold is the cumulative score at which an artifact becomes
// significant (spec 4.2). A single War or Conquest event is itself pivotal.
const significanceThreshold = 3

// eventWeights is the fixed per-category significance weight table (spec
// 4.2). Categories absent from the table — including synthetic lifecycle
// events and unknown categories — confer no weight.
var eventWeights = map[string]int{
	"War":       3,
	"Conquest":  3,
	"Diplomacy": 2,
	"Politics":  2,
	"Raid":      2,
	"Expansion": 1,
	"Disaster":  1,
	"Economy":   0,
}

// settlementLumpSums is the one-time significance award when a settlement
// acquires an artifact, keyed by the settlement's size class at acquisition
// (spec 4.4). The class is read once, at the acquisition year; later growth
// never re-awards it.
var settlementLumpSums = map[string]int{
	"MajorCity": 3,
	"City":      2,
	"Village":   1,
	"Abandoned": 0,
}

// SignificanceContext supplies world-state data for the owner-importance
// fallback (spec 4.4). The caller builds it from the world state so the pass
// stays a pure domain pass; the zero value simply disables the fallback.
type SignificanceContext struct {
	// FigureReputation maps figure IDs to their reputation deltas summed by
	// year. A year's net delta is accrued only when positive.
	FigureReputation map[string]map[int]int
	// SettlementClass maps settlement names to their size class (MajorCity,
	// City, Village, Abandoned).
	SettlementClass map[string]string
}

// significanceEvent is the weight-bearing view of an event associated with
// an artifact, captured during the stream walk.
type significanceEvent struct {
	year    int
	weight  int
	eventID string
}

// contribution is one score increment applied in chronological order.
// order breaks year ties deterministically: acquisition lump sums first,
// then events in stream order, then annual accrual.
type contribution struct {
	year    int
	order   int
	value   int
	eventID string
}

// evaluateSignificance computes the significance state of every artifact
// from its completed provenance chain and the weight-bearing events the
// stream walk associated with it. Contributions are merged chronologically
// and the monotonic latch flips at the first crossing year.
func evaluateSignificance(artifacts []Artifact, events []simulation.Event, contribs [][]significanceEvent, sig SignificanceContext) {
	horizon := 0
	for i := range events {
		if events[i].Year > horizon {
			horizon = events[i].Year
		}
	}
	for i := range artifacts {
		applySignificance(&artifacts[i], contribs[i], horizon, sig)
	}
}

// applySignificance folds the artifact's contributions into its score and
// updates the significance latch. A crossing caused by an event records that
// event as pivotal; a crossing caused by owner-importance accrual has no
// pivotal event (PivotalEventID stays empty).
func applySignificance(a *Artifact, events []significanceEvent, horizon int, sig SignificanceContext) {
	contribs := buildContributions(a, events, horizon, sig)
	sort.SliceStable(contribs, func(i, j int) bool {
		if contribs[i].year != contribs[j].year {
			return contribs[i].year < contribs[j].year
		}
		return contribs[i].order < contribs[j].order
	})

	score := a.SignificanceScore
	isSignificant := a.IsSignificant
	pivotalEventID := a.PivotalEventID
	significanceYear := a.SignificanceYear

	// All contributions are non-negative, so the cumulative score never
	// drops below its carried value (spec 3.1 clamps it at 0 by
	// construction).
	for _, c := range contribs {
		score += c.value
		if isSignificant {
			continue
		}
		if score >= significanceThreshold {
			isSignificant = true
			significanceYear = c.year
			if c.eventID != "" {
				pivotalEventID = c.eventID
			}
		}
	}

	a.SignificanceScore = score
	a.IsSignificant = isSignificant
	a.PivotalEventID = pivotalEventID
	a.SignificanceYear = significanceYear
}

// buildContributions turns the provenance chain and associated events into
// score contributions. Figure ownership accrues the owner's positive yearly
// reputation delta for each held year (acquisition year inclusive, next
// transition exclusive); settlement ownership awards the size-class lump sum
// once at the acquisition year. While the artifact is lost nothing accrues
// and event weights are skipped (spec 4.6).
func buildContributions(a *Artifact, events []significanceEvent, horizon int, sig SignificanceContext) []contribution {
	var contribs []contribution

	for i := range a.Provenance {
		entry := a.Provenance[i]
		toYear := horizon
		if i+1 < len(a.Provenance) {
			toYear = a.Provenance[i+1].Year
		}
		switch entry.Owner.Kind {
		case "figure":
			byYear := sig.FigureReputation[entry.Owner.ID]
			for y := entry.Year; y < toYear; y++ {
				if delta := byYear[y]; delta > 0 {
					contribs = append(contribs, contribution{year: y, order: 2, value: delta})
				}
			}
		case "settlement":
			if lump := settlementLumpSums[sig.SettlementClass[entry.Owner.ID]]; lump > 0 {
				contribs = append(contribs, contribution{year: entry.Year, order: 0, value: lump})
			}
		}
	}

	lost := lostRanges(a, horizon)
	for _, e := range events {
		if inLostRange(lost, e.year) {
			continue
		}
		contribs = append(contribs, contribution{year: e.year, order: 1, value: e.weight, eventID: e.eventID})
	}

	return contribs
}

// lostRanges returns the half-open year intervals during which the artifact
// is lost. Before the first provenance entry the artifact is lost only when
// its status says so (planted relics begin lost pre-timeline, and an empty
// provenance means no rediscovery happened, so the span runs to the end of
// the stream); afterwards the provenance chain governs: entries with
// Owner.Kind "lost" mark lost spans until the next entry. math.MinInt marks
// the unbounded start of the initial span.
func lostRanges(a *Artifact, horizon int) [][2]int {
	var ranges [][2]int
	first := horizon
	if len(a.Provenance) > 0 {
		first = a.Provenance[0].Year
	}
	if a.Status == "lost" {
		ranges = append(ranges, [2]int{math.MinInt, first})
	}
	for i := range a.Provenance {
		if a.Provenance[i].Owner.Kind != "lost" {
			continue
		}
		toYear := horizon
		if i+1 < len(a.Provenance) {
			toYear = a.Provenance[i+1].Year
		}
		ranges = append(ranges, [2]int{a.Provenance[i].Year, toYear})
	}
	return ranges
}

func inLostRange(ranges [][2]int, year int) bool {
	for _, r := range ranges {
		if year >= r[0] && year < r[1] {
			return true
		}
	}
	return false
}
