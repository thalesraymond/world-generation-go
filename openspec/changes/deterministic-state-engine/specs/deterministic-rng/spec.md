## ADDED Requirements

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
