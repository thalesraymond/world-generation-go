package pointcrawl

import (
	"encoding/json"
	"testing"
)

func TestNewGraphInitializesEmptyCollections(t *testing.T) {
	g := NewGraph()
	if g == nil {
		t.Fatal("NewGraph returned nil")
	}
	if g.Nodes == nil {
		t.Fatal("Nodes map is nil")
	}
	if g.Edges == nil {
		t.Fatal("Edges slice is nil")
	}
	if g.NodeCount() != 0 {
		t.Fatalf("NodeCount = %d, want 0", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Fatalf("EdgeCount = %d, want 0", g.EdgeCount())
	}
}

func TestAddNodeStoresNodeByID(t *testing.T) {
	g := NewGraph()
	n := &Node{ID: 7, X: 1, Y: 2, Name: "Tower", Kind: "landmark", Visibility: Known}
	g.AddNode(n)

	if g.NodeCount() != 1 {
		t.Fatalf("NodeCount = %d, want 1", g.NodeCount())
	}

	stored, ok := g.Nodes[7]
	if !ok {
		t.Fatal("node not found by ID")
	}
	if stored != n {
		t.Fatal("stored node is not the inserted node")
	}
}

func TestAddNodeIgnoresNil(t *testing.T) {
	g := NewGraph()
	g.AddNode(nil)
	if g.NodeCount() != 0 {
		t.Fatalf("NodeCount = %d, want 0 after nil insert", g.NodeCount())
	}
}

func TestAddEdgeAppendsEdge(t *testing.T) {
	g := NewGraph()
	g.AddEdge(1, 2, 3)

	if g.EdgeCount() != 1 {
		t.Fatalf("EdgeCount = %d, want 1", g.EdgeCount())
	}

	edge := g.Edges[0]
	if edge.From != 1 || edge.To != 2 || edge.Cost != 3 {
		t.Fatalf("unexpected edge: %+v", edge)
	}
}

func TestNodeCountNilGraph(t *testing.T) {
	var g *Graph
	if g.NodeCount() != 0 {
		t.Fatalf("nil graph NodeCount = %d, want 0", g.NodeCount())
	}
}

func TestEdgeCountNilGraph(t *testing.T) {
	var g *Graph
	if g.EdgeCount() != 0 {
		t.Fatalf("nil graph EdgeCount = %d, want 0", g.EdgeCount())
	}
}

func TestVisibilityOrdering(t *testing.T) {
	cases := []struct {
		got  Visibility
		want int
	}{
		{Known, 0},
		{Unknown, 1},
		{Hidden, 2},
	}

	for _, tc := range cases {
		if int(tc.got) != tc.want {
			t.Fatalf("visibility %d want %d", tc.got, tc.want)
		}
	}
}

func TestGraphMarshalUnmarshalRoundTrip(t *testing.T) {
	g := NewGraph()
	g.AddNode(&Node{ID: 0, X: 1, Y: 2, Name: "A", Kind: "settlement", Visibility: Known})
	g.AddNode(&Node{ID: 1, X: 4, Y: 5, Name: "B", Kind: "wilderness", Visibility: Unknown})
	g.AddEdge(0, 1, 4)
	g.AddEdge(1, 0, 4)

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	decoded, err := GraphFromJSON(data)
	if err != nil {
		t.Fatalf("GraphFromJSON error = %v", err)
	}

	if decoded.NodeCount() != g.NodeCount() {
		t.Fatalf("NodeCount mismatch: got %d want %d", decoded.NodeCount(), g.NodeCount())
	}
	if decoded.EdgeCount() != g.EdgeCount() {
		t.Fatalf("EdgeCount mismatch: got %d want %d", decoded.EdgeCount(), g.EdgeCount())
	}

	nodeA, ok := decoded.Nodes[0]
	if !ok || nodeA.Name != "A" || nodeA.Visibility != Known {
		t.Fatalf("decoded node A unexpected: %+v", nodeA)
	}
}

func TestGraphToJSONNilGraphReturnsError(t *testing.T) {
	_, err := GraphToJSON(nil)
	if err == nil {
		t.Fatal("expected error for nil graph")
	}
}

func TestGraphToJSONSuccess(t *testing.T) {
	g := NewGraph()
	g.AddNode(&Node{ID: 0, X: 1, Y: 2, Name: "A", Kind: "settlement", Visibility: Known})
	g.AddEdge(0, 0, 1)

	data, err := GraphToJSON(g)
	if err != nil {
		t.Fatalf("GraphToJSON error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty JSON output")
	}

	decoded, err := GraphFromJSON(data)
	if err != nil {
		t.Fatalf("GraphFromJSON error = %v", err)
	}

	if decoded.NodeCount() != 1 {
		t.Fatalf("decoded NodeCount = %d, want 1", decoded.NodeCount())
	}
	if decoded.EdgeCount() != 1 {
		t.Fatalf("decoded EdgeCount = %d, want 1", decoded.EdgeCount())
	}
}

func TestGraphFromJSONInvalidPayloadReturnsError(t *testing.T) {
	_, err := GraphFromJSON([]byte("not-json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGraphFromJSONInitializesMissingCollections(t *testing.T) {
	t.Run("null edges", func(t *testing.T) {
		data := []byte(`{"nodes":{},"edges":null}`)
		g, err := GraphFromJSON(data)
		if err != nil {
			t.Fatalf("GraphFromJSON error = %v", err)
		}
		if g.Edges == nil {
			t.Fatal("expected Edges slice to be initialized")
		}
		if g.NodeCount() != 0 {
			t.Fatalf("NodeCount = %d, want 0", g.NodeCount())
		}
	})

	t.Run("missing nodes", func(t *testing.T) {
		data := []byte(`{"edges":[]}`)
		g, err := GraphFromJSON(data)
		if err != nil {
			t.Fatalf("GraphFromJSON error = %v", err)
		}
		if g.Nodes == nil {
			t.Fatal("expected Nodes map to be initialized")
		}
		if g.EdgeCount() != 0 {
			t.Fatalf("EdgeCount = %d, want 0", g.EdgeCount())
		}
	})
}

func TestGetUndiscoveredNear_ReturnsExpectedNodes(t *testing.T) {
	g := NewGraph()
	g.AddNode(&Node{ID: 1, X: 0, Y: 0, Name: "KnownClose", Visibility: Known})
	g.AddNode(&Node{ID: 2, X: 3, Y: 4, Name: "UnknownAtRadius", Visibility: Unknown})
	g.AddNode(&Node{ID: 3, X: 6, Y: 8, Name: "HiddenOutOfRange", Visibility: Hidden})
	g.AddNode(&Node{ID: 4, X: 1, Y: 1, Name: "HiddenClose", Visibility: Hidden})

	nodes := g.GetUndiscoveredNear(0, 0, 5.0)
	if len(nodes) != 2 {
		t.Fatalf("len(nodes) = %d, want 2", len(nodes))
	}
	if nodes[0].ID != 2 {
		t.Errorf("nodes[0].ID = %d, want 2", nodes[0].ID)
	}
	if nodes[1].ID != 4 {
		t.Errorf("nodes[1].ID = %d, want 4", nodes[1].ID)
	}
}

func TestGetUndiscoveredNear_EmptyWhenNone(t *testing.T) {
	g := NewGraph()
	g.AddNode(&Node{ID: 1, X: 0, Y: 0, Name: "Known", Visibility: Known})
	g.AddNode(&Node{ID: 2, X: 10, Y: 10, Name: "UnknownFar", Visibility: Unknown})

	nodes := g.GetUndiscoveredNear(0, 0, 5.0)
	if len(nodes) != 0 {
		t.Fatalf("len(nodes) = %d, want 0", len(nodes))
	}
}

func TestGetUndiscoveredNear_DeterministicOrder(t *testing.T) {
	g := NewGraph()
	g.AddNode(&Node{ID: 5, X: 1, Y: 0, Visibility: Unknown})
	g.AddNode(&Node{ID: 2, X: 0, Y: 1, Visibility: Hidden})
	g.AddNode(&Node{ID: 9, X: -1, Y: 0, Visibility: Unknown})

	first := g.GetUndiscoveredNear(0, 0, 2.0)
	second := g.GetUndiscoveredNear(0, 0, 2.0)

	if len(first) != 3 {
		t.Fatalf("len(first) = %d, want 3", len(first))
	}
	if len(second) != 3 {
		t.Fatalf("len(second) = %d, want 3", len(second))
	}
	expected := []int{2, 5, 9}
	for i, id := range expected {
		if first[i].ID != id {
			t.Errorf("first[%d].ID = %d, want %d", i, first[i].ID, id)
		}
		if second[i].ID != id {
			t.Errorf("second[%d].ID = %d, want %d", i, second[i].ID, id)
		}
	}
}

func TestGetUndiscoveredNear_NilGraph(t *testing.T) {
	var g *Graph
	nodes := g.GetUndiscoveredNear(0, 0, 10.0)
	if nodes != nil {
		t.Fatalf("expected nil, got %v", nodes)
	}
}
