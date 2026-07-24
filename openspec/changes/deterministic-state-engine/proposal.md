## Why

World generation requires deterministic outcomes based on an initial seed to ensure that the same seed always produces the exact same world, no matter when or where the simulation runs. To achieve this reliably, we need to ensure strict seed segregation using `math/rand/v2` so that random generation across different components is reproducible and isolated from global state changes.

## What Changes

- Introduce a deterministic state management engine that manages seeds and PRNG instances for different simulation components.
- Implement component-specific PRNG derivation using `math/rand/v2`.
- Ensure strict seed segregation so that one component's PRNG calls do not affect the random stream of another component.
- Establish the reusable RNG foundation that later generation and simulation changes will integrate.

## Capabilities

### New Capabilities

- `deterministic-rng`: Provides reproducible, strictly segregated pseudo-random number generators based on `math/rand/v2` for different simulation components.

### Modified Capabilities

## Impact

- Adds a reusable deterministic RNG engine in the core domain layer.
- Provides the seed-derivation and PRNG-construction contract that downstream generation and simulation changes will consume.
- Defers component refactoring and end-to-end determinism verification to a follow-up integration change.
