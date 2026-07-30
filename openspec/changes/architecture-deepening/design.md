# Architecture Deepening — Design

## Sequencing Strategy

Actions are ordered by priority and dependency. Each action is independently reviewable and committable. Actions 3 and 4 may be deferred to when the relevant subsystem is next modified.

### Commit Order

1. **Action 1: Fill adapter layer** — Lowest risk, establishes the pattern
2. **Action 2: Deepen usecase** — Largest diff, benefits from adapter being in place
3. **Action 3: Merge pointcrawl** — File moves only, no logic changes (deferred)
4. **Action 4: Add export seam** — Interface definition + implementation (deferred)

**Rationale for ordering:**

- Action 1 is the smallest, most isolated change. It establishes that the adapter layer is real and occupied.
- Action 2 is the largest refactor but benefits from Action 1 being in place (the orchestrator can use the adapter's `agentEnv` rather than rebuilding it).
- Actions 3 and 4 are deferred because they touch subsystems not being actively developed. Per the design principle "one adapter means a hypothetical seam; two adapters means a real one," we defer until a second use case justifies the seam.

---

## Action 1: Fill the Adapter Layer

### Current State

- `agentEnv` struct defined in `cmd/simulate.go` (lines ~40-46)
- Implements `agent.AgentEnv` interface from `internal/domain/agent/`
- Methods: `SuitabilityFor(x, y int) float64`, `ExpansionSites(origin settlement.Settlement, count int) []settlement.Settlement`, `RandomName() string`
- `cmd/simulate.go` imports `domain/agent`, `domain/world`, `domain/pointcrawl`, `domain/settlement` directly

### Target State

- Move `agentEnv` to `internal/adapter/simulation/env.go`
- Package: `adapter/simulation` (not `adapter/agent` — the adapter is for simulation, not agent domain)
- `cmd/simulate.go` imports `adapter/simulation` instead of `domain/agent`
- `adapter/simulation` imports `domain/agent` (for interface), `domain/world`, `domain/pointcrawl`, `domain/settlement` (for data types)

### File Structure

```
internal/adapter/
├── README.md
└── simulation/
    ├── env.go          # agentEnv implementation
    └── env_test.go     # unit tests
```

### Key Design Decisions

**Decision 1: Package naming — `adapter/simulation` vs `adapter/agent`**

_Choice:_ `adapter/simulation`

_Rationale:_ The adapter is for the simulation command's environment, not the agent domain concept. The agent domain defines the interface (`agent.AgentEnv`), but the adapter provides the concrete implementation that wires the live world state to that interface. Naming by _use case_ (simulation) rather than _domain concept_ (agent) follows the principle that adapters exist to serve application needs, not to mirror domain structure.

_Alternatives considered:_

- `adapter/agent`: Implies the adapter is part of the agent domain. Confusing — the agent interface lives in `domain/agent`, but the adapter is not an agent.
- `adapter/env`: Too generic. Doesn't convey that this is for simulation.

**Decision 2: Constructor signature**

_Choice:_

```go
func NewAgentEnv(ws *world.State, g *dompointcrawl.Graph, all *[]world.Settlement, used map[string]bool) *AgentEnv
```

_Rationale:_ Explicit dependencies, no globals. The `*[]world.Settlement` pointer is preserved from the original because the simulation appends settlements dynamically (expansion), and the adapter needs to see the growing slice. This is a known wart but changing it would alter domain behavior.

**Decision 3: Exported vs unexported type**

_Choice:_ Export `AgentEnv` type and `NewAgentEnv` constructor.

_Rationale:_ The CLI needs to construct the adapter. Keeping it unexported would force the CLI to import a factory function from the adapter, adding indirection without value.

### Determinism Preservation

- No RNG logic changes. The adapter receives pre-derived RNGs from the `state.Engine`.
- `RandomName()` uses the grammar's RNG, which is derived from the master seed. No change to derivation path.

### Testing Strategy

- **Unit tests:** Test each method (`SuitabilityFor`, `ExpansionSites`, `RandomName`) with mocked `world.State`, `pointcrawl.Graph`, and settlement slices.
- **No integration tests needed:** The adapter is pure pass-through logic. Determinism is tested at the orchestrator level.

---

## Action 2: Deepen the Usecase Layer

### Current State

- `cmd/simulate.go` (~400 lines) contains:
  - CLI flag parsing and validation
  - `agentEnv` construction
  - Narrative engine construction from infra grammars
  - Figure lookup table building
  - Agent decision-loop orchestration (goroutines, channels, WaitGroups)
  - Event filtering and narration
  - JSON marshalling and file I/O for `world_state.json` and `timeline.json`
- `internal/usecase/simulation/runner.go` (35 lines) — shallow wrapper around `domain/simulation.Run()`

### Target State

- `cmd/simulate.go` (~80 lines) — CLI entrypoint only:
  - Parse flags, build config
  - Call orchestrator
  - Handle errors
  - Print summary
- `internal/usecase/simulation/orchestrator.go` (~300 lines) — deep module owning:
  - Agent setup and environment wiring
  - Simulation loop orchestration
  - Narrative engine construction and invocation
  - Figure lookup building
  - Result collection (events slice + world state)

### File Structure

```
internal/usecase/simulation/
├── runner.go           # may be removed or subsumed
├── orchestrator.go     # new: deep simulation orchestrator
├── orchestrator_test.go # new: integration tests
├── worldgen.go         # unchanged
└── worldgen_test.go    # unchanged
```

### Key Design Decisions

**Decision 1: Orchestrator interface vs concrete type**

_Choice:_ Concrete type `Orchestrator` with exported `Run()` method.

```go
type Orchestrator struct {
    config   OrchestratorConfig
    env      AgentEnvConstructor
    narrative NarrativeEngineConstructor
}

type OrchestratorConfig struct {
    Seed       int64
    StartYear  int
    EndYear    int
    WorldState *world.State
}

func (o *Orchestrator) Run(ctx context.Context) (*RunResult, error)
```

_Rationale:_ A concrete type is simpler than an interface when there's only one implementation. If a second caller emerges, we can extract an interface then. Per `AGENTS.md`: "Avoid speculative abstractions before a second concrete use case exists."

**Decision 2: Dependency injection for adapter and infra**

_Choice:_ The orchestrator receives _constructors_ (functions) for the adapter and narrative engine, not concrete instances.

```go
type AgentEnvConstructor func(ws *world.State, g *dompointcrawl.Graph, all *[]world.Settlement, used map[string]bool) agent.AgentEnv

type NarrativeEngineConstructor func(grammarPath string) (*infranarrative.Engine, error)
```

_Rationale:_ The usecase layer should not import the adapter or infra layers directly. By injecting constructors, the usecase layer depends on _abstractions_ (function signatures), not concrete packages. The CLI wires the constructors from `adapter/simulation` and `infra/narrative`.

**Decision 3: Result type vs multiple return values**

_Choice:_ Return a `RunResult` struct.

```go
type RunResult struct {
    WorldState *world.State
    Events     []domsim.Event
    Timeline   []domsim.Event
}
```

_Rationale:_ The orchestrator produces multiple related outputs. A struct is clearer than multiple return values and allows future extension without breaking the signature.

**Decision 4: Context for cancellation**

_Choice:_ `Run(ctx context.Context)` accepts a context for cancellation.

_Rationale:_ The simulation is long-running. If the user hits Ctrl-C, the orchestrator should stop gracefully. The CLI passes a context with cancel-on-signal. This is a standard Go pattern and costs nothing.

**Decision 5: Fate of `runner.go`**

_Choice:_ Remove `runner.go` and subsume its logic into `orchestrator.go`.

_Rationale:_ `runner.go` is a shallow wrapper around `domain/simulation.Run()`. The orchestrator will call `domain/simulation.Run()` directly as part of its orchestration. Keeping `runner.go` would create two layers of indirection. The deletion test confirms `runner.go` adds no value.

### Determinism Preservation

- **RNG isolation:** The orchestrator receives a `*state.Engine` (or pre-derived RNGs). All component-scoped RNGs are derived from the master seed via `engine.GetPRNG(componentID)`. No change to derivation paths.
- **Goroutine ordering:** The orchestrator preserves the existing goroutine lifecycle: one producer goroutine per settlement agent, one consumer goroutine for event streaming. Channel buffer sizes preserved.
- **Event ordering:** Events are collected in a slice, sorted by year, then written to `timeline.json`. No change to sorting logic.

### Testing Strategy

- **Unit tests:** Test orchestrator with mocked adapter and narrative engine. Verify:
  - Agent setup (correct number of agents, correct environment wiring)
  - Simulation loop (correct number of ticks, events generated)
  - Result collection (world state updated, events slice populated)
- **Integration tests:** Run full pipeline with orchestrator. Verify byte-identical output for same seed.
- **Determinism tests:** Same seed → byte-identical `world_state.json`, `timeline.json`, and all export files.

---

## Action 3: Merge geography/pointcrawl into domain/pointcrawl (Deferred)

### Current State

- `internal/geography/pointcrawl/` contains:
  - `generator.go` — pointcrawl graph generation from terrain map
  - `routing.go` — travel cost calculation between nodes
  - `pointcrawl.go` — type aliases re-exporting `domain/pointcrawl` types
  - `generator_test.go`, `routing_test.go`, `pointcrawl_test.go`
- Called by `usecase/simulation/worldgen.go`
- `internal/domain/pointcrawl/` contains domain types: `Node`, `Edge`, `Graph`, `Visibility`

### Target State

- Move `generator.go` and `routing.go` into `internal/domain/pointcrawl/`
- Delete `internal/geography/pointcrawl/`
- Update imports in `usecase/simulation/worldgen.go` and any other callers

### File Structure

```
internal/domain/pointcrawl/
├── expansion.go         # existing
├── expansion_test.go    # existing
├── generator.go         # moved from geography/pointcrawl/
├── generator_test.go    # moved from geography/pointcrawl/
├── json.go              # existing
├── routing.go           # moved from geography/pointcrawl/
├── routing_test.go      # moved from geography/pointcrawl/
├── types.go             # existing
└── types_test.go        # existing
```

### Key Design Decisions

**Decision 1: Domain vs usecase for generation logic**

_Choice:_ Move to `domain/pointcrawl/`

_Rationale:_ The generation algorithm depends only on domain types (`terrain.Map`, `world.State`, `randv2.Rand`) and performs no I/O. It's pure computation. The usecase layer is for orchestration, not pure domain logic. The domain layer is the natural home.

**Decision 2: Type aliases removal**

_Choice:_ Delete `pointcrawl.go` (the type alias file). Update all imports to use `domain/pointcrawl` directly.

_Rationale:_ The type aliases add no value. They were a workaround for the orphaned `geography/` package. Consolidating into `domain/pointcrawl/` eliminates the need for aliases.

### Determinism Preservation

- No RNG logic changes. The generation algorithm uses the same `randv2.Rand` passed from `worldgen.go`.
- No change to derivation path: `engine.GetPRNG("pointcrawl")` is still called in `worldgen.go` and passed to the generator.

### Testing Strategy

- **No new tests needed.** The existing `generator_test.go` and `routing_test.go` move with the code.
- **Verify all tests pass** after import updates.

---

## Action 4: Add an Export Seam (Deferred)

### Current State

- `cmd/export.go` calls:
  - `exporter.Export(state, cfg.Output)`
  - `exporter.ExportPointcrawl(state, cfg.Output)`
  - `exporter.ExportFigures(state, cfg.Output)` (if events exist)
- Direct calls to `infra/exporter/` without usecase interface
- `cmd/export.go` imports `infra/exporter` directly

### Target State

- Define `WorldExporter` interface in `internal/usecase/simulation/export.go`
- `infra/exporter/` implements the interface
- `cmd/export.go` calls usecase-layer export function, which delegates to the interface

### File Structure

```
internal/usecase/simulation/
├── export.go            # new: WorldExporter interface + ExportWorld function
├── export_test.go       # new: integration tests with in-memory adapter

internal/infra/exporter/
├── exporter.go          # existing: implements WorldExporter
├── ...                  # other existing files
```

### Key Design Decisions

**Decision 1: Interface location**

_Choice:_ Define `WorldExporter` interface in `usecase/simulation/export.go`

```go
type WorldExporter interface {
    Export(state *world.State, outputDir string) error
    ExportPointcrawl(state *world.State, outputDir string) error
    ExportFigures(state *world.State, outputDir string, events []domsim.Event) error
}
```

_Rationale:_ The usecase layer defines the contract; the infra layer implements it. This follows the dependency inversion principle. The CLI depends on the usecase interface, not the infra concrete type.

**Decision 2: Single orchestrating function vs multiple methods**

_Choice:_ Add an `ExportWorld(exporter WorldExporter, state *world.State, outputDir string, events []domsim.Event) error` function in the usecase layer that orchestrates the three export calls.

```go
func ExportWorld(exporter WorldExporter, state *world.State, outputDir string, events []domsim.Event) error {
    if err := exporter.Export(state, outputDir); err != nil {
        return fmt.Errorf("export world: %w", err)
    }
    if err := exporter.ExportPointcrawl(state, outputDir); err != nil {
        return fmt.Errorf("export pointcrawl: %w", err)
    }
    if len(events) > 0 {
        if err := exporter.ExportFigures(state, outputDir, events); err != nil {
            return fmt.Errorf("export figures: %w", err)
        }
    }
    return nil
}
```

_Rationale:_ The CLI should not know the order of export operations. The usecase layer owns the orchestration. If a fifth export operation is added, the CLI doesn't change.

**Decision 3: Concrete adapter vs interface injection**

_Choice:_ The CLI constructs the concrete `infra/exporter.Exporter` and passes it to `ExportWorld()`.

```go
// cmd/export.go
exp := exporter.New()
if err := ucsim.ExportWorld(exp, state, cfg.Output, events); err != nil {
    return err
}
```

_Rationale:_ Simple. The CLI knows about the concrete infra type (which is fine — the CLI is the composition root). The usecase layer depends on the interface, not the concrete type.

### Determinism Preservation

- No changes to export logic. The interface is a pass-through.
- Export determinism is tested at the infra layer (existing tests).

### Testing Strategy

- **Integration test:** Create an in-memory `WorldExporter` adapter for testing. Verify `ExportWorld()` calls the correct methods in the correct order.
- **No new infra tests needed.** Existing `exporter` tests verify the concrete implementation.
