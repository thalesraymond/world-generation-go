# ADR-0007: Settlement Generation — Suitability-Driven Placement with Agent State Initialization

## Status

ACCEPTED

## Date

2026-08-08

## Context

The world must contain named, classified settlements located in inhabitable regions, each carrying the agent state vector introduced by ADR-0001 (population, military strength, wealth, relations, goals) so the timeline simulation can drive them. Placement must be deterministic, avoid overcrowding, and reflect demographic reality.

## Decision

`internal/domain/settlement` (generator.go, types.go, conflict.go, names.go) places settlements from the world state's suitability and population layers:

1. **Candidate selection** — `findCandidates` keeps tiles with `Suitability ≥ 0.65` and `PopulationDensity ≥ 0.35`, scoring them `suitability × populationDensity`, then sorts deterministically by score descending.
2. **Distance filtering** — `filterByDistance` greedily accepts candidates at least `MinDistance (3)` apart (Euclidean), stopping at `MaxSettlements` when positive.
3. **Population and class** — `population = round(popDensity × MaxPopulation (100,000))`; `Classify` maps ≥50k → MajorCity, ≥10k → City, ≥1k → Village, else Abandoned.
4. **Agent state** — `MilitaryStrength = population × 0.1`, `Wealth = 100`, `Goals = agent.RandomGoals(rng)`, faction taken from the influence layer (or `independent`).
5. **Names** — generated from prefix+suffix tables with uniqueness enforced (`EnsureUniqueName` appends `-2`, `-3`, …).
6. **Conflict resolution** — `ResolveProximityConflicts` merges settlements closer than the merge distance, absorbing the smaller population into the larger and reclassifying the survivor.
7. **Relations** — after the final list is known, every settlement gets `world.InitRelations` (same-faction baseline +0.3), then `world.ApplyCrossFactionFriction` applies random negative friction between cross-faction, non-independent pairs so rivalries emerge.

The world pipeline (`worldgen.go`) then attaches founder figures per settlement using a derived `"figures:" + name` PRNG stream.

## Alternatives Considered

### Uniform grid placement

- **Pros:** Trivial and even spacing.
- **Cons:** Ignores suitability and demographics; places cities in deserts and mountains.
- **Rejected.**

### Poisson-disc or blue-noise sampling of random points

- **Pros:** Good spatial distribution.
- **Cons:** Decouples placement from suitability/population, losing the settlement-density correlation with the environment.
- **Rejected:** Candidates are already suitability-filtered, so greedy distance filtering suffices.

### Clustering (k-means / DBSCAN on population)

- **Pros:** Population-weighted centroids.
- **Cons:** Non-trivial to keep deterministic; overkill for the current resolution.
- **Rejected:** Simplest-that-satisfies approach preferred.

### Deterministic single settlement per region

- **Pros:** Fewer settlements.
- **Cons:** Produces sparse worlds with few agent actors to simulate.
- **Rejected:** Candidate scoring produces a healthy number of settlements on realistic maps.

## Consequences

- Settlement placement is a pure deterministic function of the suitability/population layers and the settlement RNG; `generator_test.go` asserts repeatability with fixed seeds.
- The post-merge relations initialization means every settlement has a complete relations map before simulation starts, enabling immediate diplomacy and conflict.
- Merge ordering is deterministic (sorted candidates, then sequential pairwise merge); merged settlements keep the larger's name and position.
- Tuning knobs (min suitability, min population, min distance, max population, merge distance) live in `settlement.Config`; changes shift settlement density and must be validated by determinism tests.
- Agent state (military, wealth, goals, relations) is initialized here, so `settlement` imports `agent` for goals — keeping the boundary at the domain layer.
