package pointcrawl

import (
	"encoding/json"
	"fmt"
)

// GraphToJSON serializes a graph to JSON bytes.
func GraphToJSON(graph *Graph) ([]byte, error) {
	if graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}

	data, err := json.Marshal(graph)
	if err != nil {
		return nil, fmt.Errorf("marshal pointcrawl graph: %w", err)
	}

	return data, nil
}

// GraphFromJSON deserializes a graph from JSON bytes.
func GraphFromJSON(data []byte) (*Graph, error) {
	var graph Graph
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("unmarshal pointcrawl graph: %w", err)
	}

	if graph.Nodes == nil {
		graph.Nodes = make(map[int]*Node)
	}

	if graph.Edges == nil {
		graph.Edges = make([]Edge, 0)
	}

	return &graph, nil
}
