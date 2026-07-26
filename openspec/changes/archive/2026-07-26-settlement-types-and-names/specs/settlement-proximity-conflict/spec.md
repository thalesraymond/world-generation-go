## ADDED Requirements

### Requirement: Proximity-Based Merge Resolution

The system SHALL detect pairs of settlements whose Euclidean distance is below a configured merge threshold during world generation, and SHALL merge the smaller settlement into the larger one.

#### Scenario: Close settlements trigger merge

- **WHEN** two settlements are closer than the configured merge distance threshold
- **THEN** the settlement with the larger population absorbs the smaller settlement's population
- **AND** the smaller settlement is removed from the settlement list

#### Scenario: Identical population tie-breaking

- **WHEN** two merging settlements have equal populations
- **THEN** the first-placed settlement (lower index) survives and absorbs the second

#### Scenario: Reclassification after merge

- **WHEN** a settlement absorbs another settlement's population
- **THEN** its type is re-evaluated based on the combined population

### Requirement: Merge Distance Configuration

The merge distance threshold SHALL default to the settlement generator's minimum spacing distance, ensuring only settlements that slipped through spacing constraints are merged.

#### Scenario: Default merge threshold

- **WHEN** no explicit merge distance is configured
- **THEN** the merge distance equals the settlement generator's MinDistance value