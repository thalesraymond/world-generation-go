# settlement-generation Specification

## Purpose

Define how discrete settlements are founded from demographic and suitability state, including agent state initialization.

## MODIFIED Requirements

### Requirement: Agent State Initialization at Settlement Creation

Generated settlements SHALL be initialized with agent state vector values.

#### Scenario: Military strength initialization

- **WHEN** a settlement is founded
- **THEN** MilitaryStrength SHALL be initialized to population × 0.1

#### Scenario: Wealth initialization

- **WHEN** a settlement is founded
- **THEN** Wealth SHALL be initialized to 100.0 (or config value if provided)

#### Scenario: Relations initialization

- **WHEN** a settlement is founded
- **THEN** Relations SHALL be initialized via `initRelations()` with same-faction baseline +0.3

#### Scenario: Goals randomization

- **WHEN** a settlement is founded
- **THEN** Goals SHALL be randomized to 2–3 unique values from ["grow", "defend", "expand"]
- **THEN** randomization SHALL use the settlement's figure RNG (no separate agent RNG needed at generation time)
- **THEN** same seed SHALL produce identical goals across runs
