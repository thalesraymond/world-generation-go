## 1. Core Data Structures

- [ ] 1.1 Create `internal/geography/pointcrawl/pointcrawl.go` to define the `Node` and `Edge` structs.
- [ ] 1.2 Implement the `Graph` struct to manage the collection of nodes and edges.
- [ ] 1.3 Add properties for node visibility (`Known`, `Unknown`, `Hidden`) and map coordinates.

## 2. POI Generation

- [ ] 2.1 Implement logic to extract Points of Interest from the continuous map into nodes.
- [ ] 2.2 Add spatial culling to merge close-proximity non-vital nodes to avoid clutter.

## 3. Path Generation and Cost Calculation

- [ ] 3.1 Implement simplified routing logic to connect nearby POI nodes with edges.
- [ ] 3.2 Define a friction table mapping terrain types to base travel cost (e.g., plain=1, mountain=4).
- [ ] 3.3 Create the travel cost calculator logic to compute the total "watches" for an edge based on distance and underlying terrain friction.

## 4. Integration

- [ ] 4.1 Update the main simulation loop to invoke the Pointcrawl network generation at the end of the geographic phase.
- [ ] 4.2 Validate the generated Pointcrawl graph and ensure it's ready for downstream Obsidian Markdown export.
