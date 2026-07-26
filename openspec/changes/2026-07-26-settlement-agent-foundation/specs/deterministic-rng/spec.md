# deterministic-rng Specification

## Purpose

Define the baseline deterministic random number generation behavior for component-scoped simulation PRNGs, including agent decision RNG.

## MODIFIED Requirements

### Requirement: Agent RNG Component Identifier

The deterministic state engine SHALL derive agent-specific PRNG instances using a settlement-scoped component identifier.

#### Scenario: Agent RNG derivation

- **WHEN** a settlement is initialized for simulation
- **THEN** an agent RNG SHALL be derived via `engine.GetPRNG("agent:" + settlement.Name)`
- **THEN** the agent RNG SHALL be separate from the figure RNG (`"figures:" + settlement.Name`)

#### Scenario: Agent RNG usage

- **WHEN** agent operations execute (action selection, relation shifts, expand target selection)
- **THEN** they SHALL use the settlement's agent RNG exclusively
- **THEN** figure operations SHALL NOT use the agent RNG
- **THEN** agent operations SHALL NOT use the figure RNG

### Requirement: Agent RNG Determinism

Identical initial master seeds MUST always produce identical agent RNG instances for the same settlement.

#### Scenario: Agent decision reproducibility

- **WHEN** the full simulation is executed multiple times with the same master seed
- **THEN** the agent RNG for each settlement SHALL produce the exact same sequence
- **THEN** agent decisions (actions chosen, relation shifts, expand targets) SHALL be identical across runs
