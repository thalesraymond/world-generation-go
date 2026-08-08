# ADR-0006: Demographics — Cellular-Automata Population Diffusion and Faction Spread

## Status

ACCEPTED

## Date

2026-08-08

## Context

After terrain and suitability exist, the world needs populations and cultural factions spread across the map so settlements emerge in inhabited, organized regions rather than empty wilderness. Requirements:

- Populations must concentrate where suitability is high.
- Factions must form contiguous blobs of influence, not salt-and-pepper noise.
- The process must be deterministic and require no per-iteration randomness beyond the initial seed.

## Decision

`internal/domain/demographics` (simulator.go) implements three phases:

1. **Pre-generation** — `PreGenerateSuitability` validates dimensions against the world state and stores the computed suitability layer.
2. **Seeding** — `SeedPopulationFromSuitability` sets `PopulationDensity[tile] = suitability²` and assigns a faction from the configured list (default `auric`, `verdant`, `cinder`) by the deterministic pattern `(x + y) % len(factions)`; tiles below `MinPopulation (0.05)` get no faction.
3. **Simulation** — `Simulate` runs `Iterations (8)` rounds of:
   - `diffusePopulation`: each tile keeps `1 − rate (0.3)` of its population and transfers the rest to neighboring tiles that are more suitable and less populated, weighted by neighbor suitability.
   - `spreadFactionInfluence`: each tile adopts the faction of the neighbor with the highest accumulated population influence; tiles below the minimum population become factionless.

All layers are plain `[]float64` / `[]string` slices aligned row-major with the world state grid; no RNG is consumed during diffusion or spread, so the automata are purely deterministic given the seeded input.

## Alternatives Considered

### Random-walk or agent-based population placement

- **Pros:** Can model individual migration.
- **Cons:** Consumes RNG per step, increasing determinism risk and computational cost for little observable benefit at this resolution.
- **Rejected:** Cellular automata give the same emergent patterns more cheaply and deterministically.

### Single-pass smoothing / blur

- **Pros:** Simple convolution.
- **Cons:** A blur does not enforce faction contiguity or population thresholds; it just smears values.
- **Rejected:** The two-step diffuse-and-spread automaton produces the intended borders and population fronts.

### Sparse sampling (evaluate a subset of tiles)

- **Pros:** Cheaper on large maps.
- **Cons:** Breaks neighborhood logic and contiguity guarantees.
- **Rejected:** Full-grid iteration is O(width×height×iterations), trivially fast for CLI-sized maps.

## Consequences

- Population density correlates strongly with suitability², producing clustered inhabited regions.
- Faction influence spreads outward into contiguous regions with natural borders, which later drives cross-faction settlement friction and rivalry.
- The diffusion logic depends on iteration order over `neighbors` (deterministic fixed order) and weighted-neighbor tie-breaking, so output is stable for a given seed.
- `SimulatorConfig` exposes iterations, diffusion rate, minimum population, and faction names; tuning changes settlement distribution and must be validated by determinism tests.
- `spreadFactionInfluence` reuses `state.PopulationDensity` (the *old* layer) for neighbor scoring while consuming the new population layer for thresholds — a subtle ordering dependency that is covered by tests.
