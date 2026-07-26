# world-state Specification

## Purpose

Define persistence expectations for complete world state snapshots, including terrain-adjacent demographic layers, founded settlements, and embedded historical figures.

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

Each persisted settlement SHALL include its name, grid coordinates, faction, population, type, and a list of historical figures.

#### Scenario: JSON round trip

- **WHEN** a valid state with settlements is serialized and then deserialized
- **THEN** the decoded state contains equivalent dimensions, grid layers, settlement fields, and historical figure data.

#### Scenario: Historical figures in settlement

- **WHEN** a settlement is serialized as part of world state
- **THEN** the `figures` field SHALL be present as a JSON array on the settlement object
- **THEN** each figure in the array SHALL contain id, name, birthYear, deathYear (or 0 if alive), role, faction, and relationships fields
- **THEN** if a settlement has no figures, the `figures` field SHALL be an empty JSON array `[]`

#### Scenario: Deserializing state without figure field

- **WHEN** a world state JSON without a `figures` field on settlements is deserialized
- **THEN** deserialization SHALL succeed and settlements SHALL have an empty Figures slice

### Requirement: World State Figure Validation

The world state SHALL validate that figure references are internally consistent within each settlement.

#### Scenario: Invalid figure reference

- **WHEN** a figure references a parent or child ID that does not exist within its settlement
- **THEN** validation SHALL NOT fail (figures from other settlements are not currently referenced; the scope is settlement-level)
- **THEN** the system SHALL gracefully handle dangling references during export (link to placeholder or omit)
