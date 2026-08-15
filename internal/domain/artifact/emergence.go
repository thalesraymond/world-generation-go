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

// EmergencePass extends PostProcess with emergent artifact births (spec 5.3,
// issue #72). It first runs the provenance and event-ID walk, then a second
// stream-order walk: every qualifying event that does not already involve an
// artifact performs a seeded rarity draw on the artifacts RNG lane. A passing
// draw births a historical artifact at the event, owned by the event's
// beneficiary (the aggressor settlement for Conquest spoils, the discovering
// figure for a Discovery). A failing draw falls back to the owner-importance
// rule: a figure beneficiary whose cumulative reputation crosses
// reputationThreshold births one backdated artifact (one per figure per
// pass). Settlement beneficiaries have no fallback — settlements track no
// reputation.
//
// The lane is consumed in a fixed order per qualifying event: type draw,
// rarity-gate draw, then (on a birth) a name draw. Every draw happens inside
// the stream-order walk, so identical seed and inputs produce identical
// artifacts. Artifacts born mid-walk join the provenance walk like planted
// relics: later events that terminate their owner (owner Death, Conquest or
// Raid of the owner settlement) are attached and associated by the same rule.
func EmergencePass(artifacts []Artifact, events []simulation.Event, figures []FigureContext, sigCtx SignificanceContext, rng *randv2.Rand) ([]Artifact, error) {
	if rng == nil {
		return nil, fmt.Errorf("emergence pass requires the artifacts RNG lane")
	}
	if err := PostProcess(artifacts, events, sigCtx); err != nil {
		return nil, err
	}

	byID := make(map[string]*Artifact, len(artifacts))
	for i := range artifacts {
		byID[artifacts[i].ID] = &artifacts[i]
	}
	byFigure := make(map[string]FigureContext, len(figures))
	for _, f := range figures {
		byFigure[f.ID] = f
	}

	originCounts := make(map[string]int)
	nameCounts := make(map[string]int)
	fallbackUsed := make(map[string]bool)

	for i := range events {
		event := &events[i]
		if event.ArtifactID != "" {
			continue
		}
		attachArtifactID(event, artifacts, byID)
		if event.ArtifactID != "" {
			byID[event.ArtifactID].AssociatedEventIDs = append(byID[event.ArtifactID].AssociatedEventIDs, event.ID)
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
						byID = rebuildByID(artifacts)
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
		byID = rebuildByID(artifacts)

		born := &artifacts[len(artifacts)-1]
		event.ArtifactID = born.ID
		born.AssociatedEventIDs = append(born.AssociatedEventIDs, event.ID)
	}
	return artifacts, nil
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
