package pointcrawl

import (
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()
	if g == nil {
		t.Fatal("NewGraph returned nil")
	}
	if g.Nodes == nil {
		t.Fatal("NewGraph.Nodes is nil")
	}
	if g.Edges == nil {
		t.Fatal("NewGraph.Edges is nil")
	}
	if g.NodeCount() != 0 {
		t.Fatalf("expected empty graph, got %d nodes", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Fatalf("expected empty graph, got %d edges", g.EdgeCount())
	}
}

func TestAddNode(t *testing.T) {
	g := NewGraph()
	n := &Node{
		ID:         1,
		X:          10,
		Y:          20,
		Visibility: Known,
		Name:       "Old Mill",
		Kind:       "landmark",
	}
	g.AddNode(n)

	if g.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", g.NodeCount())
	}
	got, ok := g.Nodes[1]
	if !ok {
		t.Fatal("node not found by ID")
	}
	if got != n {
		t.Fatal("stored node does not match inserted node")
	}
}

func TestAddNodeNil(t *testing.T) {
	g := NewGraph()
	g.AddNode(nil)
	if g.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes after nil insert, got %d", g.NodeCount())
	}
}

func TestAddEdge(t *testing.T) {
	g := NewGraph()
	g.AddEdge(1, 2, 2)

	if g.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", g.EdgeCount())
	}
	e := g.Edges[0]
	if e.From != 1 || e.To != 2 || e.Cost != 2 {
		t.Fatalf("unexpected edge: %+v", e)
	}
}

func TestVisibility(t *testing.T) {
	cases := []struct {
		v    Visibility
		want int
	}{
		{Known, 0},
		{Unknown, 1},
		{Hidden, 2},
	}
	for _, tc := range cases {
		if int(tc.v) != tc.want {
			t.Fatalf("expected visibility %d, got %d", tc.want, tc.v)
		}
	}
}
