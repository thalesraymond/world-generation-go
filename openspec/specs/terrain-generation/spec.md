# terrain-generation Specification

## Purpose

Define the deterministic terrain generation rules for elevation, temperature, humidity, and biome classification.

## Requirements

### Requirement: Noise Map Generation

The system SHALL generate deterministic 2D terrain layers using a provided seed and noise configuration. The configuration SHALL include octave count, persistence, and coordinate scale.

#### Scenario: Deterministic generation

- **WHEN** the noise generator is initialized with a specific seed and configuration
- **THEN** it must produce the exact same sequence of values for given coordinates

#### Scenario: Invalid noise configuration

- **WHEN** octave count, persistence, or scale is zero or negative
- **THEN** the generator replaces that value with a deterministic positive default before sampling.

### Requirement: Normalized Noise Values

Terrain noise samples and every generated environmental layer SHALL be normalized to the closed interval `[0, 1]`.

#### Scenario: Sampling a coordinate

- **WHEN** a noise layer is sampled at any integer coordinate
- **THEN** the resulting value is greater than or equal to `0` and less than or equal to `1`.

### Requirement: Elevation Generation

The system SHALL generate an elevation value for each coordinate using the noise generator.

#### Scenario: Land and Water classification

- **WHEN** generating a terrain tile
- **THEN** it must be classified as Water if the elevation is below a specific threshold, and Land otherwise

#### Scenario: Configured water threshold

- **WHEN** a full map is generated with a positive water threshold
- **THEN** tiles below that configured threshold are water
- **AND** all other biome decisions use the same threshold for that map.

### Requirement: Temperature and Humidity Generation

The system SHALL generate temperature and humidity values for each coordinate.

#### Scenario: Temperature variation

- **WHEN** calculating the temperature of a coordinate
- **THEN** the value must be influenced by both its latitude (Y-coordinate) and its elevation (higher elevation means lower temperature)

#### Scenario: Latitude baseline

- **WHEN** map height is greater than one
- **THEN** base temperature is greatest at the equator and decreases symmetrically toward both map edges.

#### Scenario: Degenerate map height

- **WHEN** a temperature baseline is requested for a map with height one or less
- **THEN** the baseline temperature is `0.5`.

### Requirement: Biome Determination

The system SHALL determine the biome of each coordinate based on its elevation, temperature, and humidity.

#### Scenario: Desert generation

- **WHEN** a land tile has high temperature and low humidity
- **THEN** its biome must be set to Desert

#### Scenario: Tundra generation

- **WHEN** a land tile has low temperature
- **THEN** its biome must be set to Tundra or Snow

### Requirement: Biome Precedence and Baseline Thresholds

Biome classification SHALL apply the first matching rule in this order: water for elevation below the active water threshold; tundra for temperature below `0.25`; desert for temperature above `0.70` and humidity below `0.30`; forest for humidity above `0.60`; and grassland otherwise. The default water threshold SHALL be `0.45`, and the default elevation cooling factor SHALL be `0.35`.

#### Scenario: Deterministic complete map

- **WHEN** the same generator configuration is used twice
- **THEN** both maps have identical dimensions, row-major tile order, environmental values, and biome classifications.

### Requirement: Map Shape and Addressing

Generated maps SHALL store tiles in row-major order and expose bounds-checked coordinate lookup.

#### Scenario: Out-of-bounds lookup

- **WHEN** a caller requests a tile with a negative coordinate or a coordinate outside the configured dimensions
- **THEN** lookup reports that no tile exists rather than indexing the tile array.
