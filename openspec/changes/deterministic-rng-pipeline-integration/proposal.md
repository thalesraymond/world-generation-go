## Why

The deterministic state engine now exists, but the generation and simulation pipeline does not yet consume it. To guarantee that the same seed yields the same generated world across terrain creation, demographic passes, and historical simulation, the downstream modules must accept isolated `math/rand/v2` instances and the simulation bootstrap must wire them consistently.

## What Changes

- Refactor terrain and climate generation code to accept injected `*rand.Rand` instances instead of relying on global randomness.
- Refactor demographic, settlement, and other simulation components to consume isolated component-specific PRNGs.
- Wire the deterministic state engine into simulation initialization so each subsystem receives its own reproducible random stream.
- Add integration and end-to-end tests that prove identical seeds produce identical generated results.

## Capabilities

### New Capabilities

- `deterministic-rng-integration`: Integrates isolated deterministic PRNG streams into world generation and simulation components.

### Modified Capabilities

- `terrain-generation`: Must consume injected deterministic PRNGs.
- `simulation-loop`: Must initialize and distribute component-specific PRNGs deterministically.
- `demographic-automata`: Must use isolated deterministic PRNGs for repeatable simulation behavior.
- `settlement-generation`: Must use isolated deterministic PRNGs for repeatable placement outcomes.

## Impact

- Depends on the terrain-generation, demographic, settlement, and simulation-loop modules existing in code.
- Changes generator and simulation package APIs to accept explicit PRNG dependencies.
- Adds regression coverage for deterministic end-to-end runs.
