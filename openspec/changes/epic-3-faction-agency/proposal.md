## Why

Factions are currently plain strings (`"Ashfield"`, `"Coldcrest"`), carrying no identity, leadership, resources, or strategic behavior. This prevents the system from generating causal history — wars have no factions behind them, alliances are meaningless, and there is no mechanism for settlements to shift allegiance. To realize the "zero-player strategy game" vision from the ADR, factions must become strategic agents with goals, resources, and decision-making capability. This is Epic 3 in the five-epic roadmap defined in ADR-0001.

## What Changes

- Create `internal/domain/faction/` package with a `Faction` domain entity containing identity, leadership, treasury, strategic goals, member tracking, and inter-faction relations.
- Implement `simulation.Entity` interface for factions, adding a strategic-level decision loop that runs each simulation year.
- Add strategic actions: declare war, form alliance, set policy — each with preconditions, consequences, and event emission.
- Enable dynamic faction membership: settlements can join or leave factions through conquest or diplomatic defection.
- Replace `world.State.FactionInfluence []string` grid with a faction entity registry; update `Settlement.Faction` from a string to a reference.
- Update Obsidian exporter to display faction dynamics: membership changes, strategic decisions, alliance maps, and faction history.

## Capabilities

### New Capabilities
- `faction-entity`: Domain model for factions — identity (cultural/ethnic grouping), leadership (figures with faction-wide authority), treasury (shared resources), strategic goals, member settlement tracking, and inter-faction relations.
- `faction-agency`: Strategic-level agent behavior — annual decision loop, war declaration, alliance formation, policy setting (expansion/defense/diplomacy), event emission, and deterministic RNG integration.
- `faction-dynamics`: Dynamic faction membership — settlement faction switching through conquest or diplomatic defection, faction collapse when empty, new faction formation from breakaway settlements, and history recording of membership changes.

### Modified Capabilities
- `settlement-generation`: Settlement faction assignment changes from copying a string from the `FactionInfluence` grid to referencing a registered `Faction` entity. Default faction for isolated settlements changes from `"independent"` to an implicit unaffiliated state.
- `obsidian-export`: Faction pages gain dynamic content — membership timeline, strategic decision log, alliance relationships, and faction identity metadata.
- `world-state`: `FactionInfluence []string` grid is replaced with a `Factions map[string]*faction.Faction` registry. World state JSON serialization gains faction entity data.

## Impact

- **New package**: `internal/domain/faction/` — all new domain types and agent logic.
- **Modified packages**: `internal/domain/world/` (state struct), `internal/domain/settlement/` (generator, types), `internal/usecase/simulation/` (worldgen pipeline, runner), `internal/infra/exporter/` (faction pages), `cmd/` (if CLI flags reference faction config).
- **JSON schema change**: `world.State` serialization changes — `factionInfluence` grid replaced by `factions` map. This is a **BREAKING** change for any persisted world state files.
- **Simulation pipeline**: Faction entities must be registered with the `simulation.Simulation` engine alongside settlement entities (Epic 1 assumption).
- **Test coverage**: New tests for faction entity, agent decisions, membership changes, and determinism. Integration test for full init→simulate→export with faction dynamics.
