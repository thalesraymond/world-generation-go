package pointcrawl

import (
	"math"
	randv2 "math/rand/v2"
	"sort"
)

// SettlementSite describes the location and faction of an existing
// settlement. It is the minimal information required to filter expansion
// targets, keeping the pointcrawl package decoupled from the world package.
type SettlementSite struct {
	Name    string
	X       int
	Y       int
	Faction string
}

// FindExpansionTarget selects an unclaimed pointcrawl node suitable for
// founding a new settlement. Candidates must be Unknown or Hidden nodes
// within maxRange of (selfX, selfY), at least minDistance away from every
// existing settlement, and must not sit inside another faction's influence
// (unless the expanding settlement is independent). Selection is a weighted
// random draw where closer nodes are more likely to be chosen, making the
// result deterministic for a given RNG state. It returns nil when no
// suitable candidate exists.
func FindExpansionTarget(g *Graph, selfX, selfY int, selfFaction string, sites []SettlementSite, maxRange, minDistance float64, rng *randv2.Rand) *Node {
	if g == nil || g.Nodes == nil || rng == nil {
		return nil
	}

	candidates := g.GetUndiscoveredNear(selfX, selfY, maxRange)
	eligible := make([]*Node, 0, len(candidates))
	for _, node := range candidates {
		if node == nil {
			continue
		}

		if tooCloseToSettlement(node, sites, minDistance) {
			continue
		}

		if insideOtherFactionInfluence(node, selfFaction, sites) {
			continue
		}

		eligible = append(eligible, node)
	}

	if len(eligible) == 0 {
		return nil
	}

	// Deterministic order before the weighted draw.
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].ID < eligible[j].ID
	})

	weights := make([]float64, len(eligible))
	total := 0.0
	for i, node := range eligible {
		d := distance(selfX, selfY, node.X, node.Y)
		// Closer nodes are better expansion targets; guard the degenerate
		// zero-distance case so every candidate keeps a positive weight.
		w := 1.0 / (1.0 + d)
		weights[i] = w
		total += w
	}

	draw := rng.Float64() * total
	for i, w := range weights {
		draw -= w
		if draw < 0 {
			return eligible[i]
		}
	}

	return eligible[len(eligible)-1]
}

func tooCloseToSettlement(node *Node, sites []SettlementSite, minDistance float64) bool {
	for _, site := range sites {
		if distance(node.X, node.Y, site.X, site.Y) < minDistance {
			return true
		}
	}
	return false
}

func insideOtherFactionInfluence(node *Node, selfFaction string, sites []SettlementSite) bool {
	if selfFaction == "independent" || selfFaction == "" {
		return false
	}

	// A node is considered inside another faction's influence when the
	// nearest settlement belongs to a different faction.
	nearest := ""
	nearestDist := math.MaxFloat64
	for _, site := range sites {
		d := distance(node.X, node.Y, site.X, site.Y)
		if d < nearestDist {
			nearestDist = d
			nearest = site.Faction
		}
	}

	return nearest != "" && nearest != selfFaction
}

func distance(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Sqrt(dx*dx + dy*dy)
}
