# world-state Delta Specification

## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: World State Figure Validation

The world state SHALL validate that figure references are internally consistent within each settlement.

#### Scenario: Invalid figure reference

- **WHEN** a figure references a parent or child ID that does not exist within its settlement
- **THEN** validation SHALL NOT fail (figures from other settlements are not currently referenced; the scope is settlement-level)
- **THEN** the system SHALL gracefully handle dangling references during export (link to placeholder or omit)