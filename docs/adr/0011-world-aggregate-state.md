# ADR-0011: World Aggregate State — Grid Layers, Settlements, Relations, Serialization

## Status

ACCEPTED

## Date

2026-08-08

## Context

Every subsystem produces a piece of the world, and the pipeline (`GenerateWorld`) must assemble them into one coherent, persistable object that downstream consumers (timeline simulation, exporter, tests) can read. The state must guarantee that grid-backed layers stay aligned, that settlements carry their agent state and figures, and that the whole thing round-trips through JSON without corruption.

## Decision

`internal/domain/world` is the aggregate root:

- `State` (state.go) holds `Width`, `Height`, and three row-major layers of equal length — `PopulationDensity []float64`, `FactionInfluence []string`, `Suitability []float64` — plus `Settlements []Settlement` and an optional `PointcrawlGraph *pointcrawl.Graph`.
- `NewState(width, height)` zero-initializes all layers; `Index(x, y)` converts coordinates with bounds checks; `CellCount()` returns `width×height` (0 for invalid dimensions).
- `SetSuitability` validates the layer length before storing a defensive copy.
- `Validate` asserts all three layers match `CellCount`, guarding serialization and every downstream consumer.
- `ToJSON` / `FromJSON` marshal/unmarshal with `Validate` on both paths, making JSON the canonical persistence format (`world_state.json`).
- `Settlement` embeds agent state from ADR-0001: `Name`, `Type`, `X`, `Y`, `Faction`, `Population`, `Figures []figures.HistoricalFigure`, `MilitaryStrength`, `Wealth`, `Relations map[string]float64`, `Goals []string`.

### Relations subsystem (relations.go)

- `InitRelations` builds each settlement's baseline map: same non-independent faction pairs start at `+0.3`, all others at `0.0`, self excluded.
- `ShiftRelations` applies a signed delta toward a target, clamping to `[−1.0, +1.0]`.
- `ApplyCrossFactionFriction` / `ApplySettlementCrossFactionFriction` inject random negative friction (`−rng.Float64() × 0.6`) between cross-faction, non-independent pairs so rivalries emerge naturally.
- Shift constants per action: raid success −0.4/−0.3, raid failure −0.2, conquest −0.8, ally +0.4, prosper +0.05.

## Alternatives Considered

### Separate per-subsystem state stores stitched only at export time

- **Pros:** No single aggregate type.
- **Cons:** No invariant enforcement (layer alignment), no single validation point, harder determinism testing.
- **Rejected:** One validated aggregate makes the pipeline and tests tractable.

### Relational database or protobuf persistence

- **Pros:** Schema evolution and tooling.
- **Cons:** Heavyweight for a CLI that saves one JSON file; JSON is human-inspectable and diffable.
- **Rejected:** JSON meets the product need with zero infrastructure.

### Relations as an external matrix rather than per-settlement maps

- **Pros:** O(1) lookup.
- **Cons:** Sparse and brittle when settlements are added mid-simulation (expansion founds new settlements); per-name maps grow naturally.
- **Rejected:** Map-based relations adapt to dynamic settlement counts.

### Copying layers on every write vs. validating at the boundary

- **Pros:** Copy-on-write guarantees immutability.
- **Cons:** Costs allocations in hot simulation paths.
- **Rejected:** Validation at the aggregate boundary (single write points) is sufficient and cheap.

## Consequences

- All consumers depend on one validated aggregate; `state_test.go` and `relations_test.go` assert layer alignment, round-trips, and relation clamping.
- `Settlement` is the shared data structure across generation, agents, figures, and export — its JSON schema is a de-facto public contract.
- Cross-faction friction is deliberately random-but-seeded (consumes the settlement RNG stream), so rivalries emerge deterministically per seed.
- Relations maps are the substrate for the agent decision loop (raid/conquer/ally preconditions) and for the exporter's Allies/Rivals sections.
- `world` imports `figures` and `pointcrawl`; it sits at the domain layer and imports no infrastructure, keeping the aggregate pure.
