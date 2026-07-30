# Architecture Deepening — Implementation Tasks

## Action 1: Fill the Adapter Layer

### 1.1 Create adapter package structure

- [ ] Create `internal/adapter/simulation/` directory
- [ ] Create `internal/adapter/simulation/env.go` with package declaration and imports
- [ ] Move `agentEnv` struct definition from `cmd/simulate.go` to `internal/adapter/simulation/env.go`
- [ ] Rename `agentEnv` to `AgentEnv` (exported)
- [ ] Add `NewAgentEnv` constructor: `func NewAgentEnv(ws *world.State, g *dompointcrawl.Graph, all *[]world.Settlement, used map[string]bool) *AgentEnv`
- [ ] Ensure all methods (`SuitabilityFor`, `ExpansionSites`, `RandomName`) are exported and match the `agent.AgentEnv` interface

### 1.2 Update cmd/simulate.go to use adapter

- [ ] Remove `agentEnv` struct definition from `cmd/simulate.go`
- [ ] Remove `domain/agent` import from `cmd/simulate.go`
- [ ] Add `adapter/simulation` import to `cmd/simulate.go`
- [ ] Update `agentEnv` construction to call `adaptersim.NewAgentEnv(...)`
- [ ] Verify `cmd/simulate.go` no longer imports any `domain/` packages directly

### 1.3 Write adapter tests

- [ ] Create `internal/adapter/simulation/env_test.go`
- [ ] Test `SuitabilityFor` with mocked `world.State` (returns correct suitability value)
- [ ] Test `ExpansionSites` with mocked `world.State`, `pointcrawl.Graph`, and settlement slice (returns correct number of sites, respects distance constraints)
- [ ] Test `RandomName` with mocked grammar (returns name, marks as used)
- [ ] Test `NewAgentEnv` constructor (all fields initialized correctly)

### 1.4 Verify Action 1

- [ ] Run `go test ./internal/adapter/simulation/...` — all tests pass
- [ ] Run `go test ./cmd/...` — all tests pass
- [ ] Run `go vet ./...` — no errors
- [ ] Run `golangci-lint run` — no errors
- [ ] Run `go test ./... -race` — no data races
- [ ] Commit: `feat(adapter): fill adapter layer with agentEnv implementation`

---

## Action 2: Deepen the Usecase Layer

### 2.1 Create orchestrator structure

- [ ] Create `internal/usecase/simulation/orchestrator.go` with package declaration and imports
- [ ] Define `OrchestratorConfig` struct:
  ```go
  type OrchestratorConfig struct {
      Seed       int64
      StartYear  int
      EndYear    int
      WorldState *world.State
      GrammarPath string
      OutputDir   string
  }
  ```
- [ ] Define `RunResult` struct:
  ```go
  type RunResult struct {
      WorldState *world.State
      Events     []domsim.Event
      Timeline   []domsim.Event
  }
  ```
- [ ] Define `AgentEnvConstructor` type:
  ```go
  type AgentEnvConstructor func(ws *world.State, g *dompointcrawl.Graph, all *[]world.Settlement, used map[string]bool) agent.AgentEnv
  ```
- [ ] Define `NarrativeEngineConstructor` type:
  ```go
  type NarrativeEngineConstructor func(grammarPath string) (*infranarrative.Engine, error)
  ```
- [ ] Define `Orchestrator` struct with config and constructor fields

### 2.2 Extract orchestration logic from cmd/simulate.go

- [ ] Move agent setup logic from `cmd/simulate.go` to `Orchestrator.setupAgents()`
- [ ] Move narrative engine construction from `cmd/simulate.go` to `Orchestrator.buildNarrativeEngine()`
- [ ] Move figure lookup table building from `cmd/simulate.go` to `Orchestrator.buildFigureLookups()`
- [ ] Move simulation loop orchestration (goroutines, channels, WaitGroups) from `cmd/simulate.go` to `Orchestrator.runSimulationLoop()`
- [ ] Move event filtering and narration from `cmd/simulate.go` to `Orchestrator.filterAndNarrateEvents()`
- [ ] Move result collection from `cmd/simulate.go` to `Orchestrator.collectResults()`
- [ ] Implement `Orchestrator.Run(ctx context.Context) (*RunResult, error)` that calls the above methods in sequence

