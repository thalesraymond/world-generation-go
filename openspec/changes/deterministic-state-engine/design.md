## Context

World generation relies on pseudo-random number generators (PRNG) to create varied but reproducible worlds based on a single seed. The current implementation relies on global or shared PRNG state (e.g., using `math/rand`). This leads to non-deterministic behavior because calls to the global PRNG from unrelated components or concurrent processes can alter the sequence of random numbers for the simulation, making the generated world impossible to reproduce exactly from the seed. `math/rand/v2` offers improved PRNG implementations (like PCG and ChaCha8) that are better suited for creating isolated random streams.

## Goals / Non-Goals

**Goals:**
- Provide a Deterministic State Management Engine that serves isolated, seed-derived PRNG instances to different components.
- Migrate to `math/rand/v2` for generating random sequences.
- Ensure strict seed segregation to prevent cross-component interference.
- Guarantee that the same global seed always produces the exact same generated world.

**Non-Goals:**
- We are not changing the core algorithms of world generation, only replacing the source of randomness they consume.
- This change does not focus on improving the performance of the world generation algorithms themselves.

## Decisions

1. **Use `math/rand/v2.New(math/rand/v2.NewPCG(seed1, seed2))`**: PCG is fast, robust, and explicitly supports initialization with two `uint64` seeds, making it ideal for creating derived PRNGs based on a master seed and a component ID.
2. **Deterministic Seed Derivation**: The State Engine will receive a master `uint64` seed. When a component requests a PRNG, the Engine will derive sub-seeds (e.g., by hashing the master seed with the component's unique string identifier) to guarantee that each component always gets the same unique, isolated PRNG stream.
3. **PRNG Injection**: Components will no longer fetch random numbers globally. Instead, they must accept a `*rand.Rand` (from `math/rand/v2`) instance via their constructors or function arguments.

## Risks / Trade-offs

- **Risk**: Refactoring existing code to accept PRNG instances could be invasive.
  - **Mitigation**: We will perform this change incrementally or create a context-based injection mechanism if the plumbing is too deep, though explicit parameter passing is preferred for clarity.
- **Risk**: Test failures across the suite due to changed random streams.
  - **Mitigation**: Tests that rely on specific random outcomes will need their expected values updated, or they should be rewritten to be less brittle (e.g., testing statistical properties instead of exact sequences where appropriate).
