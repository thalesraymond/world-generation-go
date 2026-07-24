# demographic-automata Specification

## Purpose

Define deterministic demographic automata behavior for population diffusion and faction influence expansion across the world grid.

## Requirements

### Requirement: Population Spread

The system SHALL simulate population diffusion across the map in a configured number of discrete iterations. During an iteration, each populated tile retains the portion not selected for diffusion and distributes the selected portion only to adjacent lower-density tiles with positive suitability.

#### Scenario: Population diffuses to suitable adjacent tiles

- **WHEN** an automata iteration occurs
- **THEN** population moves from high-density tiles to adjacent lower-density tiles, weighted by the destination tile's suitability score

#### Scenario: No eligible destination

- **WHEN** a populated tile has no eligible neighboring destination
- **THEN** its selected diffusion amount remains on the source tile.

### Requirement: Neighborhood and Diffusion Bounds

The automata SHALL use the eight surrounding in-bounds cells as a tile's neighborhood. Diffusion rate SHALL be clamped to `[0, 1]`; a non-positive iteration count SHALL leave the state unchanged.

#### Scenario: Edge diffusion

- **WHEN** population diffuses from an edge or corner tile
- **THEN** only in-bounds neighboring cells are considered.

### Requirement: Faction Influence

The system SHALL track faction influence as population spreads.

#### Scenario: Faction territory expansion

- **WHEN** a faction's population spreads into an unoccupied tile
- **THEN** the tile becomes part of that faction's territory

#### Scenario: Influence selection

- **WHEN** a populated tile has neighboring factions
- **THEN** it adopts the faction with the greatest neighboring population contribution
- **AND** an unpopulated tile has no faction influence.

### Requirement: Deterministic Population Seeding

The system SHALL seed population density as the square of each tile's suitability. Tiles below the configured minimum population SHALL remain unclaimed. Claimed tiles SHALL assign faction names deterministically by row-major coordinate order, cycling through configured faction names; an empty faction list SHALL use `independent`.

#### Scenario: Baseline configuration

- **WHEN** the default demographic configuration is used
- **THEN** it uses eight iterations, diffusion rate `0.3`, minimum population `0.05`, and factions `auric`, `verdant`, and `cinder`.

### Requirement: State and Terrain Shape Agreement

Suitability pre-generation SHALL reject a nil state and SHALL reject terrain maps whose dimensions do not match the world state. On success, it SHALL store one suitability value per state cell before seeding or diffusion.

#### Scenario: Mismatched dimensions

- **WHEN** suitability pre-generation receives a state and terrain map with different dimensions
- **THEN** it returns an error without storing a mismatched suitability layer.
