## ADDED Requirements

### Requirement: Deterministic Combinatorial Name Generation

The system SHALL generate unique settlement names by combining a prefix and a suffix drawn from deterministic tables, using the settlement's component RNG for reproducibility.

#### Scenario: Name is generated from prefix and suffix

- **WHEN** a settlement is placed during world generation
- **THEN** its Name is formed by concatenating a prefix and a suffix drawn from fixed tables using the settlements RNG

#### Scenario: Same seed produces same names

- **WHEN** the same master seed is used across two generation runs
- **THEN** each settlement receives the identical name in both runs

### Requirement: Name Uniqueness Enforcement

The system SHALL ensure that no two settlements receive the same name by appending a deterministic numeric suffix when a collision is detected.

#### Scenario: Collision produces suffixed name

- **WHEN** a generated name matches an already-assigned settlement name
- **THEN** a numeric suffix (`-2`, `-3`, etc.) is appended to the name

#### Scenario: No collision produces plain name

- **WHEN** a generated name does not match any existing settlement name
- **THEN** no suffix is appended