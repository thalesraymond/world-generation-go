## 1. Terrain and Climate Integration

- [ ] 1.1 Update terrain generation modules to accept injected `*rand.Rand` instances.
- [ ] 1.2 Update climate or weather-related generation modules to accept injected `*rand.Rand` instances.
- [ ] 1.3 Add or update tests for terrain and climate generators to use isolated PRNG instances.

## 2. Demographic and Simulation Integration

- [ ] 2.1 Update demographic, settlement, and other simulation components to accept injected `*rand.Rand` instances.
- [ ] 2.2 Wire the deterministic `Engine` into simulation initialization so each component receives a stable component-specific PRNG.
- [ ] 2.3 Update existing tests across the affected generation and simulation packages to supply isolated PRNG instances.

## 3. End-to-End Determinism Verification

- [ ] 3.1 Add an integration test that runs the same seed multiple times and asserts identical world output or serialized state.
- [ ] 3.2 Add coverage for differing component identifiers or subsystem seeds to confirm streams remain isolated.