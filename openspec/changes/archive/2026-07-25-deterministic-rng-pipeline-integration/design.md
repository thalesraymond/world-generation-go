## Context

The deterministic state engine provides reproducible component-specific PRNGs, but that alone does not make the application reproducible. Determinism only reaches the user-visible world once the terrain, demographic, settlement, and simulation orchestration layers all accept injected PRNGs and derive them from the same master seed through the engine.

## Goals / Non-Goals

**Goals:**

- Refactor generation and simulation modules to accept `*rand.Rand` instances from `math/rand/v2`.
- Assign a stable component identifier to each subsystem so its PRNG stream remains isolated.
- Wire the deterministic engine into the simulation bootstrap and use it to derive per-component PRNGs.
- Add integration coverage proving same-seed runs produce identical results.

**Non-Goals:**

- Replacing or redesigning terrain, demographic, or simulation algorithms beyond the randomness plumbing they require.
- Optimizing performance or introducing concurrency changes unrelated to deterministic RNG injection.

## Decisions

1. **Explicit PRNG injection**: Each generator or simulation component should accept its own `*rand.Rand` dependency through a constructor or function parameter instead of reading shared global state.
2. **Stable component identifiers**: The bootstrap layer will derive PRNGs using fixed identifiers such as `terrain`, `climate`, `demographics`, `settlements`, and `timeline` so streams remain stable across refactors.
3. **Bootstrap-owned wiring**: The simulation initialization path is responsible for creating the deterministic engine from the master seed and distributing all component-specific PRNGs.
4. **End-to-end regression checks**: Once the pipeline exists, integration tests should compare complete world outputs or serialized snapshots across repeated runs with the same seed.

## Risks / Trade-offs

- **Risk**: Upstream simulation modules may land with APIs that make PRNG injection invasive.
  - **Mitigation**: Keep identifiers and RNG ownership explicit at package boundaries as those modules are introduced.
- **Risk**: Determinism tests may become brittle if they assert unstable formatting details.
  - **Mitigation**: Compare normalized world state snapshots or other stable serialized artifacts instead of incidental console formatting.
