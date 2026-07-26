package pointcrawl

// Visibility represents how much information the player has about a node.
type Visibility int

const (
	Known Visibility = iota
	Unknown
	Hidden
)

// Node is a point of interest on the pointcrawl graph.
type Node struct {
	ID         int        `json:"id"`
	X          int        `json:"x"`
	Y          int        `json:"y"`
	Visibility Visibility `json:"visibility"`
	Name       string     `json:"name"`
	Kind       string     `json:"kind"`
}

// Edge connects two nodes and records the travel cost in watches.
type Edge struct {
	From int `json:"from"`
	To   int `json:"to"`
	Cost int `json:"cost"`
}

// Graph stores nodes and edges for a pointcrawl network.
type Graph struct {
	Nodes map[int]*Node `json:"nodes"`
	Edges []Edge        `json:"edges"`
}

// NewGraph creates an empty graph ready for nodes and edges.
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[int]*Node),
		Edges: make([]Edge, 0),
	}
}

// AddNode inserts a node into the graph.
func (g *Graph) AddNode(n *Node) {
	if n != nil {
		g.Nodes[n.ID] = n
	}
}

// AddEdge adds a directed edge between two node IDs with the given cost.
func (g *Graph) AddEdge(from, to, cost int) {
	g.Edges = append(g.Edges, Edge{From: from, To: to, Cost: cost})
}

// NodeCount returns the number of nodes in the graph.
func (g *Graph) NodeCount() int {
	if g == nil || g.Nodes == nil {
		return 0
	}

	return len(g.Nodes)
}

// EdgeCount returns the number of edges in the graph.
func (g *Graph) EdgeCount() int {
	if g == nil || g.Edges == nil {
		return 0
	}

	return len(g.Edges)
}
