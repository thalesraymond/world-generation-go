# demographic-automata Specification

## Purpose

Define deterministic demographic automata behavior for population diffusion and faction influence expansion across the world grid.

## Requirements

### Requirement: Population Spread

The system SHALL simulate the spread of population across the map over discrete iterations.

#### Scenario: Population diffuses to suitable adjacent tiles

- **WHEN** an automata iteration occurs
- **THEN** population moves from high-density tiles to adjacent lower-density tiles, weighted by the destination tile's suitability score

### Requirement: Faction Influence

The system SHALL track faction influence as population spreads.

#### Scenario: Faction territory expansion

- **WHEN** a faction's population spreads into an unoccupied tile
- **THEN** the tile becomes part of that faction's territory
