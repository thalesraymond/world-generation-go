package artifact

// abandonedClass is the settlement size class that triggers loss (spec 6.4).
// It mirrors settlement.Classify's output; the artifact domain compares the
// string literal so it stays decoupled from the settlement package.
const abandonedClass = "Abandoned"

// applyLoss runs lifecycle step 3, after the provenance walk: every artifact
// whose current owner is a settlement whose FINAL size class is Abandoned is
// recorded lost at the horizon year with the synthetic ArtifactLoss event
// type (spec 6.1) and no event ID — the loss is a post-hoc observation, not
// a stream event. The world state records no historical population, so
// abandonment is only observable at pass end; the loss year is therefore the
// horizon (the maximum event year, 0 for an empty stream).
//
// The step then propagates the lost state to status: any artifact whose last
// provenance entry is Owner{Kind: "lost"} — the abandonment loss above or a
// mid-stream degenerate death transfer (spec 6.3: no heir and no settlement)
// — is set Status "lost" so rediscovery and export see a consistent state.
// Artifacts owned by figures or standing settlements are untouched.
func applyLoss(artifacts []Artifact, horizon int, sig SignificanceContext) {
	for i := range artifacts {
		a := &artifacts[i]
		owner := CurrentOwner(*a)
		if owner.Kind == "settlement" && sig.SettlementClass[owner.ID] == abandonedClass {
			a.Provenance = append(a.Provenance, ProvenanceEntry{
				Year:      horizon,
				Owner:     Owner{Kind: "lost"},
				EventID:   "",
				EventType: "ArtifactLoss",
			})
		}
		if CurrentOwner(*a).Kind == "lost" {
			a.Status = "lost"
		}
	}
}

// LostSinceYear returns the year of the artifact's most recent lost
// provenance entry — the year its current (or last) lost span began. ok is
// false when no lost entry exists (an artifact lost before any recorded
// entry, e.g. a planted relic never found).
func LostSinceYear(a Artifact) (year int, ok bool) {
	for i := range a.Provenance {
		if a.Provenance[i].Owner.Kind == "lost" {
			year = a.Provenance[i].Year
			ok = true
		}
	}
	return year, ok
}
