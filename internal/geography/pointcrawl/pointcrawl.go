package pointcrawl

import domainpointcrawl "github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"

// Visibility represents how much information the player has about a node.
type Visibility = domainpointcrawl.Visibility

const (
	Known   = domainpointcrawl.Known
	Unknown = domainpointcrawl.Unknown
	Hidden  = domainpointcrawl.Hidden
)

// Node is a point of interest on the pointcrawl graph.
type Node = domainpointcrawl.Node

// Edge connects two nodes and records the travel cost in watches.
type Edge = domainpointcrawl.Edge

// Graph stores nodes and edges for a pointcrawl network.
type Graph = domainpointcrawl.Graph

// NewGraph creates an empty graph ready for nodes and edges.
func NewGraph() *Graph {
	return domainpointcrawl.NewGraph()
}
