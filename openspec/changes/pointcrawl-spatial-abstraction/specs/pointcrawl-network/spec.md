## ADDED Requirements

### Requirement: Generate pointcrawl nodes from map data
The system SHALL identify geographic locations of significance (e.g., capitals, ruins) and transform them into discrete nodes (Points of Interest).

#### Scenario: Nodes generation
- **WHEN** the geographic map is processed by the spatial abstraction layer
- **THEN** it generates a list of nodes categorized as Known, Unknown, or Hidden/Secret Points.

### Requirement: Generate routes connecting nodes
The system SHALL establish paths (edges) between generated nodes representing traversable routes (trade routes, fluvial paths, underground paths).

#### Scenario: Route establishment
- **WHEN** a set of POI nodes is generated
- **THEN** the system establishes logical paths connecting them to form a Pointcrawl graph.
