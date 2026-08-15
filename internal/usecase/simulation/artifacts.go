package simulation

import (
	"fmt"
	"sort"

	"github.com/thalesraymond/world-generation-go/internal/domain/agent"
	"github.com/thalesraymond/world-generation-go/internal/domain/artifact"
	"github.com/thalesraymond/world-generation-go/internal/geography/pointcrawl"
)

// genesisYear is the year world genesis happens (founders are born at year 0
// and simulation starts at year 1).
const genesisYear = 0

// artifactTypes is the deterministic planted-relic type pattern. Spec 5.6
// ranks Crown, Tome, and Relic as rare and Weapon, Armor, and Jewelry as
// common; the pattern repeats common types to mirror that weighting without
// an RNG lane.
var artifactTypes = []string{"weapon", "armor", "jewelry", "weapon", "armor", "crown", "relic", "tome"}

// GeneratePlantedRelics creates one lost, intrinsically significant artifact
// per ruin node in the pointcrawl graph. Ruins are processed in node ID
// ascending order; the function is pure and RNG-free.
func GeneratePlantedRelics(graph *pointcrawl.Graph, genesisYear int) []artifact.Artifact {
	if graph == nil {
		return nil
	}

	ruinIDs := make([]int, 0)
	for id, node := range graph.Nodes {
		if node != nil && node.Kind == "ruin" {
			ruinIDs = append(ruinIDs, id)
		}
	}
	sort.Ints(ruinIDs)

	artifacts := make([]artifact.Artifact, 0, len(ruinIDs))
	for i, id := range ruinIDs {
		node := graph.Nodes[id]
		typ := artifactTypes[node.ID%len(artifactTypes)]
		var powers []artifact.Power
		if p, ok := artifact.IntrinsicPower(typ); ok {
			powers = append(powers, p)
		}
		artifacts = append(artifacts, artifact.Artifact{
			ID:                 fmt.Sprintf("artifact-ruin-%d", i),
			Name:               "Relic of " + node.Name,
			Type:               typ,
			SignificanceSource: "intrinsic",
			Status:             "lost",
			SignificanceScore:  3,
			IsSignificant:      true,
			SignificanceYear:   genesisYear,
			Powers:             powers,
		})
	}
	return artifacts
}

// ownerKey identifies an artifact owner inside the registry's byOwner index.
type ownerKey struct {
	kind string
	id   string
}

// ArtifactRegistry provides indexed access to artifacts: lookup by ID and
// query by current owner (spec 9.1). It mirrors the FigureResolver pattern:
// built from world state at orchestration time and handed to the agent env
// as an agent.ArtifactQuerier. Queries are deterministic — each owner's
// slice is in world-state order.
type ArtifactRegistry struct {
	byOwner map[ownerKey][]artifact.Artifact
	byID    map[string]artifact.Artifact
}

var _ agent.ArtifactQuerier = (*ArtifactRegistry)(nil)

// NewArtifactRegistry builds the byOwner and byID indices from world state
// artifacts. Owners are derived from the last provenance entry (spec 3.1);
// artifacts without provenance index as lost, and destroyed artifacts index
// as destroyed (spec 6.6) — mirroring the exporter's owner resolution.
func NewArtifactRegistry(artifacts []artifact.Artifact) *ArtifactRegistry {
	r := &ArtifactRegistry{
		byOwner: make(map[ownerKey][]artifact.Artifact),
		byID:    make(map[string]artifact.Artifact),
	}
	for _, a := range artifacts {
		r.index(a)
	}
	return r
}

// index inserts a into both indices under its current owner.
func (r *ArtifactRegistry) index(a artifact.Artifact) {
	r.byID[a.ID] = a
	key := registryOwner(a)
	r.byOwner[key] = append(r.byOwner[key], a)
}

// ArtifactsFor returns the artifacts currently held by the given owner, in
// world-state order. Owners that hold nothing get an empty result.
func (r *ArtifactRegistry) ArtifactsFor(ownerKind, ownerID string) []artifact.Artifact {
	return r.byOwner[ownerKey{kind: ownerKind, id: ownerID}]
}

// Get returns the artifact with the given ID.
func (r *ArtifactRegistry) Get(id string) (artifact.Artifact, bool) {
	a, ok := r.byID[id]
	return a, ok
}

// Unlose marks a lost artifact as held by a new owner (spec 9.5). It is a
// stub with the expedition interface deferred: it validates the artifact is
// lost, flips its status to held, and re-indexes it so power queries see it
// under the new owner. Minting the rediscovery event and recording the
// provenance entry (eventID is reserved for that event) is deferred to the
// expedition implementation.
func (r *ArtifactRegistry) Unlose(id, newOwnerKind, newOwnerID, eventID string) error {
	a, ok := r.byID[id]
	if !ok {
		return fmt.Errorf("unlose artifact %q: not found", id)
	}
	if a.Status != "lost" {
		return fmt.Errorf("unlose artifact %q: status is %q, want lost", id, a.Status)
	}
	oldKey := registryOwner(a)
	r.byOwner[oldKey] = withoutArtifact(r.byOwner[oldKey], id)
	a.Status = "held"
	r.byID[id] = a
	newKey := ownerKey{kind: newOwnerKind, id: newOwnerID}
	r.byOwner[newKey] = append(r.byOwner[newKey], a)
	return nil
}

// registryOwner derives the index key for an artifact's current owner.
func registryOwner(a artifact.Artifact) ownerKey {
	if a.Status == "destroyed" {
		return ownerKey{kind: "destroyed"}
	}
	owner := artifact.CurrentOwner(a)
	if owner.Kind != "" {
		return ownerKey{kind: owner.Kind, id: owner.ID}
	}
	// No provenance: planted relics are lost by definition, and any other
	// artifact without provenance is treated as lost (fail-closed).
	return ownerKey{kind: "lost"}
}

// withoutArtifact returns entries minus the artifact with the given ID,
// preserving world-state order.
func withoutArtifact(entries []artifact.Artifact, id string) []artifact.Artifact {
	out := make([]artifact.Artifact, 0, len(entries))
	for _, e := range entries {
		if e.ID != id {
			out = append(out, e)
		}
	}
	return out
}
