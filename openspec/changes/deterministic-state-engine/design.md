## Context

World generation relies on pseudo-random number generators (PRNG) to create varied but reproducible worlds based on a single seed. The current implementation relies on global or shared PRNG state (e.g., using `math/rand`). This leads to non-deterministic behavior because calls to the global PRNG from unrelated components or concurrent processes can alter the sequence of random numbers for the simulation, making the generated world impossible to reproduce exactly from the seed. `math/rand/v2` offers improved PRNG implementations (like PCG and ChaCha8) that are better suited for creating isolated random streams.

## Goals / Non-Goals

**Goals:**
- Provide a Deterministic State Management Engine that serves isolated, seed-derived PRNG instances to different components.
- Migrate to `math/rand/v2` for generating random sequences.
- Ensure strict seed segregation to prevent cross-component interference.
- Guarantee that the same master seed always reproduces the same component-specific PRNG stream.

**Non-Goals:**
- We are not refactoring terrain, climate, demographic, or simulation modules in this change.
- We are not yet proving full end-to-end world reproducibility in the CLI pipeline.
- This change does not focus on improving the performance of world generation algorithms.

## Decisions

1. **Use `math/rand/v2.New(math/rand/v2.NewPCG(seed1, seed2))`**: PCG is fast, robust, and explicitly supports initialization with two `uint64` seeds, making it ideal for creating derived PRNGs based on a master seed and a component ID.
2. **Deterministic Seed Derivation**: The State Engine will receive a master `uint64` seed. When a component requests a PRNG, the Engine will derive sub-seeds (e.g., by hashing the master seed with the component's unique string identifier) to guarantee that each component always gets the same unique, isolated PRNG stream.
3. **Deferred PRNG Injection**: Downstream components will eventually accept a `*rand.Rand` (from `math/rand/v2`) instance via constructors or function arguments, but that refactor is tracked separately once those modules exist.

## Risks / Trade-offs

- **Risk**: The broader simulation pipeline is not implemented yet, so integration tasks could block this change.
  - **Mitigation**: Keep this change scoped to the engine contract and move module refactors into a follow-up integration change.
- **Risk**: Test failures across the suite due to changed random streams.
  - **Mitigation**: Limit this change to new engine-level tests; update broader integration tests only when the dependent modules land.
