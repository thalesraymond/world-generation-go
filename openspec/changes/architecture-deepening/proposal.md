# Architecture Deepening — Adapter Layer, Usecase Depth, and Package Reorganization

## Problem Statement

The codebase follows a clean-architecture layout (`cmd/ → adapter/ → usecase/ → domain/` with `infra/` implementing usecase interfaces) as documented in `AGENTS.md`. An architecture review surfaced four specific gaps where the intended boundaries are not earning their keep:

1. **Ghost adapter layer** — `internal/adapter/` exists in the directory tree but contains zero Go files (only a README.md). The `agent.AgentEnv` interface (defined in `internal/domain/agent/`) has its only implementation — `agentEnv` — inlined in `cmd/simulate.go`. This forces `cmd/` to import domain packages directly, violating the intended `cmd → adapter → usecase → domain` dependency chain. The deletion test confirms the adapter layer adds no value today: deleting `internal/adapter/` would change nothing.

2. **Thick CLI, thin usecase** — `cmd/simulate.go` (~400 lines) contains far more than a CLI entrypoint should: goroutine and channel management, narrative engine construction from infra grammars, figure lookup tables, agent decision-loop orchestration, event filtering and narration, and JSON marshalling with file I/O. Meanwhile `internal/usecase/simulation/runner.go` (35 lines) is so shallow that its complexity would trivially move to `cmd/` if deleted — confirming it fails the deletion test.

3. **Orphaned geography package** — `internal/geography/pointcrawl/` exists outside the five documented clean architecture layers. It contains pointcrawl generation logic called by `usecase/simulation/worldgen.go`, while `internal/domain/pointcrawl/` holds the domain types. The file `geography/pointcrawl/pointcrawl.go` re-exports domain types as type aliases with zero added behavior — a textbook shallow module.

4. **Missing export seam** — `cmd/export.go` calls `exporter.Export()`, `exporter.ExportPointcrawl()`, and `exporter.ExportFigures()` directly without any usecase interface or adapter layer. The export pipeline cannot be tested at the usecase level, switching formats requires changing the CLI, and the dependency direction rule is violated.

## Solution

We will undertake four deepening actions, in priority order. All address the architecture gaps without altering domain behavior or breaking determinism guarantees:

1. **Fill the adapter layer** (highest priority) — Move `agentEnv` struct from `cmd/simulate.go` into `internal/adapter/simulation/env.go`. This gives the adapter layer its first real occupant and restores the dependency chain.

2. **Deepen the usecase layer** (high priority) — Extract simulation orchestration from `cmd/simulate.go` into a deepened `internal/usecase/simulation/orchestrator.go`. The orchestrator owns agent setup, simulation loop orchestration, narrative engine construction, figure lookup building, and result collection. The CLI becomes a thin adapter.

3. **Merge geography/pointcrawl into domain/pointcrawl** (when touched next) — Consolidate `internal/geography/pointcrawl/generator.go` and `routing.go` into `internal/domain/pointcrawl/`. The generation algorithm depends only on domain types and performs no I/O — making it a valid domain citizen.

4. **Add an export seam** (lower priority) — Define a `WorldExporter` interface in `internal/usecase/simulation/`. Have `internal/infra/exporter/` implement it. `cmd/export.go` delegates through the usecase instead of calling infra directly.

## Success Criteria

### Architecture Compliance

- [ ] `cmd/simulate.go` imports zero `domain/` packages directly (only `adapter/` and `usecase/`)
- [ ] `cmd/export.go` calls usecase-layer export interface instead of `infra/exporter/` directly
- [ ] `internal/adapter/` contains at least one non-README Go file with tests
- [ ] No packages exist outside the five documented layers (`cmd/`, `adapter/`, `usecase/`, `domain/`, `infra/`)
- [ ] Dependency direction verified: `cmd → adapter → usecase → domain`, `infra` implements `usecase` interfaces

### Testability

- [ ] Simulation orchestrator testable without Cobra command setup or stdout parsing
- [ ] Adapter `agentEnv` testable independently of CLI
- [ ] Export pipeline testable at usecase level with in-memory adapter

### Determinism Preservation

- [ ] All existing determinism tests pass: same seed → byte-identical outputs
- [ ] `go test ./... -race` passes with no data races
- [ ] No RNG isolation logic changes during extraction

### Coverage

- [ ] Repository-wide statement coverage ≥ 80%
- [ ] `internal/domain/` coverage ≥ 90%
- [ ] `internal/usecase/` coverage ≥ 90%
- [ ] New adapter and orchestrator code covered at ≥ 90%

### Code Quality

- [ ] `go vet ./...` passes
- [ ] `golangci-lint run` passes
- [ ] No cyclic dependencies introduced
- [ ] All existing tests pass without modification (pure refactor)

## Non-Goals

This change explicitly does **NOT**:

- **Add new features** — No new simulation behavior, event types, or export formats. Pure refactoring only.
- **Modify domain logic** — No changes to biome mapping, suitability scoring, figure lifecycle, or narrative grammar rules.
- **Restructure the domain layer** — No reorganization of `internal/domain/` subpackages beyond Action 3's pointcrawl merge.
- **Implement Epic 2-5 features** — Character-driven execution, faction agency, artifact generation remain separate epics.
- **Add concurrency or performance optimizations** — No goroutine pooling, caching, or parallelism changes. Existing concurrency patterns preserved as-is.
- **Change CLI user experience** — No new flags, changed command behavior, or altered output formats.
- **Introduce new dependencies** — No new third-party libraries.
- **Refactor `cmd/init.go` or `cmd/root.go`** — Only `cmd/simulate.go` and `cmd/export.go` are in scope.
