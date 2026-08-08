# ADR-0008: Pointcrawl Graph — Domain Types, Network Generation, and Travel-Cost Routing

## Status

ACCEPTED

## Date

2026-08-08

## Context

The world needs a navigable abstraction of the map for travel, discovery, and agent expansion. The initial concept calls for a pointcrawl graph: nodes are points of interest (settlements, landmarks, wilderness sites, ruins), edges carry travel costs, and some nodes start unknown to the player and are discovered over time. Costs must reflect terrain so overland travel is meaningfully expensive.

## Decision

Pointcrawl is split across two packages, both current state:

### `internal/domain/pointcrawl` — types and rules

- `Node{ID, X, Y, Visibility, Name, Kind}`, `Edge{From, To, Cost}`, `Graph{Nodes map[int]*Node, Edges []Edge}` with `AddNode`/`AddEdge` and deterministic count helpers.
- `Visibility` is an enum: `Known` (settlements), `Unknown` (landmarks/wilderness), `Hidden` (ruins).
- `GetUndiscoveredNear` returns Unknown/Hidden nodes within a radius, sorted by ID ascending for determinism.
- `FindExpansionTarget` (expansion.go) selects an eligible unclaimed node for founding a new settlement: candidates must be undiscovered, within `maxRange`, at least `minDistance` from every existing settlement, and outside another faction's influence (unless expanding from `independent`). Selection is a weighted random draw favoring closer nodes, deterministic for a given RNG.
- `GraphToJSON` / `GraphFromJSON` handle serialization (the graph is embedded in the world state JSON).

### `internal/geography/pointcrawl` — generation and routing

- `Generate` (generator.go) builds the graph: settlement POIs are always `Known`; terrain nodes are sampled every `SampleStep (8)` tiles as `Unknown` landmarks (elevation > 0.85, non-water) or wilderness (forest); up to 5 `Hidden` ruins are randomly placed on non-water tiles; candidates are shuffled via the pointcrawl RNG, capped at `MaxWildernessNodes (20)`, then culled to keep `MinDistance (5.0)` spacing, preferring higher-visibility nodes.
- `ConnectNodes` (routing.go) connects node pairs closer than `DefaultMaxConnectionDistance (30.0)` with bidirectional edges.
- `CalculateEdgeCost` samples friction along the line between nodes and returns `ceil(distance × averageFriction)` watches. `FrictionTable` maps biomes to friction (water 10, grassland 1, forest/desert/tundra 3) with a +3 bonus above elevation 0.8.
- `pointcrawl.go` re-exports the domain types via aliases — a known shallow seam tracked in ADR-0002, Action 3.

## Alternatives Considered

### Full tile-grid pathfinding with A* per edge

- **Pros:** Precise shortest paths.
- **Cons:** Expensive at generation time and per query; cost recomputation on every map change; overkill for a static travel abstraction.
- **Rejected:** Sampling friction along the direct line is a good approximation at graph scale.

### Delaunay triangulation or k-nearest-neighbor graph

- **Pros:** Sparse, planar, well-studied structures.
- **Cons:** Adds an algorithm and data dependency; k-NN counts are brittle across map sizes.
- **Rejected:** Distance cutoff within 30 tiles is simpler and deterministic.

### Fully connected graph with Euclidean costs

- **Pros:** Trivial to build.
- **Cons:** Unrealistic — every node reachable from every node; no terrain influence.
- **Rejected:** Terrain-aware friction is a core requirement.

### Single package (no geography split)

- **Pros:** Removes the alias seam.
- **Cons:** This is the target state of ADR-0002 Action 3, not yet implemented; the split is acknowledged technical debt rather than a design choice.
- **Deferred:** Consolidate generation/routing into `domain/pointcrawl` when the package is next touched.

## Consequences

- Graph output is deterministic given the pointcrawl RNG stream; `generator_test.go` and `routing_test.go` assert reproducible nodes, edges, and costs.
- Unknown/Hidden nodes provide discovery mechanics and expansion targets without exposing the full map.
- Travel cost in watches is a stable heuristic that the exporter renders as wiki-linked edge tables in `pointcrawl/Network.md`.
- `FrictionTable`, elevation bonus, and connection distance are hardcoded tuning values; changes affect export costs and must be covered by determinism tests.
- The `geography/pointcrawl` alias file is a shallow module (ADR-0002) — flagged, not silently accepted.
