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

## Features & Implemented Capabilities

### CLI Commands

- **`init`**: Scaffolds and writes persistent project configuration (`worldgen.yaml`) with user-specified world name and size presets.
- **`simulate`**: Executes full procedural world generation (terrain, biomes, demographics, settlements, pointcrawl graph) and runs phased timeline simulation streaming CFG narrative chronicle events. Saves `world_state.json` and `timeline.json`.
- **`export`**: Translates world state and timeline data into an Obsidian-compatible Markdown vault containing YAML frontmatter metadata, sanitized filenames, wiki-links, and JSON pointcrawl graph data.

### Generation & Simulation Pipeline

- **Deterministic PRNG Engine** — Master seed engine (`internal/domain/state`) deriving isolated `math/rand/v2` streams per component to ensure 100% byte-identical outputs given identical seeds.
- **Geographical Genesis** — Perlin-noise elevation, latitude-aware climate (temperature & humidity), and biome classification (water, tundra, desert, forest, grassland).
- **Spatial Reasoning & Demographics** — Tile suitability scoring, population seeding, diffusion, and faction influence spread via cellular automata.
- **Settlements & Travel-Cost Pointcrawl** — Settlement placement based on suitability thresholds, network graph construction, nearest-neighbor edge routing, and travel cost heuristics measured in "watches" (factoring in terrain friction and elevation bonuses).
- **CFG Narrative Engine & Chronicle Streaming** — Context-free grammar engine supporting BNF grammar loading, variable expansion, and event narration during timeline simulation streaming.

### Exporters

- **Obsidian Markdown Vault** — Structured directory output (`settlements/`, `factions/`, `pointcrawl.json`, `timeline.md`) with frontmatter metadata and relational wiki-links.

## Getting Started

```bash
# Display help and usage
go run . --help

# Initialize a new world project
go run . init --name "Ashtar" --size medium

# Generate world & simulate timeline
go run . simulate --seed 42 --width 64 --height 64 --years 100 --events normal

# Export generated state to an Obsidian markdown vault
go run . export --format obsidian --output ./output
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
- `export` — `--format`, `--output`

## Determinism

Identical seeds produce identical worlds. Enforced by:

- `math/rand/v2` PRNGs created per component.
- Master-seed engine in `internal/domain/state` deriving stable sub-seeds from component identifiers.
- No package-level random state or shared global RNG.

## Testing & Quality Assurance

```bash
go test ./...                             # Run all tests
go test ./... -race                       # Test with race detector
go test ./... -coverprofile=coverage.out  # Code coverage
go tool cover -func=coverage.out
```

Coverage thresholds (enforced per [AGENTS.md](AGENTS.md)):

- Repository-wide statement coverage: ≥ 80%
- `internal/domain` and `internal/usecase`: ≥ 90% each
- Pure determinism assertion tests for all generation routines.

## Project Status

All planned specifications and architectural features defined under `openspec/specs/` are fully implemented, verified, and backed by test suites.
