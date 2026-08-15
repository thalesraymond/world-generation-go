package artifact

import randv2 "math/rand/v2"

// Earned-power parameters (spec 7.4/7.5, issue #74). The base magnitude of an
// earned combat/influence power is drawn from the artifacts RNG lane, so it
// is deterministic from the master seed; narrative earned powers carry no
// magnitude (spec 7.5) and consume no draw.
const (
	// earnedPowerMinBase is the lowest base magnitude an earned power can
	// draw.
	earnedPowerMinBase = 1
	// earnedPowerBaseSpan is the number of distinct magnitudes in the draw:
	// 1 + IntN(3) yields 1..3.
	earnedPowerBaseSpan = 3
)

// earnedNarrativeEffects maps a pivotal event category to the deterministic
// effect string of the narrative power it grants (spec 7.8). Only the
// weight-bearing categories that map to narrative powers appear here —
// War/Conquest grant combat and Diplomacy/Politics grant influence (spec
// 7.4), and Economy carries weight 0 so it can never be pivotal. The default
// covers any category outside the weight table defensively.
var earnedNarrativeEffects = map[string]string{
	"Raid":      "survived a raid, bearer gains the raider's boldness",
	"Expansion": "witnessed expansion, bearer gains a pioneer's resolve",
	"Disaster":  "survives calamity, bearer gains resilience",
}

// earnedNarrativeEffect returns the effect string for a pivotal event
// category, falling back to a generic effect for categories without an entry.
func earnedNarrativeEffect(category string) string {
	if effect, ok := earnedNarrativeEffects[category]; ok {
		return effect
	}
	return "shaped by its history, bearer gains renown"
}

// grantEarnedPower appends the artifact's earned power at the pivotal
// crossing (spec 7.4): the crossing event's category selects the power type
// — War/Conquest grant combat, Diplomacy/Politics grant influence, anything
// else grants narrative (spec 7.4) — and, for magnitude-bearing powers, one
// base magnitude is drawn from the artifacts RNG lane.
//
// Lane order (issue #74; the slots are shared with the in-flight #70 and
// #71 branches so they merge coherently): fake-discovery draws (pre-walk),
// destruction draws (in-walk), loss/rediscovery draws (post-walk), earned-
// power magnitude draws here (significance evaluation, per artifact in
// artifact order), then emergence draws (second walk, in event order). A
// draw is consumed only when a magnitude-bearing power is granted; narrative
// earned powers carry no magnitude (spec 7.5) and consume nothing.
//
// A nil lane (PostProcess called without the artifacts RNG) grants nothing:
// earned powers exist only on the pipeline path that threads the lane.
func grantEarnedPower(a *Artifact, category string, rng *randv2.Rand) {
	if rng == nil {
		return
	}
	switch category {
	case "War", "Conquest":
		a.Powers = append(a.Powers, CombatPower{Base: drawEarnedMagnitude(rng), Source: "earned"})
	case "Diplomacy", "Politics":
		a.Powers = append(a.Powers, InfluencePower{Base: drawEarnedMagnitude(rng), Source: "earned"})
	default:
		a.Powers = append(a.Powers, NarrativePower{Effect: earnedNarrativeEffect(category), Source: "earned"})
	}
}

// drawEarnedMagnitude draws one earned-power base magnitude from the
// artifacts RNG lane, uniformly in [earnedPowerMinBase,
// earnedPowerMinBase+earnedPowerBaseSpan).
func drawEarnedMagnitude(rng *randv2.Rand) int {
	return earnedPowerMinBase + rng.IntN(earnedPowerBaseSpan)
}
