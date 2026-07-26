# settlement-generation Specification

## Purpose

Define how discrete settlements are founded from demographic and suitability state in a deterministic world generation pass, including the generation of founding historical figures and settlement-scoped figure RNG.

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

Generated settlements SHALL be named using deterministic combinatorial generation drawing from prefix and suffix tables. A settlement SHALL be assigned a type based on population thresholds. A settlement SHALL inherit the source tile's faction; a tile without faction influence SHALL produce an `independent` settlement. At creation time, 3–5 founding historical figures SHALL be generated for the settlement.

#### Scenario: Named settlement generation

- **WHEN** a settlement is founded
- **THEN** it receives a generated name composed from combinatorial name tables using the settlements RNG

#### Scenario: Typed settlement generation

- **WHEN** a settlement is founded with a population value
- **THEN** it receives a type classification appropriate to its population

#### Scenario: Unclaimed settlement

- **WHEN** an eligible source tile has no faction influence
- **THEN** the generated settlement faction is `independent`

#### Scenario: Founding figures generated

- **WHEN** a settlement is founded during world generation
- **THEN** between 3 and 5 historical figures SHALL be generated as founders
- **THEN** one founder SHALL be assigned the Leader role
- **THEN** the remaining founders SHALL be assigned Explorer or no role
- **THEN** founders SHALL have the settlement's founding year as their birth year (or an offset of up to 20 years prior for older founders)

### Requirement: Settlement Figure Generation RNG

Settlement generation SHALL derive and store a figure-specific RNG for each settlement.

#### Scenario: Figure RNG derivation

- **WHEN** a settlement is generated
- **THEN** a figure-specific `*randv2.Rand` SHALL be derived from the master seed using the settlement name as part of the component identifier
- **THEN** the RNG SHALL be stored or derivable for use during simulation figure lifecycle processing
