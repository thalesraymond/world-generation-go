# ADR-0002: Architecture Deepening — Adapter Layer, Usecase Depth, and Package Reorganization

## Status

PROPOSED

## Date

2026-07-28

## Context

The codebase follows a clean-architecture layout (`cmd/ → adapter/ → usecase/ → domain/` with `infra/` implementing usecase interfaces) as documented in `AGENTS.md`. An architecture review in July 2026 surfaced several places where the intended boundaries are not earning their keep — modules that are too shallow, others that are missing entirely, and one package that sits outside the documented layer structure.

Four specific problems drive this proposal:

1. **Ghost adapter layer** — `internal/adapter/` exists in the directory tree but contains zero Go files (only a README.md). The `agent.AgentEnv` interface (defined in `internal/domain/agent/`) has its only implementation — `agentEnv` — inlined in `cmd/simulate.go`. This forces `cmd/` to import domain packages directly, violating the intended `cmd → adapter → usecase → domain` dependency chain. The deletion test confirms the adapter layer adds no value today: deleting `internal/adapter/` would change nothing. Per the design principle _"one adapter means a hypothetical seam; two adapters means a real one"_, the adapter layer remains hypothetical.

2. **Thick CLI, thin usecase** — `cmd/simulate.go` (~400 lines) contains far more than a CLI entrypoint should: goroutine and channel management, narrative engine construction from infra grammars, figure lookup tables, agent decision-loop orchestration, event filtering and narration, and JSON marshalling with file I/O. Meanwhile `internal/usecase/simulation/runner.go` (35 lines) is so shallow that its complexity would trivially move to `cmd/` if deleted — confirming it fails the deletion test.

3. **Orphaned geography package** — `internal/geography/pointcrawl/` exists outside the five documented clean architecture layers. It contains pointcrawl generation logic called by `usecase/simulation/worldgen.go`, while `internal/domain/pointcrawl/` holds the domain types. The file `geography/pointcrawl/pointcrawl.go` re-exports domain types as type aliases with zero added behavior — a textbook shallow module.

4. **Missing export seam** — `cmd/export.go` calls `exporter.Export()`, `exporter.ExportPointcrawl()`, and `exporter.ExportFigures()` directly without any usecase interface or adapter layer. The export pipeline cannot be tested at the usecase level, switching formats requires changing the CLI, and the dependency direction rule is violated.

## Decision

We will undertake four deepening actions, in priority order. All address the architecture gaps without altering domain behavior or breaking determinism guarantees.

### 1. Fill the adapter layer (Highest priority)

Move the `agentEnv` struct from `cmd/simulate.go` into a new package `internal/adapter/simulation/env.go`. This gives the adapter layer its first real occupant and restores the dependency chain:

- `cmd/simulate.go` delegates to the adapter and usecase
- `internal/adapter/simulation/env.go` implements `agent.AgentEnv`
- The adapter imports `domain/agent` for the interface and `domain/world`/`domain/pointcrawl` for the data types

A corresponding `internal/adapter/simulation/env_test.go` tests the environment adapter independently of Cobra command setup.

### 2. Deepen the usecase layer

Extract simulation orchestration from `cmd/simulate.go` into a deepened `internal/usecase/simulation/orchestrator.go`. The orchestrator owns:

- Agent setup and environment wiring
- Simulation loop orchestration (goroutine management, event streaming)
- Narrative engine construction and invocation
- Figure lookup building
- Result collection (events slice + world state)

The orchestrator exposes a small interface (e.g. `Run(config) → (State, Events, error)`) that encapsulates a complex multi-phase process — a **deep module** by design. The CLI becomes a thin adapter: parse flags, build config, call orchestrator, handle errors.

The existing `runner.go` may be replaced or subsumed by the orchestrator.

### 3. Merge geography/pointcrawl into domain/pointcrawl (When touched next)

Consolidate `internal/geography/pointcrawl/generator.go` and `routing.go` into `internal/domain/pointcrawl/`. The generation algorithm depends only on domain types (`terrain.Map`, `world.State`, `randv2.Rand`) and performs no I/O — making it a valid domain citizen. The shallow re-export file (`pointcrawl.go`) is eliminated.

**Alternative considered but rejected:** Moving to `internal/usecase/pointcrawl/` — the generation logic is pure computation without orchestration, making domain the more natural home.

### 4. Add an export seam (Lower priority)

Define a `WorldExporter` interface in `internal/usecase/simulation/export.go`. Have `internal/infra/exporter/` implement it. `cmd/export.go` delegates through the usecase instead of calling infra directly. This restores the dependency direction and makes the export pipeline testable at the usecase level with an in-memory adapter.

## Consequences

### Positive

- **Architecture compliance**: Restores the documented `cmd → adapter → usecase → domain` dependency chain. No more `cmd/` importing `domain/agent` or `infra/exporter` directly.
- **Testability**: The simulation orchestrator can be tested without setting up a Cobra command, parsing stdout, or writing files. Adapter tests exercise environment logic independently.
- **Locality**: Simulation orchestration changes concentrate in `usecase/simulation/orchestrator.go` instead of scattering across `cmd/` and `usecase/`.
- **Leverage**: One orchestrator interface serves all callers (CLI, tests, future API adapter).
- **Onboarding clarity**: Future contributors can see the adapter layer occupied and understand the pattern. No more "which layer does geography belong to?"

### Negative

- **Code churn**: Extracting the orchestrator from `cmd/simulate.go` produces a larger diff than a typical change. Careful sequencing (one commit per action) mitigates risk.
- **Learning curve**: Developers accustomed to the current thick-CLI pattern must learn to go through the usecase layer. This is mitigated by the strong convention already documented in `AGENTS.md`.
- **Export seam delay**: Action 4 is deferred to when the export format next changes — the current direct call works correctly and the seam cost isn't yet justified by a second adapter.

### Risks

- **Determinism must be preserved**: The RNG isolation logic (derived PRNGs from master seed via `state.Engine`) must not change during extraction. The orchestrator receives an `*Engine` or pre-derived RNGs to maintain byte-identical output for a given seed.
- **No behavioral change**: These are pure refactoring operations. All existing tests must pass with identical output. No new features are introduced.
- **ADRs must be kept in sync**: Once Actions 1-4 are implemented, `AGENTS.md` and this ADR should be updated to reflect the final directory structure.

## Alternatives Considered

### Do nothing

Continue with the current architecture. The system works and tests pass. However, upcoming Epics 2-5 (character-driven execution, faction agency, artifact generation, integration) would compound the existing violations by adding more orchestration logic to `cmd/simulate.go`, making it harder to test and reason about. Doing nothing now increases future refactoring cost.

### Reorganize geography into usecase instead of domain

Moving `internal/geography/pointcrawl/` into `internal/usecase/pointcrawl/` was considered but rejected because the generator and routing logic are pure computations with no I/O, orchestration, or application concerns. They belong in the domain layer alongside the types they produce.

### Create an adapter for export only

Defining an adapter interface between `cmd/export.go` and `infra/exporter/` was considered as an alternative to the usecase seam. This would fix the immediate dependency violation but would leave the export logic untestable at the usecase level. The usecase seam (Action 4) is preferred because it enables integration testing of the full export pipeline.

### Single monolithic refactor

Combining all four actions into one commit was considered but rejected. Each action is independently reviewable and reversible. Committing them separately (Action 1 → 2 → 3 → 4) makes review tractable and reduces risk of unintended interactions. Actions 3 and 4 may be deferred until the relevant subsystem is next modified.
