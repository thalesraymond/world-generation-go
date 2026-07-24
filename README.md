# world-generation-go

`world-generation-go` is a deterministic Go CLI for fantasy world generation. It is being built around Clean Architecture and an OpenSpec-driven roadmap, with the long-term goal of supporting world initialization, phased simulation, and export to an Obsidian-compatible Markdown vault.

The tool is intentionally LLM-free: it relies on constructive generation algorithms, cellular automata, context-free grammars, and deterministic pseudo-random number generation to produce reproducible worlds and histories.

## What is implemented today

The repository contains the foundations of the generation pipeline:

- **CLI entrypoint** — a Cobra-based command tree with `init`, `simulate`, and `export` subcommands.
- **Configuration loading** — typed config via Viper, supporting flags, environment variables, and optional YAML config files.
- **Clean Architecture scaffolding** — layered packages (`cmd`, `internal/adapter`, `internal/usecase`, `internal/domain`, `internal/infra`) with documented dependency rules.
- **Deterministic state engine** — derives isolated `math/rand/v2` PRNG streams from a master seed, ensuring component-level reproducibility.
- **Terrain generation** — Perlin-noise-based elevation, latitude-aware temperature, humidity, and biome classification into `water`, `tundra`, `desert`, `forest`, and `grassland`.
- **Tests** — unit and command-level coverage for CLI behavior, PRNG determinism, and terrain rules.

The CLI commands currently acknowledge and validate their inputs; the underlying generation logic lives in the `domain` layer and is being expanded toward full simulation and export workflows.

## Project structure

```
cmd/                    CLI command definitions and entrypoints
config/                 Typed configuration loading
internal/
  adapter/              Input/output translation and command handlers
  domain/               Pure business logic (no framework/infrastructure imports)
    state/              Deterministic PRNG engine
    terrain/            Noise generation, terrain maps, and biome mapping
  usecase/              Application orchestration and interfaces
  infra/                File I/O, exporters, and external integrations
openspec/               Source-of-truth design and specification documents
```

Each `internal/*` layer contains a `README.md` describing its responsibilities and dependency rules. For the authoritative architecture and engineering standards, see [AGENTS.md](AGENTS.md).

## Getting started

Run the CLI from the repository root:

```bash
go run .
```

Example commands:

```bash
go run . init --name "Ashtar" --size medium
go run . simulate --years 500 --events dense
go run . export --format obsidian --output ./vault
```

At this stage, `init`, `simulate`, and `export` parse configuration and acknowledge the requested operation. Full simulation and file export are under active development.

## Configuration

Global options:

- `--config` — path to a YAML config file
- `--seed` — deterministic seed override
- `--output` — output directory for generated artifacts

Command-specific options:

- `init` — `--name`, `--size`
- `simulate` — `--years`, `--events`
- `export` — `--format`

Configuration precedence:

```
CLI flags > environment variables > config file > defaults
```

Environment variables use the `WORLDGEN_` prefix. For example:

```bash
WORLDGEN_OUTPUT=./vault go run . export
```

A config file without an explicit path is resolved as `worldgen.yaml` in the current directory or the user's home directory.

## Roadmap

Implemented milestones are archived under `openspec/changes/archive/`. Active work is tracked under `openspec/changes/` and includes:

1. **Deterministic RNG pipeline integration** — wire the state engine into simulation bootstrapping so every subsystem uses an injected, seed-derived PRNG.
2. **Core simulation loop** — year-by-year world advancement with channel-based timeline streaming to `stdout`.
3. **Demographic automata and settlement generation** — cellular-automata population spread and suitability-driven settlement placement.
4. **CFG narrative engine** — context-free grammar parser that turns numerical events into readable mythic text.
5. **Pointcrawl spatial abstraction** — graph of Points of Interest with travel-cost heuristics measured in "watches".
6. **Obsidian Markdown export** — YAML frontmatter, wiki-links, and a vault directory structure for personal knowledge management.

For full requirements and design rationale, see the corresponding `design.md` and `spec.md` files in `openspec/`.

## Determinism

Identical seeds must produce identical worlds. The project enforces this through:

- `math/rand/v2` PRNGs created per component.
- A master-seed engine in `internal/domain/state` that derives stable sub-seeds from a component identifier.
- No package-level random state and no shared global RNG.

## Testing

The test suite covers:

- CLI help and command behavior
- deterministic PRNG streams and component isolation
- terrain generation, biome classification, and noise normalization

Run tests:

```bash
go test ./...
```

Run with the race detector:

```bash
go test ./... -race
```

Coverage expectations are defined in [AGENTS.md](AGENTS.md): repository-wide statement coverage must remain >= 80%, and `internal/domain` and `internal/usecase` each must remain >= 90%.