### 2.3 Remove runner.go

- [ ] Verify `runner.go` is not used elsewhere (search for `RunSimulation` calls)
- [ ] Delete `internal/usecase/simulation/runner.go`
- [ ] Delete `internal/usecase/simulation/runner_test.go` (if exists)
- [ ] Update any tests that used `RunSimulation` to use the orchestrator instead

### 2.4 Thin cmd/simulate.go

- [ ] Remove all orchestration logic from `cmd/simulate.go`
- [ ] Keep only:
  - CLI flag parsing and validation
  - Config construction
  - Orchestrator construction (injecting `adaptersim.NewAgentEnv` and `infranarrative.NewEngine`)
  - Call to `orchestrator.Run(ctx)`
  - Error handling and summary printing
- [ ] Remove JSON marshalling and file I/O from `cmd/simulate.go` (move to orchestrator or keep in infra)
- [ ] Verify `cmd/simulate.go` is ~80 lines or fewer

### 2.5 Write orchestrator tests

- [ ] Create `internal/usecase/simulation/orchestrator_test.go`
- [ ] Test `Orchestrator.Run()` with mocked adapter and narrative engine:
  - Verify agent setup (correct number of agents, correct environment wiring)
  - Verify simulation loop (correct number of ticks, events generated)
  - Verify result collection (world state updated, events slice populated)
- [ ] Test context cancellation: pass a cancelled context, verify orchestrator stops gracefully
- [ ] Test error handling: mock narrative engine to return error, verify orchestrator propagates error

### 2.6 Write integration tests

- [ ] Create integration test: run full pipeline with orchestrator
- [ ] Verify byte-identical output for same seed (compare `world_state.json`, `timeline.json`)
- [ ] Verify determinism: run twice with same seed, compare outputs

### 2.7 Verify Action 2

- [ ] Run `go test ./internal/usecase/simulation/...` — all tests pass
- [ ] Run `go test ./cmd/...` — all tests pass
- [ ] Run `go test ./...` — all tests pass
- [ ] Run `go vet ./...` — no errors
- [ ] Run `golangci-lint run` — no errors
- [ ] Run `go test ./... -race` — no data races
- [ ] Run `go test ./... -coverprofile=coverage.out` — verify coverage thresholds
- [ ] Commit: `feat(usecase): deepen usecase layer with simulation orchestrator`

---

## Action 3: Merge geography/pointcrawl into domain/pointcrawl (Deferred)

### 3.1 Move files

- [ ] Move `internal/geography/pointcrawl/generator.go` to `internal/domain/pointcrawl/generator.go`
- [ ] Move `internal/geography/pointcrawl/generator_test.go` to `internal/domain/pointcrawl/generator_test.go`
- [ ] Move `internal/geography/pointcrawl/routing.go` to `internal/domain/pointcrawl/routing.go`
- [ ] Move `internal/geography/pointcrawl/routing_test.go` to `internal/domain/pointcrawl/routing_test.go`

### 3.2 Update imports

- [ ] Update `internal/usecase/simulation/worldgen.go` to import `domain/pointcrawl` instead of `geography/pointcrawl`
- [ ] Update any other files that import `geography/pointcrawl`
- [ ] Delete `internal/geography/pointcrawl/pointcrawl.go` (type aliases no longer needed)
- [ ] Delete `internal/geography/pointcrawl/` directory

### 3.3 Verify Action 3

- [ ] Run `go test ./internal/domain/pointcrawl/...` — all tests pass
- [ ] Run `go test ./internal/usecase/simulation/...` — all tests pass
- [ ] Run `go test ./...` — all tests pass
- [ ] Run `go vet ./...` — no errors
- [ ] Run `golangci-lint run` — no errors
- [ ] Commit: `refactor(domain): merge geography/pointcrawl into domain/pointcrawl`

