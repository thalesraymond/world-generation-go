# ADR-0005: Spatial Suitability — Weighted Tile Scoring

## Status

ACCEPTED

## Date

2026-08-08

## Context

Settlement placement, population seeding, and agent expansion all need a single, comparable measure of "how good is this tile to live on" derived from terrain. The measure must be deterministic, a pure function of the terrain map, and cheap enough to compute for every tile.

## Decision

`internal/domain/spatial` (suitability.go) defines one scoring function and a precomputation pass:

- `EvaluateTileSuitability(tile, nearWater, elevationVariance)` returns a score in `[0,1]`:
  - Water tiles score 0.
  - **Water proximity (weight 0.4):** 1.0 if any water tile is within radius 2, else 0.2.
  - **Flatness (weight 0.3):** `1 − elevationVariance × 3`, where variance is the local max−min elevation across the 3×3 neighborhood.
  - **Biome livability (weight 0.3):** grassland 1.0, forest 0.85, tundra 0.25, desert 0.1, otherwise 0.
  - Multiplicative **height penalty:** `1 − max(0, elevation − 0.85) × 6`, penalizing very high tiles regardless of other factors.
- `CalculateSuitabilityMap(terrainMap)` precomputes a per-tile score slice once, using `hasNearbyWater(radius=2)` and `localElevationVariance(3×3)`.

The scores are stored on the world state via `demographics.PreGenerateSuitability`, then consumed by population seeding, settlement candidate selection, and the agent `AgentEnv.Suitability` lookups during simulation.

## Alternatives Considered

### Score on demand per tile

- **Pros:** No up-front pass.
- **Cons:** The 3×3 variance and radius-2 water checks are recomputed repeatedly across consumers and every simulation tick; the suitability layer is requested by the entire pipeline.
- **Rejected:** Precomputation is a single O(width×height) pass and the layer is reused many times.

### Machine-learned or example-driven suitability

- **Pros:** Could encode richer judgment.
- **Cons:** Non-deterministic, non-transparent, and against the algorithmic-only constraint.
- **Rejected.**

### Multiple separate rules scattered in each consumer

- **Pros:** No shared abstraction.
- **Cons:** Weighting and biome-livability tables would drift between settlement placement, demographics, and agents.
- **Rejected:** A single source of truth keeps behavior consistent and testable.

## Consequences

- Suitability is a pure function of terrain, deterministic and testable in isolation (`suitability_test.go`).
- Demographics seed population as `suitability²`, concentrating settlements on high-suitability tiles near water on flat ground — the intended settlement pattern.
- The weight constants (0.4/0.3/0.3, water score 0.2, threshold 0.85, penalty 6) are hardcoded tuning values; changing them changes settlement placement and must be covered by determinism tests.
- Suitability is exported as part of the world state JSON, making it observable in exports and tests.
