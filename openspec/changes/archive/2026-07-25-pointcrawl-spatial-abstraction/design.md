## Context

The world generation system currently produces a highly detailed, continuous noise-based geographical map. While structurally impressive, presenting raw, pixel-by-pixel geographic matrix data to a Dungeon Master is overwhelming and impractical for actual TTRPG play. The Pointcrawl abstraction distills this dense spatial data into key locations (Points of Interest) and the travel paths between them, vastly improving usability for gameplay and downstream Obsidian Markdown generation.

## Goals / Non-Goals

**Goals:**
- Extract discrete Points of Interest (POIs) from the continuous geographic map data.
- Generate a graph (nodes and edges) representing travel routes between these POIs.
- Calculate travel cost ("watches") for edges based on distance and underlying terrain friction.

**Non-Goals:**
- We are explicitly NOT building a hex-based movement or exploration system (hexcrawl).
- We are NOT discarding the underlying geographic noise data; the continuous map remains active in the background for narrative event resolution.

## Decisions

- **Graph Representation:** The abstraction will utilize a standard node-edge graph data structure inside a new package (e.g., `internal/geography/pointcrawl/`). Each node will represent a POI and maintain its underlying coordinates from the continuous map.
- **Node Categorization (Visibility):** Nodes will be flagged with a visibility state: `Known`, `Unknown`, or `Hidden/Secret`. This defines how they are treated during export and within the CLI narrative logs.
- **Path Cost Heuristic:** Edges will store the travel cost measured in "watches". A cost calculator will sample the terrain along the straight-line segment or simplified path between two nodes on the base grid, applying a predefined friction table (e.g., plain=1, mountain=4, forest=3) to compute the total required watches.

## Risks / Trade-offs

- **Risk: Node Clutter.** Generating too many nodes could recreate the very overwhelming complexity we are trying to solve.
  - **Mitigation:** Implement a spatial culling threshold or minimum grouping distance to merge nearby non-vital locations, keeping the network sparse and readable.
- **Risk: Pathing Computation Cost.** True pixel-perfect pathfinding (like A*) over a massive continuous map could slow down generation.
  - **Mitigation:** Use simplified line-of-sight terrain sampling heuristics between nodes for initial route generation instead of true pathfinding, maintaining the engine's high performance.
