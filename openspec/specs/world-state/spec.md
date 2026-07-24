# world-state Specification

## Purpose

Define persistence expectations for complete world state snapshots, including terrain-adjacent demographic layers and founded settlements.

## Requirements

### Requirement: Persist World State

The system SHALL serialize and deserialize a complete world-state snapshot as JSON for later loading. Filesystem ownership is outside the domain state object; the state object SHALL provide validated bytes for an outer persistence adapter.

#### Scenario: Saving demographic and settlement data

- **WHEN** the world state is saved
- **THEN** the demographic grids (population, faction influence) and the list of instantiated settlements MUST be included in the serialized output

### Requirement: Grid-Aligned State Shape

World state SHALL store width, height, population density, faction influence, suitability, and settlements. Each grid-backed layer SHALL contain exactly `width * height` entries and use row-major indexing.

#### Scenario: Constructing a valid state

- **WHEN** a world state is created with positive width and height
- **THEN** every grid-backed layer is initialized with exactly `width * height` entries.

#### Scenario: Invalid serialized layer size

- **WHEN** JSON input contains a grid-backed layer whose length differs from `width * height`
- **THEN** deserialization returns an error rather than a partially valid state.

### Requirement: Bounds-Checked Coordinate Indexing

The state SHALL provide coordinate-to-index lookup for in-bounds coordinates and SHALL reject negative or out-of-range coordinates.

#### Scenario: Coordinate outside state dimensions

- **WHEN** a caller requests an index outside the world dimensions
- **THEN** the lookup reports that no valid index exists.

### Requirement: Suitability Layer Assignment

Replacing the suitability layer SHALL validate that the supplied layer has exactly one score for each state cell and SHALL copy the supplied values so callers cannot mutate state through the original slice.

#### Scenario: Invalid suitability assignment

- **WHEN** a caller supplies a suitability slice with an incorrect length
- **THEN** assignment returns an error and does not install that slice.

### Requirement: Settlement Snapshot Fields

Each persisted settlement SHALL include its name, grid coordinates, faction, and population.

#### Scenario: JSON round trip

- **WHEN** a valid state with settlements is serialized and then deserialized
- **THEN** the decoded state contains equivalent dimensions, grid layers, and settlement fields.
