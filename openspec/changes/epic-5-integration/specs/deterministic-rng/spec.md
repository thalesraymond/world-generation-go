## ADDED Requirements

### Requirement: Deterministic RNG Audit for Agent Subsystems
The deterministic RNG system SHALL be audited to verify that all agent subsystems from Epics 1–4 use derived RNGs from the master seed and do not use package-level random state.

#### Scenario: No package-level random state in agent systems
- **WHEN** the agent subsystems (settlement agents, character executors, faction agents, artifacts) are audited
- **THEN** no agent subsystem SHALL import or use `math/rand` or `crypto/rand` directly
- **THEN** all agent subsystems SHALL use RNG instances derived from the master seed via the deterministic RNG engine

#### Scenario: Component-scoped RNG usage
- **WHEN** each agent subsystem requests an RNG instance
- **THEN** the instance SHALL be derived using a unique component identifier
- **THEN** the derivation SHALL be deterministic and reproducible across process runs

### Requirement: Integrated Determinism Regression Test
The deterministic RNG system SHALL include a regression test verifying that the full integrated pipeline produces byte-identical output when run with the same seed.

#### Scenario: Full pipeline determinism
- **WHEN** the full init → simulate → export pipeline is run twice with the same master seed
- **THEN** all generated export files SHALL be byte-identical between runs
- **THEN** the timeline event stream SHALL be byte-identical between runs
