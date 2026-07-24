# settlement-generation Specification

## Purpose

Define how discrete settlements are founded from demographic and suitability state in a deterministic world generation pass.

## Requirements

### Requirement: Settlement Instantiation

The system SHALL instantiate discrete settlements at locations meeting configured population and suitability thresholds, after applying deterministic spatial spacing.

#### Scenario: Founding a new settlement

- **WHEN** a tile's population density reaches a threshold AND its suitability score is above a minimum value AND it is far enough from existing settlements
- **THEN** a new settlement is instantiated at that location and assigned to the local faction

### Requirement: Candidate Selection and Ordering

Candidates SHALL be tiles whose suitability and population each meet their configured minimum. Candidates SHALL be considered in descending order of `suitability * population`, with stable row-major ordering used to break equal scores.

#### Scenario: More suitable candidate wins

- **WHEN** two otherwise eligible nearby candidates compete for a settlement position
- **THEN** the candidate with the higher suitability-population score is considered first.

### Requirement: Spacing and Limits

A candidate SHALL be rejected when its Euclidean distance from an already selected settlement is less than the configured minimum distance. A positive maximum-settlement limit SHALL stop selection when reached; zero or a negative limit SHALL not impose a count limit.

#### Scenario: Adjacent candidates

- **WHEN** candidates are closer together than the configured minimum distance
- **THEN** only the first selected candidate is founded.

### Requirement: Baseline Generation Configuration

The default configuration SHALL use minimum suitability `0.65`, minimum population `0.35`, minimum distance `3`, and no maximum settlement count.

#### Scenario: Default thresholds

- **WHEN** a tile has suitability below `0.65` or population below `0.35` under the default configuration
- **THEN** it is not a settlement candidate.

### Requirement: Settlement Identity and Faction

Generated settlements SHALL be named sequentially as `Settlement-001`, `Settlement-002`, and so on in selection order. A settlement SHALL inherit the source tile's faction; a tile without faction influence SHALL produce an `independent` settlement.

#### Scenario: Unclaimed settlement

- **WHEN** an eligible source tile has no faction influence
- **THEN** the generated settlement faction is `independent`.
