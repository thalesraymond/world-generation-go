package artifact

import (
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// Destruction rules (spec 6.6).
//
// The artifacts RNG lane (spec 10.4, state.Engine.GetPRNG("artifacts")) is
// consumed by the fixed stream-order walks in a canonical order shared by
// all artifact lifecycle features:
//
//  1. fake-discovery draws (issue #70, pre-walk)
//  2. destruction draws (issue #71, this file) — one IntN(100) per
//     terminated artifact, in both walks, stream order, artifact order
//     within each event
//  3. rediscovery draws (issue #70, post-walk)
//  4. earned-power draws (issue #74)
//  5. emergence draws (issue #72, second walk)
//
// Destruction draws sit inside the walk (recordTransfers), so all first-walk
// destruction draws precede the second walk, and within each second-walk
// event the destruction draws precede that event's emergence draws.
// Unresolvable spoils events — Conquest/Raid without an aggressor — are
// treated as terminating nothing and consume no lane values.

// destructionPercent is the pass probability of the destruction draw (spec
// 6.6 "low probability"): 10%. Fixed so the pass is deterministic per seed.
const destructionPercent = 10

// destroyIfDrawn performs the seeded destruction draw for an artifact whose
// owner the event terminates (spec 6.6). The draw triggers when the owner
// settlement is destroyed by Conquest, or when the owner figure dies — every
// owner-figure Death stands in for "dies in war", the war-death proxy: the
// simulation has no war category, so it cannot distinguish war deaths from
// other deaths. Raid plunder never destroys an artifact and consumes no lane
// values.
//
// On a pass the artifact becomes terminal: status "destroyed", a provenance
// entry records the destruction — the terminal timeline entry, whose Owner
// is the owner at destruction — and the artifact's powers vanish (spec
// 7.7). No transfer provenance entry is written, and no synthetic event is
// minted: the natural event itself IS the lifecycle event (spec 6.1), so it
// carries the ArtifactID and the artifact is associated with it (see
// recordTransfers). The artifact then exits all further lifecycle
// processing: later terminating events skip it.
//
// rng may be nil to disable destruction draws entirely (callers that do not
// supply the artifacts lane then get transfers only). The draw is one
// IntN(100) on the lane per terminated artifact, in artifact order, so the
// pass/fail sequence is fully deterministic for a given seed.
func destroyIfDrawn(event *simulation.Event, a *Artifact, rng *randv2.Rand) bool {
	if rng == nil || event.Category == "Raid" {
		return false
	}
	if rng.IntN(100) >= destructionPercent {
		return false
	}
	a.Status = "destroyed"
	a.Powers = nil
	a.Provenance = append(a.Provenance, ProvenanceEntry{
		Year:      event.Year,
		Owner:     CurrentOwner(*a),
		EventID:   event.ID,
		EventType: event.Category,
	})
	return true
}

// DestructionYear returns the year the artifact was destroyed — the year of
// its terminal provenance entry (spec 6.6). It returns 0 when the artifact
// is not destroyed.
func DestructionYear(a Artifact) int {
	if a.Status != "destroyed" || len(a.Provenance) == 0 {
		return 0
	}
	return a.Provenance[len(a.Provenance)-1].Year
}
