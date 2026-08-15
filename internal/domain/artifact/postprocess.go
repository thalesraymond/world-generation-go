package artifact

import (
	"fmt"

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
// attached by lifecycle engines such as rediscovery and transfers) or when it
// terminates the artifact's current owner: a Death event of the owner figure,
// or a Conquest/Raid event against the owner settlement. When several owned
// artifacts match, the first in artifact order wins, keeping the pass
// deterministic. Events that name the artifact's new owner directly (a
// Discovery event carrying the discovering figure) record a provenance entry;
// ownership changes whose destination requires lifecycle rules are recorded by
// those rules' engines, and the event is only associated with the artifact
// here. Status transitions (lost, held, destroyed) are owned by the lifecycle
// engines and are never touched by this pass.
func PostProcess(artifacts []Artifact, events []simulation.Event) error {
	byID := make(map[string]*Artifact, len(artifacts))
	for i := range artifacts {
		byID[artifacts[i].ID] = &artifacts[i]
	}

	yearCounts := make(map[int]int)
	for i := range events {
		event := &events[i]
		event.ID = fmt.Sprintf("event-%d-%d", event.Year, yearCounts[event.Year])
		yearCounts[event.Year]++

		attachArtifactID(event, artifacts, byID)
		if event.ArtifactID == "" {
			continue
		}

		a, ok := byID[event.ArtifactID]
		if !ok {
			return fmt.Errorf("event %s references unknown artifact %q", event.ID, event.ArtifactID)
		}
		a.AssociatedEventIDs = append(a.AssociatedEventIDs, event.ID)
		recordProvenance(a, event)
	}
	return nil
}

// attachArtifactID marks events that involve an artifact. An event already
// carrying ArtifactID is left untouched; otherwise the artifacts whose current
// owner the event terminates are matched in artifact order and the first
// match wins, so the result is deterministic for a given input order.
func attachArtifactID(event *simulation.Event, artifacts []Artifact, byID map[string]*Artifact) {
	if event.ArtifactID != "" {
		return
	}
	for i := range artifacts {
		a := byID[artifacts[i].ID]
		owner := CurrentOwner(*a)
		if owner.Kind == "figure" && event.Category == "Death" && event.FigureID == owner.ID {
			event.ArtifactID = a.ID
			return
		}
		if owner.Kind == "settlement" && (event.Category == "Conquest" || event.Category == "Raid") && event.TargetSettlement == owner.ID {
			event.ArtifactID = a.ID
			return
		}
	}
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
