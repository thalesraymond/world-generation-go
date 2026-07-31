## Why

The current figure system is passive. Historical figures have roles and generate events, but they lack personality (no stats), legacy (no reputation), meaningful succession (heir logic exists but is never called), dynamic role changes (transition logic exists but is never called), and their names don't appear in narrative text. Character names are injected as CFG variables but never consumed by grammar rules. The world needs figures that feel like characters — not just data records — driving events through their skills, reputations, and dynastic legacies.

## What Changes

**Dependency note:** Epic 2 builds on Epic 1 (Settlement Agent Foundation) for the "figures as executors of agent actions" pattern — when a settlement-agent decides to raid, it picks a General. However, many Epic 2 items (stats, reputation, new roles, succession, marriage, narrative grammar) can be implemented independently and will ship first.

- **New**: Stats system on `HistoricalFigure` — `Stats{Martial int, Diplomatic int, Infamy int}` (1–20 range). Generated deterministically at figure creation with role-based bias (Generals get +2 Martial, Diplomats +2 Diplomatic). Affects action success probability and magnitude.
- **New**: Reputation system — event-sourced log of notable deeds per figure (`[]ReputationEntry{Year, Event, Delta, Description}`). Influences future event outcomes and faction relationships. Preserves causal history for export provenance.
- **New**: Three new roles — `General` (leads military actions: raids, conquests), `Diplomat` (negotiates alliances, treaties), `Master Smith` (generates Settlement-category craftsmanship events; artifact forging deferred to Epic 4).
- **New**: Role transition mechanics — wire `CanTransitionTo()` into event outcomes (Explorer founds settlement → becomes Leader; Leader exiled → becomes Explorer; General wins battle → reputation gain).
- **Modified**: Succession — wire `GetHeir()` into death handling so eldest child inherits leadership, not random roleless adult. Heir gains partial stat bonuses from parent (half parent's stats, rounded down).
- **Modified**: Marriage mechanics — wire `FormMarriage()` into simulation loop. Adult figures (age 20–25) may marry within same faction. Cross-faction marriage deferred to Epic 3.
- **Modified**: Narrative engine — add grammar rules that consume `$FigureName`, `$FigureRole`, `$SettlementName` variables and produce character-driven event text ("General Cedric of Ashfield led a raid on Blackdale" not "A raid occurred").
- **Modified**: Event generation — all existing role event generators updated to produce richer event descriptions with figure context and stats-influenced outcomes.
- **Modified**: Export — update Obsidian character files with stats summary, reputation highlights, role transition history, and role-specific sections.
- **Modified**: Role storage — change `Role` field from `string` to `Role` interface with JSON marshal/unmarshal support, enabling role-specific behavior to persist.

## Capabilities

### New Capabilities

- `figure-stats`: Stats system on `HistoricalFigure` — `Martial`, `Diplomatic`, `Infamy` fields with 1–20 range, role-based generation bias, deterministic RNG derivation.
- `figure-reputation`: Event-sourced reputation tracking — `ReputationEntry` log on each figure, `AddReputation()` method, reputation influence on event outcomes.
- `figure-roles-general-diplomat-smith`: `General`, `Diplomat`, `Master Smith` role implementations. Each generates domain-specific events with stats-aware outcomes.
- `figure-role-transitions`: Wired `CanTransitionTo()` into simulation — role changes triggered by events (Explorer→Leader on settlement founding, Leader→Explorer on exile, etc.).

### Modified Capabilities

- `figure-roles`: Updated `Role` interface if needed for stats-aware generation. New entries in role registry (`NewRole`) for General, Diplomat, Master Smith.
- `figure-events`: All role event generators produce character-driven descriptions with figure names, stats-influenced outcomes, and reputation deltas.
- `figure-relationships`: Wired `FormMarriage()` into simulation loop (same-faction, adult-age gate). Wired `GetHeir()` into death handling with stat inheritance.
- `figure-export`: Character files include stats summary, reputation log, role transition history, and role-specific detail sections.
- `figure-determinism`: Stats/reputation/transition RNG isolation — new RNG scopes for stats generation and reputation calculation.
- `cfg-narrative-engine`: Grammar rules consume `$FigureName`, `$FigureRole`, `$SettlementName` variables. Character-driven production rules per event category. Fallback for events without figure references.
- `simulation-loop`: `Tick()` processes succession (heir-first), marriage attempts, role transitions, and stats-updated event generation.
- `historical-figures`: `HistoricalFigure` struct gains `Stats`, `Reputation`, `TransitionHistory` fields. Role field changes from `string` to `Role` interface.
- `world-state`: Settlement may need leader reference field for quick lookup.

## Impact

- **Domain layer**: `internal/domain/figures/` — new files for stats, reputation, new roles (General, Diplomat, Master Smith), role storage refactor, transition logic. `internal/domain/simulation/event.go` — event struct may gain stats/reputation fields.
- **Domain layer (existing)**: `internal/domain/figures/lifecycle.go` — succession wiring, marriage wiring, role transition integration. `internal/domain/figures/figure.go` — struct expansion with Stats, Reputation fields.
- **Infra layer**: `internal/infra/narrative/default_grammar.go` — new CFG production rules for figure variables. `internal/infra/exporter/figures.go` — updated character file generation with new fields.
- **Use case layer**: Integration of succession, marriage, transitions into simulation Tick orchestration.
- **Adapter layer**: Minimal — CLI wiring changes only if command flags change for new figure options.
- **Tests**: Stats generation determinism, reputation accumulation, new role event generation, succession with stat inheritance, marriage same-faction matching, role transitions, narrative grammar with figure variables, export with new fields, coverage validation (≥80% repo, ≥90% domain/usecase).
