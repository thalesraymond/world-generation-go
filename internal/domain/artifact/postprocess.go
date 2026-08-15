package artifact

import (
	"fmt"
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// PostProcess walks the finished event stream in deterministic order and
// materializes post-simulation artifact state. It is a pure domain pass: it
// makes no simulation engine hooks and only mutates the events and artifacts
// it is given.
//
// The pass assigns every event a deterministic event-{year}-{index} ID with a
// monotone index within each year, attaches ArtifactID to events that involve
// an artifact, and builds provenance chains from the ownership transitions the
// stream encodes. Current ownership is always the last provenance entry: no
// separate owner field is maintained.
//
// An event involves an artifact when it already carries ArtifactID (minted or
// attached by lifecycle engines such as rediscovery) or when it terminates the
// artifact's current owner: a Death event of the owner figure, or a
// Conquest/Raid event against the owner settlement. When several owned
// artifacts match, the first in artifact order wins, keeping the pass
// deterministic. Termination records the transfer itself (spec 6.3): every
// terminated artifact gains a provenance entry naming its new owner — the
// heir, the deceased's settlement treasury, or the aggressor settlement — and
// is associated with the event (transfers records both, see recordTransfers).
// Events that name the artifact's new owner directly (a Discovery event
// carrying the discovering figure) also record a provenance entry via
// recordProvenance. Status transitions (lost, held, destroyed) are owned by
// the lifecycle engines and are never touched by this pass.
//
// The pass then evaluates significance for every artifact (spec 4): weighted
// event contributions from the fixed category table, owner-importance accrual
// from sig, threshold crossing with a monotonic latch, and score freezing
// while lost. sig may be the zero value to disable the owner fallback, and
// transfers may be the zero value to disable heir/treasury resolution (death
// transfers then record the lost fallback).
//
// Destruction (spec 6.6) rides the same walk: per terminating event, per
// terminated artifact, destroyIfDrawn performs a seeded draw on the artifacts
// lane (rng). A passing draw marks the artifact destroyed and terminal —
// status "destroyed", a destruction provenance entry, powers cleared — and
// every subsequent event skips it entirely (no transfers, no associations,
// no significance contributions). rng may be nil to disable destruction
// draws entirely; the pipeline always supplies the artifacts lane, so
// identical seeds still produce identical outcomes.
func PostProcess(artifacts []Artifact, events []simulation.Event, sig SignificanceContext, transfers TransferContext, rng *randv2.Rand) error {
	byID := make(map[string]*Artifact, len(artifacts))
	byIdx := make(map[string]int, len(artifacts))
	for i := range artifacts {
		byID[artifacts[i].ID] = &artifacts[i]
		byIdx[artifacts[i].ID] = i
	}

	// eventContributions records, per artifact (parallel to artifacts), the
	// weight-bearing events the walk associates with it.
	eventContributions := make([][]significanceEvent, len(artifacts))

	yearCounts := make(map[int]int)
	for i := range events {
		event := &events[i]
		event.ID = fmt.Sprintf("event-%d-%d", event.Year, yearCounts[event.Year])
		yearCounts[event.Year]++

		terminated := recordTransfers(event, artifacts, byID, transfers, rng)
		if event.ArtifactID == "" {
			continue
		}

		a, ok := byID[event.ArtifactID]
		if !ok {
			return fmt.Errorf("event %s references unknown artifact %q", event.ID, event.ArtifactID)
		}
		// A destroyed artifact is terminal (spec 6.6): it exits all further
		// lifecycle processing, so no later event associates with it or
		// contributes significance. The destruction event itself was already
		// associated by recordTransfers above and keeps its significance
		// weight (spec 6.1: the natural lifecycle event keeps its weight);
		// only the association and provenance are skipped here.
		if a.Status == "destroyed" {
			if weight := eventWeights[event.Category]; weight > 0 {
				eventContributions[byIdx[event.ArtifactID]] = append(eventContributions[byIdx[event.ArtifactID]], significanceEvent{
					year:    event.Year,
					weight:  weight,
					eventID: event.ID,
				})
			}
			continue
		}
		if !containsArtifact(terminated, a) {
			a.AssociatedEventIDs = append(a.AssociatedEventIDs, event.ID)
		}
		recordProvenance(a, event)
		if weight := eventWeights[event.Category]; weight > 0 {
			eventContributions[byIdx[event.ArtifactID]] = append(eventContributions[byIdx[event.ArtifactID]], significanceEvent{
				year:    event.Year,
				weight:  weight,
				eventID: event.ID,
			})
		}
	}

	evaluateSignificance(artifacts, events, eventContributions, sig)
	return nil
}

// recordProvenance appends a provenance entry when the event encodes the
// artifact's new owner. A Discovery event naming the discovering figure moves
// the artifact to that figure; the entry's Owner is the new owner. Other
// events only associate the artifact with the event.
func recordProvenance(a *Artifact, event *simulation.Event) {
	if event.Category != "Discovery" || event.FigureID == "" {
		return
	}
	a.Provenance = append(a.Provenance, ProvenanceEntry{
		Year:      event.Year,
		Owner:     Owner{Kind: "figure", ID: event.FigureID},
		EventID:   event.ID,
		EventType: event.Category,
	})
}

// CurrentOwner returns the owner recorded by the last provenance entry.
// Artifacts without provenance (e.g. planted relics, which begin lost before
// the timeline) return the zero Owner.
func CurrentOwner(a Artifact) Owner {
	if len(a.Provenance) == 0 {
		return Owner{}
	}
	return a.Provenance[len(a.Provenance)-1].Owner
}
