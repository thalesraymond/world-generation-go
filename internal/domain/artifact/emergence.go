package artifact

import (
	"fmt"
	randv2 "math/rand/v2"
	"strings"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// Emergence parameters (spec 5.3 and 5.6, issue #72). Fixed values keep the
// pass deterministic for a given seed.
const (
	// reputationThreshold is the cumulative reputation a figure must reach
	// for the owner-importance fallback to birth a backdated artifact.
	reputationThreshold = 10
	// commonGatePercent is the rarity-gate pass probability for common
	// types (weapon, armor, jewelry): 25%.
	commonGatePercent = 25
	// rareGatePercent is the rarity-gate pass probability for rare types
	// (crown, tome, relic): 10%. Rarity is a scarcity-of-type axis (spec
	// 5.6), so rare types pass the gate less often.
	rareGatePercent = 10
)

// emergenceTypePool mirrors the planted-relic type weighting (spec 5.6):
// common types outnumber rare types two to one. The type draw drives the
// rarity gate: the drawn type selects the gate probability.
var emergenceTypePool = []string{"weapon", "armor", "jewelry", "weapon", "armor", "crown", "relic", "tome"}

// emergenceNamePool supplies the first half of emergent artifact names; the
// drawn word is combined with the artifact's origin settlement.
var emergenceNamePool = map[string][]string{
	"weapon":  {"Blade", "Spear", "Warhammer"},
	"armor":   {"Aegis", "Cuirass", "Shield"},
	"crown":   {"Crown", "Diadem", "Coronet"},
	"relic":   {"Relic", "Idol", "Talisman"},
	"tome":    {"Tome", "Grimoire", "Codex"},
	"jewelry": {"Ring", "Amulet", "Brooch"},
}

// FigureContext is the minimal figure data the emergence fallback needs: the
// figure's home settlement (the ID origin) and its reputation history (used
// to detect the threshold crossing and its year). The pass accepts this
// summary rather than the full figure model so the artifact domain stays
// decoupled from the figures package.
type FigureContext struct {
	ID         string
	Settlement string
	Reputation []ReputationDelta
}

// ReputationDelta records one reputation change for a figure.
type ReputationDelta struct {
	Year  int
	Delta int
	Event string
}

// EmergencePass is the post-processing pipeline entry: it extends PostProcess
// (the provenance and event-ID walk, which also evaluates significance) with
// the lifecycle steps and emergent artifact births (spec 5.3, issues #70,
// #71, #72, #74). The artifacts RNG lane is consumed in a fixed order so the
// in-flight branches merge coherently:
//
//  1. fake-discovery draws (issue #70): planted relics are handed to a
//     figure through the temporary DiscoveryAgent seam; the minted synthetic
//     Discovery events are PREPENDED to the stream and join the walk with
//     walk-assigned IDs (fakeDiscovery).
//  2. destruction draws (issue #71, reserved, in-walk).
//  3. post-walk loss detection (issue #70): settlement owners whose FINAL
//     class is Abandoned are recorded lost at the horizon year, and any
//     artifact whose current owner is lost gets Status "lost" (applyLoss).
//  4. rediscovery draws (issue #70): every artifact still lost draws a
//     pass/fail gate on the lane; on a pass a synthetic Discovery event is
//     APPENDED to the stream with a manually continued event ID and the
//     artifact returns to held (rediscovery).
//  5. significance evaluation (issue #74, reserved: earned powers hook
//     there; unchanged today — PostProcess evaluates before the post-walk
//     steps, which is safe because loss/rediscovery entries are dated at the
//     horizon where nothing accrues).
//  6. emergence draws (the second walk below).
//
// The second walk scans every qualifying event that does not already involve
// an artifact and performs a seeded rarity draw on the artifacts RNG lane. A
// passing draw births a historical artifact at the event, owned by the
// event's beneficiary (the aggressor settlement for Conquest spoils, the
// discovering figure for a Discovery). A failing draw falls back to the
// owner-importance rule: a figure beneficiary whose cumulative reputation
// crosses reputationThreshold births one backdated artifact (one per figure
// per pass). Settlement beneficiaries have no fallback — settlements track
// no reputation.
//
// The lane is consumed in a fixed order per qualifying event: type draw,
// rarity-gate draw, then (on a birth) a name draw. Every draw happens inside
// the stream-order walk, so identical seed and inputs produce identical
// artifacts. Artifacts born mid-walk join the provenance walk like planted
// relics: later events that terminate their owner (owner Death, Conquest or
// Raid of the owner settlement) are attached, associated, and transferred by
// the same rule (recordTransfers applies to born artifacts only — the first
// walk already handled the initial artifacts). Destruction draws (spec 6.6)
// apply to born artifacts in this walk too, by the same rule and the same
// artifacts lane: within each second-walk event the destruction draws for
// terminated born artifacts precede that event's emergence draws, keeping
// the canonical lane order (see destruction.go). transfers supplies the
// figure lifecycle data for transfer destinations (spec 6.3).
//
// The pass extends the event stream (fake-discovery events are prepended,
// rediscovery events appended), so it returns the extended slice; callers
// must use the returned stream, not the one they passed in.
func EmergencePass(artifacts []Artifact, events []simulation.Event, figures []FigureContext, sigCtx SignificanceContext, transfers TransferContext, rng *randv2.Rand) ([]Artifact, []simulation.Event, error) {
	if rng == nil {
		return nil, nil, fmt.Errorf("emergence pass requires the artifacts RNG lane")
	}
	agent := newFakeDiscoveryAgent(figures, rng)
	events = fakeDiscovery(artifacts, events, agent)
	if err := PostProcess(artifacts, events, sigCtx, transfers, rng); err != nil {
		return nil, nil, err
	}
	horizon := HorizonYear(events)
	applyLoss(artifacts, horizon, sigCtx)
	events = rediscovery(artifacts, events, agent, rng, horizon)

	byFigure := make(map[string]FigureContext, len(figures))
	for _, f := range figures {
		byFigure[f.ID] = f
	}

	originCounts := make(map[string]int)
	nameCounts := make(map[string]int)
	fallbackUsed := make(map[string]bool)

	// initial anchors the first walk's artifacts: every element at or past
	// this index was born mid-walk and must not be re-recorded by the second
	// pass over earlier events.
	initial := len(artifacts)
	byID := rebuildByID(artifacts[initial:])

	for i := range events {
		event := &events[i]
		already := event.ArtifactID != ""
		recordTransfers(event, artifacts[initial:], byID, transfers, rng)
		if already || event.ArtifactID != "" {
			continue
		}
		beneficiary, ok := emergenceBeneficiary(event)
		if !ok {
			continue
		}

		typ := emergenceTypePool[rng.IntN(len(emergenceTypePool))]
		if rng.IntN(100) >= emergenceGatePercent(typ) {
			if beneficiary.Kind == "figure" && !fallbackUsed[beneficiary.ID] {
				if fc, ok := byFigure[beneficiary.ID]; ok {
					if year, eventName, crossed := reputationCrossing(fc.Reputation, reputationThreshold); crossed {
						fallbackUsed[beneficiary.ID] = true
						a := birthEmergent(fc.Settlement, year, "", eventName, beneficiary, typ, originCounts, nameCounts, rng)
						artifacts = append(artifacts, a)
						byID = rebuildByID(artifacts[initial:])
					}
				}
			}
			continue
		}

		origin := event.SettlementName
		if origin == "" && beneficiary.Kind == "figure" {
			origin = byFigure[beneficiary.ID].Settlement
		}
		if origin == "" {
			continue
		}
		a := birthEmergent(origin, event.Year, event.ID, event.Category, beneficiary, typ, originCounts, nameCounts, rng)
		artifacts = append(artifacts, a)
		byID = rebuildByID(artifacts[initial:])

		born := &artifacts[len(artifacts)-1]
		event.ArtifactID = born.ID
		born.AssociatedEventIDs = append(born.AssociatedEventIDs, event.ID)
	}
	return artifacts, events, nil
}

// HorizonYear returns the maximum event year in the stream, or 0 when the
// stream is empty. Loss and rediscovery entries minted after the walk are
// dated at the horizon: the world state records no historical population, so
// settlement abandonment is only observable at pass end (spec 6.4 note). The
// exporter uses the same definition for the banner fallback year.
func HorizonYear(events []simulation.Event) int {
	horizon := 0
	for i := range events {
		if events[i].Year > horizon {
			horizon = events[i].Year
		}
	}
	return horizon
}

// emergenceBeneficiary resolves the owner a born artifact is transferred to.
// Conquest spoils go to the aggressor settlement (SettlementName; a conquest
// without a target yields no spoils), and a Discovery goes to the discovering
// figure. Other categories never qualify.
func emergenceBeneficiary(event *simulation.Event) (Owner, bool) {
	switch event.Category {
	case "Conquest":
		if event.SettlementName == "" || event.TargetSettlement == "" {
			return Owner{}, false
		}
		return Owner{Kind: "settlement", ID: event.SettlementName}, true
	case "Discovery":
		if event.FigureID == "" {
			return Owner{}, false
		}
		return Owner{Kind: "figure", ID: event.FigureID}, true
	}
	return Owner{}, false
}

// emergenceGatePercent maps a type to its rarity-gate pass probability.
func emergenceGatePercent(typ string) int {
	switch typ {
	case "crown", "tome", "relic":
		return rareGatePercent
	}
	return commonGatePercent
}

// reputationCrossing returns the year (and the reputation event name) at
// which the cumulative reputation first reaches the threshold, walking the
// entries in recorded order. The entry's event name backdates the fallback
// artifact's provenance entry; entries without a name fall back to
// "Reputation".
func reputationCrossing(entries []ReputationDelta, threshold int) (year int, eventName string, ok bool) {
	total := 0
	for _, e := range entries {
		total += e.Delta
		if total >= threshold {
			if e.Event == "" {
				return e.Year, "Reputation", true
			}
			return e.Year, e.Event, true
		}
	}
	return 0, "", false
}

// birthEmergent creates one historical artifact with deterministic
// `artifact-{origin}-{index}` identity, a name drawn from the artifacts lane,
// and the backdated or event-dated provenance entry that encodes its first
// transfer. The artifact begins with status "created" — the birth event is
// its creation — and the pass immediately applies the first transfer, so the
// materialized status is "held".
func birthEmergent(origin string, year int, eventID, eventType string, owner Owner, typ string, originCounts, nameCounts map[string]int, rng *randv2.Rand) Artifact {
	idx := originCounts[origin]
	originCounts[origin] = idx + 1

	base := fmt.Sprintf("%s of %s", emergenceNamePool[typ][rng.IntN(len(emergenceNamePool[typ]))], origin)
	name := base
	if n := nameCounts[base]; n > 0 {
		name = base + " " + romanNumeral(n+1)
	}
	nameCounts[base] = nameCounts[base] + 1

	var powers []Power
	if p, ok := IntrinsicPower(typ); ok {
		powers = append(powers, p)
	}

	return Artifact{
		ID:                 fmt.Sprintf("artifact-%s-%d", origin, idx),
		Name:               name,
		Type:               typ,
		SignificanceSource: "historical",
		Status:             "held",
		SignificanceScore:  0,
		IsSignificant:      false,
		Powers:             powers,
		Provenance: []ProvenanceEntry{{
			Year:      year,
			Owner:     owner,
			EventID:   eventID,
			EventType: eventType,
		}},
	}
}

// romanNumerals disambiguates repeated emergent names ("Blade of Deepcrest",
// "Blade of Deepcrest II", ...) so every exported note filename stays unique.
func romanNumeral(n int) string {
	vals := []struct {
		v int
		s string
	}{
		{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
	}
	var b strings.Builder
	for _, r := range vals {
		for n >= r.v {
			b.WriteString(r.s)
			n -= r.v
		}
	}
	return b.String()
}

// rebuildByID refreshes the artifact ID index after an append: appends may
// reallocate the backing array, which would invalidate earlier pointers.
func rebuildByID(artifacts []Artifact) map[string]*Artifact {
	byID := make(map[string]*Artifact, len(artifacts))
	for i := range artifacts {
		byID[artifacts[i].ID] = &artifacts[i]
	}
	return byID
}
