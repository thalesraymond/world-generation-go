# world-generation-go

world-generation-go is a deterministic Go CLI for fantasy world generation. The project is being built around a clean architecture and an OpenSpec-driven roadmap, with the long-term goal of supporting world initialization, simulation, and export to an Obsidian-friendly Markdown vault.

## What is implemented today

The repository already contains the foundations for a working generation pipeline:

- A Cobra-based CLI with the core commands `init`, `simulate`, and `export`
- Configuration loading through flags, environment variables, and optional config files via Viper
- A deterministic state engine that derives isolated PRNG streams from a master seed
- A terrain generation subsystem that produces elevation, temperature, humidity, and biome data from deterministic noise
- Unit and command-level tests covering the core domain behavior and CLI entrypoints

The current CLI commands are wired as entrypoints and currently acknowledge their inputs, while the underlying generation logic is implemented in the domain layer and is being expanded toward full simulation and export workflows.

## Project structure

- `cmd/` — CLI command definitions and entrypoints
- `config/` — typed configuration loading
- `internal/domain/` — pure domain logic for deterministic world generation
  - `state/` — seed-derived PRNG engine
  - `terrain/` — noise generation, terrain maps, and biome mapping
- `internal/usecase/`, `internal/adapter/`, and `internal/infra/` — architecture scaffolding for the next stages of the application
- `openspec/` — the source-of-truth design/specification documents for the project

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

## Configuration

The CLI supports these global options:

- `--config` — path to a YAML config file
- `--seed` — deterministic seed override
- `--output` — output directory for generated artifacts

Configuration precedence follows the current Viper behavior:

`CLI flags > environment variables > config file > defaults`

Environment variables use the `WORLDGEN_` prefix. For example:

```bash
WORLDGEN_OUTPUT=./vault go run . export
```

## Current development focus

The implemented domain pieces already cover the groundwork for deterministic generation, but the full end-to-end workflow is still being developed. The next major milestones in the roadmap are:

- full simulation loop orchestration
- demographic automata and settlement generation
- narrative synthesis and timeline streaming
- Obsidian-compatible export formatting

## Testing

The repository includes tests for:

- CLI help and command behavior
- deterministic PRNG behavior
- terrain generation and biome classification

Run the test suite with:

```bash
go test ./...
```