---

## Action 4: Add an Export Seam (Deferred)

### 4.1 Define WorldExporter interface

- [ ] Create `internal/usecase/simulation/export.go` with package declaration and imports
- [ ] Define `WorldExporter` interface:
  ```go
  type WorldExporter interface {
      Export(state *world.State, outputDir string) error
      ExportPointcrawl(state *world.State, outputDir string) error
      ExportFigures(state *world.State, outputDir string, events []domsim.Event) error
  }
  ```
- [ ] Define `ExportWorld` function:
  ```go
  func ExportWorld(exporter WorldExporter, state *world.State, outputDir string, events []domsim.Event) error
  ```
- [ ] Implement `ExportWorld` to call the three methods in sequence with error wrapping

### 4.2 Update cmd/export.go

- [ ] Remove direct calls to `exporter.Export()`, `exporter.ExportPointcrawl()`, `exporter.ExportFigures()`
- [ ] Remove `infra/exporter` import from `cmd/export.go`
- [ ] Construct concrete `exporter.Exporter` (or use `exporter.New()` if constructor exists)
- [ ] Call `ucsim.ExportWorld(exp, state, cfg.Output, events)`
- [ ] Verify `cmd/export.go` no longer imports `infra/exporter` directly

### 4.3 Verify infra/exporter implements interface

- [ ] Verify `infra/exporter.Exporter` implements all three methods of `WorldExporter` interface
- [ ] If method signatures differ, update `infra/exporter` to match the interface
- [ ] Add compile-time check: `var _ ucsim.WorldExporter = (*exporter.Exporter)(nil)` in `infra/exporter/exporter.go`

### 4.4 Write export seam tests

- [ ] Create `internal/usecase/simulation/export_test.go`
- [ ] Create in-memory `WorldExporter` adapter for testing
- [ ] Test `ExportWorld()` calls the correct methods in the correct order
- [ ] Test error handling: mock exporter to return error, verify `ExportWorld()` stops and returns error
- [ ] Test `ExportFigures()` is skipped when `events` is empty

### 4.5 Verify Action 4

- [ ] Run `go test ./internal/usecase/simulation/...` — all tests pass
- [ ] Run `go test ./cmd/...` — all tests pass
- [ ] Run `go test ./...` — all tests pass
- [ ] Run `go vet ./...` — no errors
- [ ] Run `golangci-lint run` — no errors
- [ ] Commit: `feat(usecase): add export seam with WorldExporter interface`

---

## Global Verification

### Final integration tests

- [ ] Run `go test ./...` — all tests pass
- [ ] Run `go test ./... -race` — no data races
- [ ] Run `go test ./... -coverprofile=coverage.out` — verify coverage thresholds
- [ ] Run `go vet ./...` — no errors
- [ ] Run `golangci-lint run` — no errors

### Determinism verification

- [ ] Run full pipeline twice with same seed — byte-identical outputs

### Architecture compliance

- [ ] `cmd/simulate.go` imports zero `domain/` packages directly
- [ ] `cmd/export.go` calls usecase-layer export interface instead of `infra/exporter/`
- [ ] `internal/adapter/` contains at least one non-README Go file with tests
- [ ] No packages outside the five documented layers
- [ ] Dependency direction: `cmd → adapter → usecase → domain`, `infra` implements `usecase` interfaces

### Documentation updates

- [ ] Update `AGENTS.md` to reflect final directory structure (if needed)
- [ ] Update `internal/adapter/README.md` to document the new `simulation` package
- [ ] Update `internal/usecase/simulation/README.md` (if exists) to document orchestrator
- [ ] Mark ADR-0002 as ACCEPTED and update with implementation date

### Final commit

- [ ] Commit all documentation updates: `docs: update architecture docs after deepening refactor`
