# world-generation-go

`world-generation-go` is a deterministic Go CLI for fantasy world generation. The long-term goal is a clean-architecture application that can initialize a world project, run simulation phases, and export results to an Obsidian-friendly Markdown vault.

## Current State

The repository currently provides the CLI foundation built with Cobra and Viper. The command surface is in place, configuration can be loaded from flags, environment variables, and config files, and the core commands are present as working placeholders:

- `init` acknowledges project initialization inputs
- `simulate` acknowledges simulation inputs and reports queued status
- `export` acknowledges export format and destination

The actual world generation, simulation pipeline, and export implementation are still in progress under the OpenSpec changes in `openspec/`.

## Usage

Run the CLI directly from the repository root:

```bash
go run .
```

Available commands:

```bash
go run . init --name "Ashtar" --size medium
go run . simulate --years 500 --events dense
go run . export --format obsidian --output ./vault
```

Common global flags:

- `--config`: path to a YAML config file
- `--seed`: deterministic seed override
- `--output`: output directory for generated artifacts

Configuration precedence follows the current Viper setup:

`CLI flags > environment variables > config file > defaults`

Environment variables use the `WORLDGEN_` prefix. For example:

```bash
WORLDGEN_OUTPUT=./vault go run . export
```
