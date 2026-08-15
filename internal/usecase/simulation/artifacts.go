package simulation

import (
	"fmt"
	"sort"

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
