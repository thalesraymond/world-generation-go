# ADR-0003: Deterministic State Engine — Component-Scoped PRNG Derivation

## Status

ACCEPTED

## Date

2026-08-08

## Context

Determinism is a hard product requirement: identical seeds must produce byte-identical worlds, timelines, and exports. The initial concept (`openspec/specs/initial_concept.md`) and `AGENTS.md` both mandate that the RNG be component-scoped and derived from a master seed, and that no package-level random state ever be used.

Each generation subsystem needs its own random stream so that changing one subsystem (or adding RNG consumers to it) cannot perturb every downstream subsystem's output. With a single shared RNG, any addition would silently re-sequence all later draws and break byte-identical reproducibility across versions.

## Decision

`internal/domain/state.Engine` (engine.go) is the single entry point for RNG creation. Current implementation:

- `NewEngine(masterSeed uint64)` records the master seed.
- `GetPRNG(componentID string) *randv2.Rand` derives two 64-bit seeds and returns a fresh `math/rand/v2` PCG stream:
  - `seed1 = FNV-1a(masterSeed ‖ componentID ‖ "stream")`
  - `seed2 = FNV-1a(masterSeed ‖ componentID ‖ "sequence")`
  - `randv2.New(randv2.NewPCG(seed1, seed2))`
- `deriveSeed` uses `hash/fnv` (FNV-1a 64-bit) over the little-endian master seed bytes, a `0x00` separator, the component ID bytes, a second `0x00` separator, and a lane byte-string.

The generation pipeline (`internal/usecase/simulation/worldgen.go`) requests one stream per subsystem: `"terrain"`, `"climate"`, `"demographics"`, `"settlements"`, `"pointcrawl"`, plus a per-settlement stream `"figures:" + settlementName`. Per-settlement figures therefore never share a stream with each other or with terrain.

## Alternatives Considered

### One global PRNG for the whole run

- **Pros:** Simplest possible implementation.
- **Cons:** Any code change that adds or removes a draw re-sequences every downstream subsystem, invalidating reproducibility guarantees for prior seeds and making determinism tests brittle.
- **Rejected:** Violates the component-isolation requirement.

### Persist per-component seeds in the world state

- **Pros:** Allows checkpointing and resuming.
- **Cons:** Adds statefulness and serialization burden; makes re-derivation impossible without stored seeds; complicates the "identical seed → identical world" contract.
- **Rejected:** Derivation from the master seed keeps the contract closed and cheap.

### Hash output chaining from a single derived seed (seeded `rand.New(seed)`)

- **Pros:** Uses only one hash, slightly cheaper.
- **Cons:** A single 64-bit seed for `math/rand` PCG leaves less stream diversity; two lanes ("stream"/"sequence") plus PCG's two state words give stronger isolation and map cleanly onto `NewPCG`'s constructor.
- **Rejected:** The two-lane derivation costs little and is more robust.

### Third-party RNG library or CSPRNG

- **Pros:** Could offer stronger statistical quality.
- **Cons:** Adds a dependency for no product need; `math/rand/v2` PCG is deterministic, well-understood, and already used across the codebase.
- **Rejected:** Determinism and reproducibility are the requirements, not cryptographic randomness.

## Consequences

- Identical seeds reproduce identical outputs across the whole pipeline, verified by determinism tests in every generator package.
- Adding a new subsystem requires only a new component ID string; existing streams are untouched.
- **Breaking constraint:** Changing the `deriveSeed` hashing scheme or the lane byte-strings would change every derived seed and break byte-identical output for all existing seeds. This is now frozen behavior.
- Component IDs are part of the public contract (`worldgen.go` enumerates them); renaming a component ID silently re-seeds that subsystem.
- `math/rand/v2` PCG is a stable standard-library primitive; the wrapper keeps the derivation isolated in one place.
