# world-generation-go

`world-generation-go` is a deterministic, LLM-free Go CLI for fantasy world generation. It uses
constructive generation algorithms, cellular automata, context-free grammars, and deterministic
pseudo-random number generation to produce reproducible worlds and histories.

The project follows Clean Architecture with an OpenSpec-driven design.

## Architecture

```
cmd/                      CLI entrypoints (Cobra)
config/                   Typed configuration loading (Viper)
internal/
  adapter/                Input/output translation (scaffolded, not yet populated)
  domain/                 Pure business entities and rules
    demographics/         Cellular-automata population diffusion and faction spread
    narrative/            CFG parser and rule-expansion engine
    pointcrawl/           Graph node/edge types and serialization
    settlement/           Suitability-driven settlement placement
    simulation/           Iterative year-by-year simulation engine with tickable entities
    spatial/              Tile suitability scoring from terrain layers
    state/                Deterministic PRNG engine with component-scoped sub-seeds
    terrain/              Perlin-noise elevation, latitude-aware climate, biome classification
    world/                Aggregate world state with serialization
  geography/              Spatial generation algorithms
    pointcrawl/           Pointcrawl network generation, node placement, and edge routing
  infra/                  File I/O and external integrations
    exporter/             Obsidian Markdown export (frontmatter + wiki-links)
    streamer/             (scaffolded, not yet populated)
  usecase/                Application orchestration
    simulation/           World generation pipeline and timeline runner
openspec/                 Source-of-truth specs and design documents
```

Dependencies point inward: `cmd` → `usecase` → `domain`. `infra` implements interfaces from `usecase`. `domain` imports no framework or infrastructure packages.

## What is implemented

### CLI

- **Cobra command tree** with `init`, `simulate`, `export` subcommands.
- **Typed config** via Viper: flags, `WORLDGEN_*` env vars, and YAML config files.
- `simulate` runs full world generation and timeline simulation. `init` and `export` currently acknowledge inputs; `export` backing logic is implemented in `internal/infra/exporter`.

### Generation pipeline

- **Deterministic RNG** — master-seed engine derives isolated `math/rand/v2` PRNG streams per component.
- **Terrain generation** — Perlin-noise elevation, latitude-aware temperature/humidity, biome classification (water, tundra, desert, forest, grassland).
- **Spatial reasoning** — tile suitability scoring from water proximity, elevation, and biome.
- **Demographic automata** — population seeding from suitability, diffusion, and faction influence spread.
- **Settlement generation** — candidate selection from suitability/population thresholds with distance filtering.
- **Pointcrawl network** — node placement from settlements, edge connection via nearest-neighbor routing.
- **Simulation loop** — iterative year-by-year engine with tickable entities and channel-based event streaming.
- **CFG narrative engine** — lexer, parser, and rule-expansion engine supporting variable injection and recursion protection.
- **World state** — aggregate snapshot with JSON serialization/deserialization and validation.

### Export

- **Obsidian Markdown vault** — directory creation, YAML frontmatter, wiki-links, and sanitized filenames for settlements and factions.

### Testing

Comprehensive unit and integration tests across all domain and use case packages, including determinism tests (same seed → identical output), CLI behavior, and export format correctness.

## Getting started

```bash
go run .                        # show help
go run . init --name "Ashtar" --size medium
go run . simulate --years 500 --events dense
go run . export --format obsidian --output ./vault
```

The `simulate` command performs full world generation and runs the timeline:

```bash
go run . simulate --seed 42 --width 64 --height 64 --years 100
```

## Configuration

| Level       | Mechanism                          |
|-------------|-------------------------------------|
| Flags       | `--seed`, `--output`, `--config`   |
| Env vars    | `WORLDGEN_SEED`, `WORLDGEN_OUTPUT` |
| Config file | `worldgen.yaml` in CWD or `$HOME`  |
| Defaults    | hardcoded in command definitions   |

Command-specific flags:

- `init` — `--name`, `--size`
- `simulate` — `--width`, `--height`, `--years`, `--events`
- `export` — `--format`

## Determinism

Identical seeds produce identical worlds. Enforced by:

- `math/rand/v2` PRNGs created per component.
- Master-seed engine in `internal/domain/state` deriving stable sub-seeds from component identifiers.
- No package-level random state or shared global RNG.

## Testing

```bash
go test ./...                     # run all tests
go test ./... -race               # with race detector
go test ./... -coverprofile=coverage.out  # coverage
go tool cover -func=coverage.out
```

Coverage thresholds (from [AGENTS.md](AGENTS.md)):

- Repository-wide: ≥ 80%
- `internal/domain` and `internal/usecase`: ≥ 90% each

## Roadmap

Completed items have moved into `openspec/changes/archive/`. Remaining work tracked in `openspec/changes/` includes:

- **`init` project scaffolding** — persistent project file generation.
- **`export` CLI integration** — wire the existing Obsidian exporter into the `export` command with a full world state input.
- **Timeline streaming with CFG narratives** — route simulation events through the narrative engine for rich mythic text output.
- **Travel cost calculator** — route-evaluation heuristics measured in "watches".

For full requirements and design rationale, see the corresponding `design.md` and `spec.md` files under `openspec/specs/`.