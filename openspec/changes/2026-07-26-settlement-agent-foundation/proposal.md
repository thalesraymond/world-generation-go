## Why

The current simulation produces disconnected random events without causal agency. Settlements passively experience "a raid occurred" or "a festival is held" with no decision-making entity driving these outcomes. This lacks the strategic depth and emergent narrative called for in the initial concept (Dwarf Fortress + Caves of Qud inspiration). Transforming settlements into active agents with state vectors, goals, and decision loops makes history a log of a zero-player strategy game — where events emerge from agent choices rather than random templates.

## What Changes

- **New**: Settlement agent state vector (MilitaryStrength, Wealth, Relations map, Goals) with decision loop logic
- **New**: Six core agent actions (Expand, Raid, Conquer, Fortify, Ally, Prosper) with preconditions, execution, consequences, and event generation
- **New**: Inter-settlement relations tracking (float −1.0 to +1.0 per settlement pair) with relation shift calculations per action type
- **Modified**: `world.Settlement` struct gains agent state fields (MilitaryStrength, Wealth, Relations, Goals)
- **Modified**: `settlementEntity.Tick()` replaces random event generation with agent decision loop
- **Modified**: Settlement generation initializes new settlements with agent state (default values, randomized goals)
- **Modified**: RNG scoping adds agent decision RNG per settlement (separate from figure RNG)
- **Modified**: Timeline events gain new categories from agent actions (Expansion, Raid, Conquest, Diplomacy, Economy)
- **Modified**: CFG narrative engine gains variable injection for agent action contexts (ActionType, TargetSettlement, Outcome)
- **Modified**: Obsidian export includes agent state in settlement files (relations summary, military strength, wealth tier)

## Capabilities

### New Capabilities

- `settlement-agents`: Core agent state vector (MilitaryStrength derived from population + fortification investments, Wealth as abstract economy value, Relations as map of settlement-name → float −1.0 to +1.0, Goals as slice of grow/defend/expand priorities). Decision loop: evaluate state → score actions by preconditions + goal alignment → select weighted random action → execute → record event. Deterministic RNG isolation via settlement-scoped agent RNG.

- `agent-actions`: Six action definitions with explicit preconditions, execution logic, consequences, and event generation:
  1. **Expand**: Found new settlement in unclaimed suitable tile (requires population > threshold, unclaimed tile in range, wealth > cost). Creates real settlement in state, reduces wealth, emits Expansion event.
  2. **Raid**: Steal wealth from hostile neighbor (requires military > target military, relations < −0.5, range ≤ max). Transfers wealth, shifts relations, emits Raid event with outcome (success/fail).
  3. **Conquer**: Military attack to absorb weaker neighbor (requires military >> target military, relations < −0.7). Absorbs settlement into faction, reduces military, emits Conquest event.
  4. **Fortify**: Invest wealth into military strength (requires wealth > threshold). Converts wealth to military, emits Economy event.
  5. **Ally**: Propose alliance with friendly settlement (requires relations > 0.5, no existing alliance). Sets alliance flag, shifts relations positive, emits Diplomacy event.
  6. **Prosper**: Passive growth of population and wealth (default action). Increases population/wealth based on suitability, emits Economy event.

- `settlement-relations`: Relations map stored per settlement (map[string]float64 keyed by settlement name). Initial values: 0.0 for all, modified by faction (same faction +0.3 baseline). Relation shifts per action: Raid −0.3 to −0.5, Conquer −0.8, Ally +0.4, Prosper +0.05 (gradual warming). Relations capped at −1.0 to +1.0. Relations used in action preconditions (can't ally with hostile, can't raid allies).

### Modified Capabilities

- `world-state`: `Settlement` struct in `internal/domain/world/state.go` gains `MilitaryStrength float64`, `Wealth float64`, `Relations map[string]float64`, `Goals []string` fields. All fields have JSON serialization tags. Existing fields (Name, Type, X, Y, Faction, Population, Figures) remain unchanged. Backward compatible: old world_state.json without agent fields deserializes with zero values.

- `simulation-loop`: `settlementEntity.Tick()` in `cmd/simulate.go` replaces steps 5–6 (core settlement random events) with agent decision loop. Figure lifecycle (steps 1–4) remains unchanged. Agent loop: (1) evaluate state vector, (2) score all six actions by preconditions + goal weights, (3) select action via weighted random using agent RNG, (4) execute action (modify state, create settlements, shift relations), (5) emit event with new category. Same seed produces identical agent decisions.

- `settlement-generation`: `internal/domain/settlement/generator.go` `Generate()` function initializes new settlements with agent state: MilitaryStrength = population × base_multiplier, Wealth = initial_value from config, Relations = map with baseline values (same faction +0.3, others 0.0), Goals = randomized slice of 2–3 priorities from [grow, defend, expand]. Uses settlement's figure RNG for goal randomization.

- `deterministic-rng`: Agent decision RNG scoped per settlement via `engine.GetPRNG("agent:" + settlement.Name)`. Separate from figure RNG (`"figures:" + settlement.Name`) to keep agent decisions independent from figure lifecycle. All agent operations (action selection, relation shifts, expand target selection) draw from agent RNG.

- `timeline-streaming`: Event categories expanded to include [Expansion, Raid, Conquest, Diplomacy, Economy]. Agent actions emit events with new categories: Expand → Expansion, Raid → Raid, Conquer → Conquest, Ally → Diplomacy, Fortify/Prosper → Economy. Event struct gains optional `TargetSettlement string` field for actions with targets.

- `cfg-narrative-engine`: Grammar engine gains new variable injection for agent actions. Variables: `$ActionType`, `$TargetSettlement`, `$Outcome`, `$Amount`. Grammar rules extended with `<AgentAction>` production. Fallback: missing variables resolve to `$name` placeholders.

- `obsidian-export`: Export adds agent state sections to settlement Markdown files: "Military Strength" (value + tier), "Wealth" (value + tier), "Relations" (top 5 allies and rivals with wiki-links), "Goals" (priority list). Existing export sections remain unchanged.

## Impact

- **Domain layer**: `internal/domain/world/state.go` — `Settlement` struct gains four new fields. `internal/domain/simulation/event.go` — `Event` struct gains optional `TargetSettlement` field. New package `internal/domain/agent/` for agent state types, action interfaces, decision loop logic, relations management.

- **Infra layer**: No new infra packages. Existing `internal/infra/exporter/` modified to include agent state in settlement export.

- **Use case layer**: No new usecase packages. Agent logic remains in domain layer (pure business rules).

- **Adapter layer**: `cmd/simulate.go` — `settlementEntity` struct gains `agentRNG *randv2.Rand` field. `Tick()` method modified to replace random events with agent decision loop. Simulation bootstrap derives agent RNG per settlement.

- **Config**: `worldgen.yaml` may gain optional tuning parameters. None required for initial implementation — reasonable defaults suffice.

- **Tests**: Determinism tests, unit tests (action preconditions, relation shifts, goal-weighted selection), integration tests (init→simulate→export with agent state). Coverage: ≥90% for new `internal/domain/agent/` package.
