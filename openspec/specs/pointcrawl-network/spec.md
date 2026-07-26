# pointcrawl-network Specification

## Purpose

Define how the world map is abstracted into a pointcrawl graph of Points of Interest (POIs) connected by traversable routes, including a query interface used by Explorer figures to discover adjacent undiscovered nodes.

## Requirements

### Requirement: Generate pointcrawl nodes from map data
The system SHALL identify geographic locations of significance (e.g., capitals, ruins) and transform them into discrete nodes (Points of Interest). Nodes SHALL be referenceable by Explorer figures during figure event generation.

#### Scenario: Nodes generation
- **WHEN** the geographic map is processed by the spatial abstraction layer
- **THEN** it generates a list of nodes categorized as Known, Unknown, or Hidden/Secret Points.

### Requirement: Generate routes connecting nodes
The system SHALL establish paths (edges) between generated nodes representing traversable routes (trade routes, fluvial paths, underground paths). These edges SHALL inform Explorer figure events about which nodes are "adjacent" to settlements.

#### Scenario: Route establishment
- **WHEN** a set of POI nodes is generated
- **THEN** the system establishes logical paths connecting them to form a Pointcrawl graph.

### Requirement: Explorer Node Query Interface
The pointcrawl graph SHALL provide a query method enabling Explorer figures to discover undiscovered nodes adjacent to a settlement's region.

#### Scenario: Querying undiscovered nodes
- **WHEN** an Explorer figure queries the pointcrawl graph for nodes near a given coordinate (settlement location)
- **THEN** the graph SHALL return all nodes whose visibility is `Unknown` or `Hidden` and whose coordinates are within a configurable radius of the settlement
- **THEN** the query SHALL be deterministic and read-only (does not modify node visibility)

#### Scenario: No undiscovered nodes
- **WHEN** an Explorer queries the graph and no undiscovered nodes exist near the settlement
- **THEN** the query SHALL return an empty list
- **THEN** the Explorer SHALL generate generic discovery events not tied to specific nodes
