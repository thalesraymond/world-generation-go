# ADR-0004: Terrain Generation — Perlin Elevation, Latitude Climate, Threshold Biomes

## Status

ACCEPTED

## Date

2026-08-08

## Context

The world needs a landmass map before any spatial reasoning, demographics, or settlement placement can happen. Requirements from the initial concept:

- Deterministic: the same seed and dimensions must produce the same terrain.
- Structurally plausible: coastlines, temperature bands, and biomes should be coherent rather than random noise per tile.
- Cheap: full-map generation must run in a single pass suitable for a CLI invocation.

## Decision

`internal/domain/terrain` implements geographical genesis in three deterministic steps (generator.go, noise.go, map.go):

1. **Elevation** — `GenerateMap` seeds a `NoiseGenerator` from `config.TerrainRNG.Int64()` and samples 4-octave Perlin noise (`github.com/aquilax/go-perlin`), normalized from `[-1,1]` to `[0,1]` in `Sample`.
2. **Climate** — `BaseTemperatureForLatitude(y, height)` sets a baseline temperature that peaks at the equator and cools toward the poles; `AdjustTemperatureForElevation` subtracts `elevation * ElevationCooling (0.35)`. Humidity is an independent 4-octave noise layer seeded from `ClimateRNG` at scale 32 (elevation scale 48, persistence 2, octaves 4).
3. **Biome classification** — `DetermineBiome` applies ordered thresholds: water below `DefaultWaterThreshold (0.45)`; tundra when `temperature < 0.25`; desert when `temperature > 0.7 && humidity < 0.3`; forest when `humidity > 0.6`; otherwise grassland.

The map is stored row-major in `Map{Tiles []Tile}` with per-tile `Elevation`, `Temperature`, `Humidity`, `Biome`, and an `TileAt(x, y)` bounds-checked accessor. The third-party `go-perlin` dependency is wrapped behind the narrow `NoiseGenerator` API so only one package knows about it.

## Alternatives Considered

### Hand-authored heightmaps or seeded image assets

- **Pros:** Full artistic control.
- **Cons:** Finite content, not reproducible from a seed, no parametric variety, storage burden.
- **Rejected:** Must support arbitrary seeds and dimensions.

### Simplex or value noise

- **Pros:** Simplex is visually smoother in some respects.
- **Cons:** `go-perlin` is a small, proven dependency; value noise looks blockier. No product requirement forces a swap.
- **Rejected:** Perlin meets the need at acceptable complexity.

### Rule-based or ecosystem-model climate (e.g., rain shadows, prevailing winds)

- **Pros:** More physically plausible biomes.
- **Cons:** Significantly more complexity with no current consumer demanding it.
- **Deferred:** Latitude-plus-elevation temperature and independent humidity noise is the simplest model satisfying the OpenSpec geography requirements; richer climate is a future refinement.

### Per-tile fully independent random sampling

- **Pros:** Trivially simple.
- **Cons:** No spatial coherence; coastlines and biomes would be salt-and-pepper noise.
- **Rejected:** Fails the structural-plausibility requirement.

## Consequences

- Terrain is a pure deterministic function of `(seed, width, height, config)`; determinism is asserted by `generator_test.go` and `noise_test.go` with fixed seeds.
- Biomes are used by `spatial` (suitability), `demographics` (faction spread ignores water implicitly via population), `settlement` (placement), and `geography/pointcrawl` (friction table and node sampling).
- Classification thresholds and noise tuning constants are hardcoded defaults but overridable via `GeneratorConfig`, keeping behavior configurable without changing the pipeline.
- Relying on `go-perlin`'s fixed algorithm: upgrading the dependency must be verified against determinism tests before merge.
