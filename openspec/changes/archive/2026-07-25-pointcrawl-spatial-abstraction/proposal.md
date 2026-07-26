## Why

To prevent players and Dungeon Masters from being overwhelmed by the high-resolution, continuous noise-based geographic map, we must compress the simulation's geographic output into an abstracted Pointcrawl network (Points of Interest interconnected by routes). This step is essential for pragmatic TTRPG gameplay, providing clear navigation and travel costs rather than tedious pixel-by-pixel map traversal.

## What Changes

- Implement a graph-based spatial abstraction layer that processes the continuous geographic map into a network of Points of Interest (POIs).
- Categorize POIs into "Known Points" (landmarks/cities), "Unknown Points" (wilderness nodes), and "Hidden/Secret Points" (mysterious dungeons).
- Implement a routing algorithm to connect these POIs with traversable paths (e.g., roads, trails, rivers).
- Develop a travel cost calculation system that determines the travel time between POIs in "watches", based on terrain friction, environment types, and distance.
- Prepare the graph data structure for export into relational Markdown artifacts for the Obsidian vault (atlas/bases).

## Capabilities

### New Capabilities
- `pointcrawl-network`: The core logic for detecting Points of Interest on the map and generating an interconnected graph representing trade routes and travel paths.
- `travel-cost-calculator`: Logic to calculate dynamic travel costs (in watches) for paths based on environmental traversal friction (e.g., standard plain trail vs. dangerous orogenic slope).

### Modified Capabilities

## Impact

- Adds new data structures (Nodes, Edges, Graph) on top of the existing geographic grid data.
- Interfaces with the existing geographic map structures to query terrain data for pathfinding and cost calculation.
- Acts as the primary geographic source of truth for the final Obsidian vault exporter.
