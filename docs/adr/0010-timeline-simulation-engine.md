# ADR-0010: Timeline Simulation Engine — Chronological Ticking and Event Streaming

## Status

ACCEPTED

## Date

2026-08-08

## Context

After the world is generated (terrain, demographics, settlements, agents, figures), history must be simulated year by year. Each settlement agent acts each year; figures age, die, marry, and rise; narrative events accumulate into a timeline that is streamed and later exported as a chronicle. The engine must guarantee chronological order, deterministic behavior, and clean channel semantics for consumers.

## Decision

`internal/domain/simulation` provides the core engine plus the event type:

- `Simulation` (engine.go) holds `startYear`, `endYear`, a list of `Entity` ticks, and an RNG. `Run(eventChan)` iterates years from start to end, and for each year ticks every entity in registration order, passing the year, the event channel, and the shared timeline RNG. It `defer close(eventChan)` so consumers can range until completion.
- `Entity` (entity.go) is the single tick interface: `Tick(year int, eventChan chan<- Event, rng *randv2.Rand)`.
- `Event` (event.go) carries `Year`, `Category`, `Description`, `FigureID`, `RelatedFigures`, `SettlementName`, `TargetSettlement` with `omitempty` JSON tags, and `FormatEvent` renders human-readable lines (e.g., `[42] (War) Ashfield → Blackdale: …`).

The usecase layer (`internal/usecase/simulation/runner.go`) wraps the engine: `RunSimulation` validates the year range, registers entities, spawns one goroutine that ranges over the event channel and writes formatted lines to an `io.Writer`, then runs the engine synchronously and waits for the goroutine. `cmd/simulate.go` supplies the entities: one settlement-agent per settlement plus figure lifecycle handling.

## Alternatives Considered

### Event-driven time-advance loop (process next event with its own timestamp)

- **Pros:** Natural for irregularly spaced events.
- **Cons:** Here every year every settlement acts, so events are inherently yearly; an event queue adds ordering complexity without benefit.
- **Rejected:** Fixed-year loop matches the simulation's structure.

### Entities push directly to a channel with no engine orchestrator

- **Pros:** Minimal indirection.
- **Cons:** No centralized year loop or deterministic entity ordering; harder to reason about chronology.
- **Rejected:** The engine centralizes the year/entity iteration and channel ownership.

### Buffered single-entity simulation (agents only, no separate figure loop)

- **Pros:** Fewer moving parts.
- **Cons:** Figure lifecycle (deaths, successions, marriages) is inherently per-settlement and interleaves with agent decisions; separating concerns keeps each domain package cohesive.
- **Rejected:** Both agent and figure systems are separate entities registered into the same engine.

### Unbounded or caller-owned channel

- **Pros:** Simpler API.
- **Cons:** Unbounded buffers hide backpressure; caller-owned close semantics risk double-close or leaks. The engine owning `close` makes lifecycle explicit.
- **Rejected:** The engine owns the channel lifecycle; consumers only range.

## Consequences

- Chronology is guaranteed: events are produced strictly in year order (entity order within a year is deterministic registration order).
- The engine closing the event channel gives consumers a clear termination signal; the runner's goroutine + `WaitGroup` guarantees no goroutine leak.
- Determinism depends on shared RNG consumption order: entity registration order and each entity's RNG draw pattern must stay stable — covered by `engine_test.go` determinism assertions.
- `Event` is the single cross-cutting contract consumed by the narrative engine (`Narrate`), the runner (`FormatEvent`), and the exporter (`ExportTimeline`); field additions are JSON `omitempty`-safe.
- The runner is deliberately thin (goroutine + format loop); the bulk of orchestration lives in `cmd/simulate.go`, which ADR-0002 Action 2 proposes moving into a usecase orchestrator.
