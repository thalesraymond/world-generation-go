package artifact

import (
	randv2 "math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
)

// TransferContext supplies the figure lifecycle data the post-processing pass
// needs to resolve transfer destinations (spec 6.3): who inherits a deceased
// owner's artifacts, and which settlement treasury receives them when no heir
// exists. Callers build it from the world state so the artifact domain stays
// decoupled from the figures package. The zero value disables heir and
// treasury resolution: death transfers then record the lost fallback (a
// degenerate case — real figures always have a home settlement).
type TransferContext struct {
	Figures []FigureLifecycle
}

// FigureLifecycle is the minimal figure summary a transfer needs: identity,
// home settlement (the treasury destination), birth/death years, and parent
// IDs (heir resolution).
type FigureLifecycle struct {
	ID         string
	Settlement string // home settlement (treasury destination)
	BirthYear  int
	DeathYear  int // 0 = alive at end of stream
	Parents    []string
}

// heirFor returns the eldest child of the deceased figure who is alive at the
// event year. A child is eligible iff DeathYear == 0 || DeathYear > year.
// Eldest means the smallest BirthYear; ties resolve to the earlier index in
// the Figures slice, so the caller must keep that slice in deterministic
// (world-state) order — never iterate a map here. Returns "" when no
// eligible child exists.
func (ctx TransferContext) heirFor(deceasedID string, year int) string {
	best := -1
	for i := range ctx.Figures {
		f := &ctx.Figures[i]
		if f.DeathYear != 0 && f.DeathYear <= year {
			continue
		}
		for _, parentID := range f.Parents {
			if parentID == deceasedID {
				if best == -1 || f.BirthYear < ctx.Figures[best].BirthYear {
					best = i
				}
				break
			}
		}
	}
	if best == -1 {
		return ""
	}
	return ctx.Figures[best].ID
}

// recordTransfers attaches the event to every artifact whose current owner it
// terminates and records each transfer's provenance entry (spec 6.3/6.1).
// The event's ArtifactID is attached to the first terminated artifact in
// artifact order when not already set, preserving first-match-wins for
// significance; the caller is responsible for associating the ArtifactID
// carrier (the walk skips the carrier's association here so each
// artifact-event pair is recorded exactly once). Returns the terminated
// artifacts.
//
// Destruction (spec 6.6) hooks this walk: per terminated artifact, in
// artifact order, destroyIfDrawn performs the seeded destruction draw on the
// artifacts lane (rng). A passing draw makes the artifact terminal — the
// destruction provenance entry replaces the transfer entry, and every
// subsequent terminating event skips it. rng may be nil to disable
// destruction draws entirely (transfers only).
func recordTransfers(event *simulation.Event, artifacts []Artifact, byID map[string]*Artifact, ctx TransferContext, rng *randv2.Rand) []*Artifact {
	// An unresolvable spoils transfer — a Conquest/Raid whose aggressor
	// (SettlementName) is unknown — must not produce a bogus entry: the
	// match rule needs only TargetSettlement, but without the aggressor
	// there is no destination to record. Such an event is treated as if it
	// terminated nothing: no ArtifactID, no association, no provenance, and
	// no destruction draw (it consumes no lane values).
	unresolvable := (event.Category == "Conquest" || event.Category == "Raid") && event.SettlementName == ""

	var terminated []*Artifact
	for i := range artifacts {
		a := byID[artifacts[i].ID]
		// A destroyed artifact is terminal (spec 6.6): it exits all further
		// lifecycle processing — no later event terminates, transfers,
		// associates, or destroys it.
		if a.Status == "destroyed" {
			continue
		}
		if !terminatesOwner(event, CurrentOwner(*a)) {
			continue
		}
		terminated = append(terminated, a)
	}
	if !unresolvable && event.ArtifactID == "" && len(terminated) > 0 {
		event.ArtifactID = terminated[0].ID
	}
	if unresolvable {
		return terminated
	}
	for _, a := range terminated {
		a.AssociatedEventIDs = append(a.AssociatedEventIDs, event.ID)
		if destroyIfDrawn(event, a, rng) {
			// The natural event is the lifecycle event (spec 6.1): the
			// artifact is already associated above, and the destruction
			// provenance entry records the terminal transition instead of a
			// transfer. No synthetic event is minted.
			continue
		}
		a.Provenance = append(a.Provenance, ProvenanceEntry{
			Year:      event.Year,
			Owner:     transferDestination(event, CurrentOwner(*a), ctx),
			EventID:   event.ID,
			EventType: event.Category,
		})
	}
	return terminated
}

// containsArtifact reports whether the slice holds the given artifact
// pointer, i.e. whether recordTransfers already associated the event with it.
func containsArtifact(list []*Artifact, a *Artifact) bool {
	for _, x := range list {
		if x == a {
			return true
		}
	}
	return false
}

// terminatesOwner reports whether the event ends the given owner's tenure:
// a Death of the owner figure, or a Conquest/Raid against the owner
// settlement (spec 6.3).
func terminatesOwner(event *simulation.Event, owner Owner) bool {
	if owner.Kind == "figure" && event.Category == "Death" && event.FigureID == owner.ID {
		return true
	}
	return owner.Kind == "settlement" && (event.Category == "Conquest" || event.Category == "Raid") && event.TargetSettlement == owner.ID
}

// transferDestination resolves the new owner for a terminated artifact (spec
// 6.3). A Death transfers to the heir alive at the event year; with no heir
// the artifact passes to the deceased figure's settlement treasury; when the
// figure is absent from the context or has no settlement the artifact is
// lost. A Conquest/Raid transfers to the aggressor settlement (event
// SettlementName); callers skip the call when that is empty.
func transferDestination(event *simulation.Event, owner Owner, ctx TransferContext) Owner {
	if owner.Kind == "figure" && event.Category == "Death" {
		if heir := ctx.heirFor(owner.ID, event.Year); heir != "" {
			return Owner{Kind: "figure", ID: heir}
		}
		for i := range ctx.Figures {
			if ctx.Figures[i].ID == owner.ID {
				if ctx.Figures[i].Settlement != "" {
					return Owner{Kind: "settlement", ID: ctx.Figures[i].Settlement}
				}
				break
			}
		}
		return Owner{Kind: "lost"}
	}
	return Owner{Kind: "settlement", ID: event.SettlementName}
}
