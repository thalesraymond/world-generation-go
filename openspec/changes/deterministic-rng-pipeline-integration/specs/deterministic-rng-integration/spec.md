## ADDED Requirements

### Requirement: Deterministic PRNG Injection Across Generation Components

Generation and simulation components SHALL accept isolated `math/rand/v2` PRNG instances instead of relying on shared global random state.

#### Scenario: Terrain generation uses an injected PRNG

- **WHEN** the terrain generation module is initialized for a simulation run
- **THEN** it receives a component-specific `*rand.Rand` derived from the deterministic state engine

#### Scenario: Demographic generation uses an injected PRNG

- **WHEN** demographic or settlement generation executes
- **THEN** it uses its own injected `*rand.Rand` and does not share random state with terrain or other components

### Requirement: Deterministic Simulation Bootstrap

The simulation initialization path SHALL construct the deterministic state engine from the master seed and distribute reproducible PRNG instances to each subsystem.

#### Scenario: Bootstrap derives per-component streams

- **WHEN** a simulation run starts with a specific master seed
- **THEN** the bootstrap layer derives stable PRNG streams for each registered subsystem using fixed component identifiers

### Requirement: End-to-End Seed Reproducibility

Once the generation and simulation pipeline is wired to the deterministic state engine, repeated runs with the same seed MUST produce identical world output.

#### Scenario: Re-running the same seed

- **WHEN** the full world generation pipeline is executed multiple times with the same master seed and equivalent configuration
- **THEN** the resulting world state or exported deterministic snapshot is identical across runs
