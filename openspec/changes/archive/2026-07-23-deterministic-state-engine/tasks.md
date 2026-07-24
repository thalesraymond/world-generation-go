## 1. Engine Core Setup

- [x] 1.1 Create the `state` package (or similar) to house the deterministic engine.
- [x] 1.2 Implement the `Engine` struct that holds the master seed.
- [x] 1.3 Implement the `NewEngine(masterSeed uint64) *Engine` constructor.

## 2. Seed Derivation and PRNG Injection

- [x] 2.1 Implement deterministic seed derivation (e.g., using a hash of the component identifier combined with the master seed).
- [x] 2.2 Implement `Engine.GetPRNG(componentID string) *rand.Rand` utilizing `math/rand/v2.NewPCG`.

## 3. Core Engine Testing

- [x] 3.1 Write unit tests to verify that identical master seeds yield identical PRNG sequences for the same component ID.
- [x] 3.2 Write unit tests to verify that different component IDs yield independent PRNG sequences.
