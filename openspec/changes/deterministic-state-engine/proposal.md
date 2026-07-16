## Why

World generation requires deterministic outcomes based on an initial seed to ensure that the same seed always produces the exact same world, no matter when or where the simulation runs. To achieve this reliably, we need to ensure strict seed segregation using `math/rand/v2` so that random generation across different components is reproducible and isolated from global state changes.

## What Changes

- Switch from `math/rand` (or any other PRNG) to `math/rand/v2` for random number generation.
- Implement a deterministic state management engine that manages seeds and PRNG instances for different simulation components.
- Ensure strict seed segregation so that one component's PRNG calls do not affect the random stream of another component.
- Remove any reliance on global PRNG state.

## Capabilities

### New Capabilities
- `deterministic-rng`: Provides reproducible, strictly segregated pseudo-random number generators based on `math/rand/v2` for different simulation components.

### Modified Capabilities

## Impact

- All world generation algorithms will need to be updated to accept and use isolated PRNG instances instead of a global generator.
- Simulation components will receive their initial state/seed from the new Deterministic State Management Engine.
- Requires updates to any existing tests expecting specific random outcomes or relying on global random state.
