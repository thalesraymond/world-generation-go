# spatial-reasoning Specification

## Purpose

Define how terrain-derived geographic factors are evaluated into settlement suitability scores for downstream simulation and settlement placement.

## Requirements

### Requirement: Suitability Scoring

The system SHALL evaluate each non-water map tile and assign a deterministic settlement-suitability score in the range `[0, 1]` based on nearby water, local elevation variance, biome livability, and an extreme-elevation penalty. Water tiles SHALL always have a score of `0`.

#### Scenario: High suitability near water and flat terrain

- **WHEN** a tile is near a fresh water source, has low elevation variance, and is in a temperate biome
- **THEN** it receives a high suitability score (e.g., > 0.8)

#### Scenario: Low suitability in harsh environments

- **WHEN** a tile is in a desert or high mountain peak
- **THEN** it receives a low suitability score (e.g., < 0.2)

### Requirement: Deterministic Suitability Inputs

Nearby water SHALL be detected within a two-tile radius. Local elevation variance SHALL be the difference between the minimum and maximum elevations in the tile's in-bounds three-by-three neighborhood.

#### Scenario: Edge tile evaluation

- **WHEN** suitability is evaluated for a tile on a map edge or corner
- **THEN** only in-bounds neighboring tiles contribute to water detection and elevation variance.

### Requirement: Baseline Suitability Weights

The baseline score SHALL combine water access with weight `0.4`, flatness with weight `0.3`, and biome livability with weight `0.3`; it SHALL then apply the high-elevation penalty and clamp the result to `[0, 1]`. Water access is `1.0` near water and `0.2` otherwise. Biome livability is grassland `1.0`, forest `0.85`, tundra `0.25`, desert `0.10`, and `0` for all other biomes.

#### Scenario: Suitability-map shape

- **WHEN** suitability is calculated for a terrain map with positive dimensions
- **THEN** the result contains exactly one row-major score for every terrain tile.

#### Scenario: Empty terrain map

- **WHEN** suitability is calculated for a terrain map with zero or negative cell count
- **THEN** the result is empty.
