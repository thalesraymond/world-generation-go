## Purpose

Define the baseline deterministic random number generation behavior for component-scoped simulation PRNGs.

## Requirements

### Requirement: Master Seed Initialization

The deterministic state management engine SHALL accept a primary master seed during initialization.

#### Scenario: Engine initialization

- **WHEN** the engine is instantiated with a specific 64-bit seed
- **THEN** it internally records this master seed for deriving component-specific PRNGs

### Requirement: Deterministic Component PRNG Derivation

The engine SHALL provide isolated, reproducible PRNG instances for requested component identifiers.

#### Scenario: Requesting a PRNG for a specific component

- **WHEN** a component requests a PRNG instance using a unique identifier (e.g., "terrain", "weather")
- **THEN** the engine returns a `math/rand/v2` PRNG initialized with a seed deterministically derived from the master seed and the identifier

### Requirement: Stable Two-Lane Derivation

The engine SHALL derive both initialization values required by the PCG generator from the master seed, component identifier, and distinct fixed derivation lanes. The derivation SHALL be stable across process runs and MUST NOT depend on iteration order or package-global random state.

#### Scenario: Distinct component streams

- **WHEN** two different component identifiers are requested from an engine with the same master seed
- **THEN** each receives a distinct deterministic PCG stream.

### Requirement: Strict Seed Segregation

The random number sequences generated for one component SHALL NOT be affected by the PRNG usage of any other component.

#### Scenario: Interleaved component generation

- **WHEN** two separate components draw random numbers from their respective PRNGs
- **THEN** the sequence of numbers drawn by each component is identical to the sequence it would draw if the other component had not drawn any numbers

### Requirement: Reproducibility Guarantee

Identical initial master seeds MUST always produce identical PRNG instances for the same component identifiers.

#### Scenario: Re-running simulations

- **WHEN** the engine is initialized twice in separate process instances with the exact same master seed
- **THEN** the PRNG instance requested for "terrain" in both runs produces the exact same sequence of random numbers

### Requirement: Engine Contract Boundary

This specification SHALL guarantee determinism for the state engine and every component that receives one of its derived PRNGs. End-to-end pipeline injection is separately specified by the active deterministic-RNG-pipeline integration change and is not implied by the engine API alone.

#### Scenario: Isolated engine verification

- **WHEN** only the deterministic state engine is exercised with a fixed seed and component identifier
- **THEN** its sequence is reproducible even if unrelated generation components are not yet wired to it.

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
