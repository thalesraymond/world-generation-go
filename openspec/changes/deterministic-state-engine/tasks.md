## 1. Engine Core Setup

- [ ] 1.1 Create the `state` package (or similar) to house the deterministic engine.
- [ ] 1.2 Implement the `Engine` struct that holds the master seed.
- [ ] 1.3 Implement the `NewEngine(masterSeed uint64) *Engine` constructor.

## 2. Seed Derivation and PRNG Injection

- [ ] 2.1 Implement deterministic seed derivation (e.g., using a hash of the component identifier combined with the master seed).
- [ ] 2.2 Implement `Engine.GetPRNG(componentID string) *rand.Rand` utilizing `math/rand/v2.NewPCG`.

## 3. Core Engine Testing

- [ ] 3.1 Write unit tests to verify that identical master seeds yield identical PRNG sequences for the same component ID.
- [ ] 3.2 Write unit tests to verify that different component IDs yield independent PRNG sequences.

## 4. Component Refactoring

- [ ] 4.1 Update terrain generation module to accept a `*rand.Rand` instance instead of using a global PRNG.
- [ ] 4.2 Update weather generation module to accept a `*rand.Rand` instance.
- [ ] 4.3 Update entity generation or any other simulation components to accept a `*rand.Rand` instance.
- [ ] 4.4 Wire up the `Engine` in the main simulation initialization to pass the specific PRNGs to each component.

## 5. Integration and Test Fixes

- [ ] 5.1 Update existing tests across the codebase to supply isolated PRNG instances.
- [ ] 5.2 Verify that the end-to-end simulation generates the exact same world when run multiple times with the same initial seed.
